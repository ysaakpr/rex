# Centralized Frontend Configuration - Complete ✅

**Date**: November 25, 2025  
**Status**: Implemented and Working

## Overview

All frontend base path configuration has been **centralized into a single environment variable**: `VITE_BASE_PATH`. This eliminates hardcoded paths across multiple files and makes it trivial to deploy the app at any URL path.

## Single Point of Configuration

### Change This ONE File: `frontend/.env`

```bash
# Current configuration (app at /demo)
VITE_BASE_PATH=/demo

# Examples of other configurations:
# VITE_BASE_PATH=/         # Run at root
# VITE_BASE_PATH=/admin    # Run at /admin
# VITE_BASE_PATH=/app      # Run at /app

# API Domain (optional - defaults to window.location.origin)
VITE_API_DOMAIN=http://localhost
```

**That's it!** Change this one variable and restart the frontend container.

## How It Works

```
frontend/.env
    ↓
VITE_BASE_PATH=/demo
    ↓
┌──────────────────┬─────────────────┬──────────────────┐
│                  │                 │                  │
vite.config.js    src/config.js    App.jsx
│                  │                 │                  │
base: '/demo/'    BASENAME          BrowserRouter
HMR path          AUTH_PATH         SuperTokens
                  API_DOMAIN        React Router
```

## Files Created

### 1. `frontend/.env`
Environment configuration (gitignored, created from .env.example)

### 2. `frontend/.env.example`
Template with documentation and examples

### 3. `frontend/src/config.js`
**Centralized configuration module** - Single source of truth

```javascript
// All paths derived from VITE_BASE_PATH
export const config = {
  basePath: '/demo',          // Base path with trailing slash
  basename: '/demo',          // React Router basename (no trailing slash)
  authPath: '/demo/auth',     // SuperTokens auth pages
  apiDomain: 'http://localhost',
  websiteDomain: 'http://localhost',
  apiBasePath: '/api/auth',   // SuperTokens API endpoints
  appName: 'Rex',
};
```

### 4. `frontend/CONFIG.md`
Complete configuration documentation and troubleshooting guide

## Files Modified

### 1. `frontend/vite.config.js`
```javascript
// Before: Hardcoded
base: '/demo/',

// After: From environment
const basePath = env.VITE_BASE_PATH || '/'
const base = basePath.endsWith('/') ? basePath : `${basePath}/`
```

### 2. `frontend/src/App.jsx`
```javascript
// Before: Hardcoded values
<BrowserRouter basename="/demo">
SuperTokens.init({
  apiDomain: window.location.origin,
  websiteBasePath: "/demo/auth",
})

// After: From centralized config
import appConfig from './config';

<BrowserRouter basename={appConfig.basename}>
SuperTokens.init({
  apiDomain: appConfig.apiDomain,
  websiteBasePath: appConfig.authPath,
})
```

## Configuration Examples

### Example 1: Run at `/demo` (Current)

```bash
# frontend/.env
VITE_BASE_PATH=/demo
```

**Result:**
- App URL: `http://localhost/demo/`
- Auth: `http://localhost/demo/auth`
- Routes: `/demo/tenants`, `/demo/roles`, etc.

### Example 2: Run at Root `/`

```bash
# frontend/.env
VITE_BASE_PATH=/
```

**Result:**
- App URL: `http://localhost/`
- Auth: `http://localhost/auth`
- Routes: `/tenants`, `/roles`, etc.

### Example 3: Run at `/admin`

```bash
# frontend/.env
VITE_BASE_PATH=/admin
```

**Result:**
- App URL: `http://localhost/admin/`
- Auth: `http://localhost/admin/auth`
- Routes: `/admin/tenants`, `/admin/roles`, etc.

## Quick Start

### Change Base Path

```bash
# 1. Edit the .env file
vim frontend/.env

# 2. Change VITE_BASE_PATH to desired path
VITE_BASE_PATH=/your-path

# 3. Restart the frontend
docker-compose restart frontend

# 4. Access the app
open http://localhost/your-path/
```

### Verify Configuration

The app logs configuration in development mode (browser console):

```javascript
[Config] Frontend configuration: {
  BASE_PATH: '/demo',
  BASENAME: '/demo',
  AUTH_PATH: '/demo/auth',
  API_DOMAIN: 'http://localhost',
  WEBSITE_DOMAIN: 'http://localhost',
  env: 'development'
}

[SuperTokens] Initializing with config: {
  apiDomain: 'http://localhost',
  websiteDomain: 'http://localhost',
  authPath: '/demo/auth',
  apiBasePath: '/api/auth'
}
```

## Path Normalization Rules

The config module handles all edge cases:

| Input | BASENAME | AUTH_PATH | Description |
|-------|----------|-----------|-------------|
| `/demo/` | `/demo` | `/demo/auth` | Trailing slash removed |
| `/demo` | `/demo` | `/demo/auth` | Works with or without slash |
| `/` | `` | `/auth` | Root becomes empty basename |
| `/admin/` | `/admin` | `/admin/auth` | Any path supported |

## Benefits

✅ **Single Source of Truth** - One variable controls everything  
✅ **No Hardcoded Paths** - All paths dynamically generated  
✅ **Easy Deployment** - Different paths for dev/staging/prod  
✅ **Type-Safe** - Centralized config prevents typos  
✅ **Self-Documenting** - Config logs explain all values  
✅ **Maintainable** - Change once, updates everywhere  

## Production Deployment

### Building for Production

```bash
# 1. Set production base path
echo "VITE_BASE_PATH=/demo" > frontend/.env

# 2. Build
cd frontend && npm run build

# 3. Output includes base path in all assets
# dist/assets/index-abc123.js loaded from /demo/assets/
```

