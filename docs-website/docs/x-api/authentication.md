# Authentication API

Complete API reference for SuperTokens authentication endpoints.

## Overview

This backend uses **SuperTokens** for authentication with the following features:
- Email/Password authentication
- Google OAuth
- Session management (cookie-based and header-based)
- Email verification
- Password reset

**SuperTokens Base URL**: `/auth`

**Backend Proxy**: All SuperTokens endpoints are proxied through the backend

:::info Frontend Integration
Use the SuperTokens React SDK for seamless integration. It handles all authentication flows automatically.
:::

## Authentication Recipes

### Email/Password Recipe

Standard email and password authentication.

**Enabled Features**:
- Sign up with email/password
- Sign in with email/password
- Email verification
- Password reset
- Session management

### Third-Party Recipe

Social authentication providers.

**Enabled Providers**:
- Google OAuth

## Sign Up

Create a new user account with email and password.

**Request**:
```http
POST /auth/signup
Content-Type: application/json
```

**Body**:
```json
{
  "formFields": [
    {
      "id": "email",
      "value": "user@example.com"
    },
    {
      "id": "password",
      "value": "SecurePassword123!"
    }
  ]
}
```

**Response** (200):
```json
{
  "status": "OK",
  "user": {
    "id": "auth0|user123",
    "email": "user@example.com",
    "timeJoined": 1700000000000
  }
}
```

**Error Response** (200):
```json
{
  "status": "FIELD_ERROR",
  "formFields": [
    {
      "id": "email",
      "error": "Email already exists"
    }
  ]
}
```

:::tip Frontend SDK
Use SuperTokens React SDK's `<EmailPasswordAuth>` component instead of calling this endpoint directly.
:::

## Sign In

Authenticate an existing user.

**Request**:
```http
POST /auth/signin
Content-Type: application/json
```

**Body**:
```json
{
  "formFields": [
    {
      "id": "email",
      "value": "user@example.com"
    },
    {
      "id": "password",
      "value": "SecurePassword123!"
    }
  ]
}
```

**Response** (200):
```json
{
  "status": "OK",
  "user": {
    "id": "auth0|user123",
    "email": "user@example.com",
    "timeJoined": 1700000000000
  }
}
```

**Sets Cookies**:
- `sAccessToken`: Access token (short-lived)
- `sRefreshToken`: Refresh token (long-lived)
- `sIdRefreshToken`: ID refresh token

**Error Responses**:

Wrong credentials:
```json
{
  "status": "WRONG_CREDENTIALS_ERROR"
}
```

Email not verified:
```json
{
  "status": "SIGN_IN_NOT_ALLOWED",
  "reason": "Email not verified"
}
```

## Sign Out

End the current session.

**Request**:
```http
POST /auth/signout
```

**Response** (200):
```json
{
  "status": "OK"
}
```

**Effect**:
- Clears all session cookies
- Invalidates session on backend
- User must sign in again

## Google OAuth

Authenticate using Google account.

### Initiate OAuth Flow

**Request**:
```http
GET /auth/authorisationurl?thirdPartyId=google
```

**Response** (200):
```json
{
  "status": "OK",
  "url": "https://accounts.google.com/o/oauth2/v2/auth?..."
}
```

**Next Steps**:
1. Redirect user to the returned URL
2. User authenticates with Google
3. Google redirects to callback URL
4. SuperTokens completes authentication
5. Session cookies are set

### Handle OAuth Callback

**Endpoint**: `/auth/callback/google`

**Handled automatically by SuperTokens backend**

:::tip Frontend SDK
Use `<ThirdPartyAuth providers={['google']}>` component. It handles the entire OAuth flow.
:::

## Session Management

### Get Session Info

Check if user has an active session.

**Request**:
```http
GET /auth/session/verify
Cookie: sAccessToken=...; sRefreshToken=...
```

**Response** (200):
```json
{
  "status": "OK",
  "sessionHandle": "session-handle-uuid",
  "userId": "auth0|user123",
  "userDataInJWT": {},
  "accessTokenPayload": {}
}
```

**Error Response** (401):
```json
{
  "status": "UNAUTHORISED"
}
```

### Refresh Session

Automatically refresh access token using refresh token.

**Request**:
```http
POST /auth/session/refresh
Cookie: sRefreshToken=...
```

**Response** (200):
```json
{
  "status": "OK"
}
```

**Effect**:
- Issues new access token
- Updates `sAccessToken` cookie
- Extends session lifetime

:::info Automatic Refresh
The SuperTokens frontend SDK automatically refreshes tokens when they expire.
:::

## Email Verification

### Send Verification Email

