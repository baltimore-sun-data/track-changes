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
	router.Get("/api/sheet/{sheetID}", getApiRequest)
	router.Post("/api/sheet/{sheetID}", postApiRequest)
	router.Get("/api/health", healthCheck)
}

func getApiRequest(w http.ResponseWriter, r *http.Request) {
	sheetID := chi.URLParam(r, "sheetID")

	data, err := globalData.get(sheetID)
	if err != nil {
		log.Printf("Error getting sheet %q: %v", sheetID, err)
		http.Error(w, "Could not get sheet", http.StatusBadGateway)
		return
	}
	jsonEncode(w, r, data)
}

func postApiRequest(w http.ResponseWriter, r *http.Request) {
	sheetID := chi.URLParam(r, "sheetID")

	if err := globalData.refresh(sheetID); err != nil {
		log.Printf("Error getting sheet %q: %v", sheetID, err)
		http.Error(w, "Could not get sheet", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Date string `json:"build_date"`
	}{applicationBuildDate}
	jsonEncode(w, r, envelope{Data: data})
}

func jsonEncode(w http.ResponseWriter, r *http.Request, data interface{}) {
	e := json.NewEncoder(w)
	if err := e.Encode(&data); err != nil {
		log.Printf("Unexpected error while encoding for %s: %v", r.URL.Path, err)
	}
}
