package main

import (
	"log"
	"net/http"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	URLMap = make(map[string]string)
	cfg    *config.Config
)

func initLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func main() {
	var err error
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()

	urlStorage := storage.NewURLStorage()
	urlMappings, err := storage.LoadURLMappings(cfg.FileStorage)
	if err != nil {
		log.Printf("Error loading URL mappings: %v", err)
	} else {
		for shortURL, originalURL := range urlMappings {
			urlStorage.AddURL(shortURL, originalURL)
		}
	}

	handlers.InitURLStorage(urlStorage)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.GzipMiddleware)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePost(cfg, w, r)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGet(cfg, w, r)
	})
	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShortenPost(cfg, w, r)
	})

	log.Printf("Server is running on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
