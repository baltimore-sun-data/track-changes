package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/andybalholm/cascadia"
	"github.com/carlmjohnson/gracefulserver"
	"github.com/pkg/errors"
)

var (
	nWorkers int
	dSleep   time.Duration
)

func main() {
	flag.IntVar(&nWorkers, "workers", 4, "how many simultaneously downloading workers to launch")
	flag.DurationVar(&dSleep, "poll", 5*time.Minute, "how often to poll for changes")
	fname := flag.String("file", "-", "File to open for jobs")
	flag.Parse()

	if jobs, err := start(*fname); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error during start up: %v\n", err)
		os.Exit(1)
	} else {
		go scheduler(jobs)
	}

	http.HandleFunc("/", handler)
	gracefulserver.Serve(http.DefaultServeMux)
}

func start(fname string) ([]job, error) {
	f := os.Stdin
	var err error
	if fname != "-" {
		f, err = os.Open(fname)
		if err != nil {
			return nil, errors.WithMessage(err, "could not open file")
		}
		defer deferClose(&err, f.Close)
	}

	var ldata map[string]struct {
		Url, Selector string
	}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&ldata); err != nil {
		return nil, errors.WithMessage(err, "could not parse JSON file")
	}

	jobs := make([]job, 0, len(ldata))
	for id := range ldata {
		url := ldata[id].Url
		selector := ldata[id].Selector

		txt, err := get(url, selector)
		if err != nil {
			return nil, err
		}

		data.Lock()
		data.m[id] = txt
		data.Unlock()

		jobs = append(jobs, job{id, url, selector})
	}

	return jobs, nil
}

func get(url, selector string) (string, error) {
	sel, err := cascadia.Compile(selector)
	if err != nil {
		return "", errors.WithMessage(err, "bad selector")
	}

	rsp, err := http.Get(url)
	if err != nil {
		return "", errors.WithMessage(err, "problem getting URL")
	}
	defer deferClose(&err, rsp.Body.Close)

	if rsp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected status for %s: %s", url, rsp.Status)
	}

	doc, err := html.Parse(rsp.Body)
	if err != nil {
		return "", errors.WithMessage(err, "could not parse document as HTML")
	}

	nn := sel.MatchAll(doc)
	ss := make([]string, 0, len(nn))
	for _, n := range nn {
		if n.FirstChild != nil {
			ss = append(ss, n.FirstChild.Data)
		}
	}

	return strings.Join(ss, ","), err
}

func deferClose(err *error, f func() error) {
	newErr := f()
	if *err == nil {
		*err = errors.WithMessage(newErr, "problem closing")
	}
}

var data = struct {
	m map[string]string
	sync.RWMutex
}{
	m: make(map[string]string),
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	e := json.NewEncoder(w)
	data.RLock()
	defer data.RUnlock()
	if err := e.Encode(&data.m); err != nil {
		log.Printf("Unexpected error: %v", err)
	}
}
