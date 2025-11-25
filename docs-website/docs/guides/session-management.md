# Session Management

Advanced session management topics with SuperTokens.

## Session Lifecycle

### Token Types

**Access Token**:
- Short-lived: 1 hour (users), 24 hours (system users)
- Contains user info and permissions
- Sent with every API request
- Automatically refreshed before expiry

**Refresh Token**:
- Longer-lived: 100 days (default)
- Used to obtain new access tokens
- Stored securely (HTTP-only cookie or secure storage)
- Rotates on each refresh

### Session States

```
Active → Expired → Revoked
  ↓
Refreshed
```

**Active**: Valid session, user authenticated  
**Expired**: Access token expired, needs refresh  
**Revoked**: Session invalidated (logout, security)

## Token Storage

### Cookie-Based (Web)

**Automatically Managed**:
```javascript
// Cookies set by SuperTokens
sAccessToken=...      // Access token
sRefreshToken=...     // Refresh token  
sIdRefreshToken=...   // Anti-CSRF token
```

**Properties**:
- `HttpOnly`: JavaScript cannot access
- `Secure`: HTTPS only (production)
- `SameSite=Lax`: CSRF protection
- `Path=/`: Available site-wide

### Header-Based (API/Mobile)

**Manual Management**:
```bash
Authorization: Bearer <access-token>
st-auth-mode: header
```

**Client Responsibilities**:
- Store tokens securely
- Refresh before expiry
- Handle 401 errors
- Clear on logout

## Token Refresh

### Automatic (Frontend SDK)

SuperTokens SDK handles refresh automatically:

```javascript
// No manual refresh needed!
const response = await fetch('/api/v1/data', {
  credentials: 'include'
});
// SDK refreshes token if near expiry
```

### Manual (Without SDK)

```javascript
async function refreshToken() {
  const response = await fetch('/api/auth/session/refresh', {
    method: 'POST',
    credentials: 'include'
  });
  
  if (response.ok) {
    // New access token in response headers
    return true;
  }
  return false;
}

// Use before API calls
if (isTokenExpiringSoon()) {
  await refreshToken();
}
await makeAPICall();
```

## Session Verification

### Backend Verification

```go
// Using SuperTokens SDK (automatic)
session.VerifySession(nil, handler)

// Get session info
sessionContainer := session.GetSessionFromRequestContext(ctx)
userID := sessionContainer.GetUserID()
```

### Manual JWT Verification

For languages without SDK:

```java
// Verify JWT signature
Claims claims = Jwts.parserBuilder()
    .setSigningKey(publicKey)
    .build()
    .parseClaimsJws(token)
    .getBody();

// Check expiry
Date expiry = claims.getExpiration();
if (expiry.before(new Date())) {
    throw new TokenExpiredException();
}

// Extract user ID
String userId = claims.getSubject();
```

## Session Revocation

### Sign Out

```javascript
// Frontend
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";

await EmailPassword.signOut();
// All sessions revoked
```

### Revoke Specific Session

```go
// Backend
session.RevokeSession(sessionHandle)
```

### Revoke All User Sessions

```go
// Revoke all sessions for user
session.RevokeAllSessionsForUser(userID)
```

## Security Features

### Anti-CSRF Protection

**Enabled by Default**:
- `sIdRefreshToken` cookie validates requests
- Prevents cross-site request forgery
- Automatic with SuperTokens

### Token Rotation

**Refresh Tokens Rotate**:
```
1. Client sends refresh token
2. Server validates and creates new tokens
3. Old refresh token invalidated
4. New tokens returned
5. Client stores new tokens
```

**Benefits**:
- Limits impact of stolen tokens
- Detects token reuse
- Enhanced security

## Configuration

### Token Expiry

```go
// In SuperTokens init
session.Init(&sessmodels.TypeInput{
    AccessTokenValidity:  3600,    // 1 hour
    RefreshTokenValidity: 8640000, // 100 days
})
```

### Cookie Settings

```go
session.Init(&sessmodels.TypeInput{
    CookieSameSite: ptrString("lax"),
    CookieSecure:   ptrBool(true), // HTTPS only
    CookieDomain:   ptrString(".yourdomain.com"),
})
```

## Best Practices

### 1. Always Use HTTPS in Production

```bash
# .env (production)
COOKIE_SECURE=true
API_DOMAIN=https://api.yourdomain.com
```

### 2. Set Appropriate Expiry Times

```
Access Token:  1 hour (users), 24 hours (system users)
Refresh Token: 100 days (default)
```

### 3. Handle Token Refresh Gracefully

```javascript
// Check expiry before important operations
if (tokenExpiringSoon()) {
  await refreshToken();
}
await criticalOperation();
```

### 4. Clear Tokens on Logout

```javascript
// Frontend
await signOut();
localStorage.clear();
```

### 5. Monitor Session Activity

```go
// Log session events
log.Info("Session created", zap.String("user_id", userID))
log.Info("Session revoked", zap.String("session_handle", handle))
```

## Next Steps

- [User Authentication](/guides/user-authentication) - Complete auth flow
- [System Users](/guides/system-users) - M2M authentication
- [Authentication API](/x-api/authentication) - API reference
