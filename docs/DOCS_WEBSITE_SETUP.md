# Documentation Website Setup Guide

**Last Updated**: November 25, 2025

## Overview

The UTM Backend now features a comprehensive VitePress documentation website served at the root URL, with the admin dashboard accessible at `/demo`. This provides a documentation-first experience for users while maintaining full access to the live demo.

## Quick Start

### Access URLs

```
http://localhost/          → Documentation (VitePress)
http://localhost/demo      → Admin Dashboard (React)
http://localhost/api       → Backend API
http://localhost/inbox     → MailHog Email Testing
```

### First-Time Setup

```bash
# 1. Copy environment configuration
cp .env.example .env

# 2. Start all services
docker-compose up -d

# 3. Wait for services (30-60 seconds)
docker-compose ps

# 4. Access the documentation
open http://localhost

# 5. Try the demo
open http://localhost/demo
```

## Architecture

### Service Ports

| Service | Internal Port | Nginx Route | Description |
|---------|--------------|-------------|-------------|
| Documentation | 5173 | `/` | VitePress docs site |
| Frontend | 3000 | `/demo` | React admin dashboard |
| Backend API | 8080 | `/api` | Go REST API |
| MailHog | 8025 | `/inbox` | Email testing UI |

### Nginx Routing

```
┌─────────────────────────────────────────┐
│           Nginx (Port 80)                │
└───────────┬─────────────────────────────┘
            │
            ├─── / ──────────→ docs:5173 (VitePress)
            │
            ├─── /demo ──────→ frontend:3000 (React)
            │
            ├─── /api ───────→ api:8080 (Go)
            │
            └─── /inbox ─────→ mailhog:8025 (MailHog)
```

## Configuration Files

### Environment Variables (.env)

Key variables for the new setup:

```bash
# API Domain - unchanged
API_DOMAIN=http://localhost

# Website Domain - now points to /demo
WEBSITE_DOMAIN=http://localhost/demo

# Invitation Base URL - includes /demo path
INVITATION_BASE_URL=http://localhost/demo/accept-invite

# SuperTokens Configuration
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=your-super-secret-api-key

# API Base Path - unchanged
API_BASE_PATH=/api/auth
```

### Frontend Configuration

**vite.config.js**:
```javascript
export default defineConfig({
  base: '/demo/',  // Base path for assets
  server: {
    hmr: {
      path: '/demo/@vite/client',  // HMR WebSocket path
    }
  }
});
```

**App.jsx** (SuperTokens Init):
```javascript
SuperTokens.init({
  appInfo: {
    apiDomain: window.location.origin,
    websiteDomain: window.location.origin,
    apiBasePath: "/api/auth",
    websiteBasePath: "/demo/auth"  // Auth pages at /demo/auth
  }
});
```

## Documentation Website

### Technology Stack

- **VitePress**: Static site generator
- **Vue 3**: Framework (VitePress uses Vue internally)
- **Markdown**: Content format
- **Node 20**: Runtime

### Directory Structure

```
docs-website/
├── docs/
│   ├── .vitepress/
│   │   ├── config.mts        # VitePress configuration
│   │   └── cache/            # Build cache
│   ├── index.md              # Homepage
│   ├── getting-started/      # Getting started guides
│   ├── guides/               # Feature guides
│   ├── api/                  # API reference
│   ├── frontend/             # Frontend integration
│   └── ...                   # Other sections
├── Dockerfile                # Multi-stage Docker build
├── package.json              # Dependencies
└── README.md                 # Documentation README
```

### Development Commands

```bash
# Inside docs-website directory

# Start dev server (with HMR)
npm run docs:dev

# Build for production
npm run docs:build

# Preview production build
npm run docs:preview
```

### Adding New Documentation

1. **Create markdown file** in appropriate directory:
   ```bash
   touch docs-website/docs/guides/my-new-guide.md
   ```

2. **Update sidebar** in `docs-website/docs/.vitepress/config.mts`:
   ```javascript
   sidebar: [
     {
       text: 'Guides',
       items: [
         { text: 'My New Guide', link: '/guides/my-new-guide' }
       ]
     }
   ]
   ```

3. **Write content** using markdown with VitePress extensions:
   ```markdown
   # My New Guide
   
   ::: tip
   This is a helpful tip
   :::
   
   ```javascript
   // Code example
   console.log('Hello');
   ```
   ```

4. **Test locally**:
   ```bash
   cd docs-website
   npm run docs:dev
   ```

## Testing the Setup

### 1. Documentation Website

```bash
# Access root URL
open http://localhost/

# Check:
# ✓ Documentation loads
# ✓ Navigation works
# ✓ Search functionality
# ✓ "Try Demo" button links to /demo
```

### 2. Admin Dashboard

```bash
# Access demo URL
open http://localhost/demo

# Test sign up
open http://localhost/demo/auth/signup

# Test sign in
open http://localhost/demo/auth/signin

# Check:
# ✓ App loads at /demo
# ✓ All routes have /demo prefix
# ✓ Authentication works
# ✓ API calls succeed
```

### 3. Hot Module Replacement (HMR)

**Docs HMR**:
1. Edit `docs-website/docs/index.md`
2. Save file
3. Check browser auto-updates

**Frontend HMR**:
1. Edit `frontend/src/App.jsx`
2. Save file
3. Check browser auto-updates at `/demo`

### 4. API Integration

```bash
# Test API endpoints
curl http://localhost/api/health

# Test protected endpoint (after login)
curl http://localhost/api/v1/tenants \
  -H "Cookie: sAccessToken=..." \
  -H "Cookie: sRefreshToken=..."
```

