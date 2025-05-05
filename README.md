# Echo Health Check

This project implements a simple health check service using the Echo framework in Go. It provides an endpoint to check the health status of the service.

## Project Structure

```
echo-health-check
├── cmd
│   └── server
│       └── main.go        # Entry point of the application
├── internal
│   ├── handlers
│   │   └── health.go      # Health check handler
│   └── routes
│       └── routes.go      # Route setup
├── pkg
│   └── health
│       └── checker.go     # Health check logic
├── go.mod                  # Module dependencies
├── go.sum                  # Module checksums
└── README.md               # Project documentation
```

## Setup Instructions

1. Clone the repository:
   ```
   git clone <repository-url>
   cd echo-health-check
   ```

2. Install the dependencies:
   ```
   go mod tidy
   ```

3. Run the application:
   ```
   go run cmd/server/main.go
   ```

## Usage

Once the server is running, you can check the health status by sending a GET request to the following endpoint:

```
GET http://localhost:8080/health
```

You should receive a JSON response indicating the service status, for example:

```json
{
  "status": "healthy"
}
```

## Contributing

Feel free to submit issues or pull requests for improvements or bug fixes.