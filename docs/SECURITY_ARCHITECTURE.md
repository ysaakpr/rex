# Security Architecture - SuperTokens Integration

**Last Updated**: November 24, 2025  
**Version**: 1.0

## Overview

This document explains the secure architecture for SuperTokens integration in the UTM Backend system.

## ⚠️ Common Security Misconception

### ❌ WRONG: Exposing SuperTokens Core Directly

Some developers might think this is correct:

```
┌──────────┐
│ Frontend │ 
└─────┬────┘
      │ /auth/signin
      ↓
┌──────────────┐
│   Nginx      │
└─────┬────────┘
      │ Proxy /auth → supertokens:3567
      ↓
┌──────────────┐
│ SuperTokens  │  ⚠️ DIRECTLY EXPOSED
│    Core      │  ⚠️ SECURITY RISK
└──────────────┘
```

**Problems**:
1. ❌ SuperTokens Core directly exposed to internet
2. ❌ No backend validation or logging
3. ❌ Bypasses your API security layer
4. ❌ Difficult to add custom auth logic
5. ❌ Can't implement rate limiting at backend level

### ✅ CORRECT: Backend as Security Gateway

The proper architecture:

```
┌──────────┐
│ Frontend │ 
└─────┬────┘
      │
      ├─ /auth (UI pages) → Frontend (React Router)
      │
      └─ /api/auth (API calls) → Backend API
                                    ↓
                          ┌──────────────────┐
                          │   Backend API    │
                          │ (SuperTokens SDK)│ ✅ Security Gateway
                          └────────┬─────────┘
                                   │ Internal only
                                   │ Docker network
                                   ↓
                          ┌──────────────────┐
                          │  SuperTokens     │
                          │     Core         │ ✅ NOT exposed
                          │    :3567         │ ✅ Protected
                          └──────────────────┘
```

**Benefits**:
1. ✅ SuperTokens Core not directly accessible
2. ✅ Backend validates all auth requests
3. ✅ Can add custom logic (logging, analytics, etc.)
4. ✅ Can implement rate limiting
5. ✅ Can add additional security checks
6. ✅ Follows security best practices

## How It Works

### Frontend Configuration

```javascript
// frontend/src/App.jsx
SuperTokens.init({
  appInfo: {
    apiBasePath: "/api/auth",       // API requests go here
    websiteBasePath: "/auth"         // UI pages (React Router)
  }
})
```

### Request Flow Breakdown

#### 1. User Visits Login Page

```
Browser: http://localhost/auth/signin
    ↓
Nginx: / → frontend:3000
    ↓
Frontend: React Router handles /auth/signin
    ↓
Renders: SuperTokens login UI component
```

**Result**: User sees login page (client-side routing)

#### 2. User Submits Login Form

```
Browser: POST http://localhost/api/auth/signin
         { email: "user@example.com", password: "..." }
    ↓
Nginx: /api → api:8080
    ↓
Backend API: SuperTokens middleware intercepts /api/auth/signin
    ↓
Backend: Validates request, forwards to SuperTokens Core (internal)
    ↓
SuperTokens Core: Validates credentials, creates session
    ↓
Backend: Receives response, sets cookies
    ↓
Browser: Receives session cookies, redirects to dashboard
```

**Result**: User is authenticated, session established

#### 3. Authenticated API Request

```
Browser: GET http://localhost/api/v1/tenants
         Cookie: sAccessToken=...; sRefreshToken=...
    ↓
Nginx: /api → api:8080
    ↓
Backend API: AuthMiddleware validates session
    ↓
Backend: session.VerifySession() → queries SuperTokens Core (internal)
    ↓
SuperTokens Core: Validates session token
    ↓
Backend: Session valid, processes request
    ↓
Browser: Receives tenant data
```

**Result**: Authenticated user accesses protected resource

## Network Security

### Docker Network Isolation

```yaml
# docker-compose.yml

services:
  # PUBLIC ACCESS
  nginx:
    ports:
      - "80:80"          # Exposed to host
    networks:
      - utm-network
  
  # INTERNAL ONLY
  api:
    expose:              # Not ports:
      - "8080"           # Only accessible within Docker network
    networks:
      - utm-network
  
  # INTERNAL ONLY - MOST SECURE
  supertokens:
    expose:              # Not ports:
      - "3567"           # Only backend can access
    networks:
      - utm-network
```

**Security Layers**:
1. **Layer 1**: Nginx (exposed port 80) - only entry point
2. **Layer 2**: Backend API (internal) - accessible from nginx only
3. **Layer 3**: SuperTokens Core (internal) - accessible from backend only

### Port Exposure Comparison

| Service | Old (Insecure) | New (Secure) |
|---------|----------------|--------------|
| Nginx | `ports: - "80:80"` ✅ | `ports: - "80:80"` ✅ |
| Frontend | `ports: - "3000:3000"` ⚠️ | `expose: - "3000"` ✅ |
| API | `ports: - "8080:8080"` ⚠️ | `expose: - "8080"` ✅ |
| SuperTokens | `ports: - "3567:3567"` ❌ | `expose: - "3567"` ✅ |

**Why This Matters**:
- `ports:` exposes service to the host machine (and potentially the internet)
- `expose:` only makes service available within the Docker network
- Only nginx needs to be exposed