## Troubleshooting

### Issue: 404 on /demo routes

**Symptom**: Frontend shows 404 errors

**Solution**:
1. Check `vite.config.js` has `base: '/demo/'`
2. Verify frontend service is running: `docker-compose ps frontend`
3. Check nginx logs: `docker-compose logs nginx`

### Issue: Authentication not working

**Symptom**: Can't sign in or session not maintained

**Solution**:
1. Verify `.env` has `WEBSITE_DOMAIN=http://localhost/demo`
2. Check `App.jsx` has `websiteBasePath: "/demo/auth"`
3. Clear browser cookies and cache
4. Restart services: `docker-compose restart frontend api`

### Issue: Assets not loading at /demo

**Symptom**: CSS/JS files show 404

**Solution**:
1. Ensure Vite base path is set correctly
2. Check browser console for path errors
3. Verify nginx rewrite rules in `nginx.conf`
4. Rebuild frontend: `docker-compose up -d --build frontend`

### Issue: Documentation not loading

**Symptom**: Root URL shows error

**Solution**:
1. Verify docs service is running: `docker-compose ps docs`
2. Check docs logs: `docker-compose logs docs`
3. Rebuild docs: `docker-compose up -d --build docs`
4. Check VitePress config: `docs-website/docs/.vitepress/config.mts`

### Issue: Invitation emails have wrong URL

**Symptom**: Email links point to wrong path

**Solution**:
1. Update `.env`: `INVITATION_BASE_URL=http://localhost/demo/accept-invite`
2. Restart API: `docker-compose restart api`
3. Test new invitation

## Production Deployment

### Environment Variables

Update `.env` for production:

```bash
# Production domains
API_DOMAIN=https://yourdomain.com
WEBSITE_DOMAIN=https://yourdomain.com/demo
INVITATION_BASE_URL=https://yourdomain.com/demo/accept-invite

# Security
APP_ENV=production
SUPERTOKENS_API_KEY=<strong-random-key-20-chars>
DB_SSL_MODE=require

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Build Process

```bash
# Build documentation
cd docs-website
npm run docs:build
# Output: docs/.vitepress/dist

# Build frontend
cd frontend
npm run build
# Output: dist/

# Backend builds in Docker
docker-compose -f docker-compose.prod.yml build
```

### Deployment Checklist

- [ ] Update all URLs in `.env`
- [ ] Set strong `SUPERTOKENS_API_KEY`
- [ ] Configure SSL certificates
- [ ] Enable HTTPS in nginx
- [ ] Set `APP_ENV=production`
- [ ] Configure database SSL
- [ ] Update CORS settings
- [ ] Test all routes work with domain
- [ ] Verify invitation emails
- [ ] Test authentication flows

## Maintenance

### Updating Documentation

```bash
# 1. Edit markdown files
vim docs-website/docs/guides/some-guide.md

# 2. Test locally
cd docs-website && npm run docs:dev

# 3. Commit changes
git add docs-website/
git commit -m "docs: update guide"

# 4. Deploy (Docker Compose rebuilds automatically)
docker-compose up -d --build docs
```

### Monitoring

```bash
# Check service health
docker-compose ps

# View logs
docker-compose logs -f docs
docker-compose logs -f frontend
docker-compose logs -f api

# Check resource usage
docker stats
```

### Backup

```bash
# Backup documentation source
tar -czf docs-backup-$(date +%Y%m%d).tar.gz docs-website/

# Backup environment config
cp .env .env.backup-$(date +%Y%m%d)
```

## Development Workflow

### Working on Documentation

```bash
# Start docs dev server directly (bypasses nginx)
cd docs-website
npm run docs:dev
# Access at http://localhost:5173

# Or use Docker Compose
docker-compose up docs
# Access at http://localhost (via nginx)
```

### Working on Frontend

```bash
# Start frontend dev server directly
cd frontend
npm run dev
# Access at http://localhost:3000

# Or use Docker Compose
docker-compose up frontend
# Access at http://localhost/demo (via nginx)
```

### Making Changes

1. **Update code** in your editor
2. **Save file** - HMR applies changes automatically
3. **Test** in browser at appropriate URL
4. **Commit** when satisfied
5. **Deploy** with `docker-compose up -d --build`

## Related Documentation

- [docs/changedoc/21-DOCS_WEBSITE_NGINX_RECONFIGURATION.md](changedoc/21-DOCS_WEBSITE_NGINX_RECONFIGURATION.md) - Complete change log
- [docs/QUICKSTART.md](QUICKSTART.md) - General quick start guide
- [docs/API_EXAMPLES.md](API_EXAMPLES.md) - API usage examples
- [docs-website/README.md](../docs-website/README.md) - VitePress documentation
- [frontend/README.md](../frontend/README.md) - Frontend documentation

## Support

For issues or questions:

1. Check the [troubleshooting section](#troubleshooting) above
2. Review [changedoc/21](changedoc/21-DOCS_WEBSITE_NGINX_RECONFIGURATION.md) for detailed information
3. Check logs: `docker-compose logs <service>`
4. Open an issue on GitHub

---

**Quick Commands Reference**

```bash
# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# Rebuild specific service
docker-compose up -d --build docs
docker-compose up -d --build frontend

# View logs
docker-compose logs -f docs
docker-compose logs -f frontend

# Restart service
docker-compose restart docs
docker-compose restart frontend

# Access services
open http://localhost           # Documentation
open http://localhost/demo      # Admin Dashboard
open http://localhost/inbox     # MailHog
```

