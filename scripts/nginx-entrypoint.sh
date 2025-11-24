#!/bin/sh
# Nginx entrypoint with SSL certificate initialization
set -e

echo "=========================================="
echo "Nginx Startup - SSL Initialization"
echo "=========================================="

SSL_DIR="/etc/nginx/ssl"
CERT_FILE="$SSL_DIR/cert.pem"
KEY_FILE="$SSL_DIR/key.pem"

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Check if certificates exist
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo ""
    echo "⚠️  SSL certificates not found. Generating temporary self-signed certificate..."
    echo ""
    
    # Get hostname for certificate CN
    HOSTNAME=$(hostname -f 2>/dev/null || echo "localhost")
    echo "Hostname: $HOSTNAME"
    
    # Generate temporary self-signed certificate
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$KEY_FILE" \
        -out "$CERT_FILE" \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=$HOSTNAME" \
        2>/dev/null || {
        echo "❌ Failed to generate certificate"
        exit 1
    }
    
    # Set proper permissions
    chmod 644 "$CERT_FILE"
    chmod 600 "$KEY_FILE"
    
    echo ""
    echo "✅ Temporary self-signed certificate created"
    echo "   Cert: $CERT_FILE"
    echo "   Key:  $KEY_FILE"
    echo ""
    echo "⚠️  Replace with Let's Encrypt certificate for production:"
    echo "   ./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL"
    echo ""
else
    echo "✅ SSL certificates found"
    echo "   Cert: $CERT_FILE"
    echo "   Key:  $KEY_FILE"
    echo ""
fi

# Test nginx configuration
echo "Testing nginx configuration..."
nginx -t || {
    echo "❌ Nginx configuration test failed"
    exit 1
}

echo "✅ Nginx configuration is valid"
echo ""
echo "Starting nginx..."
echo "=========================================="
echo ""

# Start nginx with reload loop for certificate renewal
exec /bin/sh -c "while :; do sleep 6h & wait \$!; nginx -s reload; done & nginx -g 'daemon off;'"

