package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	data.RLock()
	defer data.RUnlock()
	if err := e.Encode(&data); err != nil {
		log.Printf("Unexpected error: %v", err)
	}
}
