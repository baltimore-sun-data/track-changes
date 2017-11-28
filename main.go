package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/carlmjohnson/gracefulserver"
	"github.com/pkg/errors"
)

var (
	// Number of workers for scheduler
	nWorkers int
	// Sleep time between checks
	dSleep time.Duration
	// apiResponse data shared between poller and server
	data apiResponse
)

func main() {
	gdoc := flag.String("sheet", "", "Google Sheets ID to use for jobs")
	fname := flag.String("file", "-", "JSON file to open for jobs or '-' for stdin")
	flag.DurationVar(&dSleep, "poll", 5*time.Minute, "how often to poll for changes")
	flag.DurationVar(&http.DefaultClient.Timeout, "timeout", 10*time.Second, "how long to wait for a slow server")
	flag.IntVar(&nWorkers, "workers", 4, "how many simultaneously downloading workers to launch")
	flag.Parse()

	if err := EnvErrors(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start: %v\n", err)
		os.Exit(3)
	}

	if err := start(*gdoc, *fname); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error during start up: %v\n", err)
		os.Exit(1)
	}

	go data.jobs().start()

	http.Handle("/", http.FileServer(http.Dir("assets")))
	http.HandleFunc("/api", handler)
	gracefulserver.Serve(http.DefaultServeMux)
}

func start(gdoc, fname string) error {
	if gdoc != "" {
		return fromSheet(gdoc)
	}

	f := os.Stdin
	var err error
	if fname != "-" {
		f, err = os.Open(fname)
		if err != nil {
			return errors.WithMessage(err, "could not open file")
		}
		defer deferClose(&err, f.Close)
	}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&data); err != nil {
		return errors.WithMessage(err, "could not parse JSON file")
	}

	return nil
}

func deferClose(err *error, f func() error) {
	newErr := f()
	if *err == nil {
		*err = errors.WithMessage(newErr, "problem closing")
	}
}
