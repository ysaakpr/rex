#!/bin/bash
# Initialize SSL certificates for production deployment
# This script creates temporary self-signed certificates if none exist
# allowing nginx to start successfully

set -e

echo "=========================================="
echo "SSL Certificate Initialization"
echo "=========================================="

SSL_DIR="/etc/nginx/ssl"
CERT_FILE="$SSL_DIR/cert.pem"
KEY_FILE="$SSL_DIR/key.pem"

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Check if certificates already exist
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo "✓ SSL certificates already exist"
    echo "  Cert: $CERT_FILE"
    echo "  Key:  $KEY_FILE"
    exit 0
fi

echo ""
echo "No SSL certificates found. Generating temporary self-signed certificate..."
echo ""

# Get hostname for certificate CN
HOSTNAME=$(hostname -f 2>/dev/null || echo "localhost")
echo "Hostname: $HOSTNAME"

# Generate temporary self-signed certificate
# This allows nginx to start while you set up Let's Encrypt
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=$HOSTNAME" \
    2>/dev/null

# Set proper permissions
chmod 644 "$CERT_FILE"
chmod 600 "$KEY_FILE"

echo ""
echo "✓ Temporary self-signed certificate created"
echo "  Cert: $CERT_FILE"
echo "  Key:  $KEY_FILE"
echo ""
echo "⚠️  IMPORTANT: This is a temporary self-signed certificate!"
echo "   Replace it with Let's Encrypt certificate:"
echo "   1. Run: ./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL"
echo "   2. Or manually request certificate using certbot"
echo ""

