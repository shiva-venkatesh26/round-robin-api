package roundrobin

import (
	"log"
	"sync"
	"time"
)

type HostStatus struct {
	URL        string        // Host URL
	IsHealthy  bool          // Whether the host is healthy
	LastCheck  time.Time     // Timestamp of the last health check
	RetryAfter time.Duration // Time duration before retrying a failed host
}

type RoundRobin struct {
	Hosts []HostStatus
	Index int // Current index in the round-robin cycle
	Mutex sync.Mutex
}

// Next selects the next host in a round-robin fashion (vanilla round-robin).
func (rr *RoundRobin) Next() string {
	rr.Mutex.Lock()
	defer rr.Mutex.Unlock()

	// Get the next host
	instance := rr.Hosts[rr.Index]
	rr.Index = (rr.Index + 1) % len(rr.Hosts)
	return instance.URL
}

// NextHealthy selects the next healthy host in a round-robin fashion.
// If no healthy host is available, returns an empty string and false.
func (rr *RoundRobin) NextHealthy() (string, bool) {
	rr.Mutex.Lock()
	defer rr.Mutex.Unlock()

	// Iterate over all hosts to find the next healthy one
	for i := 0; i < len(rr.Hosts); i++ {
		rr.Index = (rr.Index + 1) % len(rr.Hosts)
		host := rr.Hosts[rr.Index]
		if host.IsHealthy {
			return host.URL, true
		}
	}
	// No healthy host found
	return "", false
}

// UpdateHealthStatus updates the health status of a given host.
func (rr *RoundRobin) UpdateHealthStatus(url string, isHealthy bool) {
	rr.Mutex.Lock()
	defer rr.Mutex.Unlock()

	for i, host := range rr.Hosts {
		if host.URL == url {
			rr.Hosts[i].IsHealthy = isHealthy
			rr.Hosts[i].LastCheck = time.Now()
			if !isHealthy {
				rr.Hosts[i].RetryAfter = 10 * time.Second // Retry after 10 seconds
			}
			return
		}
	}
}

// HealthCheckWorker periodically checks the health of all hosts.
func (rr *RoundRobin) HealthCheckWorker(interval time.Duration, check func(string) bool) {
	ticker := time.NewTicker(interval) // Set up a periodic ticker & clean up the ticker when the worker stops
	defer ticker.Stop()

	for range ticker.C { // Run at each tick
		for _, host := range rr.Hosts { // Iterate through all hosts
			healthy := check(host.URL) // Calling the user-defined health check function
			rr.UpdateHealthStatus(host.URL, healthy)

			// Optional: Log the health status
			status := "healthy"
			if !healthy {
				status = "unhealthy"
			}
			log.Printf("Host %s is %s", host.URL, status)
		}
	}
}
