BINARY_NAME=workouts-backend
IMAGE_NAME=workouts-backend

RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m

.PHONY: run build test clean deps fmt lint run-env wire help migrate-up migrate-down migrate-version migrate-force

run:
	@echo "$(BLUE)Starting server...$(NC)"
	go run ./cmd

build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	go build -o $(BINARY_NAME) ./cmd
	@echo "$(GREEN)Build completed: $(BINARY_NAME)$(NC)"

deps:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)Dependencies installed!$(NC)"

wire:
	@echo "$(BLUE)Generating Wire code...$(NC)"
	go generate ./cmd
	@echo "$(GREEN)Wire code generated!$(NC)"

test:
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...

test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -cover ./...

clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	rm -f $(BINARY_NAME)
	go clean
	@echo "$(GREEN)Clean completed!$(NC)"

fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

lint:
	@echo "$(BLUE)Linting code...$(NC)"
	golangci-lint run
	@echo "$(GREEN)Linting completed!$(NC)"

run-env:
	@echo "$(BLUE)Starting server with .env variables...$(NC)"
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs) && go run ./cmd; \
	else \
		echo "$(YELLOW).env file not found. Using default values.$(NC)"; \
		go run ./cmd; \
	fi

migrate-up:
	@echo "$(BLUE)Running migrations UP...$(NC)"
	cd ../migrate && go run main.go -command=up -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-down:
	@echo "$(BLUE)Running migrations DOWN...$(NC)"
	@read -p "Are you sure you want to rollback migrations? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	cd ../migrate && go run main.go -command=down -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-version:
	@echo "$(BLUE)Checking migration version...$(NC)"
	cd ../migrate && go run main.go -command=version -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required$(NC)"; \
		echo "$(YELLOW)Example: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	cd ../migrate && go run main.go -command=force -version=$(VERSION) -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

help:
	@echo "$(BLUE)Available commands:$(NC)"
	@echo "  $(GREEN)run$(NC)             - Start server"
	@echo "  $(GREEN)build$(NC)           - Build binary file"
	@echo "  $(GREEN)deps$(NC)            - Install dependencies"
	@echo "  $(GREEN)wire$(NC)            - Generate Wire DI code"
	@echo "  $(GREEN)test$(NC)            - Run tests"
	@echo "  $(GREEN)test-coverage$(NC)   - Run tests with coverage"
	@echo "  $(GREEN)fmt$(NC)             - Format code"
	@echo "  $(GREEN)lint$(NC)            - Lint code"
	@echo "  $(GREEN)clean$(NC)           - Clean compiled files"
	@echo "  $(GREEN)run-env$(NC)         - Run with variables from .env"
	@echo "  $(GREEN)migrate-up$(NC)      - Run database migrations (up)"
	@echo "  $(GREEN)migrate-down$(NC)    - Rollback database migrations (down)"
	@echo "  $(GREEN)migrate-version$(NC) - Show current migration version"
	@echo "  $(GREEN)migrate-force$(NC)   - Force migration version (VERSION=<n>)"
	@echo "  $(GREEN)help$(NC)            - Show this help"

.DEFAULT_GOAL := help
