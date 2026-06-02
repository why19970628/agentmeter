GO ?= go
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod
VERSION ?= dev

.PHONY: check test build build-web run validate release-snapshot

check:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*' -not -path './webui/node_modules/*'))"
	sh -n scripts/install.sh
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) vet ./...

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

build:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -ldflags "-X main.version=$(VERSION)" -o bin/agentmeter .

build-web:
	cd webui && npm run build

run:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) run . $(ARGS)

validate: build-web check test build

release-snapshot:
	goreleaser release --snapshot --clean
