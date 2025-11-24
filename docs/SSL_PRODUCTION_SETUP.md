# SSL/HTTPS Production Setup Guide

## 🚨 Production Issue: Nginx Can't Start Without Certificates

If you're seeing this error in production:
```
nginx: [emerg] cannot load certificate "/etc/nginx/ssl/cert.pem": BIO_new_file() failed
```

**This is fixed!** The system now automatically generates temporary certificates on startup.

## 🔧 Quick Fix for Existing Production Deployments

### Step 1: Update Your Deployment

Pull the latest changes:
```bash
cd ~/rex  # or your deployment directory
git pull origin main
```

### Step 2: Restart Services

```bash
docker-compose down
docker-compose up -d
```

The nginx service will now:
1. ✅ Automatically generate temporary self-signed certificates if none exist
2. ✅ Start successfully
3. ✅ Serve HTTPS (with browser warning until Let's Encrypt is set up)

### Step 3: Setup Let's Encrypt (Replace Temporary Certificates)

Once the service is running, replace the temporary certificates with Let's Encrypt:

```bash
# Get your domain/hostname
DOMAIN="rex.stage.fauda.dream11.in"  # or your actual domain
EMAIL="your-email@example.com"

# Run the Let's Encrypt setup
./scripts/setup-ssl-letsencrypt.sh $DOMAIN $EMAIL
```

This will:
- Request a trusted certificate from Let's Encrypt
- Automatically update nginx to use it
- Enable automatic renewal
- Remove browser warnings

## 📋 Detailed Production Setup Steps

### Prerequisites

1. **Domain Setup**:
   - Domain must point to your server's IP
   - DNS A record: `rex.stage.fauda.dream11.in` → `YOUR_SERVER_IP`
   - Verify: `dig +short rex.stage.fauda.dream11.in`

2. **Firewall/Security Group**:
   - Port 80 (HTTP) must be open
   - Port 443 (HTTPS) must be open
   - Verify: `curl -I http://rex.stage.fauda.dream11.in`

3. **Services Running**:
   ```bash
   docker-compose ps
   # All services should be "Up"
   ```

### Method 1: Automated Setup (Recommended)

```bash
# One command to set up Let's Encrypt
./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com
```

**What it does:**
1. Validates domain is accessible
2. Requests certificate from Let's Encrypt
3. Updates nginx.conf automatically
4. Reloads nginx
5. Sets up auto-renewal

### Method 2: Manual Setup

If the automated script fails, you can set up manually:

#### 1. Stop nginx temporarily

```bash
docker-compose stop nginx
```

#### 2. Request certificate using standalone mode

```bash
docker-compose run --rm -p 80:80 certbot certonly --standalone \
  --preferred-challenges http \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d rex.stage.fauda.dream11.in
```

#### 3. Update nginx.conf

```bash
# Backup current config
cp nginx.conf nginx.conf.backup

# Update certificate paths
sed -i \
  -e 's|ssl_certificate /etc/nginx/ssl/cert.pem;|ssl_certificate /etc/letsencrypt/live/rex.stage.fauda.dream11.in/fullchain.pem;|g' \
  -e 's|ssl_certificate_key /etc/nginx/ssl/key.pem;|ssl_certificate_key /etc/letsencrypt/live/rex.stage.fauda.dream11.in/privkey.pem;|g' \
  nginx.conf
```

#### 4. Test and start nginx

```bash
# Test configuration
docker-compose run --rm nginx nginx -t

# Start nginx
docker-compose start nginx
```

## 🔍 Verification

### Check Certificate Status

```bash
# View installed certificates
docker-compose run --rm certbot certificates

# Expected output:
# Certificate Name: rex.stage.fauda.dream11.in
#   Domains: rex.stage.fauda.dream11.in
#   Expiry Date: 2025-XX-XX
#   Certificate Path: /etc/letsencrypt/live/rex.stage.fauda.dream11.in/fullchain.pem
```

### Test HTTPS Access

```bash
# Should work without -k flag (no self-signed warning)
curl https://rex.stage.fauda.dream11.in/health

# Expected: healthy
```

### Test Browser Access

Open in browser:
- https://rex.stage.fauda.dream11.in

**No security warnings!** Lock icon should be green/secure.

### Verify Auto-Renewal

```bash
# Check certbot is running
docker-compose ps certbot

# Check renewal logs
docker-compose logs certbot | grep -i renew

# Test renewal (dry run)
docker-compose run --rm certbot renew --dry-run
```

## 🐛 Troubleshooting

### Issue: Certificate Request Fails

**Error**: "Invalid response from http://..."

**Diagnosis:**
```bash
# 1. Check domain resolves
dig +short rex.stage.fauda.dream11.in
# Should return your server's IP

# 2. Check port 80 is accessible
curl -I http://rex.stage.fauda.dream11.in/.well-known/acme-challenge/test
# Should NOT get "connection refused"

# 3. Check nginx is serving ACME challenges
docker-compose logs nginx | grep acme-challenge
```

**Solutions:**
1. **DNS not set up**: Update DNS A record to point to server IP
2. **Firewall blocking**: Open port 80 in security group
3. **Nginx not running**: `docker-compose restart nginx`

### Issue: Nginx Won't Start

**Error**: "cannot load certificate"

**Solution:**
```bash
# The init script should handle this, but if it fails:

# Generate temporary certificate manually
docker-compose run --rm nginx sh -c "
  mkdir -p /etc/nginx/ssl
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/nginx/ssl/key.pem \
    -out /etc/nginx/ssl/cert.pem \
    -subj '/CN=localhost'
"

# Restart nginx
docker-compose restart nginx
```

### Issue: "Certbot: No renewals were attempted"

This is **normal** if:
1. Certificates are more than 30 days from expiry
2. No certificates have been issued yet

**Action needed:**
- If no certificates: Run initial setup with `setup-ssl-letsencrypt.sh`
- If certificates exist but no renewal: Wait until 30 days before expiry

### Issue: HTTP/2 Deprecation Warning

**Warning**: "the listen ... http2 directive is deprecated"

This has been **fixed** in the latest nginx.conf. Update:
```bash
git pull origin main
docker-compose restart nginx
```

### Issue: Certificate Expired

**Symptoms:**
- Browser shows "Certificate expired"
- HTTPS stops working

**Solution:**
```bash
# Force renewal
docker-compose run --rm certbot renew --force-renewal

# Reload nginx
docker-compose exec nginx nginx -s reload
```

## 🔄 Certificate Renewal

### Automatic Renewal

The certbot container automatically:
- Checks for renewal every 12 hours
- Renews certificates 30 days before expiry
- Nginx reloads every 6 hours to pick up new certificates

**No manual action required!**

### Manual Renewal (if needed)

```bash
# Renew all certificates
docker-compose run --rm certbot renew

# Reload nginx
docker-compose exec nginx nginx -s reload

# Verify new certificate
docker-compose run --rm certbot certificates
```

### Set Up Monitoring

Add a cron job to alert on expiry:

```bash
# Add to crontab
crontab -e

# Add this line (check daily, alert if < 10 days)
0 3 * * * docker-compose -f /path/to/docker-compose.yml run --rm certbot certificates | grep -q "VALID: [0-9] days" && echo "Certificate expiring soon!" | mail -s "SSL Alert" your-email@example.com
```

## 📊 Certificate Lifecycle

```
Day 0:   ✅ Certificate issued (90-day validity)
Day 60:  🔄 Certbot starts attempting renewal
Day 89:  ⚠️  Last day - manual renewal if auto-renewal failed
Day 90:  ❌ Certificate expires (HTTPS stops working)
```

**Best Practice**: Monitor certbot logs and set up alerts.

## 🔐 Security Checklist

Production SSL/HTTPS checklist:

- [x] Let's Encrypt certificate installed
- [x] HTTP automatically redirects to HTTPS
- [x] HSTS header enabled
- [x] Security headers configured
- [x] TLS 1.2/1.3 only (no SSL, no TLS 1.0/1.1)
- [x] Strong cipher suites
- [x] Certificate auto-renewal enabled
- [x] Monitoring/alerts set up (manual step)
- [ ] Consider rate limiting
- [ ] Consider WAF (Web Application Firewall)

## 📞 Quick Reference

```bash
# Check nginx status
docker-compose ps nginx
docker-compose logs nginx --tail=50

# Check certbot status
docker-compose ps certbot
docker-compose logs certbot --tail=50

# List certificates
docker-compose run --rm certbot certificates

# Test renewal (dry run)
docker-compose run --rm certbot renew --dry-run

# Force renewal
docker-compose run --rm certbot renew --force-renewal

# Reload nginx
docker-compose exec nginx nginx -s reload

# Restart nginx
docker-compose restart nginx

# View certificate details
echo | openssl s_client -connect rex.stage.fauda.dream11.in:443 2>/dev/null | openssl x509 -noout -text
```

## 🚀 Deployment Workflow

For new production deployments:

```bash
# 1. Deploy application
git clone <repo>
cd <repo>
docker-compose up -d

# 2. Verify services are up
docker-compose ps

# 3. Set up Let's Encrypt (once DNS is ready)
./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL

# 4. Verify HTTPS
curl https://YOUR_DOMAIN/health

# 5. Test in browser
open https://YOUR_DOMAIN
```

## 📚 Related Documentation

- [SSL Quick Start](SSL_QUICK_START.md) - Quick reference
- [Complete SSL Guide](changedoc/20-LETSENCRYPT_SSL_SETUP.md) - Comprehensive documentation
- [HTTPS Setup Summary](HTTPS_SETUP_SUMMARY.md) - What's configured

---

**Last Updated**: November 24, 2025  
**Tested On**: Ubuntu 22.04 LTS (AWS EC2)  
**Status**: ✅ Production Ready

