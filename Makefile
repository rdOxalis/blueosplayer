# Makefile for bluesoundplayer
# Multi-platform build system

# Variables
BINARY_NAME = bluesoundplayer
SRC_DIR = src
RELEASE_DIR = release
SOURCE_FILES = $(SRC_DIR)/*.go

# Version info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date +%Y-%m-%d)
BUILD_TIME := $(shell date +%H:%M:%S)

# Default target
.PHONY: all
all: clean linux windows raspberry apple

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning old builds..."
	@rm -f $(RELEASE_DIR)/$(BINARY_NAME)-*
	@rm -f $(RELEASE_DIR)/BUILD_INFO.txt

# Linux AMD64
.PHONY: linux
linux:
	@echo "🐧 Building for Linux AMD64..."
	@cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64

# Windows AMD64
.PHONY: windows
windows:
	@echo "🪟 Building for Windows AMD64..."
	@cd $(SRC_DIR) && GOOS=windows GOARCH=amd64 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-win-amd64.exe *.go

# Raspberry Pi builds
.PHONY: raspberry
raspberry: rpi-armv6 rpi-armv7 rpi-arm64

.PHONY: rpi-armv6
rpi-armv6:
	@echo "🥧 Building for Raspberry Pi ARMv6 (Pi 1/Zero/Zero W)..."
	@cd $(SRC_DIR) && GOOS=linux GOARCH=arm GOARM=6 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-armv6 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-linux-armv6

.PHONY: rpi-armv7
rpi-armv7:
	@echo "🥧 Building for Raspberry Pi ARMv7 (Pi 2/3/4/Zero 2 W)..."
	@cd $(SRC_DIR) && GOOS=linux GOARCH=arm GOARM=7 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-armv7 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-linux-armv7

.PHONY: rpi-arm64
rpi-arm64:
	@echo "🥧 Building for Raspberry Pi ARM64 (Pi 4/5)..."
	@cd $(SRC_DIR) && GOOS=linux GOARCH=arm64 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64

# Apple builds
.PHONY: apple
apple: apple-arm64 apple-amd64

.PHONY: apple-arm64
apple-arm64:
	@echo "🍎 Building for Apple Silicon (M1/M2/M3)..."
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=arm64 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-apple-arm64 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-apple-arm64

.PHONY: apple-amd64
apple-amd64:
	@echo "🍎 Building for Intel Mac..."
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=amd64 go build -o ../$(RELEASE_DIR)/$(BINARY_NAME)-apple-amd64 *.go
	@chmod +x $(RELEASE_DIR)/$(BINARY_NAME)-apple-amd64

# Build info
.PHONY: build-info
build-info:
	@echo "📝 Creating build info..."
	@echo "Build Version: $(VERSION)" > $(RELEASE_DIR)/BUILD_INFO.txt
	@echo "Build Date: $(BUILD_DATE) $(BUILD_TIME)" >> $(RELEASE_DIR)/BUILD_INFO.txt
	@echo "" >> $(RELEASE_DIR)/BUILD_INFO.txt
	@echo "Platforms built:" >> $(RELEASE_DIR)/BUILD_INFO.txt
	@ls $(RELEASE_DIR)/$(BINARY_NAME)-* 2>/dev/null | sed 's/$(RELEASE_DIR)\//  - /' >> $(RELEASE_DIR)/BUILD_INFO.txt || echo "  No binaries found" >> $(RELEASE_DIR)/BUILD_INFO.txt

# Show built files
.PHONY: list
list:
	@echo "📦 Built executables:"
	@ls -la $(RELEASE_DIR)/$(BINARY_NAME)-* 2>/dev/null || echo "No binaries found"

# Run locally (development)
.PHONY: run
run:
	@echo "🚀 Running bluesoundplayer..."
	@cd $(SRC_DIR) && go run *.go

# Test build (just compile, don't save)
.PHONY: test
test:
	@echo "🧪 Testing build..."
	@cd $(SRC_DIR) && go build -o /tmp/$(BINARY_NAME)-test *.go && rm /tmp/$(BINARY_NAME)-test
	@echo "✅ Build test successful!"

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Checking dependencies..."
	@cd $(SRC_DIR) && go mod download 2>/dev/null || echo "No go.mod file found"

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	@cd $(SRC_DIR) && go fmt *.go

# Help
.PHONY: help
help:
	@echo "BluOS Player - Makefile targets:"
	@echo ""
	@echo "  make all          - Build for all platforms"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make linux        - Build for Linux AMD64"
	@echo "  make windows      - Build for Windows AMD64"
	@echo "  make raspberry    - Build for all Raspberry Pi variants"
	@echo "  make apple        - Build for macOS (Intel and Apple Silicon)"
	@echo "  make run          - Run locally for development"
	@echo "  make test         - Test if project builds"
	@echo "  make fmt          - Format Go code"
	@echo "  make list         - List built executables"
	@echo "  make build-info   - Create build info file"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Individual platform targets:"
	@echo "  make rpi-armv6    - Raspberry Pi 1/Zero/Zero W"
	@echo "  make rpi-armv7    - Raspberry Pi 2/3/4/Zero 2 W"
	@echo "  make rpi-arm64    - Raspberry Pi 4/5 (64-bit OS)"
	@echo "  make apple-arm64  - Apple Silicon Macs"
	@echo "  make apple-amd64  - Intel Macs"

# Set default goal
.DEFAULT_GOAL := help