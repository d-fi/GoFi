# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Name of binary
BINARY_NAME = d-fi

# Build the binary
build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "-s -w" -o $(BINARY_NAME) cmd/main.go

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Default target
default: build

.PHONY: build clean test