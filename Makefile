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

.PHONY: help all fmt check test mocks update upgrade

help: ## show available make targets
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
		awk -F':.*## ' '{ printf "  %-18s %s\n", $$1, $$2 }'

all: fmt test ## run fmt then test

fmt: ## add license headers, fix, format, and lint-fix the root library
	@echo "🧹 go mod tidy"
	@$(GO) mod tidy
	@echo "⚙️ go fix"
	@$(GO) fix ./...
	@echo "📜 addlicense"
	@$(ADDLICENSE) -f LICENCE $$(find . -name "*.go")
	@echo "🎨 golangci-lint fmt"
	@$(GOLANGCI) fmt ./...
	@echo "🔍 golangci-lint run --fix"
	@$(GOLANGCI) run --fix ./...

check: ## verify license headers and lint without making changes (CI-safe)
	@echo "📜 license check"
	@$(ADDLICENSE) -check -f LICENCE $$(find . -name "*.go" \
		-not -path "./tools/*" \
		-not -path "./_examples/*" \
		-not -path "./extras/*")
	@echo "🔍 golangci-lint run"
	@$(GOLANGCI) run ./...

test: ## run tests with race detector and coverage
	@echo "🧪 running tests"
	@$(GO) test -race -json -coverpkg ./... -parallel=4 ./... | $(TPARSE) --all

mocks: ## generate or refresh all mocks
	@echo "🤖 generating mocks"
	@$(MOCKGEN) -destination=./mocks/media/mock_media.go           -package=media         -source=./media/media.go
	@$(MOCKGEN) -destination=./mocks/user/mock_user.go             -package=user          -source=./user/user.go
	@$(MOCKGEN) -destination=./mocks/phonenumber/mock_phonenumber.go -package=phonenumber -source=./phonenumber/phonenumber.go
	@$(MOCKGEN) -destination=./mocks/qrcode/mock_qrcode.go         -package=qrcode        -source=./qrcode/qrcode.go
	@$(MOCKGEN) -destination=./mocks/webhooks/mock_webhooks_handlers.go -package=webhooks -source=./webhooks/handler.go
	@$(MOCKGEN) -destination=./mocks/auth/mock_auth.go             -package=auth          -source=./auth/auth.go
	@$(MOCKGEN) -destination=./mocks/conversation/automation/mock_automation.go -package=automation -source=./conversation/automation/automation.go
	@$(MOCKGEN) -destination=./mocks/message/mock_message.go       -package=message       -source=./message/message.go
	@$(MOCKGEN) -destination=./mocks/message/mock_status_update.go -package=message       -source=./message/status.go
	@$(MOCKGEN) -destination=./mocks/flow/mock_flow.go             -package=flow          -source=./flow/flow.go
	@$(MOCKGEN) -destination=./mocks/business/mock_business.go     -package=business      -source=./business/business.go
	@$(MOCKGEN) -destination=./mocks/business/analytics/mock_templates.go -package=analytics -source=./business/analytics/templates.go
	@$(MOCKGEN) -destination=./mocks/config/config_mock.go         -package=config        -source=./config/config.go
	@$(MOCKGEN) -destination=./mocks/http/mock_http.go             -package=http          -source=./pkg/http/http.go
	@$(MOCKGEN) -destination=./mocks/webhooks/mock_webhooks.go     -package=webhooks      -source=./webhooks/webhooks.go
	@$(MOCKGEN) -destination=./mocks/business/analytics/mock_analytics.go -package=analytics -source=./business/analytics/analytics.go

update: ## update dependencies and tidy go.mod
	@echo "⬆️  go get -u ./..."
	@$(GO) get -u ./...
	@echo "🧹  go mod tidy"
	@$(GO) mod tidy

upgrade: ## upgrade go.mod directives to match installed Go version, then update deps
	@echo "⬆️  Updating go.mod to Go $(GO_VERSION_FULL)"
	@$(GO) mod edit -go=$(GO_VERSION_MM)
	@$(GO) mod edit -toolchain=go$(GO_VERSION_FULL)
	@$(MAKE) update