Send verification email to user's registered email.

**Request**:
```http
POST /auth/user/email/verify/token
Cookie: sAccessToken=...
```

**Response** (200):
```json
{
  "status": "OK"
}
```

**Error** (if already verified):
```json
{
  "status": "EMAIL_ALREADY_VERIFIED_ERROR"
}
```

### Verify Email Token

Verify email using the token from the verification email.

**Request**:
```http
POST /auth/user/email/verify
Content-Type: application/json
```

**Body**:
```json
{
  "method": "token",
  "token": "verification-token-from-email"
}
```

**Response** (200):
```json
{
  "status": "OK",
  "user": {
    "id": "auth0|user123",
    "email": "user@example.com"
  }
}
```

**Error**:
```json
{
  "status": "EMAIL_VERIFICATION_INVALID_TOKEN_ERROR"
}
```

### Check Verification Status

Check if current user's email is verified.

**Request**:
```http
GET /auth/user/email/verify
Cookie: sAccessToken=...
```

**Response** (200):
```json
{
  "status": "OK",
  "isVerified": true
}
```

## Password Reset

### Generate Reset Token

Request password reset email.

**Request**:
```http
POST /auth/user/password/reset/token
Content-Type: application/json
```

**Body**:
```json
{
  "formFields": [
    {
      "id": "email",
      "value": "user@example.com"
    }
  ]
}
```

**Response** (200):
```json
{
  "status": "OK"
}
```

:::info Email Sent
Always returns "OK" even if email doesn't exist (security best practice).
:::

### Reset Password

Reset password using token from email.

**Request**:
```http
POST /auth/user/password/reset
Content-Type: application/json
```

**Body**:
```json
{
  "formFields": [
    {
      "id": "password",
      "value": "NewSecurePassword123!"
    }
  ],
  "token": "reset-token-from-email"
}
```

**Response** (200):
```json
{
  "status": "OK"
}
```

**Error**:
```json
{
  "status": "RESET_PASSWORD_INVALID_TOKEN_ERROR"
}
```

## Frontend Integration Examples

### React: Complete Auth Setup

```jsx
// src/main.jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import SuperTokens from 'supertokens-auth-react';
import EmailPassword from 'supertokens-auth-react/recipe/emailpassword';
import ThirdParty from 'supertokens-auth-react/recipe/thirdparty';
import Session from 'supertokens-auth-react/recipe/session';
import App from './App';

SuperTokens.init({
  appInfo: {
    appName: "Rex",
    apiDomain: "http://localhost:8080",
    websiteDomain: "http://localhost:3000",
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    ThirdParty.init({
      signInAndUpFeature: {
        providers: [ThirdParty.Google.init()]
      }
    }),
    Session.init()
  ]
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

### React: Protected Route

```jsx
import {SessionAuth} from 'supertokens-auth-react/recipe/session';

function ProtectedRoute() {
  return (
    <SessionAuth>
      <YourProtectedComponent />
    </SessionAuth>
  );
}

// With redirect
function ProtectedRouteWithRedirect() {
  return (
    <SessionAuth
      redirectToLogin={() => window.location.href = '/auth'}
    >
      <YourProtectedComponent />
    </SessionAuth>
  );
}
```

### React: Get Session in Component

```jsx
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

function UserProfile() {
  const session = useSessionContext();
  
  if (session.loading) {
    return <div>Loading...</div>;
  }
  
  if (!session.doesSessionExist) {
    return <div>Not logged in</div>;
  }
  
  return (
    <div>
      <p>User ID: {session.userId}</p>
      <p>Access Token: {session.accessTokenPayload}</p>
    </div>
  );
}
```

### React: Sign Out Button

```jsx
import {signOut} from 'supertokens-auth-react/recipe/session';

function SignOutButton() {
  const handleSignOut = async () => {
    await signOut();
    window.location.href = '/';
  };
  
  return <button onClick={handleSignOut}>Sign Out</button>;
}
```

### JavaScript: Making Authenticated Requests

```javascript
// Always include credentials: 'include' to send cookies
async function callProtectedAPI() {
  const response = await fetch('/api/v1/tenants', {
    method: 'GET',
    credentials: 'include',  // ← Critical for cookies
    headers: {
      'Content-Type': 'application/json'
    }
  });
  
  if (response.status === 401) {
    // Session expired, redirect to login
    window.location.href = '/auth';
    return;
  }
  
  return response.json();
}

