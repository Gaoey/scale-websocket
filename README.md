# Scale websocet
----

## Project Structure

```
scale-websocket
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
   cd scale-websocket
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