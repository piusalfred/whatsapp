SHELL := /bin/bash

GO        := go
GOTOOLS   := gotools exec
GOLANGCI  := $(GOTOOLS) golangci-lint
ADDLICENSE := $(GOTOOLS) addlicense
TPARSE    := $(GOTOOLS) tparse
MOCKGEN   := $(GOTOOLS) mockgen

GO_VERSION_FULL := $(shell go version | awk '{print $$3}' | sed 's/^go//')
GO_VERSION_MM   := $(shell echo $(GO_VERSION_FULL) | awk -F. '{print $$1 "." $$2}')

.DEFAULT_GOAL := help

.PHONY: help all \
        fmt mocks test clean build coverage bench \
        check-license check-fmt check-lint check-mod-tidy check-vuln check-build check-test \
        check check-all \
        update upgrade check-updates

# ── Help ──────────────────────────────────────────────────────────────────────

help: ## show available make targets
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
		awk -F':.*## ' '{ printf "  %-22s %s\n", $$1, $$2 }'

# ── Primary Entry Points ──────────────────────────────────────────────────────

all: fmt test ## run fmt then test

# ── Helpers (modify / generate / inform) ──────────────────────────────────────

fmt: ## format, fix, lint-fix, add license headers, regenerate mocks
	@echo "🧹 go mod tidy"
	@$(GO) mod tidy
	@echo "⚙️  go fix"
	@$(GO) fix ./...
	@echo "📜 addlicense"
	@$(ADDLICENSE) -f LICENCE $$(find . -name "*.go" -not -path "./tools/*" -not -path "./_examples/*" -not -path "./extras/*")
	@echo "🎨 golangci-lint fmt"
	@$(GOLANGCI) fmt ./...
	@echo "🔍 golangci-lint run --fix"
	@$(GOLANGCI) run --fix ./...
	@$(MAKE) mocks

mocks: ## regenerate all mocks
	@echo "🤖 generating mocks"
	@$(MOCKGEN) -destination=./mocks/phonenumber/mock_phonenumber.go -package=phonenumber -source=./phonenumber/phonenumber.go
	@$(MOCKGEN) -destination=./mocks/qrcode/mock_qrcode.go         -package=qrcode        -source=./qrcode/qrcode.go
	@$(MOCKGEN) -destination=./mocks/webhooks/mock_webhooks_handlers.go -package=webhooks -source=./webhooks/handler.go
	@$(MOCKGEN) -destination=./mocks/auth/mock_auth.go             -package=auth          -source=./auth/auth.go
	@$(MOCKGEN) -destination=./mocks/conversation/automation/mock_automation.go -package=automation -source=./conversation/automation/automation.go
	@$(MOCKGEN) -destination=./mocks/config/config_mock.go         -package=config        -source=./config/config.go
	@$(MOCKGEN) -destination=./mocks/http/mock_http.go             -package=http          -source=./pkg/http/http.go
	@$(MOCKGEN) -destination=./mocks/webhooks/mock_webhooks.go     -package=webhooks      -source=./webhooks/webhooks.go
	@$(MOCKGEN) -destination=./mocks/business/analytics/mock_analytics.go -package=analytics -source=./business/analytics/analytics.go

test: ## run tests with race detector
	@echo "🧪 running tests"
	@$(GO) test -race -json -coverpkg ./... -parallel=4 ./... | $(TPARSE) --all

webhook-test: ## run webhooks package tests only
	@echo "🧪 running webhooks tests"
	@$(GO) test -race -json -coverpkg ./... -parallel=4 ./webhooks/... | $(TPARSE) --all

clean: ## remove build artifacts, coverage files, and caches
	@echo "🧹 cleaning"
	@rm -f coverage.out coverage.html
	@$(GO) clean -cache -testcache

build: ## compile all packages (does not install)
	@echo "🏗️  building"
	@$(GO) build -v ./...

coverage: ## run tests and open HTML coverage report
	@echo "📊 generating coverage report"
	@$(GO) test -race -coverprofile=coverage.out -covermode=atomic -coverpkg ./... -parallel=4 ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
	@$(GO) tool cover -func=coverage.out | grep total

bench: ## run all benchmarks
	@echo "⏱️  benchmarking"
	@$(GO) test -bench=. -benchmem -benchtime=1s -count=3 ./...

check-updates: ## list available module updates
	@echo "🔎 checking for updates"
	@$(GO) list -u -m -f '{{if .Update}}{{.Path}}  {{.Version}} → {{.Update.Version}}{{end}}' all 2>/dev/null | column -t

update: ## update all dependencies and tidy go.mod
	@echo "⬆️  go get -u ./..."
	@$(GO) get -u ./...
	@echo "🧹 go mod tidy"
	@$(GO) mod tidy

upgrade: ## match go.mod directives to installed Go, then update deps
	@echo "⬆️  updating go.mod to Go $(GO_VERSION_FULL)"
	@$(GO) mod edit -go=$(GO_VERSION_MM)
	@$(GO) mod edit -toolchain=go$(GO_VERSION_FULL)
	@$(MAKE) update

# ── Sanity Checks (read-only; pre-commit + CI) ────────────────────────────────

check-license: ## verify all .go files have license headers
	@echo "📜 license check"
	@$(ADDLICENSE) -check -f LICENCE $$(find . -name "*.go" \
		-not -path "./tools/*" \
		-not -path "./_examples/*" \
		-not -path "./extras/*")

check-fmt: ## verify code is formatted (goimports, gofumpt, gci, golines)
	@echo "🎨 format check"
	@OUT=$$($(GOLANGCI) fmt --diff ./... 2>/dev/null); \
	if [ -n "$$OUT" ]; then \
		echo "$$OUT"; \
		echo "❌ code is not formatted. Run 'make fmt' locally."; \
		exit 1; \
	fi

check-lint: ## run linters (read-only, no auto-fix)
	@echo "🔍 lint check"
	@$(GOLANGCI) run ./...

check-mod-tidy: ## verify go.mod and go.sum are consistent with source
	@echo "🧹 mod-tidy check"
	@$(GO) mod tidy
	@if ! git diff --exit-code -- go.mod go.sum; then \
		echo "❌ go.mod or go.sum is not tidy. Run 'make fmt' locally."; \
		git checkout -- go.mod go.sum; \
		exit 1; \
	fi

check-vuln: ## run vulnerability scan
	@echo "🛡️  vulnerability check"
	@$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

check-build: ## verify all packages compile
	@echo "🏗️  build check"
	@$(GO) build ./...

check-test: ## run all tests (must pass)
	@echo "🧪 test check"
	@$(GO) test -race -parallel=4 ./...

# ── Meta-targets ──────────────────────────────────────────────────────────────

check: check-license check-fmt check-lint check-mod-tidy check-build ## run fast sanity checks (pre-commit)

check-all: check check-vuln check-test ## run all sanity checks (CI)
