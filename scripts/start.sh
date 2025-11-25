#!/bin/bash
set -e

echo "Starting Gophish unified container..."
echo "===================================="

# Generate SSL certificates if they don't exist
echo "Step 1: Generating SSL certificates..."
/usr/local/bin/generate-ssl.sh

# Change to Gophish directory
cd /opt/gophish
echo "Step 2: Configuring Gophish..."

# Fix permissions on Gophish directory and database
# This ensures the app user can write to the database
# Must be done as root before switching to app user
chown -R app:app /opt/gophish 2>/dev/null || true
chmod -R 775 /opt/gophish 2>/dev/null || true

# Ensure database directory is writable
mkdir -p /opt/gophish
chown -R app:app /opt/gophish
chmod 775 /opt/gophish

# Fix existing database if it exists
if [ -f /opt/gophish/gophish.db ]; then
    chown app:app /opt/gophish/gophish.db
    chmod 664 /opt/gophish/gophish.db
fi

# Create a wrapper to ensure database is created with correct permissions
# This will be called by supervisor

# Configure Gophish config.json from environment variables
if [ -n "${ADMIN_LISTEN_URL+set}" ]; then
    jq -r --arg ADMIN_LISTEN_URL "${ADMIN_LISTEN_URL}" \
        '.admin_server.listen_url = $ADMIN_LISTEN_URL' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

if [ -n "${ADMIN_USE_TLS+set}" ]; then
    jq -r --argjson ADMIN_USE_TLS "${ADMIN_USE_TLS}" \
        '.admin_server.use_tls = $ADMIN_USE_TLS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

if [ -n "${PHISH_LISTEN_URL+set}" ]; then
    jq -r --arg PHISH_LISTEN_URL "${PHISH_LISTEN_URL}" \
        '.phish_server.listen_url = $PHISH_LISTEN_URL' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

if [ -n "${PHISH_USE_TLS+set}" ]; then
    jq -r --argjson PHISH_USE_TLS "${PHISH_USE_TLS}" \
        '.phish_server.use_tls = $PHISH_USE_TLS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

if [ -n "${CONTACT_ADDRESS+set}" ]; then
    jq -r --arg CONTACT_ADDRESS "${CONTACT_ADDRESS}" \
        '.contact_address = $CONTACT_ADDRESS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

# Ensure Gophish listens on 0.0.0.0 for Docker (internal communication)
# Use port 8081 for phish server since Nginx uses port 80
jq -r '.admin_server.listen_url = "0.0.0.0:3333" | .admin_server.use_tls = false | .phish_server.listen_url = "0.0.0.0:8081" | .phish_server.use_tls = false' config.json > config.json.tmp && \
cat config.json.tmp > config.json

# Update Nginx config if IP whitelist is provided
if [ -n "${ALLOWED_IPS+set}" ]; then
    # Create allow rules for each IP
    IFS="," read -ra IPS <<< "${ALLOWED_IPS}"
    for ip in "${IPS[@]}"; do
        # Insert allow rule after the IP whitelist comment
        sed -i "/# IP whitelist enforcement - STRICT MODE/a\    allow ${ip};" /etc/nginx/conf.d/gophish.conf
    done
fi

echo "Step 3: Configuration complete!"
echo ""
echo "Starting services with supervisor..."
echo "Services will be available at:"
echo "  - Admin Dashboard: https://localhost:8443"
echo "  - Landing Pages: http://localhost:8080"
echo ""
echo "Waiting for services to start..."
sleep 2

# Start supervisor in foreground
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf -n

