.PHONY: all build install clean test run dev deps

# Binary name
BINARY_NAME=hubitat-go-mcp
INSTALL_PATH=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: deps build

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) -v

# Install binary to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

# Install systemd service
install-service: install
	@echo "Installing systemd service..."
	sudo cp hubitat-go-mcp.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable hubitat-go-mcp.service
	@echo "Service installed. Start with: sudo systemctl start hubitat-go-mcp"

# Uninstall systemd service
uninstall-service:
	@echo "Uninstalling systemd service..."
	sudo systemctl stop hubitat-go-mcp.service || true
	sudo systemctl disable hubitat-go-mcp.service || true
	sudo rm -f /etc/systemd/system/hubitat-go-mcp.service
	sudo systemctl daemon-reload

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Development mode with auto-reload (requires 'air' tool)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@echo "Starting development mode..."
	air

# Show help
help:
	@echo "Hubitat Go MCP Server - Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make deps              - Download dependencies"
	@echo "  make build             - Build the binary"
	@echo "  make install           - Install binary to $(INSTALL_PATH)"
	@echo "  make install-service   - Install and enable systemd service"
	@echo "  make uninstall-service - Uninstall systemd service"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make test              - Run tests"
	@echo "  make run               - Build and run the application"
	@echo "  make dev               - Run in development mode with auto-reload"
	@echo "  make help              - Show this help message"
