#!/bin/bash
# Wrapper script to ensure database has correct permissions before starting Gophish

# Fix permissions before starting - CRITICAL: Must be done as root
chown -R app:app /opt/gophish 2>/dev/null || true
chmod -R 775 /opt/gophish 2>/dev/null || true

# Ensure directory is writable by app user
chmod 775 /opt/gophish

# If database exists, fix its permissions immediately
if [ -f /opt/gophish/gophish.db ]; then
    chown app:app /opt/gophish/gophish.db
    chmod 664 /opt/gophish/gophish.db
fi

# Change to Gophish directory
cd /opt/gophish

# Start a background process to monitor and fix database permissions
(
    while true; do
        if [ -f /opt/gophish/gophish.db ]; then
            chown app:app /opt/gophish/gophish.db 2>/dev/null
            chmod 664 /opt/gophish/gophish.db 2>/dev/null
        fi
        sleep 1
    done
) &
MONITOR_PID=$!

# Set umask to ensure new files are created with correct permissions
# 002 = group writable (files created as 664)
umask 002

# Run Gophish as app user with proper environment
# Use exec to replace this process
exec su -s /bin/bash - app -c "cd /opt/gophish && umask 002 && /opt/gophish/gophish"

# Cleanup (should never reach here due to exec, but just in case)
kill $MONITOR_PID 2>/dev/null

