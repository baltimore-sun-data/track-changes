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
	fname := flag.String("file", "-", "file to open for jobs or '-' for stdin")
	flag.DurationVar(&dSleep, "poll", 5*time.Minute, "how often to poll for changes")
	flag.DurationVar(&http.DefaultClient.Timeout, "timeout", 10*time.Second, "how long to wait for a slow server")
	flag.IntVar(&nWorkers, "workers", 4, "how many simultaneously downloading workers to launch")
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
	for id, val := range ldata {
		j := job{id, val.Url, val.Selector}
		jobs = append(jobs, j)
		if err = data.Update(j); err != nil {
			return nil, err
		}
	}

	return jobs, nil
}

func get(url, selector string) (string, error) {
	if selector == "twitter" {
		return getTweet(url)
	}

	return getUrl(url, selector)
}

func getUrl(url, selector string) (string, error) {
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

type jsonData struct {
	Url          string     `json:"url"`
	Selector     string     `json:"selector"`
	Content      string     `json:"content"`
	LastAccessed time.Time  `json:"last_accessed"`
	LastChange   time.Time  `json:"last_change"`
	Err          string     `json:"error,omitempty"`
	LastError    *time.Time `json:"last_error,omitempty"`
}

type apiResponse struct {
	m map[string]jsonData
	sync.RWMutex
}

func (a *apiResponse) Update(j job) error {
	log.Printf("Updating %#v", j)

	txt, err := get(j.url, j.selector)
	now := time.Now()
	var (
		errStr  string
		errTime *time.Time
	)
	if err != nil {
		errStr = err.Error()
		errTime = &now
		log.Printf("Error for %s: %v", j.id, err)
	}

	a.Lock()
	defer a.Unlock()

	lastChange := a.m[j.id].LastChange
	if a.m[j.id].Content != txt {
		lastChange = now
	}

	a.m[j.id] = jsonData{
		Url:          j.url,
		Selector:     j.selector,
		Content:      txt,
		LastAccessed: now,
		LastChange:   lastChange,
		Err:          errStr,
		LastError:    errTime,
	}

	return err
}

func (a *apiResponse) MarshalJSON() ([]byte, error) {
	a.RLock()
	defer a.RUnlock()
	return json.Marshal(&a.m)
}

var data = apiResponse{m: make(map[string]jsonData)}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	e := json.NewEncoder(w)
	if err := e.Encode(&data); err != nil {
		log.Printf("Unexpected error: %v", err)
	}
}
