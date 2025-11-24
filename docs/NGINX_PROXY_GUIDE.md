# Nginx Reverse Proxy Guide

**Last Updated**: November 24, 2025  
**Version**: 1.0

## Overview

The UTM Backend uses Nginx as a reverse proxy to route requests to the appropriate services. This provides a unified entry point for all frontend, API, and authentication requests.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Client Browser                       │
└───────────────────┬─────────────────────────────────────┘
                    │
                    │ http://localhost
                    ↓
        ┌───────────────────────┐
        │   Nginx Proxy :80     │
        └───────────┬───────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
   /    │      /api │           │
        ↓           ↓           │
┌─────────────┐ ┌──────────────┴───────┐
│  Frontend   │ │      Backend API      │
│   :3000     │ │        :8080          │
│             │ │  (SuperTokens SDK)    │
└─────────────┘ └──────────┬───────────┘
                           │ Internal
                           │ Connection
                           ↓
                  ┌────────────────┐
                  │ SuperTokens    │
                  │    Core        │
                  │   :3567        │
                  │  (NOT EXPOSED) │
                  └────────────────┘
```

**Key Security Feature**: SuperTokens Core is never directly exposed to the internet. All authentication requests go through your backend API.

## Routing Rules

### 1. Frontend Routes (`/`)

**Pattern**: `/*` (default/fallback)  
**Target**: `frontend:3000` (React + Vite)  
**Example**:
```
http://localhost/            → frontend:3000/
http://localhost/tenants     → frontend:3000/tenants
http://localhost/auth        → frontend:3000/auth (SuperTokens UI)
```

**Features**:
- WebSocket support for Vite Hot Module Replacement (HMR)
- Automatic refresh on code changes
- Single Page Application (SPA) routing

### 2. API Routes (`/api`)

**Pattern**: `/api/*`  
**Target**: `api:8080` (Go Gin Backend)  
**Example**:
```
http://localhost/api/v1/tenants              → api:8080/api/v1/tenants
http://localhost/api/v1/users/me             → api:8080/api/v1/users/me
http://localhost/api/v1/authorize            → api:8080/api/v1/authorize
```

**Features**:
- CORS headers for cross-origin requests
- Cookie forwarding for session management
- Large request body support (10MB)
- Extended timeouts (60s)

### 3. Authentication Routes (`/auth`)

**Pattern**: `/auth/*`  
**Target**: `frontend:3000` (React Router - Client-side routing)  
**Example**:
```
http://localhost/auth              → Frontend React app
http://localhost/auth/signin       → Frontend React app (SuperTokens UI)
http://localhost/auth/signup       → Frontend React app (SuperTokens UI)
```

**Important**: These are **UI routes** handled by React Router, NOT API routes.

**Actual Authentication API** (`/api/auth`):
```
http://localhost/api/auth/*        → Backend API → SuperTokens Core (internal)
```

**Security Note**: ⚠️
- SuperTokens Core is **NOT** directly exposed to the internet
- Frontend calls `/api/auth/*` which goes to your Go backend
- Backend communicates with SuperTokens Core internally
- This is the secure, recommended architecture

### 4. Health Check Route (`/health`)

**Pattern**: `/health`  
**Target**: Nginx itself  
**Response**: `200 OK` with "healthy"

Used for:
- Container health checks
- Load balancer health probes
- Monitoring systems

## Configuration Files

### 1. `nginx.conf`

Main Nginx configuration file located at the project root.

**Key Sections**:

#### Upstream Definitions
```nginx
upstream frontend {
    server frontend:3000;
}

upstream api {
    server api:8080;
}

upstream supertokens {
    server supertokens:3567;
}
```

#### Server Block
```nginx
server {
    listen 80;
    server_name localhost;
    
    # ... location blocks ...
}
```

### 2. `docker-compose.yml`

Nginx service configuration:

```yaml
nginx:
  image: nginx:alpine
  container_name: utm-nginx
  ports:
    - "80:80"
  volumes:
    - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
  depends_on:
    - frontend
    - api
    - supertokens
  networks:
    - utm-network
  restart: unless-stopped
```

## CORS Configuration

### Why CORS?

Cross-Origin Resource Sharing (CORS) is required because:
1. Frontend runs on one origin (via Nginx)
2. API and SuperTokens are separate services
3. Browsers enforce same-origin policy

### Configured Headers

```nginx
# Allow requests from any origin (development mode)
Access-Control-Allow-Origin: *

# Allowed HTTP methods
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS

# Allowed request headers
Access-Control-Allow-Headers: 
  DNT, User-Agent, X-Requested-With, If-Modified-Since, 
  Cache-Control, Content-Type, Range, Authorization, 
  st-auth-mode, rid, fdi-version, anti-csrf

# Allow credentials (cookies)
Access-Control-Allow-Credentials: true

# Expose response headers to frontend
Access-Control-Expose-Headers: 
  Content-Length, Content-Range, front-token, 
  id-refresh-token, anti-csrf
```

### Preflight Requests

Nginx handles OPTIONS preflight requests automatically:

```nginx
if ($request_method = 'OPTIONS') {
    add_header 'Access-Control-Allow-Origin' '*' always;
    # ... other headers ...
    return 204;
}
```

## Development Setup

### Starting Services

```bash
# Start all services including nginx
docker-compose up -d

# View nginx logs
docker-compose logs -f nginx

# Check nginx status
docker-compose ps nginx
```

### Accessing the Application

| Service | Direct Access | Via Nginx |
|---------|--------------|-----------|
| Frontend | http://localhost:3000 | http://localhost/ |
| API | http://localhost:8080 | http://localhost/api |
| SuperTokens | http://localhost:3567 | http://localhost/auth |
| Nginx Health | N/A | http://localhost/health |

**Recommended**: Use `http://localhost` (via Nginx) for all development to match production behavior.

### Testing Routing

```bash
# Test frontend
curl http://localhost/

# Test API
curl http://localhost/api/v1/health

# Test auth endpoint
curl http://localhost/auth/hello

# Test health check
curl http://localhost/health
```

## Production Considerations

### 1. Security

**SuperTokens Core Protection** ✅:

The current configuration is already secure:
- ✅ SuperTokens Core is NOT directly exposed
- ✅ All auth requests go through `/api/auth` → Backend API → SuperTokens (internal)
- ✅ Only your Go backend can communicate with SuperTokens Core
- ✅ Frontend never directly accesses SuperTokens Core

**Why This Matters**:
- Direct exposure of SuperTokens Core would be a security risk
- Your backend acts as a security gateway
- Backend can add additional validation, logging, rate limiting
- Follows the principle of least privilege

**Enable HTTPS**:
```nginx
server {
    listen 443 ssl http2;
    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;
    
    # ... rest of configuration ...
}
```

**Restrict CORS**:
```nginx
# Replace wildcard with specific origin
add_header 'Access-Control-Allow-Origin' 'https://yourdomain.com' always;
```

**Add Security Headers**:
```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
```

### 2. Performance

**Enable Gzip Compression**:
```nginx
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript 
           application/x-javascript application/xml+rss 
           application/json application/javascript;
```

**Add Caching**:
```nginx
# Cache static assets
location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

**Connection Limits**:
```nginx
# Limit connections per IP
limit_conn_zone $binary_remote_addr zone=addr:10m;
limit_conn addr 10;
```

### 3. Monitoring

**Access Logs**:
```nginx
access_log /var/log/nginx/access.log combined;
error_log /var/log/nginx/error.log warn;
```

**Custom Log Format**:
```nginx
log_format detailed '$remote_addr - $remote_user [$time_local] '
                   '"$request" $status $body_bytes_sent '
                   '"$http_referer" "$http_user_agent" '
                   '$request_time';
```

## Troubleshooting

### Issue: 502 Bad Gateway

**Cause**: Backend service is not running or not reachable

**Solution**:
```bash
# Check if backend services are running
docker-compose ps

# Check backend logs
docker-compose logs api
docker-compose logs supertokens

# Restart services
docker-compose restart api supertokens
```

### Issue: CORS Errors

**Symptoms**: Browser console shows CORS-related errors

**Solution**:
1. Check nginx CORS headers are configured
2. Verify `Access-Control-Allow-Credentials: true` is set
3. Ensure frontend makes requests with `credentials: 'include'`
4. Check browser dev tools Network tab for response headers

**Debug**:
```bash
# Test CORS preflight
curl -X OPTIONS http://localhost/api/v1/tenants \
  -H "Origin: http://localhost" \
  -H "Access-Control-Request-Method: GET" \
  -v
```

### Issue: WebSocket Connection Failed (Vite HMR)

**Symptoms**: Hot module replacement not working in development

**Solution**:
1. Check nginx WebSocket configuration:
   ```nginx
   proxy_set_header Upgrade $http_upgrade;
   proxy_set_header Connection "upgrade";
   ```

2. Update Vite config:
   ```javascript
   hmr: {
     clientPort: 80,
   }
   ```

3. Restart nginx:
   ```bash
   docker-compose restart nginx
   ```

### Issue: Request Timeout

**Symptoms**: Requests take too long and fail

**Solution**: Increase timeout values in nginx.conf:
```nginx
proxy_connect_timeout 120s;
proxy_send_timeout 120s;
proxy_read_timeout 120s;
```

### Issue: Large File Upload Fails

**Symptoms**: 413 Request Entity Too Large

**Solution**: Increase max body size:
```nginx
client_max_body_size 50M;  # Adjust as needed
```

### Issue: Cookie Not Set

**Symptoms**: Session cookies not being stored

**Solution**:
1. Check `Access-Control-Allow-Credentials: true` is set
2. Verify frontend uses `credentials: 'include'`
3. Check cookie domain and path settings
4. Ensure SameSite policy is compatible

## Nginx Commands

### View Configuration

```bash
# View nginx config inside container
docker-compose exec nginx cat /etc/nginx/conf.d/default.conf

# Test configuration syntax
docker-compose exec nginx nginx -t
```

### Reload Configuration

```bash
# Reload nginx without downtime
docker-compose exec nginx nginx -s reload

# Or restart the container
docker-compose restart nginx
```

### View Logs

```bash
# Follow nginx access logs
docker-compose logs -f nginx

# View last 100 lines
docker-compose logs --tail=100 nginx

# View error logs only
docker-compose exec nginx tail -f /var/log/nginx/error.log
```

### Debug Mode

Enable debug logging in nginx.conf:
```nginx
error_log /var/log/nginx/error.log debug;
```

Then restart:
```bash
docker-compose restart nginx
docker-compose logs -f nginx
```

## Best Practices

### 1. Development

✅ **Use Nginx for all requests** instead of direct service access  
✅ **Enable HMR** for better development experience  
✅ **Use detailed logging** for debugging  
✅ **Test CORS** configuration early

### 2. Configuration Management

✅ **Keep nginx.conf** in version control  
✅ **Use environment variables** for dynamic values  
✅ **Document custom configurations**  
✅ **Test changes locally** before deploying

### 3. Security

✅ **Never expose backend ports** directly in production  
✅ **Use HTTPS** in production  
✅ **Implement rate limiting**  
✅ **Restrict CORS origins** in production  
✅ **Add security headers**

### 4. Performance

✅ **Enable gzip compression**  
✅ **Cache static assets**  
✅ **Set appropriate timeouts**  
✅ **Use HTTP/2** when possible

## Migration from Direct Access

If you were previously accessing services directly:

### Frontend Changes

```javascript
// ❌ Old (direct access)
const response = await fetch('http://localhost:8080/api/v1/tenants');

// ✅ New (via nginx)
const response = await fetch('/api/v1/tenants');
```

### Vite Proxy Changes

The Vite dev server now proxies through docker network, nginx handles browser requests.

### Environment Variables

No changes needed! The backend services remain the same, only the routing changes.

## Additional Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [Nginx Reverse Proxy Guide](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/)
- [Docker Networking](https://docs.docker.com/network/)
- [Vite Proxy Configuration](https://vitejs.dev/config/server-options.html#server-proxy)

---

**Need Help?**
- Check nginx logs: `docker-compose logs nginx`
- Test configuration: `docker-compose exec nginx nginx -t`
- Restart nginx: `docker-compose restart nginx`


