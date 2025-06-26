package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"time"
)

// InitMiddleware - производит подключение таких middleware, как:
//   - Logger
//   - Recover
//   - AllowContentType
//   - RequestTimeout
//   - CORS
//   - RequestID
func InitMiddleware(r *chi.Mux, timeout time.Duration) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/crmproxy/ping"))
	r.Use(middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.Timeout(timeout))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "Content-Length", "Accept-Encoding"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}
