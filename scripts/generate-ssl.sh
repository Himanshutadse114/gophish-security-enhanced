#!/bin/bash

# Generate self-signed SSL certificate for Nginx if it doesn't exist
if [ ! -f /etc/nginx/ssl/gophish.crt ] || [ ! -f /etc/nginx/ssl/gophish.key ]; then
    echo "Generating self-signed SSL certificate for Nginx..."
    openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/gophish.key \
        -out /etc/nginx/ssl/gophish.crt \
        -subj "/C=US/ST=State/L=City/O=Gophish/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1,IP:::1" 2>/dev/null || \
    openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/gophish.key \
        -out /etc/nginx/ssl/gophish.crt \
        -subj "/C=US/ST=State/L=City/O=Gophish/CN=localhost"
    chmod 600 /etc/nginx/ssl/gophish.key
    chmod 644 /etc/nginx/ssl/gophish.crt
    echo "SSL certificate generated successfully"
fi

