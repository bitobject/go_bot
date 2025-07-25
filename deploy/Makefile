.PHONY: help up up_prev start down logs build test clean

# Docker image configuration
IMAGE_NAME = go-bot
CURRENT_TAG = current
PREVIOUS_TAG = previous

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

up: rebuild-app start ## Пересобрать образ 'app' и запустить все сервисы 

up_prev: ## Запустить сервисы с предыдущей версией образа для отката
	@echo "🚀 Starting services with PREVIOUS image version..."
	@if ! [ "$(shell docker images -q $(IMAGE_NAME):$(PREVIOUS_TAG))" ]; then \
		echo "Error: No previous image to roll back to."; \
		exit 1; \
	fi
	@IMAGE_TAG=$(PREVIOUS_TAG) $(COMPOSE_CMD) up -d

start: ## Запустить все сервисы в фоновом режиме (без пересборки)
	@echo "🚀 Starting all services..."
	@IMAGE_TAG=$(CURRENT_TAG) $(COMPOSE_CMD) up -d

down: ## Остановить все сервисы
	@echo "🛑 Stopping all services..."
	@$(COMPOSE_CMD) down

rebuild-up: ## Пересобрать образы и запустить сервисы
	@echo "🏗️  Rebuilding and starting all services..."
	@$(COMPOSE_CMD) up -d --build

rebuild: ## Пересобрать образы сервисов без их запуска
	@echo "🛠️ Rebuilding images (without starting)..."
	@$(COMPOSE_CMD) build

rebuild-app: ## Пересобрать образ 'app', сохранить предыдущую версию (без запуска)
	@echo "Rotating images: current -> previous..."
	@if [ "$(shell docker images -q $(IMAGE_NAME):$(CURRENT_TAG))" ]; then \
		docker rmi $(IMAGE_NAME):$(PREVIOUS_TAG) 2>/dev/null || true; \
		docker tag $(IMAGE_NAME):$(CURRENT_TAG) $(IMAGE_NAME):$(PREVIOUS_TAG); \
	fi
	@echo "Building new 'current' image for app service (without starting)..."
	@IMAGE_TAG=$(CURRENT_TAG) $(COMPOSE_CMD) build app

rebuild-nginx: ## Пересобрать образ только для сервиса 'nginx'
	@echo "🛠️  Rebuilding nginx service image..."
	@$(COMPOSE_CMD) build nginx

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