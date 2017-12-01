package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
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
	router.With(basicAuthMiddleware).Get("/api/sheet/{sheetID}", getApiRequest)
	router.With(basicAuthMiddleware).Post("/api/sheet/{sheetID}", postApiRequest)

	// Unprotected routes
	router.Get("/api/health", healthCheck)
}

var staticManifest map[string]string

func init() {
	b, err := ioutil.ReadFile("assets/manifest.json")
	if err != nil {
		log.Fatalf("Could not read static asset manifest: %v", err)
		return
	}

	err = json.Unmarshal(b, &staticManifest)
	if err != nil {
		log.Fatalf("Could not parse static asset manifest: %v", err)
		return
	}
}

var (
	homepageTemplate = template.Must(template.ParseFiles(
		"templates/base.gohtml", "templates/index.gohtml"))
	listingTemplate = template.Must(template.ParseFiles(
		"templates/base.gohtml", "templates/listing.gohtml"))
)

func getHomepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("sheet") == "" {
		data := struct {
			Manifest        map[string]string
			BasicAuthHeader string
		}{staticManifest, baHeader}
		templateExec(w, r, homepageTemplate, &data)
		return
	}
	data := struct {
		Manifest        map[string]string
		BasicAuthHeader string
	}{staticManifest, baHeader}
	templateExec(w, r, listingTemplate, &data)
}

func templateExec(w http.ResponseWriter, r *http.Request, t *template.Template, data interface{}) {
	if err := t.Execute(w, &data); err != nil {
		log.Printf("Unexpected template error for %s: %v", r.URL.Path, err)
	}
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

// Basic Auth static vars
var (
	baUsername = GetEnv("BASIC_AUTH_USER")
	baPassword = GetEnv("BASIC_AUTH_PASSWORD")
	baRealm    = GetEnv("BASIC_AUTH_MESSAGE")
	baHeader   string
)

func basicAuthMiddleware(h http.Handler) http.Handler {
	if baUsername == "" || baPassword == "" {
		return h
	}

	baHeader = base64.StdEncoding.EncodeToString([]byte(
		baUsername + ":" + baPassword,
	))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should use HTTPS with BasicAuth
		w.Header().Add("Strict-Transport-Security", "max-age=31536000")

		u, p, ok := r.BasicAuth()
		if ok && u == baUsername && p == baPassword {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, baRealm))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%d %s\n",
			http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	})
}
