package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
)

var (
	indexTemplate *template.Template
)

func RegisterStaticRoutes(mux *http.ServeMux, templatesFS fs.FS, staticFS fs.FS) {
	// Parse index template
	var err error
	indexTemplate, err = template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		panic("failed to parse index template: " + err.Error())
	}

	// Serve static files
	staticFSSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("failed to get static sub-fs: " + err.Error())
	}
	fileServer := http.FileServerFS(staticFSSub)
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Serve index page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Only serve index on root path; API 404 for other paths is handled elsewhere
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		indexTemplate.Execute(w, nil)
	})
}
