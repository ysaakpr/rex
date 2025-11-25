# Authentication Overview

Complete guide to authentication in Rex using SuperTokens.

## What is SuperTokens?

SuperTokens is an open-source authentication solution that provides:
- User registration and login
- Session management
- Token refresh
- OAuth integration
- Password reset
- Email verification

## Authentication Modes

Rex supports two authentication modes:

### 1. Cookie-Based (Web Applications)

**Best For**: Web frontends (React, Vue, Angular)

**How It Works**:
- Session tokens stored in HTTP-only cookies
- Cookies automatically sent with requests
- XSS protection (JavaScript can't access)
- CSRF protection built-in

**Frontend Usage**:
```javascript
// Cookies sent automatically
fetch('/api/v1/tenants', {
  credentials: 'include'  // Required!
})
```

### 2. Header-Based (APIs & Mobile)

**Best For**: Mobile apps, API clients, M2M services

**How It Works**:
- Access token sent in Authorization header
- Token stored by client (secure storage)
- Manual token management required

**Client Usage**:
```bash
Authorization: Bearer <access-token>
st-auth-mode: header
```

## User Types

### Regular Users

**Purpose**: Human users who log in via web or mobile

**Authentication**: Email/password or Google OAuth

**Token Expiry**: 1 hour (access token)

**Use Cases**:
- Web application users
- Mobile app users
- Admin panel users

### System Users

**Purpose**: Service accounts for automated processes

**Authentication**: Email/password (programmatic)

**Token Expiry**: 24 hours (access token)

**Use Cases**:
- Background workers
- Scheduled jobs
- API integrations
- Data pipelines

**[Learn more about System Users →](/guides/system-users)**

## Authentication Flow

### Sign Up Flow

```
1. User submits email + password
   ↓
2. Frontend → POST /api/auth/signup
   ↓
3. SuperTokens creates user account
   ↓
4. SuperTokens creates session
   ↓
5. Session cookies set (or tokens returned)
   ↓
6. User redirected to dashboard
```

**Frontend Code**:
```javascript
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";

// SuperTokens handles the UI and flow
// User navigates to /auth
```

### Sign In Flow

```
1. User submits credentials
   ↓
2. Frontend → POST /api/auth/signin
   ↓
3. SuperTokens verifies credentials
   ↓
4. SuperTokens creates session
   ↓
5. Session cookies set
   ↓
6. User authenticated
```

### API Request Flow

```
1. Client makes API request
   ↓
2. Request includes session cookie/token
   ↓
3. Backend middleware verifies session
   ↓
4. SuperTokens validates token
   ↓
5. User ID extracted
   ↓
6. Request proceeds to handler
```

## Backend Verification

### How Backend Verifies Sessions

**With SuperTokens SDK** (Go, Node.js, Python):
```go
// Gin middleware
session.VerifySession(nil, handler)

// Inside handler
sessionContainer := session.GetSessionFromRequestContext(ctx)
userID := sessionContainer.GetUserID()
```

**Without SDK** (Java, C#):
```java
// Extract and verify JWT manually
Claims claims = Jwts.parserBuilder()
    .setSigningKey(publicKey)
    .build()
    .parseClaimsJws(token)
    .getBody();

String userId = claims.getSubject();
```

## Session Management

### Token Types

**Access Token**:
- Short-lived (1 hour for users, 24 hours for system users)
- Included in every API request
- Contains user info and permissions

**Refresh Token**:
- Longer-lived
- Used to get new access tokens
- Stored securely

### Token Refresh

**Automatic** (with SuperTokens SDK):
- Frontend SDK automatically refreshes
- Happens before token expires
- Transparent to user

**Manual** (without SDK):
```bash
POST /api/auth/session/refresh
Authorization: Bearer <refresh-token>

# Returns new access token
```

### Session Revocation

**Sign Out**:
```javascript
await EmailPassword.signOut();
// All sessions for user revoked
```

**Revoke Specific Session**:
```bash
# Platform admin can revoke sessions
POST /api/v1/platform/users/:user_id/revoke-sessions
```

## Google OAuth

### Enabling Google Login

1. **Get Google OAuth Credentials**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create OAuth 2.0 credentials
   - Set redirect URI: `https://yourdomain.com/api/auth/callback/google`

2. **Configure Backend**:
   ```bash
   # .env
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

3. **Frontend Automatically Shows Google Button**:
   ```javascript
   // SuperTokens detects Google is enabled
   // Shows Google sign-in button automatically
   ```

### OAuth Flow

```
1. User clicks "Sign in with Google"
   ↓
2. Redirected to Google login
   ↓
3. User authorizes
   ↓
4. Google redirects back with code
   ↓
5. Backend exchanges code for tokens
   ↓
6. SuperTokens creates user (if new)
   ↓
7. Session created
   ↓
8. User authenticated
```

## Security Features

### HTTP-Only Cookies

**What**: Cookies that JavaScript cannot access

**Why**: Prevents XSS attacks

**How**: Set automatically by SuperTokens

### SameSite Cookies

**What**: Cookies only sent to same-site requests

**Why**: Prevents CSRF attacks

**Value**: `Lax` (balance of security and usability)

### Secure Flag

**What**: Cookies only sent over HTTPS

**Why**: Prevents man-in-the-middle attacks

**When**: Automatic in production (HTTPS)

### Token Rotation

**What**: Refresh token changes on each refresh

**Why**: Limits impact of stolen tokens

**How**: Automatic with SuperTokens

## Configuration

### Backend Configuration

**File**: `cmd/api/main.go`

```go
func initSuperTokens(cfg *config.Config) error {
    recipeList := []supertokens.Recipe{
        emailpassword.Init(nil),
        usermetadata.Init(nil),
    }
    
    // Add Google OAuth if configured
    if cfg.IsGoogleOAuthEnabled() {
        recipeList = append(recipeList, thirdparty.Init(...))
    }
    
    // Session configuration
    recipeList = append(recipeList, session.Init(&sessmodels.TypeInput{
        CookieSameSite: ptrString("lax"),
        GetTokenTransferMethod: func(req *http.Request, ...) {
            // Support both cookie and header
            if req.Header.Get("st-auth-mode") == "header" {
                return sessmodels.HeaderTransferMethod
            }
            return sessmodels.CookieTransferMethod
        },
    }))
    
    return supertokens.Init(...)
}
```

### Frontend Configuration

**File**: `frontend/src/App.jsx`

```javascript
function initializeSuperTokens(config) {
  const recipeList = [];
  
  // Add Google OAuth if enabled
  if (config?.providers?.google) {
    recipeList.push(ThirdParty.init({
      signInAndUpFeature: {
        providers: [ThirdParty.Google.init()]
      }
    }));
  }
  
  recipeList.push(EmailPassword.init());
  recipeList.push(Session.init({
    sessionExpiredStatusCode: 401
  }));
  
  SuperTokens.init({
    appInfo: {
      apiDomain: window.location.origin,
      websiteDomain: window.location.origin,
      apiBasePath: "/api/auth",
      websiteBasePath: "/auth"
    },
    recipeList: recipeList
  });
}
```

## Common Patterns

### Protected Routes (Backend)

```go
// Require authentication
auth := v1.Group("")
auth.Use(middleware.AuthMiddleware())
{
    auth.GET("/tenants", handler.ListTenants)
}
```

### Protected Routes (Frontend)

```javascript
import { SessionAuth } from "supertokens-auth-react/recipe/session";

<Route
  path="/dashboard"
  element={
    <SessionAuth>
      <Dashboard />
    </SessionAuth>
  }
/>
```

### Getting Current User

**Backend**:
```go
func (h *Handler) GetCurrentUser(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)
    // Use userID...
}
```

**Frontend**:
```javascript
import Session from "supertokens-auth-react/recipe/session";

const userId = await Session.getUserId();
```

## Testing Authentication

### Test Sign Up

```bash
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "test@example.com"},
      {"id": "password", "value": "SecurePass123!"}
    ]
  }'
```

### Test Sign In

```bash
curl -X POST http://localhost:8080/api/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "test@example.com"},
      {"id": "password", "value": "SecurePass123!"}
    ]
  }' \
  -v  # See response headers with tokens
```

### Test Protected Endpoint

```bash
# Extract token from sign-in response headers
TOKEN="<access-token>"

curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "st-auth-mode: header"
```

## Troubleshooting

### "Session not found"

**Cause**: Cookies not being sent

**Solution**:
```javascript
// Always include credentials
fetch(url, { credentials: 'include' })
```

### CORS Errors

**Cause**: Frontend and backend on different origins

**Solution**: Configure CORS middleware with correct origins

### "Invalid token"

**Cause**: Token expired or verification failed

**Solution**: 
- Check token hasn't expired
- Verify SuperTokens configuration
- Ensure public key is correct

## Next Steps

- **[User Authentication Details](/guides/user-authentication)** - Deep dive into user auth
- **[System Users (M2M)](/guides/system-users)** - Service account authentication
- **[Session Management](/guides/session-management)** - Advanced session topics
- **[Middleware Examples](/middleware/overview)** - Implement auth middleware

