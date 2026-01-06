# Variables
APP_NAME := taskerr
BUILD_DIR := build
GO_FILES := $(shell find . -type f -name '*.go')

# Default target
all: build

# Build the application
build: $(GO_FILES)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Run the application
run: build
	@echo "Running $(APP_NAME) with arguments: $(filter-out $@,$(MAKECMDGOALS))..."
	@./$(BUILD_DIR)/$(APP_NAME) $(filter-out $@,$(MAKECMDGOALS))

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

rebuild: clean build

.PHONY: all build run clean rebuild

