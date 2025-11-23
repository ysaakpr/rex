# Authentication Implementation Guide

**Date**: November 22, 2025  
**Purpose**: Complete guide for implementing SuperTokens authentication in frontend and backend applications

## Table of Contents

- [Overview](#overview)
- [Part 1: Frontend Authentication Setup](#part-1-frontend-authentication-setup)
  - [Installation](#installation)
  - [SuperTokens Configuration](#supertokens-configuration)
  - [Protected Routes](#protected-routes)
  - [Making API Calls](#making-api-calls)
  - [Session Management](#session-management)
- [Part 2: Backend Token Verification](#part-2-backend-token-verification)
  - [Stateless vs Stateful Overview](#stateless-vs-stateful-overview)
  - [Stateless Token Verification](#stateless-token-verification)
  - [Stateful Token Verification](#stateful-token-verification)
  - [Comparison and Trade-offs](#comparison-and-trade-offs)
  - [Middleware Implementation](#middleware-implementation)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

This guide covers two main topics:

1. **Frontend Authentication**: How to integrate SuperTokens in a React application
2. **Backend Token Verification**: Understanding and implementing stateless vs stateful token verification

### What You'll Learn

- ✅ Set up SuperTokens SDK in React
- ✅ Configure authentication flows
- ✅ Protect routes and make authenticated API calls
- ✅ Understand stateless vs stateful token verification
- ✅ Choose the right verification mode for your needs
- ✅ Implement backend middleware correctly

---

## Part 1: Frontend Authentication Setup

### Installation

#### React Application

Install SuperTokens React SDK and dependencies:

```bash
npm install supertokens-auth-react
npm install react-router-dom
```

**Dependencies**:
- `supertokens-auth-react`: SuperTokens React SDK
- `react-router-dom`: For routing and protected routes

### SuperTokens Configuration

#### 1. Initialize SuperTokens

Create or update your `src/main.jsx` (Vite) or `src/index.js` (Create React App):

```javascript
import React from 'react';
import ReactDOM from 'react-dom/client';
import SuperTokens, { SuperTokensWrapper } from 'supertokens-auth-react';
import EmailPassword from 'supertokens-auth-react/recipe/emailpassword';
import Session from 'supertokens-auth-react/recipe/session';
import App from './App';

// Initialize SuperTokens
SuperTokens.init({
  appInfo: {
    appName: 'Your App Name',
    apiDomain: 'http://localhost:3000',      // Your API domain (through proxy in dev)
    websiteDomain: 'http://localhost:3000',  // Your frontend domain
    apiBasePath: '/auth',                    // SuperTokens auth endpoints
    websiteBasePath: '/auth'                 // Frontend auth pages
  },
  recipeList: [
    EmailPassword.init({
      // Customize sign-in/sign-up forms if needed
      signInAndUpFeature: {
        signUpForm: {
          formFields: [
            {
              id: 'email',
              label: 'Email',
              placeholder: 'Enter your email'
            },
            {
              id: 'password',
              label: 'Password',
              placeholder: 'Enter your password'
            }
          ]
        }
      }
    }),
    Session.init({
      // Session configuration
      tokenTransferMethod: 'cookie', // or 'header' for mobile apps
    })
  ]
});

// Render app
ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <SuperTokensWrapper>
      <App />
    </SuperTokensWrapper>
  </React.StrictMode>
);
```

#### 2. Configuration Options Explained

**appInfo**:
```javascript
{
  appName: 'Your App Name',           // Display name for your app
  apiDomain: 'http://localhost:3000',  // Backend API domain (use proxy in dev)
  websiteDomain: 'http://localhost:3000', // Frontend domain
  apiBasePath: '/auth',                // Where auth APIs are mounted (backend)
  websiteBasePath: '/auth'             // Where auth UI is rendered (frontend)
}
```

**Key Points**:
- In **development**: Use same domain for both `apiDomain` and `websiteDomain` if using Vite proxy
- In **production**: Set to actual domains (can be different for CORS setups)
- `apiBasePath`: Must match your backend SuperTokens configuration
- `websiteBasePath`: Where `/auth/signin`, `/auth/signup` will be rendered

**Session.init options**:
```javascript
Session.init({
  tokenTransferMethod: 'cookie',  // 'cookie' for web, 'header' for mobile
  // Optional: customize session timeout
  sessionTokenFrontendDomain: 'localhost', // Cookie domain (usually auto-detected)
})
```

### Protected Routes

#### 3. Set Up Routing with Protected Routes

Update your `src/App.jsx`:

```javascript
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import SuperTokens, { getSuperTokensRoutesForReactRouterDom } from 'supertokens-auth-react';
import { SessionAuth } from 'supertokens-auth-react/recipe/session';
import { EmailPasswordPreBuiltUI } from 'supertokens-auth-react/recipe/emailpassword/prebuiltui';

// Your components
import Dashboard from './components/Dashboard';
import Profile from './components/Profile';
import PublicPage from './components/PublicPage';

function App() {
  return (
    <Router>
      <Routes>
        {/* SuperTokens auth routes (sign in, sign up, etc.) */}
        {getSuperTokensRoutesForReactRouterDom(require('react-router-dom'), [EmailPasswordPreBuiltUI])}
        
        {/* Public routes */}
        <Route path="/public" element={<PublicPage />} />
        
        {/* Protected routes */}
        <Route
          path="/"
          element={
            <SessionAuth>
              <Dashboard />
            </SessionAuth>
          }
        />
        
        <Route
          path="/profile"
          element={
            <SessionAuth>
              <Profile />
            </SessionAuth>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
```

#### 4. Understanding SessionAuth

**`<SessionAuth>`** is a wrapper component that:
- ✅ Checks if user has a valid session
- ✅ Redirects to `/auth` (sign-in page) if not authenticated
- ✅ Renders child components if authenticated
- ✅ Automatically handles token refresh

**Usage**:
```javascript
<SessionAuth>
  <YourProtectedComponent />
</SessionAuth>
```

**Custom redirect**:
```javascript
<SessionAuth redirectToLogin={() => window.location.href = '/login'}>
  <YourProtectedComponent />
</SessionAuth>
```

**Access session info in components**:
```javascript
import { useSessionContext } from 'supertokens-auth-react/recipe/session';

function Dashboard() {
  const session = useSessionContext();
  
  if (session.loading) {
    return <div>Loading...</div>;
  }
  
  const userId = session.userId;
  const accessTokenPayload = session.accessTokenPayload;
  
  return (
    <div>
      <h1>Welcome, User {userId}</h1>
    </div>
  );
}
```

### Making API Calls

#### 5. API Calls with Authentication

**Important**: Always include `credentials: 'include'` in fetch calls!

```javascript
// Good: Includes cookies automatically
async function fetchTenants() {
  const response = await fetch('/api/v1/tenants', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include'  // ⭐ CRITICAL: Send cookies
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch tenants');
  }
  
  return await response.json();
}

// Good: POST request with body
async function createTenant(data) {
  const response = await fetch('/api/v1/tenants', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include',  // ⭐ CRITICAL: Send cookies
    body: JSON.stringify(data)
  });
  
  if (!response.ok) {
    throw new Error('Failed to create tenant');
  }
  
  return await response.json();
}
```

**Why `credentials: 'include'` is required**:
- Without it, cookies (including session tokens) are NOT sent
- Backend will see the request as unauthenticated
- You'll get `401 Unauthorized` errors

#### 6. Best Practice: Always Use `credentials: 'include'`

When making API calls from your React components, **always include `credentials: 'include'`** in your fetch options. This ensures that session cookies are sent with every request.

```javascript
async function fetchTenants() {
  const response = await fetch('/api/v1/tenants', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include'  // ⭐ REQUIRED: Send session cookies
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch tenants');
  }
  
  return await response.json();
}
```

**Why `credentials: 'include'` is critical**:
- ✅ Sends session cookies with every request
- ✅ Backend can verify authentication
- ✅ Works with SuperTokens cookie-based auth
- ✅ Without it, you'll get 401 Unauthorized errors

**Common mistake**:
```javascript
// ❌ BAD: Missing credentials
const response = await fetch('/api/v1/tenants');
// Result: 401 Unauthorized (cookies not sent!)

// ✅ GOOD: Include credentials
const response = await fetch('/api/v1/tenants', {
  credentials: 'include'
});
// Result: 200 OK (cookies sent, authenticated!)
```

### Session Management

#### 7. Sign Out

```javascript
import { signOut } from 'supertokens-auth-react/recipe/session';

async function handleSignOut() {
  await signOut();
  window.location.href = '/auth'; // Redirect to login page
}
```

#### 8. Check Authentication Status

```javascript
import { useSessionContext } from 'supertokens-auth-react/recipe/session';

function Header() {
  const session = useSessionContext();
  
  if (session.loading) {
    return <div>Loading...</div>;
  }
  
  if (session.doesSessionExist) {
    return (
      <div>
        <span>Logged in as: {session.userId}</span>
        <button onClick={handleSignOut}>Sign Out</button>
      </div>
    );
  }
  
  return <a href="/auth">Sign In</a>;
}
```

#### 9. Handling Session Expiry

SuperTokens automatically handles session refresh when using `<SessionAuth>` wrapper:
- Detects when access token expires
- Calls `/auth/session/refresh` with refresh token
- Gets new access token
- Updates cookies automatically

**What you need to do:**
- Wrap protected routes with `<SessionAuth>` 
- Use `credentials: 'include'` in all fetch calls
- SuperTokens handles the rest automatically

**Important Note**: If a fetch request fails with 401 during token refresh, the user will be redirected to the login page automatically by `<SessionAuth>`.

#### 10. Custom Session Events

Listen for session events:

```javascript
import Session from 'supertokens-auth-react/recipe/session';

// Listen for session events
Session.addEventListener((context) => {
  if (context.action === 'SESSION_CREATED') {
    console.log('User signed in!');
  }
  
  if (context.action === 'SIGN_OUT') {
    console.log('User signed out!');
  }
  
  if (context.action === 'REFRESH_SESSION') {
    console.log('Session refreshed!');
  }
  
  if (context.action === 'UNAUTHORISED') {
    console.log('Session expired, redirecting to login...');
  }
});
```

### Complete Frontend Example

Here's a complete, working React component:

```javascript
// src/components/Dashboard.jsx
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSessionContext, signOut } from 'supertokens-auth-react/recipe/session';

function Dashboard() {
  const navigate = useNavigate();
  const session = useSessionContext();
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  useEffect(() => {
    loadTenants();
  }, []);
  
  async function loadTenants() {
    try {
      setLoading(true);
      // Always include credentials: 'include' to send session cookies
      const response = await fetch('/api/v1/tenants', {
        credentials: 'include'  // ⭐ CRITICAL
      });
      
      if (!response.ok) {
        throw new Error('Failed to load tenants');
      }
      
      const data = await response.json();
      setTenants(data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }
  
  async function handleSignOut() {
    await signOut();
    navigate('/auth');
  }
  
  if (session.loading || loading) {
    return <div>Loading...</div>;
  }
  
  return (
    <div>
      <header>
        <h1>Dashboard</h1>
        <div>
          <span>User ID: {session.userId}</span>
          <button onClick={handleSignOut}>Sign Out</button>
        </div>
      </header>
      
      {error && <div className="error">{error}</div>}
      
      <section>
        <h2>Your Tenants</h2>
        {tenants.length === 0 ? (
          <p>No tenants found. Create one to get started!</p>
        ) : (
          <ul>
            {tenants.map(tenant => (
              <li key={tenant.id}>
                <strong>{tenant.name}</strong> - {tenant.slug}
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

export default Dashboard;
```

### Frontend Summary

**Checklist**:
- ✅ Install `supertokens-auth-react`
- ✅ Initialize SuperTokens in `main.jsx` with correct domains
- ✅ Wrap app in `<SuperTokensWrapper>`
- ✅ Add SuperTokens routes for auth UI
- ✅ Use `<SessionAuth>` to protect routes
- ✅ **Always use `credentials: 'include'` in all fetch calls**
- ✅ Use `useSessionContext()` to access session info
- ✅ Use `signOut()` for logout

---

## Part 2: Backend Token Verification

### Stateless vs Stateful Overview

SuperTokens supports two modes for access token verification:

| Feature | Stateless | Stateful |
|---------|-----------|----------|
| **Verification method** | JWT signature validation only | Database lookup + signature |
| **Speed** | Very fast (~0.1ms) | Slower (~10-50ms) |
| **Database calls** | None | One per request |
| **Token revocation** | Not possible | Immediate |
| **Logout** | Client-side only | Server-enforced |
| **Security** | Good | Better |
| **Scalability** | Excellent | Good |
| **Use case** | High-traffic APIs | High-security apps |

### Stateless Token Verification

#### What is Stateless?

In **stateless mode**:
- Access token is a **JWT (JSON Web Token)**
- Backend validates token by checking:
  - ✅ Signature (using public key)
  - ✅ Expiry timestamp
  - ✅ Issuer, audience, etc.
- **NO database call** to verify session exists

**How it works**:
```
1. Client sends request with access token
2. Backend extracts token from cookie/header
3. Backend validates JWT signature (local, fast)
4. Backend checks expiry (local, fast)
5. If valid → Grant access
6. If invalid → Return 401
```

#### Advantages of Stateless

✅ **Performance**: No database lookup, very fast  
✅ **Scalability**: No DB bottleneck, infinite horizontal scaling  
✅ **Simplicity**: No session storage needed  
✅ **Offline-capable**: Can verify tokens without network access  

#### Disadvantages of Stateless

❌ **Cannot revoke tokens**: Once issued, valid until expiry  
❌ **No instant logout**: Token remains valid even after logout (until expiry)  
❌ **Longer-lived tokens**: Need to balance security vs UX  
❌ **Token size**: JWTs are larger than session IDs  

#### When to Use Stateless

Use stateless verification when:
- ✅ High traffic (1000s of requests/second)
- ✅ Need to scale horizontally
- ✅ Acceptable for tokens to be valid for 1 hour
- ✅ Don't need instant token revocation
- ✅ Cost-sensitive (fewer DB queries = lower cost)

**Example use cases**:
- Public APIs
- SaaS applications
- Mobile apps
- Microservices

#### Configuring Stateless (Go Backend)

In your `cmd/api/main.go`:

```go
import (
    "github.com/supertokens/supertokens-golang/recipe/session"
    "github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
)

func initSuperTokens() error {
    return supertokens.Init(supertokens.TypeInput{
        // ... app info ...
        RecipeList: []supertokens.Recipe{
            session.Init(&sessmodels.TypeInput{
                // Stateless mode: Use JWT access tokens
                GetTokenTransferMethod: func(req *http.Request, forCreateNewSession bool, userContext supertokens.UserContext) sessmodels.TokenTransferMethod {
                    return sessmodels.CookieTransferMethod // or HeaderTransferMethod
                },
                
                // Access token validation: Stateless (no DB check)
                // This is the DEFAULT behavior - no special config needed
                
                // Optional: Configure token expiry
                AccessTokenValidity: ptrUint64(3600), // 1 hour (in seconds)
            }),
        },
    })
}
```

**Key points**:
- **Stateless is the default** in SuperTokens
- Access tokens are JWTs signed with RSA keys
- Backend validates signature locally (no DB call)
- Short-lived tokens (1 hour default) balance security vs UX

#### Middleware for Stateless (Go)

```go
// internal/api/middleware/auth.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/supertokens/supertokens-golang/recipe/session"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verify session (stateless: validates JWT signature only)
        session.VerifySession(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Session valid! Extract user ID
            sessionContainer := session.GetSessionFromRequestContext(r.Context())
            userID := sessionContainer.GetUserID()
            
            // Add to Gin context for use in handlers
            c.Set("userID", userID)
            c.Request = c.Request.WithContext(r.Context())
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    }
}
```

**What happens in stateless mode**:
1. `session.VerifySession()` extracts access token
2. Validates JWT signature using public key
3. Checks expiry timestamp
4. **NO database call**
5. If valid, calls the success handler
6. If invalid, returns 401

### Stateful Token Verification

#### What is Stateful?

In **stateful mode**:
- Backend validates token AND checks database
- Every request → Database lookup to verify session exists
- Enables instant token revocation and logout

**How it works**:
```
1. Client sends request with access token
2. Backend extracts token from cookie/header
3. Backend validates JWT signature (local, fast)
4. Backend checks expiry (local, fast)
5. Backend queries database: "Does this session exist?" (slow)
6. If session exists in DB → Grant access
7. If not in DB → Return 401 (revoked/logged out)
```

#### Advantages of Stateful

✅ **Instant token revocation**: Delete session from DB, token invalid immediately  
✅ **Server-enforced logout**: Backend can force logout  
✅ **Audit trail**: Track all active sessions  
✅ **Better security**: Can revoke compromised tokens  
✅ **User management**: Admin can terminate user sessions  

#### Disadvantages of Stateful

❌ **Performance overhead**: Database query on every request  
❌ **Scalability**: Database can become bottleneck  
❌ **Complexity**: Need to manage session storage  
❌ **Higher cost**: More database queries  

#### When to Use Stateful

Use stateful verification when:
- ✅ Security is paramount (banking, healthcare)
- ✅ Need instant logout/revocation
- ✅ Admin needs to terminate user sessions
- ✅ Compliance requires audit trails
- ✅ Moderate traffic (can handle DB queries)

**Example use cases**:
- Banking apps
- Healthcare systems
- Admin dashboards
- Enterprise applications

#### Configuring Stateful (Go Backend)

In your `cmd/api/main.go`:

```go
import (
    "github.com/supertokens/supertokens-golang/recipe/session"
    "github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
)

func initSuperTokens() error {
    return supertokens.Init(supertokens.TypeInput{
        // ... app info ...
        RecipeList: []supertokens.Recipe{
            session.Init(&sessmodels.TypeInput{
                // Enable stateful verification: check DB on every request
                UseDynamicAccessTokenSigningKey: ptrBool(true),
                
                // This forces SuperTokens to check the database
                // for session validity on each request
                
                // Optional: Configure token expiry (can be longer with stateful)
                AccessTokenValidity: ptrUint64(86400), // 24 hours (safer with DB checks)
            }),
        },
    })
}
```

**Key points**:
- `UseDynamicAccessTokenSigningKey: true` enables stateful mode
- Every `VerifySession()` call queries the database
- Can use longer-lived tokens (safer with DB verification)
- Session can be revoked by deleting from database

#### Middleware for Stateful (Same as Stateless!)

```go
// internal/api/middleware/auth.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/supertokens/supertokens-golang/recipe/session"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verify session (stateful: validates JWT + checks DB)
        session.VerifySession(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Session valid! Extract user ID
            sessionContainer := session.GetSessionFromRequestContext(r.Context())
            userID := sessionContainer.GetUserID()
            
            // Add to Gin context for use in handlers
            c.Set("userID", userID)
            c.Request = c.Request.WithContext(r.Context())
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    }
}
```

**What happens in stateful mode**:
1. `session.VerifySession()` extracts access token
2. Validates JWT signature using public key
3. Checks expiry timestamp
4. **Queries database** for session record
5. If session exists in DB, calls success handler
6. If not in DB (revoked), returns 401

**Note**: The middleware code is identical! SuperTokens handles stateful vs stateless internally based on configuration.

#### Revoking Sessions (Stateful Only)

```go
// internal/api/handlers/admin_handler.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/supertokens/supertokens-golang/recipe/session"
)

// Revoke all sessions for a user
func (h *AdminHandler) RevokeUserSessions(c *gin.Context) {
    userID := c.Param("user_id")
    
    // Revoke all sessions for this user
    err := session.RevokeAllSessionsForUser(userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to revoke sessions"})
        return
    }
    
    c.JSON(200, gin.H{"message": "All sessions revoked"})
}

// Revoke a specific session
func (h *AdminHandler) RevokeSession(c *gin.Context) {
    sessionHandle := c.Param("session_handle")
    
    // Revoke specific session
    err := session.RevokeSession(sessionHandle)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to revoke session"})
        return
    }
    
    c.JSON(200, gin.H{"message": "Session revoked"})
}
```

**Use cases for revocation**:
- Admin terminates user sessions
- User changes password (revoke all old sessions)
- Suspicious activity detected
- User clicks "Sign out of all devices"

### Comparison and Trade-offs

#### Performance Impact

**Benchmark example** (approximate):

| Operation | Stateless | Stateful | Difference |
|-----------|-----------|----------|------------|
| Token verification | 0.1ms | 10-50ms | 100-500x slower |
| Throughput | 10,000 req/s | 200-1000 req/s | 10-50x lower |
| Database load | 0 queries | 1 query/request | Significant |

**Real-world example**:
- **1000 requests/second** = 86,400,000 requests/day
- **Stateless**: 0 DB queries
- **Stateful**: 86,400,000 DB queries/day

#### Security Comparison

| Security Feature | Stateless | Stateful |
|------------------|-----------|----------|
| Token revocation | ❌ No (valid until expiry) | ✅ Yes (instant) |
| Forced logout | ❌ Client-side only | ✅ Server-enforced |
| Compromised token | ⚠️ Valid for hours | ✅ Can revoke immediately |
| Session audit | ❌ Not possible | ✅ Full audit trail |
| Multi-device logout | ❌ Manual per device | ✅ Server-side bulk revoke |

#### Cost Comparison

**Assume**:
- 1 million requests/month
- Database query cost: $0.001 per 1000 queries
- Load balancer cost: $0.10 per GB

**Stateless**:
- Database cost: $0 (no queries)
- Load balancer cost: $X (baseline)
- **Total**: $X

**Stateful**:
- Database cost: $1000 (1 million queries)
- Load balancer cost: $X (same)
- **Total**: $X + $1000

**Cost savings with stateless**: Significant at scale!

#### Decision Matrix

| Your Priority | Recommended Mode |
|---------------|------------------|
| **Maximum performance** | Stateless |
| **Maximum security** | Stateful |
| **High traffic (>1000 req/s)** | Stateless |
| **Low traffic (<100 req/s)** | Either (stateful for security) |
| **Banking/healthcare** | Stateful |
| **SaaS/API platform** | Stateless |
| **Need instant logout** | Stateful |
| **Cost-sensitive** | Stateless |
| **Compliance requirements** | Stateful (audit trail) |
| **Microservices** | Stateless |

### Hybrid Approach

You can also implement a **hybrid approach**:

```go
// Use stateless for most endpoints
func AuthMiddleware() gin.HandlerFunc {
    return session.VerifySession(nil, ...)
}

// Use stateful for sensitive endpoints
func StatefulAuthMiddleware() gin.HandlerFunc {
    return session.VerifySession(&sessmodels.VerifySessionOptions{
        CheckDatabase: ptrBool(true), // Force DB check even in stateless mode
    }, ...)
}

// Routes
router.GET("/api/v1/tenants", AuthMiddleware(), tenantHandler.List)         // Stateless
router.POST("/api/v1/admin/revoke", StatefulAuthMiddleware(), adminHandler.Revoke) // Stateful
```

**Benefits**:
- ✅ Fast stateless for read operations
- ✅ Secure stateful for sensitive operations
- ✅ Best of both worlds

### Middleware Implementation

#### Complete Middleware Example (Current Implementation)

Here's how our middleware is implemented in `internal/api/middleware/auth.go`:

```go
package middleware

import (
    "bytes"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/supertokens/supertokens-golang/recipe/session"
    "github.com/vyshakhp/utm-backend/internal/pkg/response"
    "go.uber.org/zap"
)

type bodyLogWriter struct {
    gin.ResponseWriter
    body       *bytes.Buffer
    statusCode int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

// AuthMiddleware verifies SuperTokens session
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        fmt.Printf("[DEBUG] AuthMiddleware: Path=%s, Host=%s, Cookies=%d\n", 
            c.Request.URL.Path, c.Request.Host, len(c.Request.Cookies()))

        verificationSucceeded := false
        
        // Capture response from SuperTokens
        blw := &bodyLogWriter{body: new(bytes.Buffer), ResponseWriter: c.Writer}
        c.Writer = blw

        // Verify session (stateless by default, stateful if configured)
        session.VerifySession(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Printf("[DEBUG] Inside VerifySession handler - session verified!\n")
            verificationSucceeded = true
            c.Request = c.Request.WithContext(r.Context())
        })).ServeHTTP(blw, c.Request)

        if !verificationSucceeded {
            fmt.Printf("[DEBUG] Session verification failed\n")
            c.Writer = blw.ResponseWriter
            c.Writer.WriteHeader(blw.statusCode)
            c.Writer.Write(blw.body.Bytes())
            c.Abort()
            return
        }

        // Extract user ID from session
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        if sessionContainer == nil {
            fmt.Printf("[DEBUG] Session container is nil\n")
            response.Unauthorized(c, "Session not found")
            c.Abort()
            return
        }

        userID := sessionContainer.GetUserID()
        fmt.Printf("[DEBUG] Session verified successfully for user: %s\n", userID)
        c.Set("userID", userID)
        c.Next()
    }
}

// GetUserID retrieves user ID from Gin context
func GetUserID(c *gin.Context) (string, error) {
    userID, exists := c.Get("userID")
    if !exists {
        return "", fmt.Errorf("user ID not found in context")
    }
    
    userIDStr, ok := userID.(string)
    if !ok {
        return "", fmt.Errorf("user ID is not a string")
    }
    
    return userIDStr, nil
}
```

**How it works**:
1. Intercepts request before handler
2. Calls `session.VerifySession()` (stateless or stateful based on config)
3. If valid, extracts user ID and adds to context
4. If invalid, returns 401 Unauthorized
5. Handlers can access user ID via `middleware.GetUserID(c)`

#### Using the Middleware

```go
// internal/api/router/router.go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/vyshakhp/utm-backend/internal/api/handlers"
    "github.com/vyshakhp/utm-backend/internal/api/middleware"
)

func SetupRouter(deps *Dependencies) *gin.Engine {
    r := gin.Default()
    
    // Public routes (no auth)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Protected routes (require auth)
    api := r.Group("/api/v1")
    api.Use(middleware.AuthMiddleware()) // Apply to all routes in group
    {
        api.GET("/tenants", deps.TenantHandler.List)
        api.POST("/tenants", deps.TenantHandler.Create)
        api.GET("/tenants/:id", deps.TenantHandler.Get)
    }
    
    return r
}
```

---

## Best Practices

### Frontend

1. **Always use `credentials: 'include'`** in all fetch calls (CRITICAL)
2. **Wrap protected routes** with `<SessionAuth>`
3. **Handle loading states** (session verification takes time)
4. **Use SuperTokens' built-in UI** (saves development time)
5. **Test sign-in, sign-out, and session expiry**
6. **Never forget**: `fetch()` without `credentials: 'include'` = 401 errors

### Backend

1. **Choose stateless for most apps** (performance + scalability)
2. **Use stateful for high-security apps** (banking, healthcare)
3. **Consider hybrid approach** (stateless + selective stateful)
4. **Set appropriate token expiry** (1 hour for stateless, 24 hours for stateful)
5. **Monitor database load** if using stateful
6. **Implement proper logging** for security audits

### Security

1. **Always use HTTPS in production** (protects cookies and tokens)
2. **Set secure cookie flags** (`Secure`, `HttpOnly`, `SameSite`)
3. **Rotate signing keys** periodically (SuperTokens handles this)
4. **Implement rate limiting** (prevent brute force attacks)
5. **Monitor for suspicious activity** (multiple failed logins, etc.)

---

## Troubleshooting

### Frontend Issues

#### "Session not found" or 401 errors

**Cause**: Cookies not being sent

**Solution**:
```javascript
// ❌ WRONG: Missing credentials
fetch('/api/v1/tenants');
// Result: 401 Unauthorized

// ✅ CORRECT: Always include credentials
fetch('/api/v1/tenants', {
  credentials: 'include'  // ⭐ CRITICAL: Send cookies
});
// Result: 200 OK (authenticated)
```

#### "Redirect loop" after sign-in

**Cause**: Frontend and backend domain mismatch

**Solution**:
```javascript
// Make sure apiDomain matches your setup
SuperTokens.init({
  appInfo: {
    apiDomain: 'http://localhost:3000',  // Use proxy domain
    websiteDomain: 'http://localhost:3000'
  }
});
```

#### CORS errors

**Cause**: Backend not configured for CORS

**Solution** (Go backend):
```go
import "github.com/gin-contrib/cors"

router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "rid", "st-auth-mode"},
    AllowCredentials: true, // ⭐ Critical for cookies!
}))
```

### Backend Issues

#### Database connection errors (Stateful mode)

**Cause**: SuperTokens can't connect to database

**Solution**:
- Check SuperTokens connection URI
- Verify database is running
- Check network connectivity
- Review SuperTokens logs

#### Performance degradation (Stateful mode)

**Cause**: Too many database queries

**Solution**:
- Switch to stateless mode for high-traffic endpoints
- Add database indexes on session tables
- Use connection pooling
- Consider caching (Redis)

#### Sessions not being revoked (Stateless mode)

**Cause**: Stateless mode doesn't check database

**Solution**:
- Switch to stateful mode (`UseDynamicAccessTokenSigningKey: true`)
- Or accept that tokens are valid until expiry (1 hour)
- Use shorter-lived tokens if needed

---

## Summary

### Frontend Setup
1. Install SuperTokens React SDK
2. Initialize with correct domains
3. Wrap app in `<SuperTokensWrapper>`
4. Use `<SessionAuth>` for protected routes
5. **Always use `credentials: 'include'` in all fetch calls**

### Backend Token Verification

**Stateless Mode** (Default):
- ✅ Fast (no DB queries)
- ✅ Scalable
- ❌ No token revocation
- **Use for**: High-traffic APIs, SaaS platforms

**Stateful Mode**:
- ✅ Instant token revocation
- ✅ Server-enforced logout
- ❌ Slower (DB query per request)
- **Use for**: Banking, healthcare, admin tools

**Configuration**:
```go
// Stateless (default)
session.Init(&sessmodels.TypeInput{
    // No special config needed
})

// Stateful
session.Init(&sessmodels.TypeInput{
    UseDynamicAccessTokenSigningKey: ptrBool(true),
})
```

### Decision Guide

**Choose Stateless** if:
- High traffic (>1000 req/s)
- Cost-sensitive
- Need to scale horizontally
- Can accept 1-hour token validity

**Choose Stateful** if:
- Security is critical
- Need instant logout
- Admin must control sessions
- Moderate traffic (<100 req/s)

---

**Last Updated**: November 22, 2025  
**Version**: 1.0

