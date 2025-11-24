#!/bin/bash
# SSL Certificate Management Helper Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

show_menu() {
    echo ""
    echo "=================================================="
    echo "  SSL Certificate Management"
    echo "=================================================="
    echo ""
    echo "1) Setup localhost (self-signed certificates)"
    echo "2) Setup Let's Encrypt (production domain)"
    echo "3) Check certificate status"
    echo "4) Renew Let's Encrypt certificates"
    echo "5) Test HTTPS connection"
    echo "6) View nginx SSL configuration"
    echo "0) Exit"
    echo ""
}

check_certificate_status() {
    echo ""
    echo "=================================================="
    echo "  Certificate Status"
    echo "=================================================="
    echo ""
    
    # Check self-signed certificates
    if [ -f "./ssl/cert.pem" ]; then
        echo -e "${GREEN}✓${NC} Self-signed certificate found (localhost)"
        echo ""
        openssl x509 -in ./ssl/cert.pem -noout -subject -issuer -dates
        echo ""
    else
        echo -e "${YELLOW}⚠${NC} No self-signed certificate found"
        echo ""
    fi
    
    # Check Let's Encrypt certificates
    echo "Let's Encrypt certificates:"
    echo ""
    docker-compose run --rm certbot certificates || echo "No Let's Encrypt certificates found"
}

renew_certificates() {
    echo ""
    echo "=================================================="
    echo "  Renewing Let's Encrypt Certificates"
    echo "=================================================="
    echo ""
    
    docker-compose run --rm certbot renew
    
    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✓${NC} Certificate renewal completed"
        echo "  Reloading nginx..."
        docker-compose exec nginx nginx -s reload
    else
        echo ""
        echo -e "${RED}✗${NC} Certificate renewal failed"
    fi
}

test_https() {
    echo ""
    echo "Enter domain to test (e.g., localhost, rex.stage.fauda.dream11.in):"
    read -r domain
    
    if [ -z "$domain" ]; then
        domain="localhost"
    fi
    
    echo ""
    echo "Testing HTTPS connection to: $domain"
    echo ""
    
    # Test with curl
    if [[ "$domain" == "localhost" ]]; then
        echo "Testing with self-signed certificate (using -k for insecure):"
        curl -k -I "https://$domain/health"
    else
        echo "Testing with Let's Encrypt certificate:"
        curl -I "https://$domain/health"
    fi
    
    echo ""
    echo "Testing SSL certificate:"
    echo ""
    
    if [[ "$domain" == "localhost" ]]; then
        openssl s_client -connect "$domain:443" -servername "$domain" < /dev/null 2>/dev/null | openssl x509 -noout -text | grep -A2 "Validity\|Subject:"
    else
        echo | openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -text | grep -A2 "Validity\|Subject:\|Issuer:"
    fi
}

view_nginx_config() {
    echo ""
    echo "=================================================="
    echo "  Nginx SSL Configuration"
    echo "=================================================="
    echo ""
    
    if [ -f "nginx.conf" ]; then
        echo "SSL certificate paths:"
        grep -A1 "ssl_certificate" nginx.conf | grep -v "^--$"
        echo ""
        echo "SSL protocols:"
        grep "ssl_protocols" nginx.conf
        echo ""
        echo "SSL ciphers:"
        grep "ssl_ciphers" nginx.conf
    else
        echo -e "${RED}✗${NC} nginx.conf not found"
    fi
}

# Main menu loop
while true; do
    show_menu
    read -p "Select an option [0-6]: " choice
    
    case $choice in
        1)
            ./scripts/setup-ssl-localhost.sh
            ;;
        2)
            echo ""
            read -p "Enter domain name: " domain
            read -p "Enter email address: " email
            ./scripts/setup-ssl-letsencrypt.sh "$domain" "$email"
            ;;
        3)
            check_certificate_status
            ;;
        4)
            renew_certificates
            ;;
        5)
            test_https
            ;;
        6)
            view_nginx_config
            ;;
        0)
            echo ""
            echo "Goodbye!"
            exit 0
            ;;
        *)
            echo ""
            echo -e "${RED}Invalid option. Please try again.${NC}"
            ;;
    esac
    
    echo ""
    read -p "Press Enter to continue..."
done

