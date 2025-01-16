
# round-robin-api

This project implements a Round Robin API that routes requests to a pool of backend Application APIs in a round-robin fashion.

## Features

- Round Robin Load Balancing
- Handles POST requests with JSON payload
- Fault Tolerance and Retry Logic (future scope)

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

3. Send POST requests to the Router API:
   ```bash
   curl -X POST http://localhost:8080/route -d '{"key":"value"}' -H "Content-Type: application/json"
   ```

## Requirements

- Go 1.19+