## Backend Security Implementation

### SuperTokens Middleware

```go
// cmd/api/main.go

// Initialize SuperTokens with internal connection
err := supertokens.Init(supertokens.TypeInput{
    Supertokens: &supertokens.ConnectionInfo{
        ConnectionURI: "http://supertokens:3567",  // Internal Docker network
        APIKey:        cfg.SuperTokens.APIKey,
    },
    // ... configuration
})

// Middleware automatically handles /api/auth/* endpoints
router.Use(middleware.SuperTokensMiddleware())
```

**What the Backend Does**:
1. **Handles auth endpoints**: `/api/auth/signin`, `/api/auth/signup`, etc.
2. **Validates sessions**: Checks tokens on protected routes
3. **Manages cookies**: Sets HttpOnly, Secure, SameSite cookies
4. **Rate limiting**: Can add rate limiting at backend level
5. **Logging**: Logs all auth events
6. **Custom logic**: Can add business logic before/after auth

### Session Validation

```go
// internal/api/middleware/auth.go

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verify session with SuperTokens Core (internal connection)
        session.VerifySession(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = c.Request.WithContext(r.Context())
        })).ServeHTTP(c.Writer, c.Request)
        
        // Extract user ID from validated session
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        // Continue with request
        c.Set("userID", userID)
        c.Next()
    }
}
```

## Nginx Configuration

### Correct Configuration

```nginx
# nginx.conf

upstream frontend {
    server frontend:3000;
}

upstream api {
    server api:8080;
}

# NO upstream for SuperTokens - it's internal only ✅

server {
    listen 80;
    
    # API routes (includes /api/auth)
    location /api {
        proxy_pass http://api;
        # ... headers ...
    }
    
    # Frontend routes (includes /auth UI pages)
    location / {
        proxy_pass http://frontend;
        # ... headers ...
    }
    
    # NO /auth proxy to SuperTokens ✅
}
```

## Security Checklist

### Development Environment

- [x] SuperTokens Core uses `expose` not `ports`
- [x] Backend API uses `expose` not `ports` (except through nginx)
- [x] Frontend uses `expose` not `ports` (except through nginx)
- [x] Only nginx exposes port 80
- [x] All services in same Docker network
- [x] SuperTokens connection URI uses internal hostname
- [x] Frontend apiBasePath is `/api/auth` (not `/auth`)

### Production Environment

Additional considerations:

- [ ] Enable HTTPS on nginx
- [ ] Use environment secrets (not hardcoded)
- [ ] Enable CORS restrictions (not wildcard)
- [ ] Implement rate limiting at nginx/backend
- [ ] Enable request logging
- [ ] Set up intrusion detection
- [ ] Regular security audits
- [ ] Keep SuperTokens Core updated

## Testing Security

### Verify SuperTokens is NOT Exposed

```bash
# This should FAIL (timeout or connection refused)
curl http://localhost:3567/hello

# This should work (through backend)
curl http://localhost/api/auth/hello
```

### Verify Docker Network Isolation

```bash
# From host - should fail
curl http://localhost:3567

# From inside backend container - should work
docker exec utm-api curl http://supertokens:3567/hello
```

### Verify Authentication Flow

```bash
# 1. Try accessing protected resource without auth (should fail)
curl http://localhost/api/v1/tenants

# 2. Sign up
curl -X POST http://localhost/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234!"}'

# 3. Sign in (get session cookies)
curl -X POST http://localhost/api/auth/signin \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"test@example.com","password":"Test1234!"}'

# 4. Access protected resource with cookies (should work)
curl http://localhost/api/v1/tenants -b cookies.txt
```

## Common Mistakes to Avoid

### ❌ Mistake 1: Direct SuperTokens Exposure

```yaml
# DON'T DO THIS
supertokens:
  ports:
    - "3567:3567"  # ❌ EXPOSED TO INTERNET
```

### ❌ Mistake 2: Wrong Nginx Proxy

```nginx
# DON'T DO THIS
location /auth {
    proxy_pass http://supertokens:3567;  # ❌ BYPASSES BACKEND
}
```

### ❌ Mistake 3: Wrong Frontend Config

```javascript
// DON'T DO THIS
SuperTokens.init({
  apiBasePath: "/auth",  // ❌ WRONG - should be /api/auth
})
```

### ❌ Mistake 4: External SuperTokens URI

```go
// DON'T DO THIS
ConnectionURI: "http://localhost:3567",  // ❌ Won't work in Docker
```

## References

- [SuperTokens Architecture](https://supertokens.com/docs/community/architecture)
- [Docker Network Security](https://docs.docker.com/network/network-tutorial-standalone/)
- [Nginx Security Best Practices](https://www.nginx.com/blog/nginx-security-best-practices/)

## Summary

✅ **Current Architecture is Secure**:
- SuperTokens Core is NOT exposed to the internet
- All authentication goes through the backend API
- Backend acts as a security gateway
- Proper Docker network isolation
- Frontend never directly accesses SuperTokens Core

❌ **What NOT to Do**:
- Never expose SuperTokens Core directly
- Never bypass backend for auth requests
- Never use `ports:` for internal services
- Never hardcode secrets in configuration

---

**Security First**: Always route authentication through your backend API, never expose authentication infrastructure directly to the internet.