### Docker Deployment

The `.env` file is automatically mounted in Docker Compose:

```yaml
frontend:
  volumes:
    - ./frontend:/app  # .env file included
```

### Nginx Configuration

Match nginx location to base path:

```nginx
# For /demo
location /demo {
    proxy_pass http://frontend;
}

# For root /
location / {
    proxy_pass http://frontend;
}
```

## Troubleshooting

### Issue: SuperTokens "Please provide apiDomain" Error

**Cause**: Variable shadowing in initializeSuperTokens function

**Fix**: Renamed function parameter from `config` to `authProviderConfig` to avoid shadowing the imported `appConfig`

```javascript
// Before (broken)
function initializeSuperTokens(config) {
  // config parameter shadows imported config
  SuperTokens.init({ apiDomain: config.apiDomain })
}

// After (fixed)
function initializeSuperTokens(authProviderConfig) {
  // Uses imported appConfig properly
  SuperTokens.init({ apiDomain: appConfig.apiDomain })
}
```

### Issue: Changes Not Taking Effect

**Solution:**
```bash
docker-compose restart frontend
```

### Issue: Assets 404

**Check:**
1. Verify `VITE_BASE_PATH` in `.env`
2. Check browser Network tab - assets should load from base path
3. Restart frontend if .env was changed

### Issue: Routing Not Working

**Check:**
1. Browser console for config logs
2. React Router basename matches base path
3. All internal links use relative paths

## Testing Different Paths

```bash
# Test script
for path in "/" "/demo" "/admin"; do
  echo "Testing with VITE_BASE_PATH=$path"
  echo "VITE_BASE_PATH=$path" > frontend/.env
  docker-compose restart frontend
  sleep 5
  curl -s http://localhost$path/ | grep "<title>"
done
```

## Migration Guide

### From Multiple Hardcoded Paths → Single Config

**Before:**
```javascript
// vite.config.js
base: '/demo/'

// App.jsx
<BrowserRouter basename="/demo">
websiteBasePath: "/demo/auth"

// Multiple places to change!
```

**After:**
```bash
# Only one place to change
# frontend/.env
VITE_BASE_PATH=/demo
```

## Environment Variables Reference

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `VITE_BASE_PATH` | `/` | Base path for app | `/demo`, `/`, `/admin` |
| `VITE_API_DOMAIN` | `window.location.origin` | Backend API domain | `http://localhost` |

## Configuration Object API

Exported from `frontend/src/config.js`:

```typescript
{
  // Path configuration
  basePath: string,          // '/demo/' - with trailing slash
  basename: string,          // '/demo' - without trailing slash
  authPath: string,          // '/demo/auth' - SuperTokens path
  
  // API configuration
  apiDomain: string,         // 'http://localhost'
  websiteDomain: string,     // 'http://localhost'
  apiBasePath: string,       // '/api/auth'
  
  // App info
  appName: string,           // 'Rex'
}
```

## Real-World Scenarios

### Scenario 1: Marketing Site + App

**Setup:**
- Marketing site at root: `/`
- App at: `/demo`

**Configuration:**
```bash
# frontend/.env
VITE_BASE_PATH=/demo
```

**Nginx:**
```nginx
location / {
    proxy_pass http://marketing_site;
}
location /demo {
    proxy_pass http://frontend;
}
```

### Scenario 2: Docs + Demo + Admin

**Setup:**
- Documentation at: `/`
- Public demo at: `/demo`
- Admin app at: `/admin`

**Two frontend instances:**

```bash
# frontend-demo/.env
VITE_BASE_PATH=/demo

# frontend-admin/.env
VITE_BASE_PATH=/admin
```

### Scenario 3: Multi-Environment

**Development:**
```bash
VITE_BASE_PATH=/
```

**Staging:**
```bash
VITE_BASE_PATH=/staging
```

**Production:**
```bash
VITE_BASE_PATH=/
```

## Best Practices

1. ✅ **Always use environment variable** - Never hardcode paths
2. ✅ **Keep .env in .gitignore** - Use .env.example for templates
3. ✅ **Document your choice** - Add comments in .env explaining the path
4. ✅ **Test before deploying** - Verify all routes work with new path
5. ✅ **Match nginx config** - Ensure nginx location matches base path
6. ✅ **Update backend .env** - Keep WEBSITE_DOMAIN in sync

## Current Status

✅ **Configuration**: Fully centralized  
✅ **Documentation**: Complete  
✅ **Testing**: Verified working  
✅ **HMR**: Hot reload functional  
✅ **SuperTokens**: Properly configured  
✅ **React Router**: Basename set correctly  

## Quick Reference

**Change base path:**
```bash
# Edit one line
vim frontend/.env
# VITE_BASE_PATH=/your-path

# Restart
docker-compose restart frontend
```

**View current config (browser console):**
```javascript
// Development mode automatically logs:
[Config] Frontend configuration: { ... }
```

**Files to check:**
- `frontend/.env` - Your configuration
- `frontend/src/config.js` - Configuration logic
- `frontend/CONFIG.md` - Full documentation

## Support

For issues or questions:
1. Check browser console for config logs
2. Review `frontend/CONFIG.md` for detailed guide
3. Verify `.env` file exists and has `VITE_BASE_PATH`
4. Check that nginx location matches your base path

---

**Implementation Date**: November 25, 2025  
**Configuration Complexity**: ⭐ Simple (1 variable)  
**Flexibility**: ⭐⭐⭐⭐⭐ Maximum  
**Maintainability**: ⭐⭐⭐⭐⭐ Excellent

