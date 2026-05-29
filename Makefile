GOCMD ?= go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test

BINARY_NAME ?= d-fi
CLI_PACKAGE ?= ./cmd/d-fi
BUILD_DIR ?= build
LDFLAGS ?= -s -w

build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CLI_PACKAGE)

pkg: clean-pkg
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CLI_PACKAGE)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CLI_PACKAGE)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 $(CLI_PACKAGE)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 $(CLI_PACKAGE)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-win-amd64.exe $(CLI_PACKAGE)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-win-arm64.exe $(CLI_PACKAGE)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-linux-amd64 $(BINARY_NAME) && zip -q $(BINARY_NAME)-linux.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-linux-arm64 $(BINARY_NAME) && zip -q $(BINARY_NAME)-linux-arm64.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-macos-amd64 $(BINARY_NAME) && zip -q $(BINARY_NAME)-macos.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-macos-arm64 $(BINARY_NAME) && zip -q $(BINARY_NAME)-macos-arm64.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-win-amd64.exe $(BINARY_NAME).exe && cp ../scripts/windows/$(BINARY_NAME).bat $(BINARY_NAME).bat && zip -q $(BINARY_NAME)-win.zip $(BINARY_NAME).exe $(BINARY_NAME).bat && rm $(BINARY_NAME).exe $(BINARY_NAME).bat
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-win-arm64.exe $(BINARY_NAME).exe && cp ../scripts/windows/$(BINARY_NAME).bat $(BINARY_NAME).bat && zip -q $(BINARY_NAME)-win-arm64.zip $(BINARY_NAME).exe $(BINARY_NAME).bat && rm $(BINARY_NAME).exe $(BINARY_NAME).bat
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 $(BUILD_DIR)/$(BINARY_NAME)-win-amd64.exe $(BUILD_DIR)/$(BINARY_NAME)-win-arm64.exe
	du -sh $(BUILD_DIR)/*.zip

clean-pkg:
	rm -rf $(BUILD_DIR)

clean: clean-pkg
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	$(GOCLEAN) -testcache
	$(GOTEST) -v ./...

.PHONY: build pkg clean-pkg clean test
