# Makefile for building Docker images locally.
# This should be run from the deploy/build/ directory.

# Use the build-specific compose file and the .env file from the parent directory.
COMPOSE_CMD = docker-compose -f docker-compose.build.yml --env-file ../.env

.PHONY: all build deliver deploy app nginx help

.DEFAULT_GOAL := help

help: ## Показать эту справку
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[1;33m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: build ## Собрать все образы (действие по умолчанию)

build: ## Build and save all images to .tar files
	@echo "🛠️  Building all images locally..."
	@$(COMPOSE_CMD) build --parallel
	@echo "📦  Saving images to .tar files..."
	@docker save $(APP_IMAGE_NAME):latest -o $(APP_IMAGE_NAME).tar
	@docker save $(NGINX_IMAGE_NAME):latest -o $(NGINX_IMAGE_NAME).tar
	@echo "✅ Images saved to $(APP_IMAGE_NAME).tar and $(NGINX_IMAGE_NAME).tar"

app: ## Собрать и сохранить образ приложения в .tar файл
	@echo "🛠️  Building app image locally..."
	@$(COMPOSE_CMD) build app
	@echo "📦  Saving app image to $(APP_IMAGE_NAME).tar..."
	@docker save $(APP_IMAGE_NAME):latest -o $(APP_IMAGE_NAME).tar
	@echo "✅  Image saved to $(APP_IMAGE_NAME).tar"

nginx: ## Собрать и сохранить образ nginx в .tar файл
	@echo "🛠️  Building nginx image locally..."
	@$(COMPOSE_CMD) build nginx
	@echo "📦  Saving nginx image to $(NGINX_IMAGE_NAME).tar..."
	@docker save $(NGINX_IMAGE_NAME):latest -o $(NGINX_IMAGE_NAME).tar
	@echo "✅  Image saved to $(NGINX_IMAGE_NAME).tar"

# ====================================================================================
# DEPLOYMENT
# ====================================================================================

deliver: build ## Deliver built images to the remote server
	@echo "🚚 Delivering images to $(SSH_HOST)..."
	@scp $(APP_IMAGE_NAME).tar $(NGINX_IMAGE_NAME).tar $(SSH_USER)@$(SSH_HOST):$(DEPLOY_PATH)
	@echo "✅ Images delivered successfully to $(DEPLOY_PATH) on $(SSH_HOST)!"

deploy: ## Deploy the application on the remote server using delivered images
	@echo "🔄 Deploying and restarting services on $(SSH_HOST)..."
	@ssh $(SSH_USER)@$(SSH_HOST) '
		set -e; \
		cd $(DEPLOY_PATH); \
		echo "- Loading app image..."; \
		docker load < $(APP_IMAGE_NAME).tar; \
		echo "- Loading nginx image..."; \
		docker load < $(NGINX_IMAGE_NAME).tar; \
		echo "- Removing temporary .tar files..."; \
		rm $(APP_IMAGE_NAME).tar $(NGINX_IMAGE_NAME).tar; \
		echo "- Restarting services with docker-compose..."; \
		docker-compose --env-file .env up -d --force-recreate; \
	'
	@echo "✅ Deployment finished successfully!"
