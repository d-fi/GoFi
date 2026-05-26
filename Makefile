GOCMD ?= go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test

BINARY_NAME ?= d-fi
CLI_PACKAGE ?= ./cmd/d-fi
BUILD_DIR ?= build
PKG_ARCH ?= amd64
LDFLAGS ?= -s -w

build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CLI_PACKAGE)

pkg: clean-pkg
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=$(PKG_ARCH) CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CLI_PACKAGE)
	GOOS=darwin GOARCH=$(PKG_ARCH) CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(CLI_PACKAGE)
	GOOS=windows GOARCH=$(PKG_ARCH) CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-win.exe $(CLI_PACKAGE)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-linux $(BINARY_NAME) && zip -q $(BINARY_NAME)-linux.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-macos $(BINARY_NAME) && zip -q $(BINARY_NAME)-macos.zip $(BINARY_NAME) && rm $(BINARY_NAME)
	cd $(BUILD_DIR) && cp $(BINARY_NAME)-win.exe $(BINARY_NAME).exe && zip -q $(BINARY_NAME)-win.zip $(BINARY_NAME).exe && rm $(BINARY_NAME).exe
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-linux $(BUILD_DIR)/$(BINARY_NAME)-macos $(BUILD_DIR)/$(BINARY_NAME)-win.exe
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
