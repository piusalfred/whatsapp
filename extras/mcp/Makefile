SHELL := /bin/bash

GO        := go
GOTOOLS   := gotools exec
GOLANGCI  := $(GOTOOLS) golangci-lint
ADDLICENSE := $(GOTOOLS) addlicense
TPARSE    := $(GOTOOLS) tparse

GO_VERSION_FULL := $(shell go version | awk '{print $$3}' | sed 's/^go//')
GO_VERSION_MM   := $(shell echo $(GO_VERSION_FULL) | awk -F. '{print $$1 "." $$2}')

.DEFAULT_GOAL := help

.PHONY: help all fmt check build test update upgrade

help: ## show available make targets
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
		awk -F':.*## ' '{ printf "  %-18s %s\n", $$1, $$2 }'

all: fmt test ## run fmt then test

fmt: ## add license headers, fix, format, and lint-fix
	@echo "🔧 go mod tidy + build"
	@$(GO) mod tidy && $(GO) build ./...
	@echo "🔧 go fix"
	@$(GO) fix ./...
	@echo "📜 addlicense"
	@$(ADDLICENSE) -f ../../LICENCE $$(find . -name "*.go" -not -path "./tools/*")
	@echo "🎨 golangci-lint fmt"
	@$(GOLANGCI) fmt
	@echo "🔍 golangci-lint run --fix"
	@$(GOLANGCI) run --fix ./...

check: ## verify license headers, formatting, and lint without making changes
	@echo "📜 license check"
	@$(ADDLICENSE) -check -f ../../LICENCE $$(find . -name "*.go" -not -path "./tools/*")
	@echo "🔍 golangci-lint run"
	@$(GOLANGCI) run ./...

build: ## build the package
	@echo "🏗️  go build ./..."
	@$(GO) build ./...

test: ## run tests with race detector and coverage
	@echo "🧪 running tests"
	@$(GO) test -race -json -coverpkg ./... -parallel=4 ./... | $(TPARSE) --all

update: ## update dependencies and tidy go.mod
	@echo "⬆️  go get -u ./..."
	@$(GO) get -u ./...
	@echo "🧹 go mod tidy"
	@$(GO) mod tidy

upgrade: ## upgrade go.mod directives to match installed Go version, then update deps
	@echo "⬆️  Updating go.mod to Go $(GO_VERSION_FULL)"
	@$(GO) mod edit -go=$(GO_VERSION_MM)
	@$(GO) mod edit -toolchain=go$(GO_VERSION_FULL)
	@$(MAKE) update
