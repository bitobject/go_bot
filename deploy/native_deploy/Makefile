# Load environment variables from .env file located in the parent directory
ifneq (,$(wildcard ../.env))
	include ../.env
	export

# Strip whitespace from variables that might have it from the .env file
DOMAIN_NAME := $(strip $(DOMAIN_NAME))
endif

# Variables
UPSTREAM_CONFIG_SRC = ../nginx/conf.d/upstream.conf.template
SITE_CONFIG_SRC = ../nginx/conf.d/site.conf.template
TMP_UPSTREAM_CONFIG = ./upstream.conf.tmp
TMP_SITE_CONFIG = ./site.conf.tmp

# Define the remote script for deployment
# This script is designed to be idempotent and safe.
# It checks for directories, backs up existing configs, and then applies changes.
export DEPLOY_SCRIPT
define DEPLOY_SCRIPT
set -e;
echo "- Ensuring Nginx directories exist...";
sudo mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled /etc/nginx/conf.d;

echo "- Preparing upstream config...";
UPSTREAM_CONF=/etc/nginx/conf.d/go_app.conf;
if [ -f "$$UPSTREAM_CONF" ]; then
    echo "  -> Backing up existing upstream config to $$UPSTREAM_CONF.bak";
    sudo cp "$$UPSTREAM_CONF" "$$UPSTREAM_CONF.bak";
fi;
sudo mv /tmp/go_app_upstream.conf "$$UPSTREAM_CONF";

echo "- Preparing site config...";
SITE_CONF=/etc/nginx/sites-available/$(DOMAIN_NAME);
if [ -f "$$SITE_CONF" ]; then
    echo "  -> Backing up existing site config to $$SITE_CONF.bak";
    sudo cp "$$SITE_CONF" "$$SITE_CONF.bak";
fi;
sudo mv "/tmp/$(DOMAIN_NAME).conf" "$$SITE_CONF";

echo "- Enabling site...";
sudo ln -sf "$$SITE_CONF" "/etc/nginx/sites-enabled/$(DOMAIN_NAME)";

echo "- Disabling default Nginx config (if it exists)...";
sudo rm -f /etc/nginx/sites-enabled/default;

echo "- Checking Nginx config syntax...";
sudo nginx -t;

echo "- Restarting Nginx service...";
sudo systemctl restart nginx;

echo "- Checking Nginx status...";
sudo systemctl status nginx --no-pager;
endef

.PHONY: help deploy-nginx clean update-nginx check-nginx-config check-ssl-certs check-nginx-status patch-nginx-conf

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deploy-nginx       Prepares, delivers, and enables the Nginx config on the remote server."
	@echo "  update-nginx       Updates Nginx on the remote server to the latest version."
	@echo "  clean              Remove temporary and leftover files."
	@echo "  patch-nginx-conf   Ensures nginx.conf includes sites-enabled directory."
	@echo "  check-nginx-config Check full Nginx configuration on the remote server."
	@echo "  check-ssl-certs    Check for SSL certs on the remote server."
	@echo "  check-nginx-status Check Nginx service status on the remote server."

deploy-nginx: clean patch-nginx-conf
	@echo "🚀 Starting native Nginx deployment to $(SSH_HOST)..."
	@if [ -z "$(DOMAIN_NAME)" ]; then \
		echo "Error: DOMAIN_NAME is not set. Please set it in deploy/.env"; \
		exit 1; \
	fi
	# Step 1: Prepare local config files
	@echo "1. Preparing Nginx config files..."
	@cat "$(UPSTREAM_CONFIG_SRC)" > "$(TMP_UPSTREAM_CONFIG)"
	@cat "$(SITE_CONFIG_SRC)" | sed "s/{{DOMAIN_NAME}}/$(DOMAIN_NAME)/g" > "$(TMP_SITE_CONFIG)"
	# Step 2: Deliver config files to the server
	@echo "2. Delivering config files to $(SSH_HOST)..."
	@scp "$(TMP_UPSTREAM_CONFIG)" "$(SSH_USER)@$(SSH_HOST):/tmp/go_app_upstream.conf"
	@scp "$(TMP_SITE_CONFIG)" "$(SSH_USER)@$(SSH_HOST):/tmp/$(DOMAIN_NAME).conf"
	# Step 3: Execute remote commands
	@echo "3. Configuring Nginx on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" 'bash -c "$$DEPLOY_SCRIPT"'
	# Step 4: Clean up local temporary files
	@echo "4. Cleaning up local temporary files..."
	@rm -f "$(TMP_UPSTREAM_CONFIG)" "$(TMP_SITE_CONFIG)"
	@echo "✅ Native Nginx deployment finished successfully!"

clean:
	@echo "Cleaning up temporary and leftover files..."
	@rm -f *.tmp *.conf

clean-remote:
	@echo "🧹 Cleaning up legacy artifacts on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" 'sudo rm -f "/etc/nginx/sites-enabled/$(DOMAIN_NAME) "'

