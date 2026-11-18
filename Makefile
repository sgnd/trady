# Common tasks. Run `make help` to list them.
BIN := bin/trady
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/sgnd/trady/internal/cli.Version=$(VERSION)
GOBUILD := CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"

.PHONY: help build install test clean

help: ## Show this help
	@grep -E '^[a-z0-9-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-15s %s\n", $$1, $$2}'

build: ## Build ./bin/trady
	$(GOBUILD) -o $(BIN) ./cmd/trady

install: ## Install trady into $GOPATH/bin
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/trady

test: ## Fast unit tests
	go test ./...

clean: ## Remove build output
	rm -rf bin/
