# Nginx Reverse Proxy with HTTPS - All-in-One Mode

## Overview

The all-in-one deployment now includes nginx as a reverse proxy with HTTPS support, consolidating all services behind a single port (80/443) with path-based routing.

## Architecture

```
Internet (Users)
    ↓
HTTPS (Port 443) / HTTP (Port 80)
    ↓
Nginx Reverse Proxy
    ├─ /api/* → api:8080 (HTTP - internal)
    ├─ /auth/* → supertokens:3567 (HTTP - internal)
    └─ /health → api:8080/health
    ↓
Docker Network (Internal HTTP)
    ├─ API Container
    ├─ SuperTokens Container
    ├─ Worker Container
    ├─ PostgreSQL
    └─ Redis
```

## Key Benefits

✅ **Single Port**: Only port 80/443 exposed (simpler security group)
✅ **HTTPS Enabled**: End-to-end encryption for users
✅ **Path-Based Routing**: Clean URLs (`/api`, `/auth`)
✅ **Internal HTTP**: No SSL overhead within Docker network
✅ **Auto HTTPS Redirect**: HTTP automatically redirects to HTTPS
✅ **Let's Encrypt Ready**: Easy SSL certificate management
✅ **Security Headers**: HSTS, X-Frame-Options, etc.

## SSL Certificate Options

### Option 1: Self-Signed Certificate (Default)

**Pros:**
- Works immediately
- No configuration needed
- Free

**Cons:**
- Browser security warnings
- Not suitable for production
- Users must accept certificate

**Setup:** Automatic - already configured!

### Option 2: Let's Encrypt (Recommended for Production)

**Pros:**
- Trusted by all browsers
- Free
- Auto-renewal
- Professional

**Cons:**
- Requires public domain/hostname
- Requires port 80 accessible from internet
- Initial setup step required

**Setup Steps:**

```bash
# 1. SSH into instance
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID

# 2. Get your public hostname
cd /app
PUBLIC_HOSTNAME=$(curl -s http://169.254.169.254/latest/meta-data/public-hostname)
echo "Hostname: $PUBLIC_HOSTNAME"

# 3. Request certificate
docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d $PUBLIC_HOSTNAME

# 4. Update nginx.conf
# Edit /app/nginx.conf and update SSL certificate paths:
sed -i "s|/etc/nginx/ssl/cert.pem|/etc/letsencrypt/live/$PUBLIC_HOSTNAME/fullchain.pem|g" nginx.conf
sed -i "s|/etc/nginx/ssl/key.pem|/etc/letsencrypt/live/$PUBLIC_HOSTNAME/privkey.pem|g" nginx.conf

# 5. Test and reload nginx
docker-compose exec nginx nginx -t
docker-compose restart nginx

# 6. Test HTTPS
curl https://$PUBLIC_HOSTNAME/health
```

**Auto-Renewal:** Certbot container automatically renews certificates every 12 hours.

### Option 3: Custom Certificate

If you have your own certificate:

```bash
# 1. Copy certificate files to instance
scp fullchain.pem key.pem ec2-user@INSTANCE_IP:/app/ssl/

# 2. Update nginx.conf
vi /app/nginx.conf
# Update SSL paths to:
# ssl_certificate /etc/nginx/ssl/fullchain.pem;
# ssl_certificate_key /etc/nginx/ssl/key.pem;

# 3. Restart nginx
docker-compose restart nginx
```

## Nginx Configuration Details

### HTTP Server (Port 80)

```nginx
server {
    listen 80;
    
    # Let's Encrypt challenges (must be HTTP)
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    # Health check (allow HTTP for monitoring)
    location /health {
        proxy_pass http://api_backend/health;
    }
    
    # Redirect all other traffic to HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}
```

### HTTPS Server (Port 443)

```nginx
server {
    listen 443 ssl http2;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/DOMAIN/privkey.pem;
    
    # Modern SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000";
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    
    # Proxy to backend services (internal HTTP)
    location /api/ {
        proxy_pass http://api:8080/api/;
        proxy_set_header X-Forwarded-Proto https;
        # ... other headers
    }
    
    location /auth/ {
        proxy_pass http://supertokens:3567/;
        proxy_set_header X-Forwarded-Proto https;
        # ... other headers
    }
}
```

## Testing

### Test HTTP to HTTPS Redirect

```bash
PUBLIC_DNS=$(pulumi stack output allInOnePublicDns)

# This should redirect to HTTPS
curl -I http://$PUBLIC_DNS/api

# Expected: HTTP/1.1 301 Moved Permanently
# Location: https://...
```

### Test HTTPS Access

```bash
# With self-signed cert (use -k to skip verification)
curl -k https://$PUBLIC_DNS/health

# With Let's Encrypt cert (no -k needed)
curl https://$PUBLIC_DNS/health
```

### Test API Endpoints

```bash
# API
curl -k https://$PUBLIC_DNS/api/v1/health

# SuperTokens
curl -k https://$PUBLIC_DNS/auth/hello

# Admin login
curl -k -X POST https://$PUBLIC_DNS/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "admin@platform.local"},
      {"id": "password", "value": "admin"}
    ]
  }' \
  -c cookies.txt
```

### Test Security Headers

```bash
curl -k -I https://$PUBLIC_DNS/api | grep -E "(Strict-Transport|X-Frame|X-Content)"
```

Expected headers:
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`

## Frontend Configuration

### Environment Variables

Update your frontend to use the new unified base URL:

```javascript
// .env or environment variables
VITE_API_URL=https://YOUR-PUBLIC-DNS
VITE_SUPERTOKENS_URL=https://YOUR-PUBLIC-DNS

