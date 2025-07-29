# Makefile for who-exporter
BINARY_NAME=who-exporter
DOCKER_IMAGE=who-exporter
GO=go
CGO_ENABLED=0
LDFLAGS=-w -s

# Default host and port
HOST=localhost
PORT=9101

.PHONY: all build run test docker-build docker-run clean

all: build

# Build the static binary
build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/who-exporter

# Run the exporter with configurable host and port
run:
	./$(BINARY_NAME) -host=$(HOST) -port=$(PORT)

# Run tests (placeholder, no tests implemented yet)
test:
	$(GO) test ./...

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	docker run --rm -p $(PORT):$(PORT) $(DOCKER_IMAGE) -host=0.0.0.0 -port=$(PORT)

# Clean up generated files
clean:
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE) || true
