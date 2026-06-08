package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/config"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/server"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/pkg/internalsql"
	"github.com/go-playground/validator/v10"
)

const (
	serverReadHeaderTimeout = 2 * time.Second
	serverReadTimeout       = 10 * time.Second
	serverWriteTimeout      = 15 * time.Second
	serverIdleTimeout       = 60 * time.Second
	shutdownTimeout         = 10 * time.Second
)

func main() {
	validate := validator.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := internalsql.ConnectMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	router := server.NewRouter(validate, server.Dependencies{
		DB:            db,
		HealthChecker: db,
	})

	httpServer := &http.Server{
		Addr:              cfg.ServerAddress(),
		Handler:           router,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("starting http server on %s", cfg.ServerAddress())
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		close(serverErrors)
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("http server stopped unexpectedly: %v", err)
		}
	case <-shutdownSignal.Done():
		log.Print("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown http server: %v", err)
	}

	log.Print("server stopped gracefully")
}
