#!/bin/bash
# Monitor and fix database permissions in real-time
# This script watches the database file and fixes permissions if needed

DB_FILE="/opt/gophish/gophish.db"
DB_DIR="/opt/gophish"

# Function to fix permissions
fix_permissions() {
    if [ -f "$DB_FILE" ]; then
        chown app:app "$DB_FILE" 2>/dev/null
        chmod 664 "$DB_FILE" 2>/dev/null
        echo "$(date): Fixed permissions on $DB_FILE" >> /var/log/gophish/db-fix.log
    fi
    chown -R app:app "$DB_DIR" 2>/dev/null
    chmod -R 775 "$DB_DIR" 2>/dev/null
}

# Fix permissions initially
fix_permissions

# Monitor and fix every 2 seconds
while true; do
    if [ -f "$DB_FILE" ]; then
        # Check if file is writable by app user
        if ! su -s /bin/bash - app -c "test -w $DB_FILE" 2>/dev/null; then
            fix_permissions
        fi
        # Check ownership
        OWNER=$(stat -c '%U' "$DB_FILE" 2>/dev/null || stat -f '%Su' "$DB_FILE" 2>/dev/null)
        if [ "$OWNER" != "app" ]; then
            fix_permissions
        fi
    fi
    sleep 2
done

