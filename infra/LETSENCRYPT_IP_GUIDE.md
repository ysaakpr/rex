# Let's Encrypt with IP Address - Complete Guide

## Overview

As of 2024, Let's Encrypt supports certificates for IP addresses! This guide covers how to set up a trusted SSL certificate for your Elastic IP.

## Requirements

✅ **Public IP Address** (Elastic IP)
✅ **Port 80 accessible** from internet
✅ **Certbot** (included in docker-compose)

## Important Notes

### IP-Based Certificates

**Supported Methods:**
- ✅ **Standalone Mode**: Certbot runs its own web server
- ✅ **HTTP-01 Challenge**: Proves you control the IP
- ❌ **Webroot Mode**: Not supported for IPs (requires domain)

**Limitations:**
- Must stop nginx temporarily during certificate request
- Certificate valid for 90 days (auto-renewal supported)
- Some older browsers may not trust IP certificates

## Setup Methods

### Method 1: Let's Encrypt with IP (Standalone Mode)

This is the **simplest method** for IP-based certificates:

```bash
# 1. SSH to instance
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID

# 2. Get your Elastic IP
cd /app
ELASTIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
echo "Elastic IP: $ELASTIC_IP"

# 3. Stop nginx temporarily (certbot needs port 80)
docker-compose stop nginx

# 4. Request certificate for IP
docker-compose run --rm -p 80:80 certbot certonly --standalone \
  --preferred-challenges http \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d $ELASTIC_IP

# 5. Update nginx.conf to use Let's Encrypt certificate
sed -i "s|/etc/nginx/ssl/cert.pem|/etc/letsencrypt/live/$ELASTIC_IP/fullchain.pem|g" nginx.conf
sed -i "s|/etc/nginx/ssl/key.pem|/etc/letsencrypt/live/$ELASTIC_IP/privkey.pem|g" nginx.conf

# 6. Start nginx with new certificate
docker-compose start nginx

# 7. Verify HTTPS works
curl https://$ELASTIC_IP/health
```

**Expected Output:**
```
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Account registered.
Requesting a certificate for X.X.X.X

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/X.X.X.X/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/X.X.X.X/privkey.pem
This certificate expires on 2025-XX-XX.
```

### Method 2: Let's Encrypt with Domain (Recommended)

If you have a domain, this is **better for production**:

```bash
# 1. Add DNS A record
# Point api.yourdomain.com → YOUR_ELASTIC_IP

# 2. Wait for DNS propagation
dig +short api.yourdomain.com
# Should return: YOUR_ELASTIC_IP

# 3. Request certificate (no nginx downtime)
docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d api.yourdomain.com

# 4. Update nginx.conf server_name
sed -i 's/server_name _;/server_name api.yourdomain.com;/g' nginx.conf

# 5. Update certificate paths
sed -i "s|DOMAIN_PLACEHOLDER|api.yourdomain.com|g" nginx.conf

# 6. Reload nginx
docker-compose exec nginx nginx -t
docker-compose restart nginx
```

## Automatic Renewal

### For IP-Based Certificates

Automatic renewal with standalone mode requires special setup:

```bash
# Create renewal hook to stop/start nginx
cat > /app/certbot-renew-ip.sh <<'RENEW_EOF'
#!/bin/bash
docker-compose stop nginx
docker-compose run --rm -p 80:80 certbot renew
docker-compose start nginx
RENEW_EOF

chmod +x /app/certbot-renew-ip.sh

# Add to crontab
(crontab -l 2>/dev/null; echo "0 3 * * * /app/certbot-renew-ip.sh >> /var/log/certbot-renew.log 2>&1") | crontab -
```

### For Domain-Based Certificates

Renewal works automatically with the existing certbot container:

```yaml
# Already configured in docker-compose.yml
certbot:
  entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
```

No additional setup needed!

## Comparison

| Feature | IP Certificate | Domain Certificate |
|---------|---------------|-------------------|
| **Setup Time** | 5 minutes | 10 minutes (DNS wait) |
| **Downtime** | 1-2 minutes (during request) | None (webroot mode) |
| **Auto-Renewal** | Requires cron + downtime | Automatic, no downtime |
| **Browser Trust** | Modern browsers | All browsers |
| **Portability** | Tied to IP | Portable to any IP |
| **Best For** | Quick setup, dev/test | Production, multiple IPs |

## Troubleshooting

### Issue: "Invalid response from http://X.X.X.X/.well-known/acme-challenge/..."

**Cause:** Port 80 not accessible or nginx still running

**Solution:**
```bash
# Check port 80 is open
nc -zv YOUR_IP 80

# Ensure nginx is stopped
docker-compose stop nginx

# Check security group allows port 80
aws ec2 describe-security-groups --filters "Name=tag:Name,Values=*allinone-sg*"
```

### Issue: "Timeout during connect"

