#!/bin/bash
# Setup Let's Encrypt SSL Certificates for Production Domain
# Use this for staging/production with a real domain name

set -e

echo "=================================================="
echo "Let's Encrypt SSL Certificate Setup"
echo "=================================================="

# Check if domain is provided
if [ -z "$1" ]; then
    echo ""
    echo "Usage: $0 <domain> [email]"
    echo ""
    echo "Examples:"
    echo "   $0 rex.stage.fauda.dream11.in admin@dream11.in"
    echo "   $0 api.example.com"
    echo ""
    echo "Note: The domain must be publicly accessible on port 80"
    echo "      and point to this server's IP address."
    echo ""
    exit 1
fi

DOMAIN=$1
EMAIL=${2:-""}

# Prompt for email if not provided
if [ -z "$EMAIL" ]; then
    read -p "Enter your email for Let's Encrypt notifications: " EMAIL
fi

# Validate email
if [[ ! "$EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    echo "❌ Invalid email address: $EMAIL"
    exit 1
fi

echo ""
echo "Domain: $DOMAIN"
echo "Email:  $EMAIL"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install it first."
    exit 1
fi

# Check if nginx container is running
if ! docker-compose ps nginx | grep -q "Up"; then
    echo "⚠️  Nginx container is not running."
    echo "   Starting services..."
    docker-compose up -d
    sleep 5
fi

echo "Testing domain accessibility..."
echo ""

# Test if port 80 is accessible
if ! curl -s -o /dev/null -w "%{http_code}" "http://$DOMAIN/.well-known/acme-challenge/test" | grep -q "200\|404\|301"; then
    echo "⚠️  WARNING: Could not reach http://$DOMAIN"
    echo "   Make sure:"
    echo "   1. Domain points to this server's IP"
    echo "   2. Port 80 is open in firewall"
    echo "   3. nginx is running: docker-compose ps nginx"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""
echo "Requesting certificate from Let's Encrypt..."
echo "This may take a minute..."
echo ""

# Request certificate using webroot method (no downtime)
docker-compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d "$DOMAIN"

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Certificate request failed!"
    echo ""
    echo "Common issues:"
    echo "   1. Domain doesn't point to this server"
    echo "   2. Port 80 is not accessible from internet"
    echo "   3. nginx is not serving /.well-known/acme-challenge/"
    echo ""
    echo "Debug steps:"
    echo "   1. Check domain DNS: dig +short $DOMAIN"
    echo "   2. Test HTTP access: curl -I http://$DOMAIN/.well-known/acme-challenge/test"
    echo "   3. Check nginx logs: docker-compose logs nginx"
    echo "   4. Check certbot logs: docker-compose logs certbot"
    echo ""
    exit 1
fi

echo ""
echo "✅ Certificate obtained successfully!"
echo ""

# Backup current nginx.conf
if [ -f "nginx.conf" ]; then
    cp nginx.conf "nginx.conf.backup.$(date +%Y%m%d_%H%M%S)"
    echo "📄 Backed up nginx.conf"
fi

# Update nginx.conf to use Let's Encrypt certificates
if grep -q "/etc/nginx/ssl/cert.pem" nginx.conf; then
    echo "🔧 Updating nginx.conf to use Let's Encrypt certificates..."
    
    sed -i.bak \
        -e "s|ssl_certificate /etc/nginx/ssl/cert.pem;|ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;|g" \
        -e "s|ssl_certificate_key /etc/nginx/ssl/key.pem;|ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;|g" \
        nginx.conf
    
    echo "✅ nginx.conf updated"
else
    echo "⚠️  Could not automatically update nginx.conf"
    echo "   Please manually update the SSL certificate paths to:"
    echo "   ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;"
    echo "   ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;"
fi

# Test nginx configuration
echo ""
echo "Testing nginx configuration..."
docker-compose exec nginx nginx -t

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Nginx configuration test failed!"
    echo "   Please check nginx.conf for errors."
    echo "   Backup available: nginx.conf.backup.*"
    exit 1
fi

# Reload nginx
echo ""
echo "Reloading nginx..."
docker-compose exec nginx nginx -s reload

echo ""
echo "=================================================="
echo "✅ HTTPS Setup Complete!"
echo "=================================================="
echo ""
echo "Your site is now secured with Let's Encrypt SSL:"
echo ""
echo "   🔒 HTTPS URL: https://$DOMAIN"
echo "   📧 Renewal notifications: $EMAIL"
echo "   📅 Certificate valid for: 90 days"
echo "   🔄 Auto-renewal: Enabled (checks every 12 hours)"
echo ""
echo "Next steps:"
echo "   1. Test HTTPS access: curl https://$DOMAIN/health"
echo "   2. Update your frontend to use HTTPS URLs"
echo "   3. Monitor certificate expiry: docker-compose run --rm certbot certificates"
echo ""
echo "Certificate auto-renewal is configured in docker-compose.yml."
echo "No manual action needed - certificates will renew automatically!"
echo ""

