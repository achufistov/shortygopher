package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func BenchmarkHandlePost(b *testing.B) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		url := "https://example.com/test" + string(rune(i))
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(url))
		req.Header.Set("Content-Type", "text/plain")

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		b.StartTimer()

		handlers.HandlePost(cfg, w, req)
	}
}

func BenchmarkHandleShortenPost(b *testing.B) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reqBody := handlers.ShortenRequest{
			OriginalURL: "https://example.com/test" + string(rune(i)),
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		b.StartTimer()

		handlers.HandleShortenPost(cfg, w, req)
	}
}

func BenchmarkHandleGet(b *testing.B) {
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)

	shortURL := "testurl"
	storageInstance.AddURL(shortURL, "https://example.com", "test-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", shortURL)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		b.StartTimer()

		handlers.HandleGet(w, req)
	}
}

func BenchmarkHandleGetUserURLs(b *testing.B) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)

	userID := "test-user"
	for i := 0; i < 100; i++ {
		shortURL := "short" + string(rune(i))
		originalURL := "https://example.com/" + string(rune(i))
		storageInstance.AddURL(shortURL, originalURL, userID)
	}

	handler := handlers.HandleGetUserURLs(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		b.StartTimer()

		handler(w, req)
	}
}
