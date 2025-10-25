package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roblesvargas97/estimago/internal/config"
	"github.com/roblesvargas97/estimago/internal/db"
	httpx "github.com/roblesvargas97/estimago/internal/http"
)

func main() {
	// 1️⃣ Cargar configuración
	cfg := config.Load()
	port := cfg.Port

	// 2️⃣ Crear conexión a la base de datos
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error al conectar a la base de datos: %v", err)
	}
	defer pool.Close()

	// 3️⃣ Crear router principal (Chi)
	r := httpx.NewRouter(pool)

	// 4️⃣ Servidor HTTP
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 5️⃣ Goroutine para escuchar
	go func() {
		log.Printf("🚀 Servidor iniciado en http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error del servidor: %v", err)
		}
	}()

	// 6️⃣ Esperar señal para apagar
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Apagando servidor...")

	// 7️⃣ Cierre elegante
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Error al apagar el servidor: %v", err)
	}
	log.Println("👋 Servidor detenido correctamente")
}
