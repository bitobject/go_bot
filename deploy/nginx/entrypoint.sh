#!/bin/sh
set -e

# Use sed to replace the domain name in the template.
# This is more robust than envsubst as it doesn't require extra packages
# and won't accidentally replace other shell variables.
cat /etc/nginx/templates/default.conf.template | sed "s/\${DOMAIN_NAME}/${DOMAIN_NAME}/g" > /etc/nginx/conf.d/default.conf

# Execute the command passed to this script (e.g., nginx -g 'daemon off;')
exec "$@"
