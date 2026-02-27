VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod

BINARY_NAME = gdrv
BINARY_DIR = bin
COVERAGE_DIR = .artifacts/coverage

OFFICIAL_BUILD ?= false

BUILD_SOURCE_LDFLAGS = -X github.com/milcgroup/gdrv/internal/auth.BundledBuildSource=source
ifeq ($(OFFICIAL_BUILD),true)
	BUILD_SOURCE_LDFLAGS = -X github.com/milcgroup/gdrv/internal/auth.BundledBuildSource=official
endif

OAUTH_LDFLAGS =
ifdef GDRV_CLIENT_ID
	OAUTH_LDFLAGS += -X github.com/milcgroup/gdrv/internal/auth.BundledOAuthClientID=$(GDRV_CLIENT_ID)
endif
ifdef GDRV_CLIENT_SECRET
	OAUTH_LDFLAGS += -X github.com/milcgroup/gdrv/internal/auth.BundledOAuthClientSecret=$(GDRV_CLIENT_SECRET)
endif

LDFLAGS = -ldflags "-X github.com/milcgroup/gdrv/pkg/version.Version=$(VERSION) \
	-X github.com/milcgroup/gdrv/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/milcgroup/gdrv/pkg/version.BuildTime=$(BUILD_TIME) \
	$(BUILD_SOURCE_LDFLAGS) \
	$(OAUTH_LDFLAGS)"

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-official build-all clean test deps tidy lint security checksums version install help format install-hooks

all: deps build

build:
	@echo "Building $(BINARY_NAME) from source..."
	@echo "Note: Source builds require GDRV_CLIENT_ID and GDRV_CLIENT_SECRET env vars."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gdrv

build-official: export OFFICIAL_BUILD = true
build-official:
	@echo "Building official $(BINARY_NAME) with bundled OAuth credentials..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gdrv
	@echo "Built official $(BINARY_DIR)/$(BINARY_NAME) - OAuth credentials bundled"

build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,) ./cmd/gdrv; \
		echo "Built $(BINARY_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

deps:
	@echo "Installing dependencies..."
	$(GOMOD) download

tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy

test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -rf $(COVERAGE_DIR)

security:
	@echo "Running govulncheck..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

checksums:
	@echo "Generating checksums..."
	@cd $(BINARY_DIR) && rm -f checksums.txt && \
	for file in $(BINARY_NAME)*; do \
		if [ -f "$$file" ]; then \
			shasum -a 256 "$$file" >> checksums.txt; \
		fi \
	done
	@echo "Checksums written to $(BINARY_DIR)/checksums.txt"

version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Build Type: $(if $(OFFICIAL_BUILD),official,source)"
	@if [ "$(OFFICIAL_BUILD)" = "true" ]; then \
		echo "OAuth: Bundled company credentials"; \
	else \
		echo "OAuth: Requires GDRV_CLIENT_ID/GDRV_CLIENT_SECRET env vars"; \
	fi

run:
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/gdrv
	@./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

format:
	@echo "Formatting code..."
	@gofmt -w -s .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
	@echo "Hooks installed from .githooks/"

help:
	@echo "Company Google Drive CLI (gdrv) Makefile"
	@echo ""
	@echo "This is a company fork. Two build modes:"
	@echo "  - Source build: requires GDRV_CLIENT_ID/GDRV_CLIENT_SECRET env vars"
	@echo "  - Official build (make build-official): bundles OAuth credentials"
	@echo ""
	@echo "Targets:"
	@echo "  all           - Build the binary (default, requires env vars)"
	@echo "  build         - Build for current platform (requires env vars)"
	@echo "  build-official - Build with bundled OAuth credentials"
	@echo "  build-all     - Build for all platforms"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy go modules"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  checksums     - Generate SHA256 checksums"
	@echo "  version       - Show version info"
	@echo "  run           - Build and run (use ARGS=... for arguments)"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build (requires env vars)"
	@echo "  make build-official           # Build with bundled OAuth (CI use)"
	@echo "  make build-all                # Build for all platforms"
	@echo "  make test                     # Run tests"
	@echo "  make run ARGS='auth login'    # Build and run auth command"
	@echo ""
	@echo "Official builds are for CI/CD only - download from GitHub Releases."
