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
	fname := flag.String("file", "-", "file to open for jobs or '-' for stdin")
	flag.DurationVar(&dSleep, "poll", 5*time.Minute, "how often to poll for changes")
	flag.DurationVar(&http.DefaultClient.Timeout, "timeout", 10*time.Second, "how long to wait for a slow server")
	flag.IntVar(&nWorkers, "workers", 4, "how many simultaneously downloading workers to launch")
	flag.Parse()

	if err := start(*fname); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error during start up: %v\n", err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.Dir("assets")))
	http.HandleFunc("/api", handler)
	gracefulserver.Serve(http.DefaultServeMux)
}

func start(fname string) error {
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

	go scheduler(data.Jobs())

	return nil
}

func deferClose(err *error, f func() error) {
	newErr := f()
	if *err == nil {
		*err = errors.WithMessage(newErr, "problem closing")
	}
}
