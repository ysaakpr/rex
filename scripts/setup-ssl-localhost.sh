#!/bin/bash
# Setup Self-Signed SSL Certificates for Localhost Development
# Use this for local development with Docker Compose

set -e

echo "=================================================="
echo "Setting up Self-Signed SSL for Localhost"
echo "=================================================="

SSL_DIR="./ssl"
CERT_FILE="$SSL_DIR/cert.pem"
KEY_FILE="$SSL_DIR/key.pem"

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Check if certificates already exist
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo ""
    echo "⚠️  SSL certificates already exist!"
    echo "   Cert: $CERT_FILE"
    echo "   Key:  $KEY_FILE"
    echo ""
    read -p "Do you want to regenerate them? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing certificates."
        exit 0
    fi
fi

echo ""
echo "Generating self-signed certificate..."
echo ""

# Generate self-signed certificate valid for 365 days
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Development/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1"

# Set proper permissions
chmod 644 "$CERT_FILE"
chmod 600 "$KEY_FILE"

echo ""
echo "✅ Self-signed SSL certificates created successfully!"
echo ""
echo "   Certificate: $CERT_FILE"
echo "   Private Key: $KEY_FILE"
echo "   Valid for: 365 days"
echo ""
echo "⚠️  NOTE: These are self-signed certificates."
echo "   Your browser will show a security warning."
echo "   This is normal for local development."
echo ""
echo "To use these certificates:"
echo "   1. Make sure docker-compose.yml is configured (already done)"
echo "   2. Start services: docker-compose up -d"
echo "   3. Access via HTTPS: https://localhost"
echo ""
echo "To trust the certificate in your browser:"
echo "   - Chrome/Edge: Click 'Advanced' → 'Proceed to localhost'"
echo "   - Firefox: Click 'Advanced' → 'Accept the Risk'"
echo "   - Safari: Click 'Show Details' → 'visit this website'"
echo ""
echo "Or add the certificate to your system's trusted certificates:"
echo "   macOS: sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_FILE"
echo "   Linux: sudo cp $CERT_FILE /usr/local/share/ca-certificates/localhost.crt && sudo update-ca-certificates"
echo ""

