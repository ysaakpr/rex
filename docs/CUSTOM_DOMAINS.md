# Custom Domains & Host Configuration

**Last Updated**: November 24, 2025  
**Version**: 1.0

## Overview

This guide explains how to configure the application to work with custom domains, staging environments, and non-localhost hostnames.

## Problem: Blocked Host Errors

When accessing the application through a custom domain or hostname, you might see:

```
Error - Blocked request. This host ("your-custom-domain.local") is not allowed.

To allow this host, add "your-custom-domain.local" to `server.allowedHosts` in vite.config.js.
```

This happens because:
1. Vite dev server has host validation for security
2. By default, it only allows `localhost`
3. Custom domains, staging URLs, or internal hostnames are blocked

## Solution

The application is now configured to accept requests from any hostname.

### 1. Vite Configuration

**File**: `frontend/vite.config.js`

```javascript
export default defineConfig({
  server: {
    host: '0.0.0.0',              // Listen on all network interfaces
    allowedHosts: 'all',          // ✅ Accept requests from any hostname
    // ... other config
  }
})
```

**Options**:
- `allowedHosts: 'all'` - Accept all hostnames (recommended for internal apps)
- `allowedHosts: ['domain1.com', 'domain2.com']` - Whitelist specific domains
- `allowedHosts: ['.example.com']` - Allow all subdomains of example.com

### 2. Nginx Configuration

**File**: `nginx.conf`

```nginx
server {
    listen 80;
    server_name _;  # ✅ Accept requests to any server name
    
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;  # ✅ Forward original hostname
        # ... other headers
    }
}
```

**What changed**:
- `server_name _` - Wildcard that matches any hostname
- `proxy_set_header Host $host` - Forwards the original Host header to Vite

## Use Cases

### 1. Custom Local Domains

**Scenario**: Using `/etc/hosts` for local development

```bash
# /etc/hosts
127.0.0.1 myapp.local
127.0.0.1 api.myapp.local
```

**Access**:
```
http://myapp.local/
```

✅ Works without additional configuration

### 2. Staging Environments

**Scenario**: Internal staging server

```
http://utm-backend-staging.company.internal
http://posthog-foss-test.dream11-stag.local
```

✅ Works with `allowedHosts: 'all'`

### 3. Docker Container Hostnames

**Scenario**: Accessing through Docker container name

```
http://utm-frontend:3000
```

✅ Works within Docker network

### 4. Cloud Development Environments

**Scenario**: AWS Cloud9, GitHub Codespaces, GitPod, etc.

```
http://abc123-3000.preview.cloud9.amazon.com
http://workspace-3000.gitpod.io
```

✅ Works with dynamic preview URLs

## Security Considerations

### Development Environment

Using `allowedHosts: 'all'` is **safe** for development because:
- ✅ Application is behind firewall/VPN
- ✅ Not exposed to public internet
- ✅ Allows flexibility for testing
- ✅ Simplifies team collaboration

### Production Environment

For production, you should:

**Option 1**: Restrict to specific domains
```javascript
// vite.config.js (production build)
export default defineConfig({
  server: {
    allowedHosts: [
      'yourdomain.com',
      'www.yourdomain.com',
      '.yourdomain.com',  // All subdomains
    ],
  }
})
```

**Option 2**: Use environment-based config
```javascript
export default defineConfig({
  server: {
    allowedHosts: process.env.NODE_ENV === 'production' 
      ? ['yourdomain.com']  // Production: strict
      : 'all',              // Development: permissive
  }
})
```

**Note**: For production, you typically don't run Vite dev server. You build static assets and serve them through nginx or CDN.

## Configuration Options

### Vite `allowedHosts` Options

```javascript
// Option 1: Allow all (most flexible)
allowedHosts: 'all'

// Option 2: Allow specific hosts
allowedHosts: ['localhost', 'myapp.local']

// Option 3: Allow domain and subdomains
allowedHosts: ['.example.com']  // Matches *.example.com

// Option 4: Auto-detect (default - restrictive)
// allowedHosts: undefined  // Only allows localhost
```

### Nginx `server_name` Options

```nginx
# Option 1: Accept all (wildcard)
server_name _;

# Option 2: Specific domain
server_name yourdomain.com;

# Option 3: Multiple domains
server_name yourdomain.com www.yourdomain.com;

# Option 4: Wildcard subdomain
server_name *.yourdomain.com;

# Option 5: Multiple patterns
server_name yourdomain.com *.yourdomain.com localhost;
```

## Testing Different Hosts

### 1. Add Host Entry

```bash
# Edit /etc/hosts (macOS/Linux) or C:\Windows\System32\drivers\etc\hosts (Windows)
sudo nano /etc/hosts

# Add entry
127.0.0.1 custom-domain.local
```

