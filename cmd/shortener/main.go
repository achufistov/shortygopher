package main

import (
	"log"
	"net/http"
	"os"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/controllers"
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
	} else if cfg.FileStorage != "" {
		fileStorage, err := storage.NewFileStorage(cfg.FileStorage)
		if err != nil {
			log.Printf("Error initializing file storage: %v", err)
			os.Exit(1)
		}
		defer fileStorage.Close()
		storageInstance = fileStorage
	} else {
		log.Println("No storage configured, using in-memory storage")
		storageInstance = storage.NewURLStorage()
	}

	// Initialize controller
	urlController := controllers.NewURLController(storageInstance)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.AuthMiddleware(cfg))

	// Routes
	r.Post("/", urlController.HandleShorten)
	r.Get("/{id}", urlController.HandleGet)
	r.Post("/api/shorten", urlController.HandleShorten)
	r.Post("/api/shorten/batch", urlController.HandleBatchShorten)
	r.Get("/api/user/urls", urlController.HandleGetUserURLs)
	r.Delete("/api/user/urls", urlController.HandleDeleteUserURLs)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Server is running on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
