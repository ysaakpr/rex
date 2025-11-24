# 🚨 Production SSL Fix - Immediate Action Required

## Problem

Your production nginx service is failing with:
```
nginx: [emerg] cannot load certificate "/etc/nginx/ssl/cert.pem": BIO_new_file() failed
```

**Root Cause**: SSL certificates don't exist in production yet.

## ✅ Quick Fix (3 Minutes)

Run these commands on your production server:

```bash
# 1. Navigate to your deployment directory
cd ~/rex  # or wherever you deployed

# 2. Pull the latest fixes
git pull origin main

# 3. Run the quick fix script
./scripts/fix-production-ssl.sh
```

**This will:**
- ✅ Generate temporary self-signed certificates
- ✅ Restart nginx
- ✅ Get your service running immediately

## 📋 Detailed Steps

### Step 1: SSH to Production Server

```bash
ssh ubuntu@10.20.146.127  # or your server IP
```

### Step 2: Navigate to Deployment Directory

```bash
cd ~/rex  # or your deployment path
```

### Step 3: Pull Latest Changes

```bash
git pull origin main
```

**What's new:**
- Auto-generates temporary certificates on startup
- Fixed http2 deprecation warning
- Added production fix scripts

### Step 4: Run Quick Fix

```bash
./scripts/fix-production-ssl.sh
```

**Output will look like:**
```
==========================================
Production SSL Quick Fix
==========================================

This script will:
1. Generate temporary self-signed certificates
2. Restart nginx
3. Get your service running with HTTPS

Continue? (y/N): y

Step 1: Creating ssl directory...
Step 2: Generating temporary self-signed certificate...
Step 3: Restarting nginx...

==========================================
✅ SUCCESS! Nginx is now running
==========================================
```

### Step 5: Verify Service is Running

```bash
# Check all services
docker-compose ps

# Should see nginx as "Up"

# Test HTTPS (will have browser warning with temp cert)
curl -k https://rex.stage.fauda.dream11.in/health
# Expected: healthy
```

## 🔐 Setup Let's Encrypt (Replace Temporary Certificates)

Once the service is running, set up proper Let's Encrypt certificates:

```bash
./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com
```

**Prerequisites:**
- Domain must point to server IP: `dig +short rex.stage.fauda.dream11.in`
- Port 80 must be accessible: `curl -I http://rex.stage.fauda.dream11.in`

**This will:**
- Request trusted certificate from Let's Encrypt (free)
- Update nginx configuration automatically
- Enable auto-renewal
- Remove browser security warnings

## 🔄 Alternative: Manual Fix

If the script doesn't work, manually generate certificates:

```bash
# 1. Create ssl directory
mkdir -p ./ssl

# 2. Generate certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ./ssl/key.pem \
  -out ./ssl/cert.pem \
  -subj "/CN=$(hostname)"

# 3. Set permissions
chmod 644 ./ssl/cert.pem
chmod 600 ./ssl/key.pem

# 4. Restart nginx
docker-compose restart nginx

# 5. Verify
docker-compose ps nginx
```

## 🐛 Troubleshooting

### Issue: Script fails with "command not found"

```bash
# Make script executable
chmod +x ./scripts/fix-production-ssl.sh

# Run again
./scripts/fix-production-ssl.sh
```

### Issue: Nginx still won't start

```bash
# Check nginx logs
docker-compose logs nginx --tail=50

# Common issues:
# 1. Port already in use
sudo netstat -tulpn | grep :443

# 2. Config syntax error
docker-compose run --rm nginx nginx -t

# 3. Permissions issue
ls -la ./ssl/
```

### Issue: Can't pull from git

```bash
# Check git status
git status

# If you have local changes, stash them
git stash

# Pull
git pull origin main

# Reapply changes if needed
git stash pop
```

## ✅ Verification Checklist

After running the fix:

- [ ] All services running: `docker-compose ps`
- [ ] Nginx is "Up"
- [ ] Health check works: `curl -k https://rex.stage.fauda.dream11.in/health`
- [ ] SSL certificates exist: `ls -la ./ssl/`
- [ ] No nginx errors: `docker-compose logs nginx --tail=20`

## 📞 Quick Commands

```bash
# Check all services
docker-compose ps

# View nginx logs
docker-compose logs nginx --tail=50

# Restart nginx
docker-compose restart nginx

# Check certificate files
ls -la ./ssl/

# Test HTTPS (with self-signed cert)
curl -k https://rex.stage.fauda.dream11.in/health

# View nginx config
docker-compose exec nginx cat /etc/nginx/nginx.conf | grep ssl
```

## 🚀 After Fix is Applied

1. ✅ Service should be accessible at `https://rex.stage.fauda.dream11.in`
2. ⚠️ Browser will show security warning (expected with temporary cert)
3. 🔐 Run Let's Encrypt setup to remove warning
4. 🔄 Auto-renewal will be enabled

## 📚 Documentation

- **Production Setup**: [docs/SSL_PRODUCTION_SETUP.md](docs/SSL_PRODUCTION_SETUP.md)
- **Complete Guide**: [docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md](docs/changedoc/20-LETSENCRYPT_SSL_SETUP.md)
- **Quick Start**: [docs/SSL_QUICK_START.md](docs/SSL_QUICK_START.md)

## 🆘 Still Having Issues?

### Check nginx error in detail:

```bash
docker-compose logs nginx | grep -i error
```

### Common errors and fixes:

| Error | Fix |
|-------|-----|
| "cannot load certificate" | Run `./scripts/fix-production-ssl.sh` |
| "port 443 already in use" | `sudo netstat -tulpn \| grep 443` and kill process |
| "permission denied" | `sudo chown -R ubuntu:ubuntu ./ssl` |
| "No such file or directory" | Ensure you're in the correct directory |

### Get help:

1. Check the logs: `docker-compose logs nginx --tail=100`
2. Check the setup: `docker-compose config`
3. Verify DNS: `dig +short rex.stage.fauda.dream11.in`
4. Check firewall: `curl -I http://rex.stage.fauda.dream11.in`

---

**Time to Fix**: 3-5 minutes  
**Impact**: Gets your service running immediately  
**Next Step**: Set up Let's Encrypt for trusted certificates  

**Need immediate help?** Share the output of:
```bash
docker-compose ps
docker-compose logs nginx --tail=50
ls -la ./ssl/
```

