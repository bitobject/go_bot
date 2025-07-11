.PHONY: help build up down logs clean restart status

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Production targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'If you need to prepare .env: make prepare-env'

build: ## Build Docker images
	docker compose -f docker-compose.yml build

up: ## Start all services
	docker compose -f docker-compose.yml up -d

down: ## Stop all services
	docker compose -f docker-compose.yml down

logs: ## Show logs from all services
	docker compose -f docker-compose.yml logs -f

logs-app: ## Show logs from app service
	docker compose -f docker-compose.yml logs -f app

logs-postgres: ## Show logs from postgres service
	docker compose -f docker-compose.yml logs -f postgres

logs-nginx: ## Show logs from nginx service
	docker compose -f docker-compose.yml logs -f nginx

restart: ## Restart all services
	docker compose -f docker-compose.yml restart

status: ## Show status of all services
	docker compose -f docker-compose.yml ps

clean: ## Remove all containers, networks, and volumes
	docker compose -f docker-compose.yml down -v --remove-orphans
	docker system prune -f

# Production environment preparation
prepare-env: ## Copy .env.example to .env if .env does not exist
	@if [ ! -f .env ]; then cp .env.example .env; fi

# Development commands
dev: ## Start services in development mode
	docker compose -f docker-compose.yml up -d

dev-build: ## Build and start services in development mode
	docker compose -f docker-compose.yml up -d --build

# Database commands

# Загружаем переменные из .env, чтобы они были доступны для migrate
include .env
export

MIGRATE_CMD = migrate -path ./migrations -database 'postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable'

migrate-create: ## Создать новый файл миграции (e.g., make migrate-create NAME=add_new_feature)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	$(MIGRATE_CMD) create -ext sql -dir ./migrations -seq $(NAME)

migrate-up: ## Применить все миграции
	$(MIGRATE_CMD) up

migrate-down: ## Откатить последнюю миграцию
	$(MIGRATE_CMD) down

db-shell: ## Connect to PostgreSQL shell
	docker compose -f docker-compose.yml exec postgres psql -U goooo_user -d goooo

db-backup: ## Create database backup
	docker compose -f docker-compose.yml exec postgres pg_dump -U goooo_user goooo > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Application commands
app-shell: ## Connect to app container shell
	docker compose -f docker-compose.yml exec app sh

# Nginx commands
nginx-reload: ## Reload nginx configuration
	docker compose -f docker-compose.yml exec nginx nginx -s reload

nginx-test: ## Test nginx configuration
	docker compose -f docker-compose.yml exec nginx nginx -t

# SSL commands
ssl-check: ## Check SSL certificate status
	@echo "Checking SSL certificate for body-architect.ru..."
	@docker compose -f docker-compose.yml exec nginx openssl x509 -in /etc/nginx/ssl/fullchain1.pem -text -noout | grep -E "(Subject:|Not Before:|Not After:)"

ssl-renew: ## Renew SSL certificate (if using certbot)
	@echo "Renewing SSL certificate..."
	@sudo certbot renew --cert-name body-architect.ru
	@make nginx-reload 