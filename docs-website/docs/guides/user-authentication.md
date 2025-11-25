# User Authentication

Deep dive into user authentication with SuperTokens.

## Overview

Rex uses SuperTokens for all user authentication, providing:
- Email/password authentication
- Google OAuth (optional)
- Session management
- Token refresh
- Password reset
- Email verification (optional)

## Sign Up Flow

### Frontend Implementation

SuperTokens provides pre-built UI components:

```javascript
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";
import ThirdParty from "supertokens-auth-react/recipe/thirdparty";

SuperTokens.init({
  appInfo: {
    appName: "Rex",
    apiDomain: window.location.origin,
    websiteDomain: window.location.origin,
    apiBasePath: "/api/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),  // Email/password
    ThirdParty.init({      // Google OAuth
      signInAndUpFeature: {
        providers: [ThirdParty.Google.init()]
      }
    }),
    Session.init()
  ]
});
```

### Sign Up Endpoint

**URL**: `POST /api/auth/signup`

**Request**:
```json
{
  "formFields": [
    {"id": "email", "value": "user@example.com"},
    {"id": "password", "value": "SecurePass123!"}
  ]
}
```

**Response Headers**:
```
st-access-token: eyJhbGc...
st-refresh-token: eyJhbGc...
Set-Cookie: sAccessToken=...; HttpOnly; Secure; SameSite=Lax
Set-Cookie: sRefreshToken=...; HttpOnly; Secure; SameSite=Lax
```

**Response Body**:
```json
{
  "status": "OK",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "timeJoined": 1732272000000
  }
}
```

## Sign In Flow

### Frontend

```javascript
// Sign in handled by SuperTokens UI
// User navigates to /auth
// Or use programmatic API:

import EmailPassword from "supertokens-auth-react/recipe/emailpassword";

async function signIn(email, password) {
  let response = await EmailPassword.signIn({
    formFields: [
      {id: "email", value: email},
      {id: "password", value: password}
    ]
  });
  
  if (response.status === "OK") {
    // Sign in successful
    window.location.href = "/dashboard";
  } else if (response.status === "WRONG_CREDENTIALS_ERROR") {
    window.alert("Invalid email or password");
  }
}
```

### Sign In Endpoint

**URL**: `POST /api/auth/signin`

**Request/Response**: Same format as sign up

## Session Management

### Cookie-Based Sessions

**Automatically Managed**:
- `sAccessToken` - Short-lived (1 hour)
- `sRefreshToken` - Long-lived (used to refresh access token)
- `sIdRefreshToken` - Anti-CSRF token

**Making API Calls**:
```javascript
// Cookies sent automatically
const response = await fetch('/api/v1/tenants', {
  credentials: 'include'  // Required!
});
```

### Header-Based Sessions

**For APIs and Mobile Apps**:

```javascript
// Get tokens from SuperTokens
const accessToken = await Session.getAccessToken();

// Use in requests
const response = await fetch('/api/v1/tenants', {
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'st-auth-mode': 'header'
  }
});
```

## Token Refresh

### Automatic (Recommended)

SuperTokens SDK handles refresh automatically:

```javascript
// No manual refresh needed!
// SDK refreshes token before it expires
const response = await fetch('/api/v1/tenants', {
  credentials: 'include'
});
```

### Manual Refresh

**Endpoint**: `POST /api/auth/session/refresh`

**Request**:
```
Cookie: sRefreshToken=...
```

**Response**:
```
st-access-token: (new token)
Set-Cookie: sAccessToken=... (new cookie)
```

## Sign Out

### Frontend

```javascript
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";

async function signOut() {
  await EmailPassword.signOut();
  window.location.href = "/auth";
}
```

### Endpoint

**URL**: `POST /api/auth/signout`

**Effect**: All session tokens invalidated

## Google OAuth

### Setup

1. **Get Credentials**: [Google Cloud Console](https://console.cloud.google.com/)
2. **Configure Backend**:
   ```bash
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```
3. **Frontend Auto-Detects**: Shows Google button automatically

### OAuth Flow

```
1. User clicks "Sign in with Google"
2. Redirected to Google
3. User authorizes
4. Google redirects back with code
5. Backend exchanges code for tokens
6. User account created (if new)
7. Session established
```

## Password Reset

### Request Reset

```javascript
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";

await EmailPassword.sendPasswordResetEmail({
  formFields: [{
    id: "email",
    value: "user@example.com"
  }]
});
```

### Reset Flow

```
1. User requests reset
2. Email sent with reset link
3. User clicks link
4. SuperTokens shows reset form
5. User enters new password
6. Password updated
```

## Protected Routes

### Backend

```go
// Require authentication
auth := router.Group("")
auth.Use(middleware.AuthMiddleware())
{
    auth.GET("/tenants", handler.ListTenants)
}
```

### Frontend

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

## Getting Current User

### Backend

```go
func GetCurrentUser(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)
    userEmail, _ := middleware.GetUserEmail(c)
    
    // Use userID and userEmail...
}
```

### Frontend

```javascript
import Session from "supertokens-auth-react/recipe/session";

const userId = await Session.getUserId();
const accessTokenPayload = await Session.getAccessTokenPayloadSecurely();
```

## Security

### Password Requirements

Enforced by SuperTokens:
- Minimum 8 characters
- Can be customized in SuperTokens config

### Session Security

- **HTTP-Only Cookies**: JavaScript can't access
- **Secure Flag**: HTTPS only in production
- **SameSite**: CSRF protection
- **Token Rotation**: Refresh tokens rotate

### Rate Limiting

Implement at infrastructure level (Nginx, API Gateway)

## Next Steps

- **[Session Management](/guides/session-management)** - Advanced session topics
- **[System Users](/guides/system-users)** - M2M authentication
- **[Frontend Integration](/frontend/react-setup)** - Complete frontend guide

