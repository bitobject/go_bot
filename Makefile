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

