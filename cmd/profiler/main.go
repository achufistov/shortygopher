package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type ShortenRequest struct {
	OriginalURL string `json:"url"`
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run cmd/profiler/main.go <server_url> <profile_name>")
	}

	serverURL := os.Args[1]
	profileName := os.Args[2]

	log.Printf("Starting load generation for %s", serverURL)
	log.Printf("Profile will be saved as profiles/%s.pprof", profileName)

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go generateLoad(serverURL, &wg, stopChan, i)
	}

	time.Sleep(5 * time.Second)

	log.Println("Collecting memory profile...")
	err := collectMemoryProfile(serverURL, profileName)
	if err != nil {
		log.Fatalf("Failed to collect memory profile: %v", err)
	}

	close(stopChan)
	wg.Wait()

	log.Printf("Profile saved to profiles/%s.pprof", profileName)
}

func generateLoad(serverURL string, wg *sync.WaitGroup, stopChan chan bool, workerID int) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	counter := 0
	for {
		select {
		case <-stopChan:
			return
		default:
			reqBody := ShortenRequest{
				OriginalURL: fmt.Sprintf("https://example.com/worker%d/url%d", workerID, counter),
			}
			jsonBody, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", serverURL+"/api/shorten", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			counter++

			time.Sleep(10 * time.Millisecond)
		}
	}
}

func collectMemoryProfile(serverURL string, profileName string) error {
	err := os.MkdirAll("profiles", 0755)
	if err != nil {
		return fmt.Errorf("failed to create profiles directory: %v", err)
	}

	resp, err := http.Get(serverURL + "/debug/pprof/heap")
	if err != nil {
		return fmt.Errorf("failed to get memory profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get memory profile, status: %d", resp.StatusCode)
	}

	profilePath := fmt.Sprintf("profiles/%s.pprof", profileName)
	file, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save profile: %v", err)
	}

	return nil
}
