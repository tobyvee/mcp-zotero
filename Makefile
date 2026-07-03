# mcp-zotero — build & dev tasks. Binaries are written to ./bin.

BINARY   := mcp-zotero
PKG      := ./cmd/mcp-zotero
BIN_DIR  := bin
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)

# Platforms for `make cross`, as GOOS/GOARCH pairs.
PLATFORMS := darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64

# Where `make install-darwin` puts the binary.
INSTALL_DIR ?= $(HOME)/.local/bin

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the server for the host platform into ./bin
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(PKG)

.PHONY: cross
cross: ## Cross-compile for every platform in $(PLATFORMS) into ./bin
	@mkdir -p $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		out=$(BIN_DIR)/$(BINARY)-$$os-$$arch$$ext; \
		echo "  $$out"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -ldflags "$(LDFLAGS)" -o $$out $(PKG) || exit 1; \
	done

.PHONY: install-darwin
install-darwin: ## Build for macOS (host arch) and install to ~/.local/bin (override INSTALL_DIR)
	@mkdir -p $(INSTALL_DIR)
	GOOS=darwin CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(INSTALL_DIR)/$(BINARY) $(PKG)
	@echo "installed $(BINARY) $(VERSION) -> $(INSTALL_DIR)/$(BINARY)"

.PHONY: run
run: ## Run the server over stdio (talks to Zotero's local API)
	go run $(PKG)

.PHONY: test
test: ## Run all tests with the race detector
	go test -race ./...

.PHONY: cover
cover: ## Run tests and open an HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run go vet, gofmt, and golangci-lint (if installed)
	go vet ./...
	@gofmt_out=$$(gofmt -l .); \
	if [ -n "$$gofmt_out" ]; then echo "gofmt needed:"; echo "$$gofmt_out"; exit 1; fi
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; ran go vet + gofmt only"; \
	fi

.PHONY: fmt
fmt: ## Format all Go source
	gofmt -w .

.PHONY: tidy
tidy: ## Sync go.mod/go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out

.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
