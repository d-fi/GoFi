# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Binary names
BINARY_NAME = d-fi
NEW_BINARY_NAME = gofi

# Installation directory
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

# Detect OS for installation
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    INSTALL_CMD = ln -sf
else
    INSTALL_CMD = ln -sf
endif

# Build the legacy binary
build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "-s -w" -o $(BINARY_NAME) cmd/main.go

# Build the new CLI binary
build-cli:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "-s -w -X github.com/d-fi/GoFi/cmd/gofi/cmd.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o $(NEW_BINARY_NAME) cmd/gofi/main.go

# Build all binaries
build-all: build build-cli

# Install the new CLI binary
install: build-cli
	@echo "Installing $(NEW_BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	@if [ -w $(BINDIR) ]; then \
		cp $(NEW_BINARY_NAME) $(BINDIR)/$(NEW_BINARY_NAME); \
		chmod 755 $(BINDIR)/$(NEW_BINARY_NAME); \
		echo "✓ Installed $(NEW_BINARY_NAME) to $(BINDIR)"; \
	else \
		echo "Installing to $(BINDIR) (requires sudo)..."; \
		sudo cp $(NEW_BINARY_NAME) $(BINDIR)/$(NEW_BINARY_NAME); \
		sudo chmod 755 $(BINDIR)/$(NEW_BINARY_NAME); \
		echo "✓ Installed $(NEW_BINARY_NAME) to $(BINDIR)"; \
	fi
	@echo "Run 'gofi --help' to get started"

# Install using symlink (for development)
install-dev: build-cli
	@echo "Creating symlink for $(NEW_BINARY_NAME) in $(BINDIR)..."
	@mkdir -p $(BINDIR)
	@if [ -w $(BINDIR) ]; then \
		$(INSTALL_CMD) $(PWD)/$(NEW_BINARY_NAME) $(BINDIR)/$(NEW_BINARY_NAME); \
		echo "✓ Symlinked $(NEW_BINARY_NAME) to $(BINDIR)"; \
	else \
		echo "Creating symlink in $(BINDIR) (requires sudo)..."; \
		sudo $(INSTALL_CMD) $(PWD)/$(NEW_BINARY_NAME) $(BINDIR)/$(NEW_BINARY_NAME); \
		echo "✓ Symlinked $(NEW_BINARY_NAME) to $(BINDIR)"; \
	fi
	@echo "Run 'gofi --help' to get started"

# Uninstall
uninstall:
	@echo "Removing $(NEW_BINARY_NAME) from $(BINDIR)..."
	@if [ -w $(BINDIR)/$(NEW_BINARY_NAME) ]; then \
		rm -f $(BINDIR)/$(NEW_BINARY_NAME); \
	else \
		sudo rm -f $(BINDIR)/$(NEW_BINARY_NAME); \
	fi
	@echo "✓ Uninstalled $(NEW_BINARY_NAME)"

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(NEW_BINARY_NAME)

# Run tests
test:
	$(GOCLEAN) -testcache
	$(GOTEST) -v ./...

# Default target
default: build-cli

.PHONY: build build-cli build-all install install-dev uninstall clean test default