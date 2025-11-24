# ✅ HTTPS/SSL Setup Complete!

## What's Been Configured

Your UTM Backend now has comprehensive HTTPS support:

### 🔐 For Localhost Development
- ✅ Self-signed SSL certificates generated
- ✅ Valid for 365 days
- ✅ Certificate: `ssl/cert.pem`
- ✅ Private Key: `ssl/key.pem`
- ✅ Access at: **https://localhost**

### 🌐 For Production (rex.stage.fauda.dream11.in)
- ✅ Let's Encrypt support ready
- ✅ Automatic certificate renewal configured
- ✅ Zero-downtime certificate issuance
- ✅ Run: `./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com`

## Changes Made

### 1. Docker Compose (`docker-compose.yml`)
- Added port 443 for HTTPS
- Added certbot service for Let's Encrypt
- Added volumes for certificates
- Configured nginx to reload every 6 hours

### 2. Nginx Configuration (`nginx.conf`)
- Complete rewrite to support HTTP/2 and HTTPS
- HTTP (port 80) redirects to HTTPS (port 443)
- Let's Encrypt ACME challenge support
- Modern TLS configuration (TLS 1.2/1.3)
- Security headers (HSTS, X-Frame-Options, etc.)

### 3. SSL Setup Scripts
- `scripts/setup-ssl-localhost.sh` - Generate self-signed certificates
- `scripts/setup-ssl-letsencrypt.sh` - Request Let's Encrypt certificates
- `scripts/manage-ssl.sh` - Interactive SSL management tool

### 4. Documentation
- `docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md` - Comprehensive guide
- `docs/SSL_QUICK_START.md` - Quick reference
- Updated `docs/INDEX.md` and `README.md`

### 5. Security
- Added `/ssl/*.pem` to `.gitignore`
- Private keys NOT committed to git

## How to Use

### Start Services

```bash
docker-compose up -d
```

### Access via HTTPS

**Localhost:**
```bash
# Frontend
open https://localhost

# API
curl -k https://localhost/api/v1/health

# Health check
curl -k https://localhost/health
```

**Production (after Let's Encrypt setup):**
```bash
# Frontend
open https://rex.stage.fauda.dream11.in

# API
curl https://rex.stage.fauda.dream11.in/api/v1/health
```

### Browser Warning

For localhost, you'll see a security warning. This is **normal** for self-signed certificates.

**Click through:**
- Chrome: "Advanced" → "Proceed to localhost (unsafe)"
- Firefox: "Advanced" → "Accept the Risk"
- Safari: "Show Details" → "visit this website"

### Setup Production Certificates

When ready for production:

```bash
./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com
```

This will:
1. Request certificate from Let's Encrypt
2. Update nginx.conf automatically
3. Reload nginx
4. Enable auto-renewal

## Management

### Interactive Tool

```bash
./scripts/manage-ssl.sh
```

### Check Certificate Status

```bash
# Localhost
openssl x509 -in ./ssl/cert.pem -noout -dates

# Let's Encrypt
docker-compose run --rm certbot certificates
```

### Renew Certificates

**Automatic:** Certbot checks every 12 hours and renews 30 days before expiry.

**Manual:**
```bash
docker-compose run --rm certbot renew
docker-compose exec nginx nginx -s reload
```

## Testing

### Test HTTP to HTTPS Redirect

```bash
curl -I http://localhost/
# Expected: 301 Moved Permanently
# Location: https://localhost/
```

### Test HTTPS Access

```bash
# Localhost (with -k for self-signed)
curl -k https://localhost/health

# Production (no -k needed)
curl https://rex.stage.fauda.dream11.in/health
```

### Test Security Headers

```bash
curl -k -I https://localhost/ | grep -E "(Strict-Transport|X-Frame|X-Content)"
```

Expected:
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`

## Frontend Updates

Update your frontend environment variables:

**Development (.env.development):**
```env
VITE_API_BASE_URL=https://localhost
```

**Production (.env.production):**
```env
VITE_API_BASE_URL=https://rex.stage.fauda.dream11.in
```

## Troubleshooting

### Nginx won't start

```bash
# Check configuration
docker-compose exec nginx nginx -t

# View logs
docker-compose logs nginx

# Restart
docker-compose restart nginx
```

### Certificate not found

```bash
# Check certificate exists
ls -la ssl/cert.pem ssl/key.pem

# Regenerate if needed
./scripts/setup-ssl-localhost.sh
```

### Let's Encrypt fails

```bash
# Check domain resolves
dig +short rex.stage.fauda.dream11.in

# Check port 80 accessible
curl -I http://rex.stage.fauda.dream11.in

# View certbot logs
docker-compose logs certbot
```

## Documentation

For complete information:

- **Quick Start**: [docs/SSL_QUICK_START.md](docs/SSL_QUICK_START.md)
- **Full Guide**: [docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md](docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md)
- **Documentation Index**: [docs/INDEX.md](docs/INDEX.md)

## Next Steps

1. ✅ HTTPS is ready for localhost - just start using `https://localhost`
2. ⏳ For production: Run Let's Encrypt setup when domain is ready
3. ⏳ Update frontend to use HTTPS URLs
4. ⏳ Test with real domain
5. ⏳ Set up monitoring for certificate expiry

## Security Notes

- ✅ TLS 1.2 and 1.3 enabled (no SSL, no TLS 1.0/1.1)
- ✅ Strong cipher suites configured
- ✅ HTTP Strict Transport Security (HSTS) enabled
- ✅ Security headers implemented
- ✅ Automatic certificate renewal
- ✅ Private keys excluded from git

## Cost

**Total Additional Cost: $0**

- Self-signed certificates: Free
- Let's Encrypt: Free
- Certbot: Free, open-source
- Auto-renewal: Free, automated

---

**Last Updated**: November 24, 2025  
**Status**: ✅ Ready to Use  
**TLS Version**: 1.2 / 1.3  
**Certificate Renewal**: Automatic

