package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"round-robin-api/internal/roundrobin"
)

type Config struct {
	Hosts               []string `yaml:"hosts"`
	HealthCheckInterval string   `yaml:"health_check_interval"`
}

// Load configuration from the YML file
func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Vanilla round-robin handler
func routeHandler(rr *roundrobin.RoundRobin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get the next host using vanilla round-robin
		nextHost := rr.Next()
		resp, err := forwardRequest(nextHost, r)
		if err != nil {
			log.Printf("Failed to forward request to host %s: %v", nextHost, err)
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Return the response to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// Fault-tolerant round-robin handler
func advancedRoutingHandler(rr *roundrobin.RoundRobin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Retry logic with fault-tolerant round-robin
		var resp *http.Response
		var err error
		for attempt := 0; attempt < len(rr.Hosts); attempt++ {
			nextHost, ok := rr.NextHealthy()
			if !ok {
				http.Error(w, "No healthy hosts available", http.StatusServiceUnavailable)
				return
			}

			// Forward request to the next healthy host
			resp, err = forwardRequest(nextHost, r)
			if err == nil {
				break
			}

			log.Printf("Host %s failed: %v. Retrying...", nextHost, err)
			rr.UpdateHealthStatus(nextHost, false) // Mark host as unhealthy
		}

		if err != nil {
			http.Error(w, "Failed to forward request after retries", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Return the response to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// Forward the request to the specified host with timeout handling
func forwardRequest(host string, r *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 2 * time.Second} // Timeout for slow hosts

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, host+"/echo", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	// Forward the request
	return client.Do(req)
}

// Check the health of a given host
func checkHealth(url string) bool {
	log.Printf("Health check triggered for host: %s", url)
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func main() {
	// Load configuration
	configPath := filepath.Join("configs", "router-config.yml")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse health check interval
	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Invalid health_check_interval: %v", err)
	}

	// Initialize RoundRobin with hosts
	rr := &roundrobin.RoundRobin{
		Hosts: func() []roundrobin.HostStatus {
			hosts := make([]roundrobin.HostStatus, len(config.Hosts))
			for i, host := range config.Hosts {
				hosts[i] = roundrobin.HostStatus{
					URL:       host,
					IsHealthy: true, // Assume all hosts are healthy at startup
				}
			}
			return hosts
		}(),
		Index: 0,
		Mutex: sync.Mutex{},
	}
	log.Printf("Initialized hosts: %v", config.Hosts)

	// Start health check worker
	go rr.HealthCheckWorker(healthCheckInterval*time.Second, checkHealth)

	// Set up endpoints
	http.HandleFunc("/route", routeHandler(rr))                      // Vanilla round-robin
	http.HandleFunc("/advanced_routing", advancedRoutingHandler(rr)) // Fault-tolerant round-robin

	// Start the server
	log.Println("Round Robin Router running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
