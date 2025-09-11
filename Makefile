TASK_BIN := go tool -modfile=./tools/go.mod task

.PHONY: mocks tools all task tools-init

mocks tools clean fmt add-license test:
	@$(TASK_BIN) $@

task:
	@if [ -z "$(cmd)" ]; then \
		echo "No task specified. Use: make task cmd=your-task"; \
	else \
		$(TASK_BIN) $(cmd); \
	fi
