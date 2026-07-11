.PHONY: swag run test build

# Generate swagger docs
swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o internal/swagger

# Format swagger doc comments
swag-format:
	go run github.com/swaggo/swag/cmd/swag@latest fmt - format swag comments

# Run the API locally (without docker)
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Build binary
build:
	go build -o bin/api ./cmd/api
