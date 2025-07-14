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

echo "- Preparing main Nginx config...";
MAIN_NGINX_CONF=/etc/nginx/nginx.conf;
if [ -f "$$MAIN_NGINX_CONF" ]; then
    echo "  -> Backing up existing main Nginx config to $$MAIN_NGINX_CONF.bak";
    sudo cp "$$MAIN_NGINX_CONF" "$$MAIN_NGINX_CONF.bak";
fi;
sudo mv /tmp/nginx.conf "$$MAIN_NGINX_CONF";


echo "- Preparing upstream config...";
UPSTREAM_CONF=/etc/nginx/conf.d/go_app.conf;
if [ -f "$$UPSTREAM_CONF" ]; then
    echo "  -> Backing up existing upstream config to $$UPSTREAM_CONF.bak";
    sudo cp "$$UPSTREAM_CONF" "$$UPSTREAM_CONF.bak";
fi;
sudo mv /tmp/go_app_upstream.conf "$$UPSTREAM_CONF";

echo "- Preparing site config...";
SITE_CONF=/etc/nginx/sites-available/$(DOMAIN_NAME).conf;
if [ -f "$$SITE_CONF" ]; then
    echo "  -> Backing up existing site config to $$SITE_CONF.bak";
    sudo cp "$$SITE_CONF" "$$SITE_CONF.bak";
fi;
sudo mv "/tmp/$(DOMAIN_NAME).conf" "$$SITE_CONF";

echo "- Enabling site...";
sudo ln -sf "$$SITE_CONF" "/etc/nginx/sites-enabled/$(DOMAIN_NAME).conf";

echo "- Disabling default Nginx config (if it exists)...";
sudo rm -f /etc/nginx/sites-enabled/default;

echo "- Checking Nginx config syntax...";
sudo nginx -t;

echo "- Restarting Nginx service...";
sudo systemctl restart nginx;

echo "- Checking Nginx status...";
sudo systemctl status nginx --no-pager;
endef

.PHONY: help deploy-nginx clean update-nginx check-nginx-config check-ssl-certs check-nginx-status check-firewall view-nginx-config reset-firewall

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  deploy-nginx       Prepares, delivers, and enables the Nginx config on the remote server."
	@echo "  update-nginx       Updates Nginx on the remote server to the latest version."
	@echo "  clean              Remove temporary and leftover files."
	@echo "  check-nginx-config Check full Nginx configuration on the remote server."
	@echo "  check-ssl-certs    Check for SSL certs on the remote server."
	@echo "  check-nginx-status Check Nginx service status and listening ports on the remote server."
	@echo "  check-firewall     Check firewall status on the remote server."
	@echo "  view-nginx-config  View main and site Nginx configs on the remote server."
	@echo "  reset-firewall     Resets iptables and re-applies UFW rules on the remote server."

deploy-nginx: clean
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
	@echo "  -> Copying main nginx.conf..."
	@scp "../nginx/nginx.conf" "$(SSH_USER)@$(SSH_HOST):/tmp/nginx.conf"
	# Step 3: Execute remote commands
	@echo "3. Configuring Nginx on $(SSH_HOST)..."
	@echo "$$DEPLOY_SCRIPT" > deploy_script.sh
	@chmod +x deploy_script.sh
	@scp deploy_script.sh "$(SSH_USER)@$(SSH_HOST):/tmp/deploy_script.sh"
	@ssh "$(SSH_USER)@$(SSH_HOST)" "sudo /tmp/deploy_script.sh && rm /tmp/deploy_script.sh"
	@rm deploy_script.sh
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

check-firewall:
	@echo "🔥 Checking firewall status on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" 'sudo ufw status verbose'

view-nginx-config:
	@echo "📄 Viewing main Nginx config on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" 'sudo cat /etc/nginx/nginx.conf'
	@echo "\n📄 Viewing site config for $(DOMAIN_NAME) on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" 'sudo cat /etc/nginx/sites-available/$(DOMAIN_NAME)'

reset-firewall:
	@echo "🔥 Resetting firewall (iptables) and re-applying UFW rules on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" '		\
		sudo ufw disable && \
		sudo iptables -F && \
		sudo iptables -X && \
		sudo iptables -t nat -F && \
		sudo iptables -t nat -X && \
		sudo iptables -t mangle -F && \
		sudo iptables -t mangle -X && \
		sudo iptables -P INPUT ACCEPT && \
		sudo iptables -P FORWARD ACCEPT && \
		sudo iptables -P OUTPUT ACCEPT && \
		echo "y" | sudo ufw enable && \
		sudo ufw allow '\''$(SSH_PORT)'\''/tcp && \
		sudo ufw allow 80/tcp && \
		sudo ufw allow 443/tcp && \
		sudo systemctl restart nginx && \
		echo "\n✅ Nginx restarted. Final firewall status:" && \
		sudo ufw status verbose \
	'

check-nginx-status:
	@echo "🕵️ Checking Nginx service status and listening ports on $(SSH_HOST)..."
	@ssh "$(SSH_USER)@$(SSH_HOST)" ' \
		echo "--- Nginx Service Status ---"; \
		sudo systemctl status nginx --no-pager -l; \
		echo "\n--- Listening on Ports 80/443 ---"; \
		sudo ss -tlpn "(sport = :80 or sport = :443)" \
	'

