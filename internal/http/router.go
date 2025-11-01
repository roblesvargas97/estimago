package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roblesvargas97/estimago/internal/clients"
)

func NewRouter(pool *pgxpool.Pool) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"*"},
		MaxAge:         300,
	}))

	r.Get("/api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Hello, World!"}`))
	})

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"db_down"}`))
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1/clients", func(r chi.Router) {
		r.Post("/", clients.PostClient(pool))
		r.Get("/", clients.ListClients(pool))
		r.Get("/{id}", clients.GetClient(pool))
	})

	return r
}
