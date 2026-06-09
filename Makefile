# LittleSecrets Makefile

# Parameters
BINARY_NAME=littlesecrets
BIN_DIR=bin
PORT?=8080

.PHONY: all build run test fmt vet lint clean help

all: help

## build: Build the server binary
build:
	@echo "Building binary..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) main.go
	@echo "Binary built successfully at $(BIN_DIR)/$(BINARY_NAME)"

## run: Run the server locally using go run
run:
	@echo "Starting server on port $(PORT)..."
	PORT=$(PORT) go run main.go

## run-bin: Build and run the compiled binary
run-bin: build
	@echo "Starting compiled binary on port $(PORT)..."
	PORT=$(PORT) ./$(BIN_DIR)/$(BINARY_NAME)

## test: Run the test suite
test:
	@echo "Running tests..."
	go test -v -race ./...

## fmt: Format codebase
fmt:
	@echo "Formatting Go files..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Vetting Go files..."
	go vet ./...

## lint: Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, falling back to go vet..."; \
		go vet ./...; \
	fi

## clean: Clean build binaries and temp files
clean:
	@echo "Cleaning build directory and temporary files..."
	rm -rf $(BIN_DIR)
	rm -f *.out *.test *.prof *.cov
	@echo "Clean completed."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
