.PHONY: build run test clean docker-build docker-run docker-compose lint

# Application name
APP_NAME=user-management-service

# Build the application
build:
	go build -o bin/$(APP_NAME) ./cmd/server

# Run the application
run:
	go run ./cmd/server/main.go

# Run tests
test:
	go test -v ./tests/...

# Run all tests including coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Build Docker image
docker-build:
	docker build -t $(APP_NAME) .

# Run Docker container
docker-run:
	docker run -p 8080:8080 --name $(APP_NAME)-container --rm $(APP_NAME)

# Run with Docker Compose
docker-compose:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down

# Run linter
lint:
	go vet ./...

# Create necessary directories for the project
init:
	mkdir -p api/handlers api/middleware internal/models internal/repositories internal/services config utils cmd/server tests

# Help target
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  run                - Run the application"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  clean              - Clean build artifacts"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-compose     - Run with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose services"
	@echo "  lint               - Run linter"
	@echo "  init               - Create necessary directories" 