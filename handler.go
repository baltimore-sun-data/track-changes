package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var router = chi.NewRouter()

func init() {
	// Basic Auth protected routes
	router.With(basicAuthMiddleware).Get("/", getHomepage)
	// Note: Don't need http.StripPrefix because we serve the parent dir
	router.With(
		basicAuthMiddleware,
		middleware.DefaultCompress,
		middleware.SetHeader("Cache-Control", "public, max-age=365000000, immutable"),
	).Get(
		"/static/*",
		http.FileServer(http.Dir("assets")).ServeHTTP,
	)

	// Unprotected routes
	router.Get("/api/sheet/{sheetID}", getApiRequest)
	router.Post("/api/sheet/{sheetID}", postApiRequest)
	router.Get("/api/health", healthCheck)
}

func getHomepage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/index.html")
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

func basicAuthMiddleware(h http.Handler) http.Handler {
	username := GetEnv("BASIC_AUTH_USER")
	password := GetEnv("BASIC_AUTH_PASSWORD")
	realm := GetEnv("BASIC_AUTH_MESSAGE")

	if username == "" || password == "" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should use HTTPS with BasicAuth
		w.Header().Add("Strict-Transport-Security", "max-age=31536000")

		u, p, ok := r.BasicAuth()
		if ok && u == username && p == password {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%d %s\n",
			http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	})
}
