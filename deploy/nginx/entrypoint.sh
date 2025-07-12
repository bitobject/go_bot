#!/bin/sh
set -e

# Use envsubst to replace environment variables in the template file
# and output the result to the final Nginx configuration file.
# The list of variables to substitute is provided to prevent accidental substitution of other Nginx variables.
envsubst '${DOMAIN_NAME}' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf

# Execute the command passed to this script (e.g., nginx -g 'daemon off;')
exec "$@"