// POST request example
async function createTenant(data) {
  const response = await fetch('/api/v1/tenants', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  
  return response.json();
}
```

## Backend Integration

### Go: Session Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/supertokens/supertokens-golang/recipe/session"
)

// Require authentication
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        session.VerifySession(nil)(c.Writer, c.Request, func(w http.ResponseWriter, r *http.Request) {
            c.Request = c.Request.WithContext(r.Context())
            c.Next()
        })
    }
}

// Get user ID from session
func GetUserIDFromContext(c *gin.Context) string {
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    return sessionContainer.GetUserID()
}
```

### Go: Using Session in Handler

```go
func GetCurrentUserHandler(c *gin.Context) {
    // Get session (already verified by middleware)
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    userID := sessionContainer.GetUserID()
    
    // Fetch user details from SuperTokens
    userInfo, err := emailpassword.GetUserByID(userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get user"})
        return
    }
    
    c.JSON(200, gin.H{
        "success": true,
        "data": gin.H{
            "id": userInfo.ID,
            "email": userInfo.Email,
            "time_joined": userInfo.TimeJoined,
        },
    })
}
```

## Cookie Configuration

### Development (localhost)

```go
supertokens.Init(supertokens.TypeInput{
    Supertokens: &supertokens.ConnectionInfo{
        ConnectionURI: os.Getenv("SUPERTOKENS_CONNECTION_URI"),
    },
    AppInfo: supertokens.AppInfo{
        AppName:       "Rex",
        APIDomain:     "http://localhost:8080",
        WebsiteDomain: "http://localhost:3000",
    },
    RecipeList: []supertokens.Recipe{
        session.Init(&sessmodels.TypeInput{
            CookieSecure: ptrBool(false), // HTTP allowed for localhost
            CookieSameSite: ptrString("lax"),
            CookieDomain: ptrString("localhost"),
        }),
    },
})
```

### Production (HTTPS)

```go
session.Init(&sessmodels.TypeInput{
    CookieSecure: ptrBool(true), // HTTPS only
    CookieSameSite: ptrString("lax"),
    CookieDomain: ptrString(".yourdomain.com"), // Subdomain support
    SessionExpiredStatusCode: ptrInt(401),
})
```

## Security Best Practices

### Cookie Security

✅ **Production Settings**:
- `secure: true` (HTTPS only)
- `httpOnly: true` (JavaScript can't access)
- `sameSite: lax` (CSRF protection)
- `domain: .yourdomain.com` (Subdomain sharing)

### Password Requirements

Default SuperTokens validation:
- Minimum 8 characters
- No maximum length (within reason)

**Custom validation** (add in frontend):
```javascript
EmailPassword.init({
  signUpFeature: {
    formFields: [
      {
        id: "password",
        label: "Password",
        validate: async (value) => {
          if (value.length < 12) {
            return "Password must be at least 12 characters";
          }
          if (!/[A-Z]/.test(value)) {
            return "Password must contain uppercase letter";
          }
          if (!/[a-z]/.test(value)) {
            return "Password must contain lowercase letter";
          }
          if (!/[0-9]/.test(value)) {
            return "Password must contain number";
          }
          if (!/[^A-Za-z0-9]/.test(value)) {
            return "Password must contain special character";
          }
          return undefined; // Valid
        }
      }
    ]
  }
})
```

### Email Verification

**Enforce verification** before tenant access:
```go
func RequireEmailVerification() gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        user, _ := emailpassword.GetUserByID(userID)
        if !user.Email.IsVerified {
            c.JSON(403, gin.H{
                "success": false,
                "error": "Email verification required",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Rate Limiting

Protect authentication endpoints:
- Sign up: 5 attempts per IP per hour
- Sign in: 10 attempts per IP per 15 minutes
- Password reset: 3 requests per email per hour

## Troubleshooting

### Cookies Not Set

**Problem**: Session cookies not appearing in browser

**Solutions**:
- Check `credentials: 'include'` in fetch calls
- Verify CORS allows credentials
- Ensure frontend and backend on same domain (or proper CORS setup)
- For localhost: Use `localhost` (not `127.0.0.1`)

### 401 Unauthorized

**Problem**: Protected endpoints return 401

**Solutions**:
- Check session cookie is present
- Verify token hasn't expired (check `sAccessToken`)
- Call `/auth/session/refresh` to refresh token
- Sign in again if refresh fails

### CORS Errors

**Problem**: CORS policy blocking requests

**Solution** (backend):
```go
import "github.com/gin-contrib/cors"

router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"content-type"},
    AllowCredentials: true, // ← Critical for cookies
}))
```

## Related Resources

- [SuperTokens Documentation](https://supertokens.com/docs/guides)
- [User Authentication Guide](/guides/user-authentication)
- [Frontend Integration](/guides/frontend-integration)
- [Session Management](/guides/session-management)
