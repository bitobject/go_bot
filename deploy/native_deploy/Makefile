# Makefile for managing native Nginx deployment

# Load environment variables from ../.env
# This makes SSH_USER, SSH_HOST, DEPLOY_PATH, and DOMAIN_NAME available
ifneq (,$(wildcard ../.env))
    include ../.env
    export
endif

# Variables
NGINX_CONFIG_SRC = ../nginx/conf.d/default.conf.template
NGINX_CONFIG_DEST = /etc/nginx/sites-available/$(DOMAIN_NAME)

.PHONY: help deploy-nginx

.DEFAULT_GOAL := help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deploy-nginx   Prepares, delivers, and enables the Nginx config on the remote server."

deploy-nginx:
	@echo "🚀 Starting native Nginx deployment to $(SSH_HOST)..."
	# Step 1: Prepare config by substituting environment variables
	@echo "1. Preparing Nginx config for domain $(DOMAIN_NAME)..."
	@cat $(NGINX_CONFIG_SRC) | envsubst '\$${DOMAIN_NAME}' > ./$(DOMAIN_NAME).conf

	# Step 2: Copy the prepared config to the server
	@echo "2. Copying config to $(SSH_HOST)..."
	@scp ./$(DOMAIN_NAME).conf $(SSH_USER)@$(SSH_HOST):/tmp/

	# Step 3: Place config, enable site, and restart Nginx on the server
	@echo "3. Enabling site and restarting Nginx on server..."
	@ssh $(SSH_USER)@$(SSH_HOST) '
		set -e; \
		sudo mv /tmp/$(DOMAIN_NAME).conf $(NGINX_CONFIG_DEST); \
		sudo ln -sf $(NGINX_CONFIG_DEST) /etc/nginx/sites-enabled/$(DOMAIN_NAME); \
		sudo rm -f /etc/nginx/sites-enabled/default; \
		echo "- Checking Nginx config syntax..."; \
		sudo nginx -t; \
		echo "- Restarting Nginx service..."; \
		sudo systemctl restart nginx; \
		echo "- Checking Nginx status..."; \
		sudo systemctl status nginx --no-pager; \
	'

	# Step 4: Clean up local temporary file
	@echo "4. Cleaning up..."
	@rm ./$(DOMAIN_NAME).conf

	@echo "✅ Native Nginx deployment finished successfully!"
