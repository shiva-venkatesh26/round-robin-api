package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
)

func main() {
	port := "8081"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		log.Printf("Received payload: %s", string(body))

		w.WriteHeader(http.StatusOK)
		log.Printf("Got a hit on an instance of Application API running on port %s", port)
		w.Write(body)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Randomly determine health status for testing
		isHealthy := rand.IntN(2) == 0 // 0 for healthy, 1 for unhealthy

		var response map[string]string

		if isHealthy {
			response = map[string]string{"status": "healthy"}
			w.WriteHeader(http.StatusOK)
		} else {
			response = map[string]string{"status": "unhealthy"}
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
			http.Error(w, `{"status":"error","message":"Failed to encode response"}`, http.StatusInternalServerError)
		}
	})
	log.Printf("Application API running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
