# Production Deployment Checklist

## 📋 Pre-Deployment

- [ ] Decide on basePath (e.g., `/demo`, `/`, `/admin`)
- [ ] Confirm domain name (e.g., `https://yourdomain.com`)
- [ ] Have SSL certificates ready (if using HTTPS)
- [ ] Database is backed up

## 🔧 Configuration Steps

### 1. Frontend Build

Choose ONE method:

#### Method A: Using Deployment Script (Recommended)
```bash
./scripts/deploy-with-basepath.sh /demo https://yourdomain.com
```

#### Method B: Manual Docker Build
```bash
docker build \
  --build-arg VITE_BASE_PATH=/demo \
  --build-arg VITE_API_DOMAIN=https://yourdomain.com \
  -f frontend/Dockerfile.prod \
  -t utm-frontend:prod \
  ./frontend
```

**Verification:**
- [ ] Build completes without errors
- [ ] Image tagged correctly

### 2. Backend Configuration

Edit `.env` file:

```bash
# Invitation URLs (MUST match frontend basePath!)
INVITATION_BASE_URL=https://yourdomain.com/demo/accept-invite

# SuperTokens
SUPERTOKENS_WEBSITE_DOMAIN=https://yourdomain.com
SUPERTOKENS_API_DOMAIN=https://yourdomain.com

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_secure_password
DB_NAME=utm_backend

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Email (for invitations)
EMAIL_PROVIDER=smtp  # or mailhog for testing
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASSWORD=your_smtp_password
EMAIL_FROM_ADDRESS=noreply@yourdomain.com
```

**Verification:**
- [ ] `INVITATION_BASE_URL` matches frontend basePath
- [ ] All sensitive values are set (not defaults)
- [ ] Email configuration is correct

### 3. Nginx Configuration

For basePath `/demo`:

```nginx
location /demo {
    proxy_pass http://frontend:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location /api {
    proxy_pass http://api:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Cookie $http_cookie;
}
```

**Verification:**
- [ ] Nginx config syntax is valid: `nginx -t`
- [ ] Locations match basePath

### 4. Deploy Services

```bash
# Using docker-compose
docker-compose up -d --build

# Or with the deployment script
./scripts/deploy-with-basepath.sh /demo https://yourdomain.com
```

**Verification:**
- [ ] All containers are running: `docker-compose ps`
- [ ] No errors in logs: `docker-compose logs -f`

## ✅ Post-Deployment Verification

### Frontend Checks

Visit: `https://yourdomain.com/demo/` (or your basePath)

- [ ] Frontend loads without errors
- [ ] Assets load (check browser DevTools → Network, no 404s)
- [ ] Console shows correct configuration (in dev mode)
- [ ] Favicon and images display correctly

### Authentication Flow

- [ ] Can access auth page: `/demo/auth`
- [ ] Can sign up / sign in
- [ ] After login, redirects to `/demo/tenants` (not `/demo/demo/tenants`)
- [ ] Can sign out
- [ ] After signout, redirects to `/demo/auth` (not just `/auth`)

### Invitation System

Create a test invitation:

- [ ] Go to tenant → Users → Invite user
- [ ] Copy invitation link from UI
- [ ] Link format is correct: `https://yourdomain.com/demo/accept-invite?token=...`
- [ ] Check email (if email is configured)
- [ ] Email link matches UI link
- [ ] Clicking invitation link loads accept page correctly

### API Endpoints

Test a few endpoints:

```bash
# Health check
curl https://yourdomain.com/api/health

# Auth config (should return JSON)
curl https://yourdomain.com/api/v1/auth/config
```

- [ ] API responds correctly
- [ ] No CORS errors
- [ ] Cookies are set/sent correctly

### Database & Services

```bash
# Check database connection
docker-compose exec postgres psql -U utmuser -d utm_backend -c "SELECT version();"

# Check Redis
docker-compose exec redis redis-cli ping

# Check all services are healthy
docker-compose ps
```

- [ ] Database is accessible
- [ ] Redis is running
- [ ] All services show "Up" status
- [ ] No services are restarting continuously

## 🔒 Security Checks

- [ ] SSL/TLS is enabled (HTTPS)
- [ ] SuperTokens cookies have `secure` flag (in production)
- [ ] Database credentials are strong
- [ ] `.env` file is not committed to git
- [ ] SuperTokens API key is set and secure
- [ ] CORS is configured correctly

## 📊 Monitoring Setup

- [ ] Set up log aggregation
- [ ] Configure uptime monitoring
- [ ] Set up error alerting
- [ ] Database backup schedule is active
- [ ] Disk space monitoring is configured

## 📝 Documentation Updates

- [ ] Document the chosen basePath
- [ ] Update team wiki/docs with access URLs
- [ ] Document any production-specific configuration
- [ ] Update runbooks with production commands

## 🚨 Rollback Plan

If something goes wrong:

```bash
# Stop services
docker-compose down

# Restore from backup (if needed)
# ... your backup restoration process ...

# Rebuild with previous working configuration
docker build --build-arg VITE_BASE_PATH=/old-path ...

# Restart
docker-compose up -d
```

## 📞 Support Contacts

Document who to contact for:

- [ ] Infrastructure issues
- [ ] Application issues  
- [ ] Database issues
- [ ] Domain/DNS issues

## ✨ Success Criteria

Deployment is successful when:

- ✅ All services are running
- ✅ Users can sign up/sign in
- ✅ Invitations work correctly (UI and email)
- ✅ No 404 errors on assets
- ✅ No double-path issues (e.g., `/demo/demo/`)
- ✅ All URLs in the app use correct basePath
- ✅ Cookies and authentication work properly
- ✅ Backend invitation URLs match frontend

## 📚 Reference Documents

- `BASEPATH_QUICK_REFERENCE.md` - Quick commands
- `PRODUCTION_BASEPATH_CONFIG.md` - Detailed guide
- `frontend/CONFIG.md` - Frontend config details
- `scripts/deploy-with-basepath.sh` - Deployment script

---

**Last Updated:** Check git log for this file  
**Review Frequency:** Before each production deployment
