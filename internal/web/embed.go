// Package web serves the embedded operator console and static assets.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

// Register mounts static UI and fallback routes on mux.
func Register(mux *http.ServeMux) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("web static fs: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("GET /{$}", fileServer)
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /app.js", fileServer)
	mux.Handle("GET /style.css", fileServer)
}

// Assets returns embedded filesystem for tests.
func Assets() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return staticFS
	}
	return sub
}
