GO ?= go
BIN_DIR ?= bin
CMD := ./cmd/bbc
SOURCES := $(shell find cmd internal pkg -name '*.go')

VERSION ?= $(shell \
	if git describe --tags --exact-match >/dev/null 2>&1; then \
		git describe --tags --exact-match; \
	else \
		short=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
		if git diff-index --quiet HEAD 2>/dev/null; then \
			echo "dev-$$short"; \
		else \
			echo "dev-$$short-dirty"; \
		fi; \
	fi \
)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/ghoseb/bb/internal/build.versionFromLdflags=$(VERSION) \
	-X github.com/ghoseb/bb/internal/build.commitFromLdflags=$(COMMIT) \
	-X github.com/ghoseb/bb/internal/build.dateFromLdflags=$(BUILD_DATE)

.PHONY: build fmt lint test tidy release snapshot clean

build: $(BIN_DIR)/bbc

$(BIN_DIR)/bbc: $(SOURCES) go.mod go.sum
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/bbc $(CMD)

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

release:
	goreleaser release --clean

snapshot:
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed. Run: brew install goreleaser"; exit 1; }
	goreleaser release --snapshot --clean --skip=publish

clean:
	rm -rf $(BIN_DIR) dist/
