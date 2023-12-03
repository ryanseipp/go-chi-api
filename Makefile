# Simple Makefile for a Go project

# Build the application
all: build

swagger:
	@if command -v swag > /dev/null; then \
		echo "Generating swagger docs..." \
	    swag init -d cmd/api,internal/server; \
	else \
	    read -p "Go's 'swag' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/swaggo/swag/cmd/swag@latest; \
			echo "Generating swagger docs..." \
			swag init -d cmd/api,internal/server; \
	    else \
	        echo "You chose not to install swag. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

build: swagger
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run: build
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/cosmtrek/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

.PHONY: all build swagger run test clean
