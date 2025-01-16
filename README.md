
# round-robin-api

This project implements a Round Robin API that routes requests to a pool of backend Application APIs in a round-robin fashion.

## Features

### Vanilla Round Robin
- Routes requests to application APIs in a round-robin manner.
- No consideration for the health status or responsiveness of application APIs.
- Simple and fast, but lacks fault tolerance.

### Advanced Routing
- Adds fault tolerance with the following features:
   - **Health Check Worker**: Periodically checks the health of application APIs using a `/health` endpoint.
   - **Retry Mechanism**: Skips unhealthy or unresponsive hosts and retries with the next healthy host.
   - **Timeout Handling**: Avoids delays by setting timeouts for requests to slow hosts.
- Automatically re-adds hosts to the routing pool when they recover with a worker that checks the status of hosts every X seconds.

---


## Project Structure

- `cmd/`: Contains entry points for the Application API and Router API.
- `internal/`: Houses core logic for the APIs and round-robin.
- `configs/`: Configuration files for APIs.
- `pkg/`: Utility packages like logging and HTTP client.
- `tests/`: Unit and integration tests.

## Usage

1. Start the Application API:
   ```bash
   go run cmd/application/main.go
   ```

2. Start the Router API:
   ```bash
   go run cmd/router/main.go
   ```

3. Send POST requests to the Vanilla Router API:
   ```bash
   curl -X POST http://localhost:8080/route -d '{"key":"value"}' -H "Content-Type: application/json"
   ```
3. Send POST requests to the Advanced Router API:
   ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"key":"value"}' http://localhost:8080/advanced_routing
   ```

## Requirements

- Go 1.19+
