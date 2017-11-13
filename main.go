package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	url := flag.Arg(0)
	selector := flag.Arg(1)

	if txt, err := get(url, selector); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println(txt)
	}
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