**Cause:** Security group doesn't allow inbound port 80

**Solution:**
```bash
# The security group already allows port 80 in the Pulumi code
# Verify with:
curl -I http://YOUR_ELASTIC_IP/
```

### Issue: "Certificate renewal failed"

**For IP certificates:**
```bash
# Manual renewal
cd /app
docker-compose stop nginx
docker-compose run --rm -p 80:80 certbot renew --standalone
docker-compose start nginx
```

**For domain certificates:**
```bash
# Check renewal logs
docker-compose logs certbot | grep -i renew

# Force renewal
docker-compose run --rm certbot renew --force-renewal
```

### Issue: "nginx: [emerg] cannot load certificate"

**Cause:** Certificate path incorrect in nginx.conf

**Solution:**
```bash
# List certificates
docker-compose run --rm certbot certificates

# Update nginx.conf with correct path
vi /app/nginx.conf

# Test config
docker-compose exec nginx nginx -t
```

## Testing

### Verify Certificate

```bash
# Check certificate details
echo | openssl s_client -servername YOUR_IP \
  -connect YOUR_IP:443 2>/dev/null | \
  openssl x509 -noout -text

# Check expiry
echo | openssl s_client -servername YOUR_IP \
  -connect YOUR_IP:443 2>/dev/null | \
  openssl x509 -noout -dates

# Verify it's from Let's Encrypt
echo | openssl s_client -servername YOUR_IP \
  -connect YOUR_IP:443 2>/dev/null | \
  openssl x509 -noout -issuer
# Should see: issuer=C = US, O = Let's Encrypt...
```

### Browser Test

```bash
# Open in browser (no -k flag needed!)
curl https://YOUR_ELASTIC_IP/health

# Should work without certificate warning
```

## Production Checklist

- [ ] Request Let's Encrypt certificate (IP or domain)
- [ ] Update nginx.conf with certificate paths
- [ ] Test HTTPS access (no browser warnings)
- [ ] Set up automatic renewal
- [ ] Test renewal process
- [ ] Monitor certificate expiry (CloudWatch alarm)
- [ ] Document renewal procedure
- [ ] Set up backup of /etc/letsencrypt

## Cost

**Let's Encrypt:** FREE
**Elastic IP:** $0 (when associated with running instance)
**Total:** $0 for trusted HTTPS!

## Script: One-Command Setup

Save this as `/app/setup-letsencrypt-ip.sh`:

```bash
#!/bin/bash
set -e

echo "=== Let's Encrypt IP Certificate Setup ==="

# Get Elastic IP
ELASTIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
echo "Elastic IP: $ELASTIC_IP"

# Prompt for email
read -p "Enter your email for Let's Encrypt: " EMAIL

# Stop nginx
echo "Stopping nginx..."
docker-compose stop nginx

# Request certificate
echo "Requesting certificate from Let's Encrypt..."
docker-compose run --rm -p 80:80 certbot certonly --standalone \
  --preferred-challenges http \
  --email "$EMAIL" \
  --agree-tos \
  --no-eff-email \
  --non-interactive \
  -d "$ELASTIC_IP"

if [ $? -eq 0 ]; then
    echo "✓ Certificate obtained successfully!"
    
    # Backup original nginx.conf
    cp nginx.conf nginx.conf.backup
    
    # Update nginx.conf
    sed -i "s|/etc/nginx/ssl/cert.pem|/etc/letsencrypt/live/$ELASTIC_IP/fullchain.pem|g" nginx.conf
    sed -i "s|/etc/nginx/ssl/key.pem|/etc/letsencrypt/live/$ELASTIC_IP/privkey.pem|g" nginx.conf
    sed -i "s/DOMAIN_PLACEHOLDER/$ELASTIC_IP/g" nginx.conf
    
    echo "✓ Nginx configuration updated"
    
    # Start nginx
    docker-compose start nginx
    
    echo "✓ Nginx started with Let's Encrypt certificate"
    echo ""
    echo "✓ HTTPS is now active with trusted certificate!"
    echo "  Access: https://$ELASTIC_IP"
    echo ""
    echo "Certificate expires in 90 days. Set up auto-renewal:"
    echo "  Run: ./setup-auto-renewal.sh"
else
    echo "✗ Certificate request failed"
    echo "  Check logs above for details"
    echo "  Starting nginx with self-signed certificate..."
    docker-compose start nginx
    exit 1
fi
```

## Related Documentation

- [ALLINONE_QUICKSTART.md](./ALLINONE_QUICKSTART.md) - Quick start guide
- [NGINX_HTTPS_GUIDE.md](./NGINX_HTTPS_GUIDE.md) - Nginx configuration
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)

---

**Last Updated**: November 23, 2025  
**Let's Encrypt**: Supports IP certificates since 2024  
**Certbot Version**: Latest

