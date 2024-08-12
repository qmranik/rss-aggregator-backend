# Directories
ROOT_DIR := $(shell pwd)
SCHEMA_DIR := $(ROOT_DIR)/sql/schema
BIN_DIR := $(ROOT_DIR)/bin

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build

# Tools
SQLC := sqlc
MIGRATE := goose

# Database connection string
DB_URL := "postgres://postgres:@localhost:5432/postgres?sslmode=disable"

# Build the project
build:
	$(GOBUILD) -o $(BIN_DIR)/main .

# Run the built binary
run: build
	./$(BIN_DIR)/main

# Generate code with sqlc
sqlc-gen:
	$(SQLC) generate

# Run database migrations up
migrate-up:
	cd $(SCHEMA_DIR) && $(MIGRATE) postgres $(DB_URL) up

# Rollback database migrations down
migrate-down:
	cd $(SCHEMA_DIR) && $(MIGRATE) postgres $(DB_URL) down

