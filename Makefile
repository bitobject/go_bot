.PHONY: help test test-unit test-integration test-benchmark test-coverage build run clean

# Переменные
BINARY_NAME=goooo-bot
MAIN_PATH=cmd/bot/main.go
TEST_PATH=./...

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

test: ## Запустить все тесты
	@echo "$(GREEN)Запуск всех тестов...$(NC)"
	go test -v $(TEST_PATH)

test-unit: ## Запустить только unit тесты
	@echo "$(GREEN)Запуск unit тестов...$(NC)"
	go test -v ./internal/api/handlers/ -run "^TestAdminHandler"

test-integration: ## Запустить только интеграционные тесты
	@echo "$(GREEN)Запуск интеграционных тестов...$(NC)"
	go test -v ./internal/api/ -run "^TestAPI"

test-benchmark: ## Запустить benchmark тесты
	@echo "$(GREEN)Запуск benchmark тестов...$(NC)"
	go test -bench=. -benchmem $(TEST_PATH)

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	go test -coverprofile=coverage.out $(TEST_PATH)
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

test-short: ## Запустить короткие тесты
	@echo "$(GREEN)Запуск коротких тестов...$(NC)"
	go test -short $(TEST_PATH)

test-race: ## Запустить тесты на race conditions
	@echo "$(GREEN)Запуск тестов на race conditions...$(NC)"
	go test -race $(TEST_PATH)

build: ## Собрать приложение
	@echo "$(GREEN)Сборка приложения...$(NC)"
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Запустить приложение
	@echo "$(GREEN)Запуск приложения...$(NC)"
	go run $(MAIN_PATH)

run-dev: ## Запустить в режиме разработки
	@echo "$(GREEN)Запуск в режиме разработки...$(NC)"
	LOG_LEVEL=debug go run $(MAIN_PATH)

deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	go mod tidy
	go mod download

deps-test: ## Установить тестовые зависимости
	@echo "$(GREEN)Установка тестовых зависимостей...$(NC)"
	go get github.com/stretchr/testify
	go get gorm.io/driver/sqlite

lint: ## Запустить линтер
	@echo "$(GREEN)Проверка кода линтером...$(NC)"
	golangci-lint run

format: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	go fmt $(TEST_PATH)
	go vet $(TEST_PATH)

clean: ## Очистить артефакты сборки
	@echo "$(GREEN)Очистка артефактов...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html

create-admin: ## Создать администратора (требует параметры)
	@echo "$(YELLOW)Использование: make create-admin LOGIN=admin PASSWORD=secure_password$(NC)"
	@if [ -z "$(LOGIN)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "$(RED)Ошибка: LOGIN и PASSWORD обязательны$(NC)"; \
		exit 1; \
	fi
	go run scripts/create_admin.go -login=$(LOGIN) -password=$(PASSWORD)

test-api: ## Тестировать API (требует запущенный сервер)
	@echo "$(GREEN)Тестирование API...$(NC)"
	@if [ ! -f "./test_admin_login.sh" ]; then \
		echo "$(RED)Ошибка: файл test_admin_login.sh не найден$(NC)"; \
		exit 1; \
	fi
	./test_admin_login.sh

docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t $(BINARY_NAME) .

docker-run: ## Запустить в Docker
	@echo "$(GREEN)Запуск в Docker...$(NC)"
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME)

# Команды для CI/CD
ci-test: deps test-unit test-integration test-race test-coverage ## Полный набор тестов для CI
	@echo "$(GREEN)CI тесты завершены успешно!$(NC)"

ci-build: deps build ## Сборка для CI
	@echo "$(GREEN)CI сборка завершена успешно!$(NC)"

# Команды для разработки
dev-setup: deps deps-test ## Настройка окружения разработки
	@echo "$(GREEN)Окружение разработки настроено!$(NC)"

dev-test: test-unit test-integration ## Быстрые тесты для разработки
	@echo "$(GREEN)Тесты разработки завершены!$(NC)"

install-migrate: ## Установить утилиту golang-migrate
	@echo "$(GREEN)Установка golang-migrate...$(NC)"
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest