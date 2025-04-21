TASK_BIN := go tool -modfile=go.tool.mod task

.PHONY: mocks tools all task tools-init

mocks tools clean:
	@$(TASK_BIN) $@

task:
	@if [ -z "$(cmd)" ]; then \
		echo "No task specified. Use: make task cmd=your-task"; \
	else \
		$(TASK_BIN) $(cmd); \
	fi

tools-init:
	@if [ -f go.tool.mod ]; then \
		echo "⚠️  go.tool.mod already exists. Skipping init..."; \
	else \
		echo "🔧 initializing tools module..."; \
		go mod init -modfile=go.tool.mod github.com/piusalfred/whatsapp/tools; \
	fi
	@echo "📦 installing task tool..."
	@go get -tool -modfile=go.tool.mod github.com/go-task/task/v3/cmd/task@latest
	@echo "✅ task installed, running 'task tools'..."
	@$(TASK_BIN) tools