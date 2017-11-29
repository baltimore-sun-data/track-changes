package main

import (
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
	// dataStore that links sheet IDs to apiResponses
	globalData dataStore
	// Set during Docker build process
	applicationBuildDate = "N/A"
)

func main() {
	flag.DurationVar(&dSleep, "poll", 5*time.Minute, "how often to poll for changes")
	flag.DurationVar(&http.DefaultClient.Timeout, "timeout", 10*time.Second, "how long to wait for a slow server")
	flag.IntVar(&nWorkers, "workers", 4, "how many simultaneously downloading workers to launch")
	flag.Parse()

	if err := EnvErrors(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start: %v\n", err)
		os.Exit(1)
	}

	gracefulserver.Serve(router)
}

func deferClose(err *error, f func() error) {
	newErr := f()
	if *err == nil {
		*err = errors.WithMessage(newErr, "problem closing")
	}
}
