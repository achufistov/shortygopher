package main

import (
	"log"
	"net/http"
	"os"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()

	var storageInstance storage.Storage
	if cfg.DatabaseDSN != "" {
		dbStorage, err := storage.NewDBStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Error initializing database storage: %v", err)
			if dbStorage != nil {
				dbStorage.Close()
			}
			os.Exit(1)
		}
		defer dbStorage.Close()
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
			err := storageInstance.AddURL(shortURL, originalURL, "system")
			if err != nil {
				log.Printf("Error adding URL mapping (short: %s, original: %s): %v", shortURL, originalURL, err)
			}
		}
	}

	handlers.InitStorage(storageInstance)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.AuthMiddleware(cfg))

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
