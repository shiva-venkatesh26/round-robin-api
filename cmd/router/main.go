package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"round-robin-api/internal/roundrobin"
)

type Config struct {
	Hosts []string `yaml:"hosts"`
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func routeHandler(rr *roundrobin.RoundRobin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		nextInstance := rr.Next()

		resp, err := forwardRequest(nextInstance, r)
		if err != nil {
			log.Printf("Failed to forward request: %v", err)
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		defer resp.Body.Close()
	}
}

func forwardRequest(instance string, r *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, instance+"/echo", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	client := &http.Client{}
	return client.Do(req)
}

func main() {
	configPath := filepath.Join("configs", "router-config.yml")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rr := &roundrobin.RoundRobin{
		Instances: config.Hosts,
		Index:     0,
		Mutex:     sync.Mutex{},
	}
	log.Printf("Initialized instances: %v", rr.Instances)

	http.HandleFunc("/route", routeHandler(rr))
	log.Println("Round Robin Router running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
