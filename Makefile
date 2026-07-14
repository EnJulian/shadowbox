BINARY := shadowbox
PKG := ./cmd/shadowbox
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/EnJulian/shadowbox/internal/cmd.version=$(VERSION) \
	-X github.com/EnJulian/shadowbox/internal/cmd.commit=$(COMMIT) \
	-X github.com/EnJulian/shadowbox/internal/cmd.date=$(DATE)

.PHONY: build install test lint vet tidy clean snapshot run

build: ## Build a static binary into ./shadowbox
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) $(PKG)

install: ## Install the binary into GOBIN/GOPATH
	CGO_ENABLED=0 go install -trimpath -ldflags="$(LDFLAGS)" $(PKG)

run: ## Run the interactive interface
	go run $(PKG)

test: ## Run all tests with the race detector
	go test -race ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy

snapshot: ## Build a local cross-platform snapshot via GoReleaser
	goreleaser build --snapshot --clean

clean: ## Remove build artifacts
	rm -rf $(BINARY) dist/

release: ## Cut a release (usage: make release BUMP=patch|minor|major|X.Y.Z)
	@if [ -z "$(BUMP)" ]; then echo "usage: make release BUMP=patch|minor|major|X.Y.Z"; exit 1; fi
	./scripts/release.sh $(BUMP)
