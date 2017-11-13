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

	"golang.org/x/net/html"

	"github.com/andybalholm/cascadia"
	"github.com/carlmjohnson/gracefulserver"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	id := flag.Arg(0)
	url := flag.Arg(1)
	selector := flag.Arg(2)

	txt, err := get(url, selector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}

	data.Lock()
	data.m[id] = txt
	data.Unlock()

	http.HandleFunc("/", handler)
	gracefulserver.Serve(http.DefaultServeMux)
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
		*err = newErr
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
