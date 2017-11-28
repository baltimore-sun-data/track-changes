package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

var router = chi.NewRouter()

func init() {
	router.Get("/*", http.FileServer(http.Dir("assets")).ServeHTTP)
	router.Get("/api", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	data.RLock()
	defer data.RUnlock()
	if err := e.Encode(&data); err != nil {
		log.Printf("Unexpected error: %v", err)
	}
}
