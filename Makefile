SHELL := /bin/bash
TASK_BIN := go tool -modfile=./tools/go.mod task --silent


.DEFAULT_GOAL := all

# Allow: `make task cmd=<name>` or `make TASK=<name>`
TASK ?= $(cmd)
TASK ?= all

.PHONY: help list task run-example fmt test mocks add-license


help: ## show documented make targets and Taskfile tasks
	@echo ""
	@echo "Make targets:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | \
		awk -F':.*## ' '{ printf "  %-18s %s\n", $$1, $$2 }'
	@echo ""
	@echo "Tasks as fetched from Taskfile.yml):"
	@{ \
		$(TASK_BIN) --list-all 2>/dev/null || \
		$(TASK_BIN) --list 2>/dev/null     || \
		$(TASK_BIN) -l 2>/dev/null         || \
		echo "  (no task list available — is Taskfile.yml present?)"; \
	} | sed 's/^/  /'


list: help ## alias for help


all: ## run add-license -> mocks -> fmt -> test -> fmt-examples
	@$(TASK_BIN) all


fmt: ## format code and update dependencies
	@$(TASK_BIN) fmt

test: ## run tests
	@$(TASK_BIN) test


mocks: ## generate or refresh mocks
	@$(TASK_BIN) mocks


add-license: ## add or update license to all go files
	@$(TASK_BIN) add-license

task: ## run arbitrary Taskfile task (usage: make task cmd=<name> [program=...] [args="..."])
	@if [ -z "$(cmd)" ]; then \
		echo "Usage: make task cmd=<name> [program=...] [args=\"...\"]"; \
		exit 2; \
	fi
	@$(TASK_BIN) $(cmd) $(if $(program),program=$(program)) $(if $(args),args=$(args))

run-example: ## run example program (usage: make run-example program=block [args="..."])
	@$(TASK_BIN) run $(if $(program),program=$(program)) $(if $(args),args=$(args))

lint-check:
	@$(TASK_BIN) lint-check

lint-fix:
	@$(TASK_BIN) lint-fix

license-check:
	@$(TASK_BIN) license-check

.PHONY: update upgrade

# go version => "go version go1.25.5 darwin/arm64"
GO_VERSION_FULL := $(shell go version | awk '{print $$3}' | sed 's/^go//')          # 1.25.5
GO_VERSION_MM   := $(shell echo $(GO_VERSION_FULL) | awk -F. '{print $$1 "." $$2}') # 1.25

update-go:
	@echo "==> go get -u ./..."
	@go get -u ./...
	@echo "==> go mod tidy"
	@go mod tidy

upgrade-go:
	@echo "==> Updating go.mod directives to match installed Go $(GO_VERSION_FULL)"
	@echo "    - setting: go $(GO_VERSION_MM)"
	@go mod edit -go=$(GO_VERSION_MM)
	@echo "    - setting: toolchain go$(GO_VERSION_FULL)"
	@go mod edit -toolchain=go$(GO_VERSION_FULL)
	@$(MAKE) update
