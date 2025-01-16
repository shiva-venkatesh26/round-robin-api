package integration

import (
	"net/http"
	"os/exec"
	"testing"
	"time"
)

func TestAdvancedRouting(t *testing.T) {
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

	// Test payload
	payload := map[string]interface{}{
		"game":    "Mobile Legends",
		"gamerID": "GYUTDTE",
		"points":  42,
	}

	// Simulate unhealthy hosts by shutting down one Application API
	unhealthyPort := "8082"
	for _, cmd := range appCmds {
		if cmd.Args[len(cmd.Args)-1] == unhealthyPort {
			cmd.Process.Kill()
		}
	}

	// Wait 10 seconds to allow the health check worker to detect the unhealthy host
	time.Sleep(10 * time.Second)

	// Sequentially send requests to the advanced routing endpoint
	expectedResponses := []string{
		`{"message":"Echo from Application API on port 8081"}`, // First healthy host
		`{"message":"Echo from Application API on port 8083"}`, // Skips 8082, uses 8083
		`{"message":"Echo from Application API on port 8081"}`, // Wraps back to 8081
	}

	for i, expectedResponse := range expectedResponses {
		resp, err := sendPostRequest("http://localhost:8080/advanced_routing", payload)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		defer resp.Body.Close()

		// Verify the response status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, resp.StatusCode)
		}

		// Verify the response body
		responseBody := readResponseBody(resp)
		if responseBody != expectedResponse {
			t.Errorf("Request %d: Expected response %s, got %s", i+1, expectedResponse, responseBody)
		}
	}
}
