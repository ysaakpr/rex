# Middleware Overview

This section provides authentication middleware implementations for multiple programming languages, both with and without SuperTokens SDK support.

## What is Middleware?

Middleware are functions that intercept HTTP requests before they reach your route handlers. They're used for:

- **Authentication**: Verify user is logged in
- **Authorization**: Check user has required permissions
- **Logging**: Record request details
- **CORS**: Handle cross-origin requests
- **Rate Limiting**: Prevent abuse

## Authentication Flow

```
HTTP Request
    ↓
Middleware extracts token (cookie or header)
    ↓
Token verified with SuperTokens
    ↓
User information extracted
    ↓
Request continues to handler
```

## Token Types

### Cookie-Based (Web Applications)

**Used By**: Web frontends (React, Vue, Angular)

**How It Works**:
- Session cookies automatically sent with requests
- HTTP-only cookies (XSS protection)
- Secure flag for HTTPS
- SameSite policy for CSRF protection

**Example**:
```javascript
fetch('/api/v1/tenants', {
  credentials: 'include'  // Send cookies
})
```

### Header-Based (APIs & Mobile)

**Used By**: Mobile apps, API clients, M2M services

**How It Works**:
- Access token sent in Authorization header
- Manual token management required
- Token refresh must be implemented

**Example**:
```bash
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
st-auth-mode: header
```

## Supported Languages

### With SuperTokens SDK

These languages have official SuperTokens SDKs that handle token verification automatically:

| Language | Framework | SDK | Docs Link |
|----------|-----------|-----|-----------|
| **Go** | Gin, Echo, Chi | ✅ Official | [Go Middleware](/middleware/go) |
| **Node.js** | Express, Koa, Fastify | ✅ Official | [Node.js Middleware](/middleware/nodejs) |
| **Python** | Flask, FastAPI, Django | ✅ Official | [Python Middleware](/middleware/python) |
| **PHP** | Laravel, Symfony | ✅ Community | - |

**Benefits**:
- Automatic token verification
- Built-in session management
- Easy integration
- Maintained by SuperTokens team

### Without SuperTokens SDK

These languages don't have official SDKs, so we verify JWT tokens manually:

| Language | Framework | Implementation | Docs Link |
|----------|-----------|----------------|-----------|
| **Java** | Spring Boot | Manual JWT | [Java Middleware](/middleware/java) |
| **C#/.NET** | ASP.NET Core | Manual JWT | [C# Middleware](/middleware/csharp) |
| **Rust** | Actix, Rocket | Manual JWT | - |
| **Ruby** | Rails, Sinatra | Manual JWT | - |

**What You Need**:
- JWT library for token parsing
- Public key from SuperTokens for verification
- Custom middleware implementation

## JWT Token Structure

### Access Token Claims

```json
{
  "sub": "user-id",               // User ID
  "exp": 1732358400,              // Expiry timestamp
  "iat": 1732272000,              // Issued at
  "sessionHandle": "...",         // Session identifier
  
  // System user specific (if applicable)
  "is_system_user": true,
  "service_name": "background-worker",
  "service_type": "worker"
}
```

### How to Verify JWT

1. **Extract token** from cookie or header
2. **Parse JWT** using library
3. **Verify signature** using SuperTokens public key
4. **Check expiry** (`exp` claim)
5. **Extract user info** from claims

## Getting SuperTokens Public Key

**Option 1: From SuperTokens Core (Recommended)**

```bash
# Get JWKS (JSON Web Key Set)
curl http://localhost:3567/recipe/jwt/jwks
```

Response:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-id",
      "n": "public-key-n",
      "e": "AQAB",
      "alg": "RS256",
      "use": "sig"
    }
  ]
}
```

**Option 2: Environment Variable**

For development, you can use a shared secret (less secure):

```bash
SUPERTOKENS_JWT_SECRET=your-secret-key
```

::: warning Production Security
In production, always fetch the public key from SuperTokens JWKS endpoint. Don't use a shared secret for production JWT verification.
:::

## Middleware Responsibilities

### 1. Authentication Middleware

**Purpose**: Verify user is logged in

**What It Does**:
- Extract token from cookie or header
- Verify token with SuperTokens (or validate JWT)
- Extract user ID
- Store user info in request context
- Return 401 if invalid

**Example** (Pseudocode):
```
function authMiddleware(request, response, next):
    token = extractToken(request)
    
    if not token:
        return response.status(401).json("No token")
    
    user = verifyToken(token)
    
    if not user:
        return response.status(401).json("Invalid token")
    
    request.userId = user.id
    next()
