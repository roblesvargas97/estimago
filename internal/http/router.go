package http

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roblesvargas97/estimago/internal/auth"
	"github.com/roblesvargas97/estimago/internal/clients"
	"github.com/roblesvargas97/estimago/internal/quotes"
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

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method_not_allowed","path":"` + r.URL.Path + `","method":"` + r.Method + `"}`))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not_found","path":"` + r.URL.Path + `","method":"` + r.Method + `"}`))
	})

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

	authCfg := auth.Config{
		JWTSecret:   strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")),
		JWTTTLHours: parseEnvInt(os.Getenv("AUTH_JWT_TTL_HOURS"), 24),
	}

	r.Route("/api/v1/auth", func(sub chi.Router) {
		auth.RegisterRoutes(sub, pool, authCfg)
	})

	r.Group(func(priv chi.Router) {
		priv.Use(auth.JWTMiddleware(authCfg))
		priv.Get("/api/v1/me", auth.MeHandler(pool))
	})

	r.Route("/api/v1/clients", func(r chi.Router) {
		r.Post("/", clients.PostClient(pool))
		r.Get("/", clients.ListClients(pool))
		r.Get("/{id}", clients.GetClient(pool))
	})

	r.Route("/api/v1/quotes", func(r chi.Router) {
		r.Post("/", quotes.PostQuote(pool))
		r.Get("/", quotes.ListQuotes(pool))
		r.Get("/{id}", quotes.GetQuote(pool))
		r.Patch("/{id}", quotes.PatchQuote(pool))
		r.Post("/{id}/send", quotes.SendQuote(pool))
	})

	chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		println(method, route)
		return nil
	})
	return r
}

func parseEnvInt(value string, fallback int) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}

	n, err := strconv.Atoi(trimmed)
	if err != nil || n <= 0 {
		return fallback
	}

	return n
}