update-nginx:
	@echo "🚀 Preparing Nginx update script..."
	@rm -f ./update_nginx.sh
	@echo '#!/bin/bash' > ./update_nginx.sh
	@echo 'set -e' >> ./update_nginx.sh
	@echo 'export DEBIAN_FRONTEND=noninteractive' >> ./update_nginx.sh
	@echo 'echo "- Cleaning up old/broken config files..."' >> ./update_nginx.sh
	@echo 'sudo rm -f /etc/nginx/sites-enabled/$(DOMAIN_NAME) /etc/apt/sources.list.d/nginx.list /etc/apt/preferences.d/99nginx' >> ./update_nginx.sh
	@echo 'echo "- Step 1: Installing prerequisites..."' >> ./update_nginx.sh
	@echo 'sudo apt-get update' >> ./update_nginx.sh
	@echo 'sudo apt-get install -y curl gnupg2 ca-certificates lsb-release ubuntu-keyring' >> ./update_nginx.sh
	@echo 'echo "- Step 2: Importing Nginx signing key..."' >> ./update_nginx.sh
	@echo 'curl -fsSL https://nginx.org/keys/nginx_signing.key | sudo gpg --dearmor -o /usr/share/keyrings/nginx-archive-keyring.gpg' >> ./update_nginx.sh
	@echo 'echo "- Step 3: Adding official Nginx repository..."' >> ./update_nginx.sh
	@echo 'echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/ubuntu $$(lsb_release -cs) nginx" | sudo tee /etc/apt/sources.list.d/nginx.list > /dev/null' >> ./update_nginx.sh
	@echo 'echo "- Step 4: Setting up repository pinning..."' >> ./update_nginx.sh
	@echo 'echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900" | sudo tee /etc/apt/preferences.d/99nginx > /dev/null' >> ./update_nginx.sh
	@echo 'echo "- Step 5: Updating package lists and installing latest Nginx..."' >> ./update_nginx.sh
	@echo 'sudo apt-get update' >> ./update_nginx.sh
	@echo 'sudo apt-get install -y nginx' >> ./update_nginx.sh
	@echo 'echo "- Step 6: Reloading systemd and restarting Nginx..."' >> ./update_nginx.sh
	@echo 'sudo systemctl daemon-reload' >> ./update_nginx.sh
	@echo 'sudo systemctl restart nginx' >> ./update_nginx.sh
	@echo 'echo "- Step 7: Verifying Nginx version..."' >> ./update_nginx.sh
	@echo 'nginx -v' >> ./update_nginx.sh
	@echo 'echo "✅ Nginx has been successfully installed from the official repository!"' >> ./update_nginx.sh
	@chmod +x ./update_nginx.sh
	@echo "🚀 Delivering and executing Nginx update script on $(SSH_HOST)..."
	@scp ./update_nginx.sh "$(SSH_USER)@$(SSH_HOST):/tmp/update_nginx.sh"
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo /tmp/update_nginx.sh && rm /tmp/update_nginx.sh"
	@rm ./update_nginx.sh

check-nginx-config:
	@echo "🔍 Checking full Nginx configuration on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo nginx -T"

check-ssl-certs:
	@echo "🔍 Checking for SSL certs in /etc/letsencrypt/live/$(DOMAIN_NAME) on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo ls -l /etc/letsencrypt/live/$(DOMAIN_NAME)"

check-nginx-status:
	@echo "🔍 Checking Nginx service status on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo systemctl status nginx --no-pager"

patch-nginx-conf:
	@echo "🔧 Preparing Nginx patch script..."
	@rm -f ./patch_nginx.sh
	@echo '#!/bin/bash' > ./patch_nginx.sh
	@echo 'set -e' >> ./patch_nginx.sh
	@echo 'CONF_FILE=/etc/nginx/nginx.conf' >> ./patch_nginx.sh
	@echo 'INCLUDE_LINE="include /etc/nginx/sites-enabled/*;"' >> ./patch_nginx.sh
	@echo 'if ! sudo grep -qF "$$INCLUDE_LINE" "$$CONF_FILE"; then' >> ./patch_nginx.sh
	@echo '    echo "   -> Patching $$CONF_FILE..."' >> ./patch_nginx.sh
	@echo '    sudo sed -i "/include \/etc\/nginx\/conf.d\/\*\.conf;/a \    $$INCLUDE_LINE" "$$CONF_FILE"' >> ./patch_nginx.sh
	@echo '    echo "   -> Patched successfully."' >> ./patch_nginx.sh
	@echo 'else' >> ./patch_nginx.sh
	@echo '    echo "   -> Already patched. Skipping."' >> ./patch_nginx.sh
	@echo 'fi' >> ./patch_nginx.sh
	@chmod +x ./patch_nginx.sh
	@echo "🚀 Delivering and executing Nginx patch script on $(SSH_HOST)..."
	@scp ./patch_nginx.sh "$(SSH_USER)@$(SSH_HOST):/tmp/patch_nginx.sh"
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo /tmp/patch_nginx.sh && rm /tmp/patch_nginx.sh"
	@rm ./patch_nginx.sh
