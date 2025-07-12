# Use the official Nginx image as a base
FROM nginx:alpine

# Install wget for healthchecks
# Install necessary packages
RUN apk add --no-cache wget gettext

# Copy the main Nginx configuration file. This is the critical fix.
COPY ./nginx.conf /etc/nginx/nginx.conf

# Copy the Nginx configuration template and the entrypoint script
COPY ./conf.d/default.conf.template /etc/nginx/templates/default.conf.template
COPY ./entrypoint.sh /docker-entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /docker-entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]

# The default command to run when the container starts
CMD ["nginx", "-g", "daemon off;"]
