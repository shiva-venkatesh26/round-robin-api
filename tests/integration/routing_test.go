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
	time.Sleep(2 * time.Second) // Allow server to start
	return cmd
}

// Helper function to start Application APIs on different ports
func startApplicationAPI(port string) *exec.Cmd {
	cmd := exec.Command("go", "run", "../../cmd/application/main.go", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start Application API on port %s: %v", port, err)
	}
	time.Sleep(1 * time.Second) // Allow server to start
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

func TestRoutingWorkflow(t *testing.T) {
	// Start Application APIs
	appPorts := []string{"8081", "8082", "8083"}
	appCmds := []*exec.Cmd{}
	for _, port := range appPorts {
		cmd := startApplicationAPI(port)
		appCmds = append(appCmds, cmd)
	}
	defer func() {
		for _, cmd := range appCmds {
			cmd.Process.Kill()
		}
	}()

	// Start the Router API
	routerCmd := startRouterAPI()
	defer routerCmd.Process.Kill()

	payload := map[string]interface{}{
		"game":   "Mobile Legends",
		"points": 42,
	}

	expectedResponses := []string{
		`{"message":"Echo from Application API on port 8081"}`,
		`{"message":"Echo from Application API on port 8082"}`,
		`{"message":"Echo from Application API on port 8083"}`,
	}

	// Send requests sequentially and verify the round-robin behavior
	for _, expectedResponse := range expectedResponses {
		resp, err := sendPostRequest("http://localhost:8080/route", payload)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		responseBody := readResponseBody(resp)
		if responseBody != expectedResponse {
			t.Errorf("Expected response: %s, got: %s", expectedResponse, responseBody)
		}
	}
}
