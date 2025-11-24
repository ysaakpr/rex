# 🚨 Production SSL Fix - Version 2 (Automatic)

## The Problem

Nginx fails to start with:
```
nginx: [emerg] cannot load certificate "/etc/nginx/ssl/cert.pem": BIO_new_file() failed
```

## ✅ The Solution (Automatic Certificate Generation)

The system now **automatically generates** temporary SSL certificates when nginx starts if they don't exist!

## 🚀 Quick Fix (2 Commands)

On your production server:

```bash
# 1. Pull the latest changes
cd ~/rex && git pull origin main

# 2. Restart the service
docker-compose up -d
```

**That's it!** The nginx entrypoint script will:
- ✅ Auto-generate certificates if missing
- ✅ Start nginx successfully
- ✅ Serve HTTPS immediately

## 📋 Detailed Steps

### Step 1: SSH to Production

```bash
ssh ubuntu@10.20.146.127
```

### Step 2: Navigate and Update

```bash
cd ~/rex
git pull origin main
```

### Step 3: Restart Services

```bash
docker-compose down
docker-compose up -d
```

### Step 4: Watch Nginx Start

```bash
docker-compose logs -f nginx
```

**You should see:**
```
==========================================
Nginx Startup - SSL Initialization
==========================================

⚠️  SSL certificates not found. Generating temporary self-signed certificate...

Hostname: ip-10-20-146-127
✅ Temporary self-signed certificate created
   Cert: /etc/nginx/ssl/cert.pem
   Key:  /etc/nginx/ssl/key.pem

Testing nginx configuration...
✅ Nginx configuration is valid

Starting nginx...
==========================================
```

### Step 5: Verify

```bash
# Check nginx is running
docker-compose ps nginx

# Test HTTPS (with temporary cert)
curl -k https://rex.stage.fauda.dream11.in/health
# Expected: healthy

# Check certificate was created
ls -la ./ssl/
# Should see cert.pem and key.pem
```

## 🔐 Setup Let's Encrypt (Remove Browser Warnings)

Once the service is running, replace temporary certificates:

```bash
./scripts/setup-ssl-letsencrypt.sh rex.stage.fauda.dream11.in your-email@example.com
```

**This will:**
- Request trusted certificate from Let's Encrypt
- Update nginx automatically
- Remove browser security warnings
- Enable auto-renewal

## 🔧 What Changed

### New Entrypoint Script

Created `scripts/nginx-entrypoint.sh` that runs **before** nginx starts:

1. **Checks** if certificates exist
2. **Generates** temporary self-signed certificates if missing
3. **Tests** nginx configuration
4. **Starts** nginx

### Updated docker-compose.yml

Changed from:
```yaml
command: "/bin/sh -c '...'"
```

To:
```yaml
entrypoint: ["/bin/sh", "/docker-entrypoint.sh"]
```

This ensures certificate generation happens **before** nginx tries to load them.

## 🐛 Troubleshooting

### Issue: Still getting certificate error

**Check entrypoint is being used:**
```bash
docker-compose config | grep -A 3 nginx

# Should show:
#   entrypoint:
#   - /bin/sh
#   - /docker-entrypoint.sh
```

**Force recreate container:**
```bash
docker-compose up -d --force-recreate nginx
```

### Issue: Permission denied creating certificates

**Fix permissions:**
```bash
# On host
mkdir -p ./ssl
chmod 755 ./ssl

# Restart
docker-compose restart nginx
```

### Issue: Nginx configuration invalid

**Test config:**
```bash
docker-compose run --rm nginx nginx -t
```

**View logs:**
```bash
docker-compose logs nginx --tail=50
```

## ✅ Verification Checklist

After the fix:

- [ ] `docker-compose ps` shows nginx as "Up"
- [ ] `ls -la ./ssl/` shows cert.pem and key.pem
- [ ] `curl -k https://rex.stage.fauda.dream11.in/health` returns "healthy"
- [ ] `docker-compose logs nginx` shows successful startup
- [ ] No certificate errors in logs

## 📞 Quick Commands

```bash
# Pull latest changes
git pull origin main

# Restart everything
docker-compose down && docker-compose up -d

# Check nginx startup
docker-compose logs nginx | grep -A 20 "SSL Initialization"

# Verify certificates exist
ls -la ./ssl/

# Test HTTPS
curl -k https://rex.stage.fauda.dream11.in/health

# Watch logs live
docker-compose logs -f nginx
```

## 🔄 Migration Path

### From Old Setup (Manual Certificates)

If you previously generated certificates manually:

**Option 1: Keep existing certificates**
- Do nothing! The script detects existing certificates and uses them

**Option 2: Regenerate with new setup**
```bash
# Remove old certificates
rm -rf ./ssl/*.pem

# Restart (will auto-generate)
docker-compose restart nginx
```

### From Let's Encrypt Setup

If you already have Let's Encrypt certificates:

**No action needed!** The nginx.conf will use Let's Encrypt certificates if the paths are updated. The temporary certificate generation only runs if `/etc/nginx/ssl/cert.pem` doesn't exist.

## 🚀 Production Deployment Workflow

For new deployments:

```bash
# 1. Deploy application
git clone <repo>
cd <repo>

# 2. Configure environment
cp .env.example .env
vim .env  # Update with production values

# 3. Start services
docker-compose up -d

# 4. Verify (nginx will auto-generate certificates)
docker-compose ps
curl -k https://YOUR_DOMAIN/health

# 5. Setup Let's Encrypt (once DNS is ready)
./scripts/setup-ssl-letsencrypt.sh YOUR_DOMAIN YOUR_EMAIL

# 6. Done!
curl https://YOUR_DOMAIN/health  # No -k needed!
```

## 📊 What Happens on Startup

```
┌─────────────────────────────────────┐
│ docker-compose up -d nginx          │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ nginx-entrypoint.sh starts          │
└────────────┬────────────────────────┘
             │
             ▼
┌─────────────────────────────────────┐
│ Check: Do certificates exist?       │
└────────┬────────────────┬───────────┘
         │ NO             │ YES
         ▼                ▼
┌──────────────────┐  ┌──────────────┐
│ Generate temp    │  │ Use existing │
│ self-signed cert │  │ certificates │
└────────┬─────────┘  └──────┬───────┘
         │                   │
         └────────┬──────────┘
                  ▼
      ┌──────────────────────┐
      │ Test nginx config    │
      └──────────┬───────────┘
                 ▼
      ┌──────────────────────┐
      │ Start nginx          │
      │ (with reload loop)   │
      └──────────────────────┘
```

## 💡 Benefits

- ✅ **Zero-touch deployment**: Certificates generated automatically
- ✅ **No manual steps**: Just `docker-compose up -d`
- ✅ **Idempotent**: Safe to run multiple times
- ✅ **Fail-safe**: Tests config before starting nginx
- ✅ **Transparent**: Clear logging of what happens

## 📚 Related Documentation

- **Complete Guide**: [changedoc/20-LETSENCRYPT_SSL_SETUP.md](changedoc/20-LETSENCRYPT_SSL_SETUP.md)
- **Quick Start**: [SSL_QUICK_START.md](SSL_QUICK_START.md)
- **Production Setup**: [SSL_PRODUCTION_SETUP.md](SSL_PRODUCTION_SETUP.md)

---

**Last Updated**: November 24, 2025  
**Version**: 2.0 (Automatic)  
**Status**: ✅ Production Ready  
**Time to Deploy**: 2 minutes

