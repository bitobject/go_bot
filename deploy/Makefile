.PHONY: help up down logs build test clean

# Load environment variables from .env file and export them
include .env
export

# ===================================================================================
# Help
# ===================================================================================

help: ## Показывает эту справку
	@echo "Usage: make [target]"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[1;33m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ===================================================================================
# Docker-based Environment Commands
# ===================================================================================

COMPOSE_FILE = docker-compose.yml
COMPOSE_CMD = docker compose -f $(COMPOSE_FILE) --env-file .env

up: ## Запустить все сервисы в фоновом режиме
	@echo "🚀 Starting all services..."
	@$(COMPOSE_CMD) up -d

down: ## Остановить все сервисы
	@echo "🛑 Stopping all services..."
	@$(COMPOSE_CMD) down

rebuild-up: ## Пересобрать образы и запустить сервисы
	@echo "🏗️  Rebuilding and starting all services..."
	@$(COMPOSE_CMD) up -d --build

rebuild-nginx: ## Пересобрать и перезапустить сервис nginx
	@echo "🔄 Rebuilding and restarting nginx service..."
	@$(COMPOSE_CMD) up -d --build nginx

build: ## Собрать или пересобрать образы сервисов
	@echo "🛠️ Building images..."
	@$(COMPOSE_CMD) build

restart: ## Перезапустить все сервисы
	@echo "🔄 Restarting all services..."
	@$(COMPOSE_CMD) restart

logs: ## Показать логи всех сервисов
	@echo "📜 Tailing logs..."
	@$(COMPOSE_CMD) logs -f

logs-app: ## Показать логи только сервиса 'app'
	@echo "📜 Tailing logs for app..."
	@$(COMPOSE_CMD) logs -f app

ps: ## Показать статус контейнеров
	@echo "📊 Showing container status..."
	@$(COMPOSE_CMD) ps

clean: ## Остановить и удалить все контейнеры, сети и volumes
	@echo "🧹 Cleaning up the environment..."
	@$(COMPOSE_CMD) down -v --remove-orphans

# ===================================================================================
# Database Migration Commands
# ===================================================================================

# Construct the database URL from exported env vars
DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATE_CMD = $(COMPOSE_CMD) run --rm migrate -path /migrations -database "$(DATABASE_URL)"

migrate-create: ## Создать новый файл миграции (e.g., make migrate-create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	@echo "✍️ Creating migration file: $(NAME)..."
	docker run --rm -v $(shell pwd)/deploy/migrations:/migrations migrate/migrate:v4.17.1 create -ext sql -dir /migrations -seq $(NAME)

migrate-status: ## Показать статус миграций
	@echo "📊 Checking migration status..."
	@$(MIGRATE_CMD) version

migrate-up: ## Применить все доступные миграции
	@echo "⬆️ Applying all up migrations..."
	@$(MIGRATE_CMD) up

migrate-down: ## Откатить последнюю примененную миграцию
	@echo "⬇️ Reverting last migration..."
	@$(MIGRATE_CMD) down

# ===================================================================================
# Utility Commands
# ===================================================================================

db-shell: ## Подключиться к оболочке PostgreSQL внутри контейнера
	@echo "🗄️ Connecting to PostgreSQL shell..."
	@$(COMPOSE_CMD) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

app-shell: ## Подключиться к оболочке 'app' контейнера (не работает с 'scratch')
	@echo "🐚 Connecting to app shell (Note: will fail with 'scratch' image)..."
	@$(COMPOSE_CMD) exec app sh

nginx-reload: ## Перезагрузить конфигурацию Nginx
	@echo " reloading Nginx configuration..."
	@$(COMPOSE_CMD) exec nginx nginx -s reload