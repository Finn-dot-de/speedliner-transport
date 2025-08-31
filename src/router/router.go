package router

import (
	"net/http"
	"path/filepath"
	"speedliner-server/src/handler"
	"speedliner-server/src/middleware"

	"github.com/go-chi/chi/v5"
)

// NewRouter erstellt einen neuen Router mit allen Routen und Middleware
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware einhängen
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.NoCacheMiddleware)

	// API-Routen
	r.Route("/app/", func(sub chi.Router) {
		handler.DefineApiRoutes(sub)
	})

	// Dynamisch zusammengesetzter relativer Pfad
	buildDir := filepath.Join("frontend")
	fs := http.FileServer(http.Dir(buildDir))
	r.Handle("/*", fs)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(buildDir, "index.html"))
	})

	// Root-Route (z.B. für deine index.html)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend")
	})

	return r
}
