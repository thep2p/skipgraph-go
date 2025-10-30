# Minimum Go version
GO_MIN_VERSION := 1.24.0
LINT_VERSION := v1.64.5
MOCKERY_VERSION := v2.43.0

# Dynamically detect OS (e.g., darwin, linux) and architecture (amd64, arm64)
GO_OS := $(shell uname -s | tr A-Z a-z)
GO_ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/arm64/arm64/')

# Download URL for the Go package
GO_DOWNLOAD_URL := https://golang.org/dl/go$(GO_MIN_VERSION).$(GO_OS)-$(GO_ARCH).tar.gz

# Go installation directory and binary path
GO_INSTALL_DIR := /usr/local/go
GO_BIN := $(shell which go)

# Hook to check Go version before running any target
.PHONY: check-go-version
check-go-version:
	@if [ -x "$(GO_BIN)" ]; then \
		CURRENT=$$($(GO_BIN) version | grep -o 'go[0-9]\+\(\.[0-9]\+\)*' | sed 's/go//'); \
		DESIRED="$(GO_MIN_VERSION)"; \
		if [ "$$(printf '%s\n' "$$DESIRED" "$$CURRENT" | sort -V | head -n1)" != "$$DESIRED" ]; then \
			echo "⚠️  Current Go version ($$CURRENT) does not meet the minimum required version ($$DESIRED)."; \
			exit 1; \
		else \
			echo "✅ Current Go version ($$CURRENT) meets or exceeds the required version ($$DESIRED)."; \
		fi; \
	else \
		echo "❌ Go is not installed."; \
		exit 1; \
	fi

# Install tools target with a dependency on Go version check
.PHONY: install-tools
install-tools: check-go-version
	@echo "Installing other tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "🔧 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINT_VERSION); \
	else \
		VERSION=$$(golangci-lint --version --format "{{.Version}}"); \
		if [[ "$${VERSION}" != "$(LINT_VERSION)" ]]; then \
			echo "🔄 Updating/Downgrading golangci-lint to $(LINT_VERSION)..."; \
			go clean -i github.com/golangci/golangci-lint/cmd/golangci-lint; \
			go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINT_VERSION); \
		else \
			echo "✅ golangci-lint $(LINT_VERSION) is already installed."; \
		fi; \
	fi
	@if ! command -v mockery >/dev/null 2>&1; then \
		echo "🔧 Installing mockery..."; \
		go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION); \
	else \
		VERSION=$$(mockery --version | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -n1); \
		if [[ "$${VERSION}" != "$(MOCKERY_VERSION)" ]]; then \
			echo "🔄 Updating/Downgrading mockery to $(MOCKERY_VERSION)..."; \
			go clean -i github.com/vektra/mockery/v2; \
			go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION); \
		else \
			echo "✅ mockery $(MOCKERY_VERSION) is already installed."; \
		fi; \
	fi
	@echo "✅ All tools installed successfully."



# Linting target that replicates CI behavior (reports issues without fixing)
.PHONY: lint
lint: check-go-version tidy
	@golangci-lint run --config ./integration/golangci-lint.yml ./...

# Linting target with auto-fix enabled
.PHONY: lint-fix
lint-fix: check-go-version tidy
	@golangci-lint run --fix --config ./integration/golangci-lint.yml ./...

.PHONY: tidy
tidy: check-go-version
	@go mod tidy

.PHONY: build
build: check-go-version tidy
	@go build ./...

# Generate mocks for all interfaces
# Note: Currently mocks are manually maintained in unittest/mock/
# Future enhancement: automate with mockery when package loading issues are resolved
.PHONY: generate-mocks
generate-mocks: check-go-version
	@echo "Checking for required mocks..."
	@if [ ! -f ./unittest/mock/immutable_lookup_table.go ]; then \
		echo "❌ Missing mock file: unittest/mock/immutable_lookup_table.go"; \
		echo "   Please regenerate mocks manually or check the codebase."; \
		exit 1; \
	fi
	@echo "✅ All mocks are present."

.PHONY: test
test: check-go-version tidy generate-mocks
	@go test -v ./...
