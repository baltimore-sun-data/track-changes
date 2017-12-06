package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func getHTML(url, selector string) (string, error) {
	// TODO: Can we cache these selectors?
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
		return "", fmt.Errorf("unexpected status for %s: %s", url, rsp.Status)
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

func getFeed(url string) (string, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return "", errors.WithMessage(err, "problem getting URL")
	}
	defer deferClose(&err, rsp.Body.Close)

	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status for %s: %s", url, rsp.Status)
	}

	return firstFeedTitle(rsp.Body)
}

func firstFeedTitle(r io.Reader) (string, error) {
	d := xml.NewDecoder(r)
	var seenItem, seenTitle bool
	for {
		t, err := d.Token()
		if err != nil {
			return "", errors.WithMessage(err, "problem processing feed")
		}

		switch tok := t.(type) {
		case xml.StartElement:
			switch {
			// RSS uses <item> and Atom uses <entry>
			case tok.Name.Local == "item" || tok.Name.Local == "entry":
				seenItem = true
			// Ignore overall <title> and return first <item> title
			case seenItem && tok.Name.Local == "title":
				seenTitle = true
			}
		case xml.CharData:
			if seenTitle {
				return string(tok), nil
			}
		}
	}
}
