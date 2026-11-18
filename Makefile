# Common tasks. Run `make help` to list them.
BIN := bin/trady
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/sgnd/trady/internal/cli.Version=$(VERSION)
GOBUILD := CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"

.PHONY: help build install lint test security dist clean

PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

help: ## Show this help
	@grep -E '^[a-z0-9-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-15s %s\n", $$1, $$2}'

build: ## Build ./bin/trady
	$(GOBUILD) -o $(BIN) ./cmd/trady

install: ## Install trady into $GOPATH/bin
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/trady

lint: ## gofmt check + go vet
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)
	go vet ./...

test: ## Fast unit tests
	go test ./...

security: ## Known module vulnerabilities (govulncheck)
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

dist: ## Build release archives and checksums.txt into ./dist
	rm -rf dist && mkdir -p dist
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; name=trady_$${os}_$${arch}; \
	  exe=trady; [ $$os = windows ] && exe=trady.exe; \
	  echo "  $$name"; mkdir -p dist/$$name && \
	  GOOS=$$os GOARCH=$$arch $(GOBUILD) -o dist/$$name/$$exe ./cmd/trady && \
	  cp LICENSE README.md dist/$$name/ && \
	  if [ $$os = windows ]; then (cd dist/$$name && zip -qX ../$$name.zip *); \
	  else tar -czf dist/$$name.tar.gz -C dist/$$name .; fi && rm -rf dist/$$name || exit 1; \
	done
	cd dist && { sha256sum *.tar.gz *.zip 2>/dev/null || shasum -a 256 *.tar.gz *.zip; } > checksums.txt

clean: ## Remove build output
	rm -rf bin/ dist/
