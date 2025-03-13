package main

import (
	"log"
	"net/http"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"

	"github.com/go-chi/chi/v5"
)

var (
	URLMap = make(map[string]string)
	cfg    *config.Config
)

func main() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePost(cfg, w, r)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGet(cfg, w, r)
	})

	log.Printf("Server is running on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
