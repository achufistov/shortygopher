package handlers_test

import (
	"strconv"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/storage"
)

func BenchmarkURLStorageAddURL(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		storageInstance.AddURL(shortURL, originalURL, "user"+strconv.Itoa(i%10))
	}
}

func BenchmarkURLStorageGetURL(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	for i := 0; i < 1000; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		storageInstance.AddURL(shortURL, originalURL, "user"+strconv.Itoa(i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortURL := "short" + strconv.Itoa(i%1000)
		_, _, _ = storageInstance.GetURL(shortURL)
	}
}

func BenchmarkURLStorageGetShortURLByOriginalURL(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	for i := 0; i < 1000; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		storageInstance.AddURL(shortURL, originalURL, "user"+strconv.Itoa(i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		originalURL := "https://example.com/" + strconv.Itoa(i%1000)
		_, _ = storageInstance.GetShortURLByOriginalURL(originalURL)
	}
}

func BenchmarkURLStorageGetURLsByUser(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	for i := 0; i < 1000; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		userID := "user" + strconv.Itoa(i%10)
		storageInstance.AddURL(shortURL, originalURL, userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := "user" + strconv.Itoa(i%10)
		_, _ = storageInstance.GetURLsByUser(userID)
	}
}

func BenchmarkURLStorageGetAllURLs(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	for i := 0; i < 1000; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		storageInstance.AddURL(shortURL, originalURL, "user"+strconv.Itoa(i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storageInstance.GetAllURLs()
	}
}

func BenchmarkURLStorageAddURLs(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		urls := make(map[string]string)
		for j := 0; j < 10; j++ {
			idx := i*10 + j
			shortURL := "short" + strconv.Itoa(idx)
			originalURL := "https://example.com/" + strconv.Itoa(idx)
			urls[shortURL] = originalURL
		}
		userID := "user" + strconv.Itoa(i%10)
		b.StartTimer()

		storageInstance.AddURLs(urls, userID)
	}
}

func BenchmarkURLStorageDeleteURLs(b *testing.B) {
	storageInstance := storage.NewURLStorage()

	for i := 0; i < 1000; i++ {
		shortURL := "short" + strconv.Itoa(i)
		originalURL := "https://example.com/" + strconv.Itoa(i)
		userID := "user" + strconv.Itoa(i%10)
		storageInstance.AddURL(shortURL, originalURL, userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		shortURLs := make([]string, 5)
		for j := 0; j < 5; j++ {
			idx := (i*5 + j) % 1000
			shortURLs[j] = "short" + strconv.Itoa(idx)
		}
		userID := "user" + strconv.Itoa(i%10)
		b.StartTimer()

		storageInstance.DeleteURLs(shortURLs, userID)
	}
}

func BenchmarkURLStorageMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	b.Run("AddURL", func(b *testing.B) {
		storageInstance := storage.NewURLStorage()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			shortURL := "short" + strconv.Itoa(i)
			originalURL := "https://example.com/" + strconv.Itoa(i)
			storageInstance.AddURL(shortURL, originalURL, "user")
		}
	})

	b.Run("GetURL", func(b *testing.B) {
		storageInstance := storage.NewURLStorage()
		for i := 0; i < 1000; i++ {
			shortURL := "short" + strconv.Itoa(i)
			originalURL := "https://example.com/" + strconv.Itoa(i)
			storageInstance.AddURL(shortURL, originalURL, "user")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			shortURL := "short" + strconv.Itoa(i%1000)
			_, _, _ = storageInstance.GetURL(shortURL)
		}
	})

	b.Run("GetAllURLs", func(b *testing.B) {
		storageInstance := storage.NewURLStorage()
		for i := 0; i < 100; i++ {
			shortURL := "short" + strconv.Itoa(i)
			originalURL := "https://example.com/" + strconv.Itoa(i)
			storageInstance.AddURL(shortURL, originalURL, "user")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = storageInstance.GetAllURLs()
		}
	})
}
