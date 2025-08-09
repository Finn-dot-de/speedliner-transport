package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"path/filepath"
	"speedliner-server/src/handler"
	"speedliner-server/src/middleware"
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
	buildDir := filepath.Join("dist")
	fs := http.FileServer(http.Dir(buildDir))
	r.Handle("/*", fs)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(buildDir, "index.html"))
	})

	// Root-Route (z.B. für deine index.html)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dist")
	})

	return r
}
