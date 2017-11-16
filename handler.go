package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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
