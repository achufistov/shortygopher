package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	buildInfo := map[string]string{
		"Build version": buildVersion,
		"Build date":    buildDate,
		"Build commit":  buildCommit,
	}

	for key, value := range buildInfo {
		if value == "" {
			value = "N/A"
		}
		fmt.Printf("%s: %s\n", key, value)
	}
}

func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func main() {
	printBuildInfo()

	// Parse command line flags
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Printf("Error syncing logger: %v", syncErr)
		}
	}()

	var storageInstance storage.Storage
	if cfg.DatabaseDSN != "" {
		dbStorage, dbErr := storage.NewDBStorage(cfg.DatabaseDSN)
		if dbErr != nil {
			log.Printf("Error initializing database storage: %v", dbErr)
			if dbStorage != nil {
				if closeErr := dbStorage.Close(); closeErr != nil {
					log.Printf("Error closing database storage: %v", closeErr)
				}
			}
			log.Fatalf("Failed to initialize database storage: %v", dbErr)
		}
		defer func() {
			if closeErr := dbStorage.Close(); closeErr != nil {
				log.Printf("Error closing database storage: %v", closeErr)
			}
		}()
		storageInstance = dbStorage
	} else {
		log.Println("Database DSN is empty, using in-memory storage")
		storageInstance = storage.NewURLStorage()
	}

	urlMappings, err := storage.LoadURLMappings(cfg.FileStorage)
	if err != nil {
		log.Printf("Error loading URL mappings: %v", err)
	} else {
		for shortURL, originalURL := range urlMappings {
			if addErr := storageInstance.AddURL(shortURL, originalURL, "system"); addErr != nil {
				log.Printf("Error adding URL mapping (short: %s, original: %s): %v", shortURL, originalURL, addErr)
			}
		}
	}

	handlers.InitStorage(storageInstance)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.AuthMiddleware(cfg))

	// Add pprof routes for profiling
	r.Mount("/debug/pprof", http.DefaultServeMux)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePost(cfg, w, r)
	})
	r.Get("/{id}", handlers.HandleGet)
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShortenPost(cfg, w, r)
	})
	r.Post("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleBatchShortenPost(cfg, w, r)
	})
	r.Get("/ping", handlers.HandlePing(storageInstance))
	r.Get("/api/user/urls", handlers.HandleGetUserURLs(cfg))
	r.Delete("/api/user/urls", handlers.HandleDeleteUserURLs(cfg))

	// Internal stats endpoint with trusted subnet middleware
	r.With(middleware.TrustedSubnetMiddleware(cfg.TrustedSubnet)).Get("/api/internal/stats", handlers.HandleGetStats())

	// Create server with timeouts
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server is running on %s", cfg.Address)
		if cfg.EnableHTTPS {
			log.Printf("HTTPS enabled, using certificate: %s and key: %s", cfg.CertFile, cfg.KeyFile)
			serverErrors <- srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			serverErrors <- srv.ListenAndServe()
		}
	}()

	// Blocking select waiting for either a signal or an error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	case <-ctx.Done():
		log.Printf("Start shutdown. Signal: %v", ctx.Err())

		// Give outstanding requests a deadline for completion
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("Error during server shutdown: %v", err)

			// If shutdown times out, force close
			err = srv.Close()
			if err != nil {
				log.Printf("Error closing server: %v", err)
			}
		}

		// If using file storage, ensure all data is saved
		if cfg.FileStorage != "" {
			// Get all URLs from storage
			urlMap := storageInstance.GetAllURLs()

			// Save to file
			if err := storage.SaveURLMappings(cfg.FileStorage, urlMap); err != nil {
				log.Printf("Error saving URL mappings during shutdown: %v", err)
			} else {
				log.Printf("Successfully saved %d URL mappings to file", len(urlMap))
			}
		}

		// Ensure database connection is properly closed
		if dbStorage, ok := storageInstance.(*storage.DBStorage); ok {
			if err := dbStorage.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			}
		}

		log.Println("Server shutdown completed")
	}
}
