.PHONY: all build docker-build docker-up docker-down logs test test-e2e clean deps help

# Variables
PROJECT_NAME=shortener
DOCKER_COMPOSE=docker compose
ENV_FILE=.env

# Read the system name
ifeq ($(OS),Windows_NT)
    PLATFORM = Windows
else
    PLATFORM := $(shell uname -s)
endif

# Make command to extract DB_USE_IN_MEMORY value from env file
ifeq ($(PLATFORM),Windows)
    READ_DB_ENV_CMD = powershell -Command "(Get-Content $(ENV_FILE) | Select-String '^DB_USE_IN_MEMORY=') -replace '^DB_USE_IN_MEMORY=', ''"
    READ_PORT_ENV_CMD = powershell -Command "(Get-Content $(ENV_FILE) | Select-String '^SERVER_PORT=') -replace '^SERVER_PORT=', ''"
else
    READ_DB_ENV_CMD = grep -E '^DB_USE_IN_MEMORY=' $(ENV_FILE) | cut -d '=' -f2
    READ_PORT_ENV_CMD = grep -E '^SERVER_PORT=' $(ENV_FILE) | cut -d '=' -f2
endif

# Acquire value and set default
DB_USE_IN_MEMORY := $(shell $(READ_DB_ENV_CMD))

ifeq ($(DB_USE_IN_MEMORY),)
    DB_USE_IN_MEMORY := true
endif

# Make profile argument for compose
ifeq ($(DB_USE_IN_MEMORY),true)
    DOCKER_PROFILE :=
else
    DOCKER_PROFILE := --profile with-postgres
endif

# Acquire value and set default
SERVER_PORT := $(shell $(READ_PORT_ENV_CMD))

ifeq ($(SERVER_PORT),)
    SERVER_PORT := 8080
endif

all: docker-up

# Build project
build:
	@echo "Building $(PROJECT_NAME)..."
	@go build -o bin/$(PROJECT_NAME) ./cmd/shortener
	@echo "Build complete."

# Build Docker image
docker-build: build
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build --build-arg SERVER_PORT=$(SERVER_PORT)
	@echo "Docker images built."

# Start Docker Compose
docker-up: docker-build
	@echo "Starting services with DB_USE_IN_MEMORY=$(DB_USE_IN_MEMORY)..."
	@$(DOCKER_COMPOSE) $(DOCKER_PROFILE) up -d
	@echo "Services started. Shortener available at http://localhost:$(SERVER_PORT)"

# Stop Docker Compose
docker-down:
	@echo "Stopping Docker Compose services..."
	-@$(DOCKER_COMPOSE) down
	-@$(DOCKER_COMPOSE) --profile with-postgres down
	@echo "Services stopped."

# View logs
logs:
	@$(DOCKER_COMPOSE) logs -f

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./... --cover

test-e2e: docker-build
	@echo "Running e2e tests..."
	@go test -v -tags=e2e ./tests/e2e/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Cleaning complete."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded."

# Help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make docker-build - Build the Docker image"
	@echo "  make docker-up    - Start service"
	@echo "  make docker-down  - Stop all containers"
	@echo "  make logs         - View compose logs"
	@echo "  make test         - Run unit tests"
	@echo "  make test-e2e     - Run e2e tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make help         - Show this help"
