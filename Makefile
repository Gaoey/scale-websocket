build:
	go build -o scale-websocket ./cmd/server/main.go

test:
	go test ./...

clean:
	go clean
	rm -f scale-websocket

run:
	go run ./cmd/server/main.go