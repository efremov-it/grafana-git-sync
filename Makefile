.PHONY: build clean test test-coverage run

BINARY_NAME=grafana-git-sync
BUILD_DIR=bin
CMD_DIR=cmd/grafana-git-sync

build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

test:
	@echo "Testing..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

install:
	@echo "Installing..."
	@go install ./$(CMD_DIR)

.DEFAULT_GOAL := build
