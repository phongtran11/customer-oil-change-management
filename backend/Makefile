.PHONY: swag run test build

# Generate swagger docs
swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o internal/swagger

# Run the API locally (without docker)
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Build binary
build:
	go build -o bin/api ./cmd/api