// In your code
const API_BASE = import.meta.env.VITE_API_URL; // https://example.com
const API_ENDPOINT = `${API_BASE}/api/v1/users`; // https://example.com/api/v1/users
const SUPERTOKENS_ENDPOINT = `${API_BASE}/auth`; // https://example.com/auth
```

### SuperTokens Configuration

```javascript
SuperTokens.init({
  appInfo: {
    appName: "Rex Backend",
    apiDomain: "https://YOUR-PUBLIC-DNS",  // Base URL
    websiteDomain: "https://YOUR-FRONTEND-DOMAIN",
    apiBasePath: "/auth",  // SuperTokens path
  },
  // ...
});
```

## Troubleshooting

### Issue: Browser Shows Security Warning

**Cause:** Self-signed certificate

**Solutions:**
1. Accept the warning (development only)
2. Install Let's Encrypt certificate (recommended)
3. Add custom trusted certificate

### Issue: Let's Encrypt Certificate Request Fails

**Check:**
```bash
# 1. Port 80 accessible from internet?
curl -I http://YOUR-PUBLIC-DNS/.well-known/acme-challenge/test

# 2. Domain resolves correctly?
dig YOUR-PUBLIC-DNS

# 3. Certbot logs
docker-compose logs certbot
```

**Common causes:**
- Security group doesn't allow port 80
- Domain doesn't resolve to instance IP
- Previous certificate exists

### Issue: HTTPS Redirect Not Working

**Check nginx config:**
```bash
docker-compose exec nginx nginx -t
docker-compose logs nginx
```

**Verify HTTP redirect:**
```bash
curl -I http://YOUR-PUBLIC-DNS/
# Should see: 301 Moved Permanently
```

### Issue: API Returns 502 Bad Gateway

**Check backend services:**
```bash
docker-compose ps
docker-compose logs api
docker-compose logs supertokens
```

**Test internal connectivity:**
```bash
docker-compose exec nginx wget -O- http://api:8080/health
docker-compose exec nginx wget -O- http://supertokens:3567/hello
```

### Issue: Certificate Expired

**Manual renewal:**
```bash
docker-compose run --rm certbot renew
docker-compose restart nginx
```

**Check auto-renewal:**
```bash
docker-compose logs certbot | grep -i renew
```

## Security Recommendations

### Production Checklist

- [ ] Enable Let's Encrypt certificate
- [ ] Restrict security group to specific IPs (if possible)
- [ ] Enable CloudWatch logging
- [ ] Set up monitoring/alerts
- [ ] Regular security updates (`apt update && apt upgrade`)
- [ ] Use strong admin password
- [ ] Enable fail2ban for brute force protection
- [ ] Consider adding rate limiting in nginx
- [ ] Set up regular backups
- [ ] Use AWS WAF if budget allows

### Enhanced SSL Configuration

For better security score, update nginx.conf:

```nginx
# Stronger SSL configuration
ssl_protocols TLSv1.3;
ssl_ciphers 'TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384';
ssl_prefer_server_ciphers off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
ssl_trusted_certificate /etc/letsencrypt/live/DOMAIN/chain.pem;

# Longer HSTS
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
```

## Monitoring

### Check SSL Certificate Expiry

```bash
# On instance
cd /app
docker-compose run --rm certbot certificates

# Or via openssl
echo | openssl s_client -servername YOUR-PUBLIC-DNS \
  -connect YOUR-PUBLIC-DNS:443 2>/dev/null | \
  openssl x509 -noout -dates
```

### Check Nginx Access Logs

```bash
docker-compose logs nginx | tail -100
```

### Monitor Certificate Renewal

```bash
# Check certbot logs
docker-compose logs certbot | grep -A 5 "Renewing"

# Check renewal timer
docker-compose logs certbot | tail -50
```

## Cost Impact

**Additional Costs:** $0

- Nginx: Included in Docker Compose (free)
- Let's Encrypt: Free
- Self-signed certificates: Free
- Certbot container: Minimal resource usage

**Total Monthly Cost:** Still ~$10 (no change)

## Migration from Direct Port Access

If you previously used direct port access (`:8080`, `:3567`):

**Frontend Updates:**
```diff
- const API_URL = 'http://example.com:8080/api';
+ const API_URL = 'https://example.com/api';

- const SUPERTOKENS_URL = 'http://example.com:3567';
+ const SUPERTOKENS_URL = 'https://example.com/auth';
```

**API Clients:**
```diff
- POST http://example.com:8080/api/v1/auth/signin
+ POST https://example.com/api/v1/auth/signin

- GET http://example.com:3567/hello
+ GET https://example.com/auth/hello
```

**Security Group:** Remove rules for ports 8080 and 3567 (now only need 80, 443, 22)

## Related Documentation

- [ALLINONE_QUICKSTART.md](./ALLINONE_QUICKSTART.md) - Quick start guide
- [LOWCOST_ALLINONE.md](./LOWCOST_ALLINONE.md) - Architecture overview
- [ADMIN_INITIALIZATION.md](./ADMIN_INITIALIZATION.md) - Admin setup
- [NETWORKING_ARCHITECTURE.md](./NETWORKING_ARCHITECTURE.md) - Network details

---

**Last Updated**: November 23, 2025  
**Nginx Version**: alpine (latest)  
**Certbot Version**: latest  
**SSL**: TLS 1.2/1.3

