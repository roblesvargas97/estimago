package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roblesvargas97/estimago/internal/auth"
	"github.com/roblesvargas97/estimago/internal/config"
	"github.com/roblesvargas97/estimago/internal/db"
	httpx "github.com/roblesvargas97/estimago/internal/http"
)

func main() {
	cfg := config.Load()
	port := cfg.Port

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error al conectar a la base de datos: %v", err)
	}
	defer pool.Close()

	authCfg := auth.Config{
		JWTSecret:   cfg.AuthJWTSecret,
		JWTTTLHours: cfg.AuthJWTTTLHrs,
	}

	r := httpx.NewRouter(pool, authCfg)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("🚀 Servidor iniciado en http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error del servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Apagando servidor...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Error al apagar el servidor: %v", err)
	}
	log.Println("👋 Servidor detenido correctamente")
}
