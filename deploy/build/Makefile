# Makefile for building Docker images locally.
# This should be run from the deploy/build/ directory.

# Use the build-specific compose file and the .env file from the parent directory.
COMPOSE_CMD = docker-compose -f docker-compose.build.yml --env-file ../.env

.PHONY: all build app nginx help

.DEFAULT_GOAL := help

help: ## Показать эту справку
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[1;33m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: build ## Собрать все образы (действие по умолчанию)

build: ## Собрать все образы и сохранить их в .tar файлы
	@echo "🛠️  Building all images locally..."
	@$(COMPOSE_CMD) build --parallel
	@echo "📦  Saving images to .tar files..."
	@docker save goooo-app:latest -o goooo-app.tar
	@docker save goooo-nginx:latest -o goooo-nginx.tar
	@echo "✅  Images saved to goooo-app.tar and goooo-nginx.tar"

app: ## Собрать и сохранить образ приложения в .tar файл
	@echo "🛠️  Building app image locally..."
	@$(COMPOSE_CMD) build app
	@echo "📦  Saving app image to goooo-app.tar..."
	@docker save goooo-app:latest -o goooo-app.tar
	@echo "✅  Image saved to goooo-app.tar"

nginx: ## Собрать и сохранить образ nginx в .tar файл
	@echo "🛠️  Building nginx image locally..."
	@$(COMPOSE_CMD) build nginx
	@echo "📦  Saving nginx image to goooo-nginx.tar..."
	@docker save goooo-nginx:latest -o goooo-nginx.tar
	@echo "✅  Image saved to goooo-nginx.tar"

# ====================================================================================
# DEPLOYMENT
# ====================================================================================
REMOTE_HOST = vps
REMOTE_IMAGE_DIR = /etc/images

deploy: build ## Собрать, доставить и загрузить образы на удаленный сервер
	@echo "🚀 Deploying images to $(REMOTE_HOST)..."
	@echo "1. Creating remote directory $(REMOTE_IMAGE_DIR) on $(REMOTE_HOST)..."
	@ssh $(REMOTE_HOST) 'sudo mkdir -p $(REMOTE_IMAGE_DIR) && sudo chown $$(whoami) $(REMOTE_IMAGE_DIR)'
	@echo "2. Copying .tar files to $(REMOTE_HOST):$(REMOTE_IMAGE_DIR)..."
	@scp goooo-app.tar goooo-nginx.tar $(REMOTE_HOST):$(REMOTE_IMAGE_DIR)/
	@echo "3. Loading images into Docker on $(REMOTE_HOST)..."
	@ssh $(REMOTE_HOST) 'docker load -i $(REMOTE_IMAGE_DIR)/goooo-app.tar && docker load -i $(REMOTE_IMAGE_DIR)/goooo-nginx.tar'
	@echo "4. Cleaning up remote .tar files..."
	@ssh $(REMOTE_HOST) 'rm $(REMOTE_IMAGE_DIR)/goooo-app.tar $(REMOTE_IMAGE_DIR)/goooo-nginx.tar'
	@echo "✅  Deployment complete!"
