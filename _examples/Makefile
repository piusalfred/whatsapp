SHELL   := /bin/bash
GOTOOLS := gotools exec

.PHONY: fmt

fmt: ## add license headers, fix code, format, and lint-fix
	@echo "==> addlicense"
	@$(GOTOOLS) addlicense -f ../LICENCE $$(find . -name "*.go" -not -path "./tools/*")
	@echo "==> go fix"
	@go fix ./...
	@echo "==> golangci-lint fmt"
	@$(GOTOOLS) golangci-lint fmt
	@echo "==> golangci-lint run --fix"
	@$(GOTOOLS) golangci-lint run --fix ./...
