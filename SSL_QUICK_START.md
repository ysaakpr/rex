# SSL/HTTPS Quick Start Guide

## 🔒 HTTPS is Now Enabled!

Your UTM Backend now supports HTTPS with automatic SSL certificate management.

## For Localhost Development (Default Setup)

✅ **Already configured!** Self-signed certificates have been generated.

### Access Your Application

- Frontend: **https://localhost** (not http://)
- API: **https://localhost/api**
- Health: **https://localhost/health**

### Browser Security Warning

You'll see a security warning because we're using self-signed certificates. This is **normal and expected** for local development.

**To proceed:**
- **Chrome/Edge**: Click "Advanced" → "Proceed to localhost (unsafe)"
- **Firefox**: Click "Advanced" → "Accept the Risk and Continue"
- **Safari**: Click "Show Details" → "visit this website"

### Trust Certificate (Optional)

To avoid the warning, add the certificate to your system's trust store:

**macOS:**
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./ssl/cert.pem
```

**Linux:**
```bash
sudo cp ./ssl/cert.pem /usr/local/share/ca-certificates/localhost.crt
sudo update-ca-certificates
```

## For Production (rex.stage.fauda.dream11.in)

### Prerequisites

1. Domain must point to your server's IP address
2. Port 80 and 443 must be accessible from the internet
3. Docker services must be running

### Get Let's Encrypt Certificate

Run the setup script:

```bash
./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com
```

The script will:
- ✅ Request certificate from Let's Encrypt
- ✅ Automatically update nginx configuration
- ✅ Reload nginx with new certificate
- ✅ Set up automatic renewal (every 12 hours)

### Access Your Application

- Frontend: **https://rex.stage.fauda.dream11.in**
- API: **https://rex.stage.fauda.dream11.in/api**
- Health: **https://rex.stage.fauda.dream11.in/health**

**No browser warnings!** Let's Encrypt certificates are trusted by all browsers.

## Management Tools

### Interactive SSL Manager

```bash
./scripts/manage-ssl.sh
```

**Menu options:**
1. Setup localhost (self-signed)
2. Setup Let's Encrypt (production)
3. Check certificate status
4. Renew certificates
5. Test HTTPS connection
6. View nginx SSL configuration

### Check Certificate Status

```bash
# Self-signed certificate
openssl x509 -in ./ssl/cert.pem -noout -dates

# Let's Encrypt certificates
docker-compose run --rm certbot certificates
```

### Manual Renewal (if needed)

```bash
docker-compose run --rm certbot renew
docker-compose exec nginx nginx -s reload
```

## Troubleshooting

### Issue: "Your connection is not private"

**For localhost:** This is expected with self-signed certificates. Click through the warning or trust the certificate (see above).

**For production:** Run the Let's Encrypt setup script.

### Issue: Let's Encrypt fails

Check:
1. Domain DNS: `dig +short rex.stage.fauda.dream11.in`
2. Port 80 access: `curl -I http://rex.stage.fauda.dream11.in`
3. Nginx logs: `docker-compose logs nginx`

### Issue: HTTP not redirecting to HTTPS

```bash
# Test redirect
curl -I http://localhost/
# Should return: 301 Moved Permanently

# Restart nginx
docker-compose restart nginx
```

## Frontend Updates

Update your frontend to use HTTPS URLs:

**For localhost:**
```javascript
const API_BASE_URL = 'https://localhost';
```

**For production:**
```javascript
const API_BASE_URL = 'https://rex.stage.fauda.dream11.in';
```

## What Changed?

1. ✅ Nginx now listens on ports 80 (HTTP) and 443 (HTTPS)
2. ✅ HTTP automatically redirects to HTTPS
3. ✅ Self-signed certificates generated for localhost
4. ✅ Certbot container added for Let's Encrypt
5. ✅ Automatic certificate renewal configured
6. ✅ Security headers added (HSTS, X-Frame-Options, etc.)
7. ✅ TLS 1.2 and 1.3 enabled with modern ciphers

## Documentation

For comprehensive documentation, see:

- **[20-LETSENCRYPT_SSL_SETUP.md](docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md)** - Complete SSL setup guide
- **[docs/INDEX.md](docs/INDEX.md)** - Full documentation index

## Quick Commands

```bash
# Start with HTTPS
docker-compose up -d

# Check nginx is running
docker-compose ps nginx

# View nginx logs
docker-compose logs nginx

# Test HTTPS (localhost)
curl -k https://localhost/health

# Test HTTPS (production)
curl https://rex.stage.fauda.dream11.in/health

# Regenerate localhost certificates
./scripts/setup-ssl-localhost.sh

# Setup production certificates
./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL
```

## Security Notes

- ✅ Private keys are NOT committed to git (in .gitignore)
- ✅ Self-signed certificates valid for 365 days
- ✅ Let's Encrypt certificates valid for 90 days (auto-renew at 30 days)
- ✅ HSTS enabled (forces HTTPS for 1 year after first visit)
- ✅ Modern TLS configuration (TLS 1.2/1.3 only)
- ✅ Strong cipher suites configured

---

**Questions?** Check the [full documentation](docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md) or run `./scripts/manage-ssl.sh` for interactive management.

