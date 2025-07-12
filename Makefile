# ===================================================================================
# Local Development & Testing Commands
# ===================================================================================

.PHONY: test lint tidy

TEST_PATH=./...

test: ## Запустить все Go тесты
	@echo "🧪 Running all tests..."
	@go test -v -race -cover $(TEST_PATH)

lint: ## Запустить golangci-lint
	@echo "🔍 Linting code..."
	@golangci-lint run

tidy: ## Привести в порядок go.mod и go.sum
	@echo "🧹 Tidying go modules..."
	@go mod tidy

# ===================================================================================
# Include deployment commands
# ===================================================================================

# Include the Makefile from the deploy directory to make its targets available here.
# This allows running commands like 'make up' or 'make migrate-up' from the root.
include deploy/Makefile
