#!/bin/sh
set -e

echo "[DEBUG] Entrypoint script started."

echo "[DEBUG] Available environment variables:"
printenv
echo "------------------------------------"

TEMPLATE_PATH="/etc/nginx/templates/default.conf.template"
CONFIG_PATH="/etc/nginx/conf.d/default.conf"

echo "[DEBUG] Generating config from $TEMPLATE_PATH to $CONFIG_PATH..."

# Replace variable and create the final config file
cat "${TEMPLATE_PATH}" | sed "s/\${DOMAIN_NAME}/${DOMAIN_NAME}/g" > "${CONFIG_PATH}"

echo "[DEBUG] Config file generated. Content of $CONFIG_PATH:"
echo "------------------------------------"
cat "${CONFIG_PATH}"
echo "------------------------------------"

echo "[DEBUG] Starting Nginx..."
# Execute the command passed to this script (e.g., nginx -g 'daemon off;')
exec "$@"
