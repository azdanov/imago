# Variables
APP_NAME = imago
BUILD_DIR = build
MAIN_FILE = main.go
ENV_FILE = .env
GOOSE_DIR = database/migrations
DB_CONN_STRING = postgres "user=$(DB_USER) password=$(DB_PASSWORD) host=$(DB_HOST) port=$(DB_PORT) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)"

# Include environment variables
-include $(ENV_FILE)

# Docker related variables
DOCKER_COMPOSE = docker-compose
DOCKER_COMPOSE_FILE = docker-compose.yaml

# Set default shell to bash
SHELL := /bin/bash

.PHONY: help build run dev test migrate-create migrate-fix migrate-up migrate-down migrate-reset docker-up docker-down docker-logs init clean deps lint

.DEFAULT_GOAL := help

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

run: ## Run the application
	@go run $(MAIN_FILE)

dev: ## Run the application with hot-reload using air
	@command -v air >/dev/null 2>&1 || { echo "air is not installed. Installing..."; go install github.com/cosmtrek/air@latest; }
	@air

test: ## Run tests
	@go test -v ./...

migrate-create: ## Create a new migration file (usage: make migrate-create name=migration_name)
	@[ -z "$(name)" ] && echo "Error: name parameter is required" && exit 1 || go tool goose -dir $(GOOSE_DIR) create $(name) sql

migrate-fix: ## Fix the migration files to use incremental prefix
	@go tool goose -dir $(GOOSE_DIR) fix $(name)

migrate-up: ## Apply all migrations
	@go tool goose -dir $(GOOSE_DIR) $(DB_CONN_STRING) up

migrate-down: ## Revert the last migration
	@go tool goose -dir $(GOOSE_DIR) $(DB_CONN_STRING) down

migrate-reset: ## Revert all migrations
	@go tool goose -dir $(GOOSE_DIR) $(DB_CONN_STRING) reset

docker-up: ## Start Docker containers
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

docker-down: ## Stop Docker containers
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## Show Docker logs
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

init: ## Initialize the project (start Docker and apply migrations)
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(ENV_FILE) not found. Please create it first."; \
		exit 1; \
	fi
	@make docker-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@make migrate-up
	@echo "Project initialized successfully!"

clean: ## Clean build artifacts
	@rm -rf $(BUILD_DIR)
	@go clean

deps: ## Install or update dependencies
	@go mod tidy
	@go mod download

lint: ## Run linters
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is not installed. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@golangci-lint run ./...
	@echo "Lint completed"
