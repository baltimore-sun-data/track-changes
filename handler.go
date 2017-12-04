package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/carlmjohnson/gracefulserver"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var router = chi.NewRouter()

func init() {
	gracefulserver.Logger = middleware.Logger

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

type fileTemplate struct {
	files []string
	t     *template.Template
}

func newFileTemplate(files ...string) fileTemplate {
	return fileTemplate{files, template.Must(template.ParseFiles(files...))}
}

var (
	homepageT = newFileTemplate("templates/base.gohtml", "templates/index.gohtml")
	listingT  = newFileTemplate("templates/base.gohtml", "templates/listing.gohtml")
)

func (t fileTemplate) Exec(w http.ResponseWriter, r *http.Request, data interface{}) {
	if reload {
		t.t = template.Must(template.ParseFiles(t.files...))
	}
	if err := t.t.Execute(w, &data); err != nil {
		log.Printf("Unexpected template error for %s: %v", r.URL.Path, err)
	}
}

var sheetRe = regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9-_]+)`)

func getHomepage(w http.ResponseWriter, r *http.Request) {
	sheet := r.URL.Query().Get("sheet")
	if sheet == "" {
		data := struct {
			Manifest map[string]string
		}{staticManifest}
		homepageT.Exec(w, r, &data)
		return
	}

	// Strip URLs down to the final Google Docs identifier
	if m := sheetRe.FindString(sheet); m != "" {
		q := r.URL.Query()
		q.Set("sheet", m[len("/spreadsheets/d/"):])
		r.URL.RawQuery = q.Encode()
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}

	data := struct {
		Manifest        map[string]string
		BasicAuthHeader string
		SheetID         string
	}{staticManifest, baHeader, sheet}
	listingT.Exec(w, r, &data)
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
