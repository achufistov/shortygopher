package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

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
	log.Printf("Server is running on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
