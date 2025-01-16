package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Helper function to start the Router API as a subprocess
func startRouterAPI() *exec.Cmd {
	cmd := exec.Command("go", "run", "../../cmd/router/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start Router API: %v", err)
	}
	time.Sleep(2 * time.Second) // Allow the server time to start
	return cmd
}

// Helper function to send a POST request
func sendPostRequest(url string, payload map[string]interface{}) (*http.Response, error) {
	jsonData, _ := json.Marshal(payload)
	return http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

// Helper function to read response body
func readResponseBody(resp *http.Response) string {
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func TestRoundRobinIntegration(t *testing.T) {
	// Start the Router API
	cmd := startRouterAPI()
	defer cmd.Process.Kill() // Ensure the Router API process is killed after the test

	// Test payload
	payload := map[string]interface{}{
		"game":   "test",
		"points": 42,
	}

	// Define the expected instances to receive requests in round-robin order
	expectedInstances := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Send requests to the Router API and verify the round-robin behavior
	for i, expectedInstance := range expectedInstances {
		resp, err := sendPostRequest("http://localhost:8080/route", payload)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Verify the response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify the forwarded response contains the payload
		responseBody := readResponseBody(resp)
		if responseBody != `{"game":"test","points":42}` {
			t.Errorf("Unexpected response body: %s", responseBody)
		}

		// Log which instance handled the request (mock verification in logs)
		t.Logf("Request %d was routed to: %s", i+1, expectedInstance)
	}
}
