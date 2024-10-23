APP					:= wwand
VERSION			?= $(shell git describe --tags --always --dirty)
LOG_FORMAT	?= console
GOFLAGS			?= -ldflags '-X main.version=$(VERSION)'
APN					?= "bredband.oister.dk"

# Canonicalized names for target architecture.
ARCH				?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
	override ARCH=amd64
else ifeq ($(ARCH),aarch64)
	override ARCH=arm64
endif

define HELP_HEADER
Usage:	make <target>

Targets:
endef

export HELP_HEADER

help: ## List all targets.
	@echo "$$HELP_HEADER"
	@grep -E '^[a-zA-Z0-9%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'


.PHONY: run
run: ## Run the application.
	LOG_FORMAT=$(LOG_FORMAT) go run $(GOFLAGS) cmd/$(APP)/main.go $(APN)

build: bin/$(APP)-linux-$(ARCH) ## Build the application binary.

bin/$(APP)-linux-$(ARCH): cmd/$(APP)/main.go $(shell find . -name '*.go')
	CGO_ENABLED=1 GOOS=linux GOARCH=$(ARCH) go build $(GOFLAGS) -o $@ $<

.PHONY: test
test: ## Run tests.
	go test -v ./...

LINTER := $(GOPATH)/bin/golangci-lint
$(LINTER):
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: $(LINTER) ## Run linter.
	$< run

FORMATTER := $(GOPATH)/bin/gofumpt
$(FORMATTER):
	go install mvdan.cc/gofumpt@latest

.PHONY: format
format: $(FORMATTER) ## Format the code.
	$< -l -w .