### 2. Restart Docker Services

```bash
docker-compose restart frontend nginx
```

### 3. Test Access

```bash
# Test nginx
curl -H "Host: custom-domain.local" http://localhost/

# Or open in browser
open http://custom-domain.local/
```

### 4. Verify Logs

```bash
# Check Vite accepts the host
docker-compose logs frontend | grep "custom-domain.local"

# Check nginx forwards correctly
docker-compose logs nginx | grep "custom-domain.local"
```

## Common Scenarios

### Scenario 1: Team Development with VPN

**Setup**: Team members access dev server through VPN

```
http://utm-backend-dev.vpn.company.com
```

**Configuration**: Already works with `allowedHosts: 'all'`

### Scenario 2: Preview Deployments

**Setup**: Each PR gets a preview URL

```
http://pr-123.utm-backend.staging.company.com
```

**Configuration**: Use wildcard in allowedHosts:
```javascript
allowedHosts: ['.staging.company.com']
```

### Scenario 3: Multi-Region Testing

**Setup**: Different regions use different domains

```
http://utm-backend-us.company.com
http://utm-backend-eu.company.com
http://utm-backend-asia.company.com
```

**Configuration**: Use environment variable:
```javascript
allowedHosts: process.env.ALLOWED_HOSTS?.split(',') || 'all'
```

Then in `.env`:
```bash
ALLOWED_HOSTS=utm-backend-us.company.com,utm-backend-eu.company.com,utm-backend-asia.company.com
```

## Troubleshooting

### Issue: Still Getting "Blocked Host" Error

**Solution 1**: Restart frontend container
```bash
docker-compose restart frontend
```

**Solution 2**: Verify configuration
```bash
docker-compose exec frontend cat /app/vite.config.js | grep allowedHosts
```

**Solution 3**: Check nginx is forwarding Host header
```bash
docker-compose exec nginx cat /etc/nginx/conf.d/default.conf | grep "proxy_set_header Host"
```

### Issue: Host Works But Sessions Don't Persist

**Problem**: Cookies are domain-specific

**Solution**: Update SuperTokens configuration in backend

```go
// cmd/api/main.go
session.Init(&sessmodels.TypeInput{
    CookieDomain: ".yourdomain.com",  // Wildcard for subdomains
    // ... other config
})
```

### Issue: CORS Errors with Custom Domain

**Problem**: Backend doesn't recognize the origin

**Solution**: Update CORS configuration in backend

```go
// internal/api/middleware/cors.go
AllowOrigins: []string{
    "http://localhost",
    "http://custom-domain.local",
    "http://*.staging.company.com",
}
```

Or use environment variable:
```bash
CORS_ALLOWED_ORIGINS=http://localhost,http://custom-domain.local
```

## Production Deployment

### Static Build (Recommended)

For production, build static assets:

```bash
# Build frontend
cd frontend
npm run build

# Serve with nginx (no Vite dev server)
```

In this case, `allowedHosts` doesn't apply because you're serving static files, not running Vite dev server.

### Vite Dev Server in Production (Not Recommended)

If you must run Vite dev server in production:

```javascript
export default defineConfig({
  server: {
    allowedHosts: [
      'yourdomain.com',
      'www.yourdomain.com',
    ],
  }
})
```

## Best Practices

### Development
✅ Use `allowedHosts: 'all'` for flexibility  
✅ Test with actual hostnames your team uses  
✅ Document custom domains in README  
✅ Use `.local` TLD for local domains

### Staging
✅ Use wildcard subdomains: `.staging.company.com`  
✅ Restrict to internal network  
✅ Enable logging for debugging  
✅ Test session persistence across domains

### Production
✅ Build static assets (don't use Vite dev server)  
✅ If using dev server, whitelist specific domains  
✅ Use HTTPS with proper certificates  
✅ Configure CDN for static assets  
✅ Set appropriate CORS policies

## Summary

**Current Configuration** ✅:
- Vite accepts requests from any hostname (`allowedHosts: 'all'`)
- Nginx forwards all hostnames (`server_name _`)
- Host header is properly forwarded to Vite
- Works with localhost, custom domains, staging URLs, etc.

**No Additional Configuration Needed** for:
- Custom local domains (`myapp.local`)
- Staging environments (`app.staging.company.com`)
- Cloud development environments
- VPN-based development servers
- Docker container hostnames

**Restart Required**:
```bash
docker-compose restart frontend nginx
```

---

**Questions or Issues?**
- Check logs: `docker-compose logs frontend`
- Verify config: `docker-compose exec frontend cat /app/vite.config.js`
- Test connectivity: `curl -H "Host: custom-domain.local" http://localhost/`


