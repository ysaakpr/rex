#!/bin/bash
# Quick fix for production SSL certificate issue
# Run this on your production server to generate temporary certificates

set -e

echo "=========================================="
echo "Production SSL Quick Fix"
echo "=========================================="
echo ""
echo "This script will:"
echo "1. Generate temporary self-signed certificates"
echo "2. Restart nginx"
echo "3. Get your service running with HTTPS"
echo ""
echo "After this, you should run setup-ssl-letsencrypt.sh"
echo "to replace with proper Let's Encrypt certificates."
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
fi

echo ""
echo "Step 1: Creating ssl directory..."
mkdir -p ./ssl

echo "Step 2: Pulling latest docker-compose configuration..."
docker-compose pull nginx

echo "Step 3: Restarting nginx (will auto-generate certificates)..."
docker-compose up -d nginx

echo ""
echo "Waiting for nginx to start..."
sleep 3

# Check if nginx is running
if docker-compose ps nginx | grep -q "Up"; then
    echo ""
    echo "=========================================="
    echo "✅ SUCCESS! Nginx is now running"
    echo "=========================================="
    echo ""
    echo "Your service is accessible at:"
    echo "  - http://$(hostname -f 2>/dev/null || curl -s ifconfig.me)"
    echo "  - https://$(hostname -f 2>/dev/null || curl -s ifconfig.me)"
    echo ""
    echo "⚠️  IMPORTANT: You're using a temporary self-signed certificate."
    echo "   Browsers will show a security warning."
    echo ""
    echo "Next steps:"
    echo "  1. Verify service is working: docker-compose ps"
    echo "  2. Set up Let's Encrypt for trusted certificate:"
    echo "     ./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL"
    echo ""
else
    echo ""
    echo "❌ Nginx failed to start. Check logs:"
    echo "   docker-compose logs nginx"
    exit 1
fi

