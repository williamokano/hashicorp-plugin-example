.DEFAULT_GOAL := help
.PHONY: help all build clean proto install test test-verbose test-coverage test-coverage-report test-race generate deps run-example dev-setup

VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X github.com/williamokano/hashicorp-plugin-example/internal/version.CLIVersion=$(VERSION) \
	-X github.com/williamokano/hashicorp-plugin-example/internal/version.CLIBuildTime=$(BUILD_TIME) \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m\033[0m"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

all: proto build ## Generate protobuf and build everything

proto: ## Generate protobuf files from .proto definitions
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/protocol/plugin.proto

build: build-cli build-plugins ## Build CLI and all plugins

build-cli: ## Build the CLI binary only
	@echo "Building CLI..."
	go build $(LDFLAGS) -o bin/plugin-cli cmd/cli/main.go

build-plugins: build-dummy-plugin build-filter-plugin build-converter-plugin build-uploader-plugin ## Build all plugin binaries

build-dummy-plugin: ## Build the dummy example plugin
	@echo "Building dummy plugin..."
	go build $(LDFLAGS) -o bin/plugin-dummy plugins/dummy/main.go

build-filter-plugin: ## Build the message filter plugin
	@echo "Building filter plugin..."
	go build $(LDFLAGS) -o bin/plugin-filter plugins/filter/main.go

build-converter-plugin: ## Build the media converter plugin
	@echo "Building converter plugin..."
	go build $(LDFLAGS) -o bin/plugin-converter plugins/converter/main.go

build-uploader-plugin: ## Build the file uploader plugin
	@echo "Building uploader plugin..."
	go build $(LDFLAGS) -o bin/plugin-uploader plugins/uploader/main.go

##@ Installation

install: build ## Install CLI to system and plugins to local .plugins directory
	@echo "Creating local .plugins directory (like .terraform)..."
	mkdir -p .plugins
	@echo "Installing plugins to .plugins/..."
	cp bin/plugin-dummy .plugins/ 2>/dev/null || true
	cp bin/plugin-filter .plugins/ 2>/dev/null || true
	cp bin/plugin-converter .plugins/ 2>/dev/null || true
	cp bin/plugin-uploader .plugins/ 2>/dev/null || true
	chmod +x .plugins/plugin-* 2>/dev/null || true
	@echo "Installing CLI to /usr/local/bin..."
	sudo cp bin/plugin-cli /usr/local/bin/
	sudo chmod +x /usr/local/bin/plugin-cli
	@echo ""
	@echo "Installation complete!"
	@echo "CLI installed to: /usr/local/bin/plugin-cli"
	@echo "Plugins installed to: ./.plugins/"
	@echo ""
	@echo "The .plugins directory is project-local (like .terraform)."
	@echo "Add it to .gitignore to avoid committing binaries."

install-local: build ## Install plugins to local .plugins directory (no sudo required)
	@echo "Creating local .plugins directory..."
	mkdir -p .plugins
	@echo "Installing plugins to .plugins/..."
	cp bin/plugin-dummy .plugins/ 2>/dev/null || true
	cp bin/plugin-filter .plugins/ 2>/dev/null || true
	cp bin/plugin-converter .plugins/ 2>/dev/null || true
	cp bin/plugin-uploader .plugins/ 2>/dev/null || true
	chmod +x .plugins/plugin-*
	@echo ""
	@echo "Local installation complete!"
	@echo "Plugins installed to: ./.plugins/"
	@echo "CLI binary available at: ./bin/plugin-cli"
	@echo "Run with: ./bin/plugin-cli"

uninstall: ## Uninstall CLI from system directory
	@echo "Removing CLI from /usr/local/bin..."
	sudo rm -f /usr/local/bin/plugin-cli
	@echo "CLI uninstalled successfully"

clean: ## Remove build artifacts and generated files
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

clean-all: clean ## Remove everything including .plugins directory
	@echo "Removing .plugins directory..."
	rm -rf .plugins/

##@ Testing

test: ## Run all tests
	@echo "Running tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v ./...

test-coverage: ## Run tests with coverage analysis
	@echo "Running tests with coverage..."
	go test -cover ./...

test-coverage-report: ## Generate HTML coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-race: ## Run tests with race condition detector
	@echo "Running tests with race detector..."
	go test -race ./...

##@ Development

generate: ## Generate mocks and other code
	@echo "Generating mocks..."
	go generate ./...

deps: ## Download and install dependencies
	@echo "Installing dependencies..."
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

run-example: build ## Build and run example with dummy plugin
	@echo "Running example..."
	./bin/plugin-cli run -p dummy -m "Hello from Makefile!"

dev-setup: deps proto build ## Complete development environment setup
	@echo "Development environment ready!"
	@echo "Run 'make install' to install the CLI and plugins"
	@echo "Run 'make run-example' to test the dummy plugin"