BINARY_NAME=workouts-backend
IMAGE_NAME=workouts-backend
CONTAINER_CONFIG_FILE=.container-config.json

RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m

.PHONY: run build test clean deps fmt lint run-env wire help migrate-up migrate-down migrate-version migrate-force docker-login docker-build docker-clean docker-push docker-registry-list docker-registry-create docker-image-list deploy-revision save-container-config deploy

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
	cd ../../migrate && go run main.go -command=up -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-down:
	@echo "$(BLUE)Running migrations DOWN...$(NC)"
	@read -p "Are you sure you want to rollback migrations? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	cd ../../migrate && go run main.go -command=down -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-version:
	@echo "$(BLUE)Checking migration version...$(NC)"
	cd ../../migrate && go run main.go -command=version -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required$(NC)"; \
		echo "$(YELLOW)Example: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	cd ../../migrate && go run main.go -command=force -version=$(VERSION) -schema=$${DB_SCHEMA:-workouts} -path=../workouts/backend/migrations

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
	@echo "  $(GREEN)docker-login$(NC)    - Login to Yandex Cloud Container Registry"
	@echo "  $(GREEN)docker-build$(NC)    - Build Docker image"
	@echo "  $(GREEN)docker-clean$(NC)    - Clean unused Docker resources"
	@echo "  $(GREEN)docker-push$(NC)     - Push Docker image to registry"
	@echo "  $(GREEN)docker-registry-list$(NC) - List registries"
	@echo "  $(GREEN)docker-registry-create$(NC) - Create new registry"
	@echo "  $(GREEN)docker-image-list$(NC) - List images in registry"
	@echo "  $(GREEN)deploy-revision$(NC) - Create new serverless container revision"
	@echo "    (options: CONTAINER_NAME, REGISTRY_ID, SERVICE_ACCOUNT_NAME, SERVICE_ACCOUNT_ID)"
	@echo "  $(GREEN)save-container-config$(NC) - Save current container configuration to file"
	@echo "  $(GREEN)deploy$(NC)          - Full deployment (docker-push + deploy-revision)"
	@echo "  $(GREEN)help$(NC)            - Show this help"

docker-login:
	@echo "$(BLUE)Logging in to Yandex Cloud Container Registry...$(NC)"
	@echo "$(YELLOW)Make sure you have run 'yc init' previously$(NC)"
	docker login cr.yandex --username iam --password $$(yc iam create-token)
	@echo "$(GREEN)Login successful!$(NC)"

docker-build:
	@echo "$(BLUE)Building Docker image for linux/amd64...$(NC)"
	@if docker buildx version > /dev/null 2>&1; then \
		echo "$(BLUE)Using Docker Buildx...$(NC)"; \
		docker buildx build --platform linux/amd64 -t $(IMAGE_NAME) -f docker/Dockerfile . --load; \
	else \
		echo "$(YELLOW)Buildx not available, using legacy builder with explicit platform...$(NC)"; \
		docker build --platform linux/amd64 -t $(IMAGE_NAME) -f docker/Dockerfile .; \
	fi
	@echo "$(GREEN)Docker image built!$(NC)"

docker-push: docker-build
	@echo "$(BLUE)Pushing image to Yandex Cloud Registry...$(NC)"
	@if [ -z "$(REGISTRY_ID)" ]; then \
		echo "$(YELLOW)REGISTRY_ID not specified, trying to get automatically...$(NC)"; \
		REGISTRY_ID=$$(yc container registry list --format=json | jq -r '.[0].id' 2>/dev/null); \
		if [ -z "$$REGISTRY_ID" ] || [ "$$REGISTRY_ID" = "null" ]; then \
			echo "$(RED)Error: Failed to get registry ID automatically$(NC)"; \
			echo "$(YELLOW)Please specify registry ID via REGISTRY_ID variable$(NC)"; \
			echo "$(YELLOW)Or create a new registry: make docker-registry-create$(NC)"; \
			echo "$(YELLOW)Example: make docker-push REGISTRY_ID=your-registry-id$(NC)"; \
			exit 1; \
		else \
			echo "$(GREEN)Found registry with ID: $$REGISTRY_ID$(NC)"; \
			echo "$(BLUE)Tagging image...$(NC)"; \
			docker tag $(IMAGE_NAME) cr.yandex/$$REGISTRY_ID/$(IMAGE_NAME):latest; \
			echo "$(BLUE)Pushing image to registry...$(NC)"; \
			docker push cr.yandex/$$REGISTRY_ID/$(IMAGE_NAME):latest; \
		fi; \
	else \
		echo "$(GREEN)Using specified registry ID: $(REGISTRY_ID)$(NC)"; \
		echo "$(BLUE)Tagging image...$(NC)"; \
		docker tag $(IMAGE_NAME) cr.yandex/$(REGISTRY_ID)/$(IMAGE_NAME):latest; \
		echo "$(BLUE)Pushing image to registry...$(NC)"; \
		docker push cr.yandex/$(REGISTRY_ID)/$(IMAGE_NAME):latest; \
	fi
	@echo "$(GREEN)Image successfully pushed to registry!$(NC)"

docker-registry-list:
	@echo "$(BLUE)List of registries in Yandex Cloud:$(NC)"
	@yc container registry list || echo "$(RED)Error: Failed to get registry list. Make sure you are logged in to Yandex Cloud (yc init)$(NC)"

docker-registry-create:
	@echo "$(BLUE)Creating new registry...$(NC)"
	@yc container registry create --name workouts-registry || echo "$(RED)Error: Failed to create registry$(NC)"
	@echo "$(GREEN)Registry created! Use 'make docker-registry-list' to get ID$(NC)"

docker-image-list:
	@echo "$(BLUE)List of images in Yandex Cloud Registry:$(NC)"
	@yc container image list || echo "$(RED)Error: Failed to get image list. Make sure you are logged in to Yandex Cloud (yc init)$(NC)"

docker-clean:
	@echo "$(BLUE)Cleaning unused Docker resources...$(NC)"
	@echo "$(YELLOW)This will remove:$(NC)"
	@echo "$(YELLOW)  - All stopped containers$(NC)"
	@echo "$(YELLOW)  - All unused images (not just dangling)$(NC)"
	@echo "$(YELLOW)  - All unused networks$(NC)"
	@echo "$(YELLOW)  - All build cache$(NC)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	docker system prune -a --volumes -f
	@echo "$(GREEN)Docker cleanup completed!$(NC)"

deploy-revision:
	@IMAGE_NAME=$(IMAGE_NAME) CONTAINER_NAME=$(CONTAINER_NAME) CONTAINER_CONFIG_FILE=$(CONTAINER_CONFIG_FILE) REGISTRY_ID=$(REGISTRY_ID) SERVICE_ACCOUNT_ID=$(SERVICE_ACCOUNT_ID) SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME) ./scripts/deploy-revision.sh

save-container-config:
	@CONTAINER_NAME=$(CONTAINER_NAME) CONTAINER_CONFIG_FILE=$(CONTAINER_CONFIG_FILE) ./scripts/save-container-config.sh

deploy:
	@IMAGE_NAME=$(IMAGE_NAME) CONTAINER_NAME=$(CONTAINER_NAME) CONTAINER_CONFIG_FILE=$(CONTAINER_CONFIG_FILE) REGISTRY_ID=$(REGISTRY_ID) SERVICE_ACCOUNT_ID=$(SERVICE_ACCOUNT_ID) SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME) ./scripts/deploy.sh

.DEFAULT_GOAL := help