```

### 2. Tenant Access Middleware

**Purpose**: Verify user has access to tenant

**What It Does**:
- Get tenant ID from URL parameter
- Check if user is member of tenant (or platform admin)
- Check member status is active
- Return 403 if no access

**Example** (Pseudocode):
```
function tenantAccessMiddleware(request, response, next):
    tenantId = request.params.tenantId
    userId = request.userId
    
    // Check platform admin
    if isPlatformAdmin(userId):
        next()
        return
    
    // Check tenant membership
    member = getMembership(tenantId, userId)
    
    if not member or member.status != "active":
        return response.status(403).json("Access denied")
    
    request.member = member
    next()
```

### 3. RBAC Middleware

**Purpose**: Verify user has required permission

**What It Does**:
- Get required permission (service:entity:action)
- Check user's role in tenant
- Get role's policies
- Check if permission exists in policies
- Return 403 if permission denied

**Example** (Pseudocode):
```
function rbacMiddleware(requiredPermission):
    return function(request, response, next):
        userId = request.userId
        tenantId = request.tenantId
        
        hasPermission = checkPermission(
            tenantId, userId, requiredPermission
        )
        
        if not hasPermission:
            return response.status(403).json("Permission denied")
        
        next()
```

## Implementation Examples

Each language implementation includes:

1. **Basic auth middleware** - Verify user is logged in
2. **Header-based middleware** - For APIs
3. **Optional auth middleware** - Don't block if not logged in
4. **Full examples** - Complete, working code

### Quick Links

- **[Go (Reference Implementation)](/middleware/go)** - Using SuperTokens SDK
- **[Node.js/Express](/middleware/nodejs)** - Using SuperTokens SDK
- **[Python/Flask](/middleware/python)** - Using SuperTokens SDK
- **[Java/Spring Boot](/middleware/java)** - Manual JWT verification
- **[C#/.NET](/middleware/csharp)** - Manual JWT verification

## Testing Your Middleware

### Test Authentication

```bash
# Should return 401 (no token)
curl http://localhost:8080/api/v1/tenants

# Should work (with valid token)
curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer <token>" \
  -H "st-auth-mode: header"
```

### Test Tenant Access

```bash
# Should return 403 (not a member)
curl http://localhost:8080/api/v1/tenants/other-tenant-id/members \
  -H "Authorization: Bearer <token>"

# Should work (is a member)
curl http://localhost:8080/api/v1/tenants/your-tenant-id/members \
  -H "Authorization: Bearer <token>"
```

## Best Practices

### Security

1. **Always use HTTPS** in production
2. **Validate JWT signature** using public key
3. **Check token expiry** before accepting
4. **Use HTTP-only cookies** when possible
5. **Implement rate limiting** to prevent abuse

### Performance

1. **Cache public keys** from JWKS endpoint
2. **Avoid database calls** in auth middleware when possible
3. **Use connection pooling** for database queries
4. **Set appropriate timeouts**

### Error Handling

1. **Return proper status codes** (401, 403, 500)
2. **Don't leak sensitive info** in error messages
3. **Log authentication failures** for security monitoring
4. **Provide clear error messages** for debugging

## Common Issues

### "Session not found"

**Cause**: Cookies not being sent

**Solution**:
```javascript
// Frontend: Always include credentials
fetch(url, { credentials: 'include' })
```

### "Invalid token"

**Cause**: Token expired or incorrect verification key

**Solution**:
- Check token expiry
- Verify public key is correct
- Ensure clock sync between servers

### CORS errors

**Cause**: Frontend and backend on different origins

**Solution**:
```go
// Add CORS middleware
router.Use(middleware.CORS())
```

## Next Steps

- **[Go Implementation](/middleware/go)** - Reference implementation
- **[Java Implementation](/middleware/java)** - Manual JWT verification
- **[System Auth Library](/system-auth/overview)** - M2M authentication

