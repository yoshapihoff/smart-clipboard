# Simple Makefile for building smart-clipboard
# -------------------------------------------
# Targets:
#   make            – default, build with CGO enabled (system-tray version)
#   make build      – same as default
#   make headless   – build without CGO (works in minimal containers)
#   make clean      – remove build artifacts in ./bin

BINARY_NAME := smart-clipboard
BIN_DIR     := bin
PACKAGE     := ./cmd

# Allow overriding GOOS / GOARCH from the command line, eg:
#   make GOOS=windows GOARCH=amd64 headless
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# -------------------------------------------
# Default target
# -------------------------------------------
.PHONY: default
default: build

# -------------------------------------------
# Build with CGO (system tray enabled)
# -------------------------------------------
.PHONY: build
build:
	@echo "[make] Building $(PACKAGE) for $(GOOS)/$(GOARCH) with CGO enabled → $(BIN_DIR)/$(BINARY_NAME)"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -o $(BIN_DIR)/$(BINARY_NAME) $(PACKAGE)

# -------------------------------------------
# Headless build without CGO (tray UI disabled)
# -------------------------------------------
.PHONY: headless
headless:
	@echo "[make] Building headless version for $(GOOS)/$(GOARCH) with CGO disabled → $(BIN_DIR)/$(BINARY_NAME)-headless"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -o $(BIN_DIR)/$(BINARY_NAME)-headless $(PACKAGE)

# -------------------------------------------
# Clean artifacts
# -------------------------------------------
.PHONY: clean
clean:
	@echo "[make] Removing $(BIN_DIR) directory"
	@rm -rf $(BIN_DIR)
