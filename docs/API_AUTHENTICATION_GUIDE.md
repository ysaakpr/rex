# API Authentication Guide - Using curl

**Date**: November 21, 2025  
**Purpose**: Complete guide for authenticating and using the UTM Backend APIs with curl

## Table of Contents

- [Overview](#overview)
- [Authentication Methods](#authentication-methods)
- [Authentication Flow](#authentication-flow)
- [Sign Up](#sign-up)
- [Sign In](#sign-in)
- [Method 1: Cookie-Based Authentication](#method-1-cookie-based-authentication-recommended)
- [Method 2: Header-Based Authentication (Bearer Token)](#method-2-header-based-authentication-bearer-token)
- [Understanding SuperTokens Tokens](#understanding-supertokens-tokens)
- [Testing Protected APIs](#testing-protected-apis)
- [Complete Examples](#complete-examples)

## Overview

The UTM Backend uses **SuperTokens** for authentication and supports **two authentication methods**:
1. **Cookie-based** (recommended for browsers)
2. **Header-based with Bearer tokens** (recommended for API clients, mobile apps, postman)

This guide shows you how to:
1. Create an account (sign up)
2. Sign in and get authentication credentials
3. Use cookies OR bearer tokens to call protected APIs

### Important Notes

- **Two auth modes**: Cookie-based or header-based (Bearer token)
- **Base URL**: `http://localhost:3000` (goes through Vite proxy in dev)
- **Session expiry**: Tokens expire after 1 hour (configurable)
- **Choose based on use case**: 
  - Cookies → Web browsers, frontend apps
  - Headers → API clients, mobile apps, CI/CD, testing

## Authentication Methods

### Cookie-Based (Default)
- **Best for**: Web browsers, frontend applications
- **Pros**: Automatic cookie management, more secure against XSS
- **Cons**: CSRF protection needed, not ideal for API clients
- **Usage**: Save cookies with `-c` flag, send with `-b` flag

### Header-Based (Bearer Token)
- **Best for**: API clients, mobile apps, Postman, automated testing
- **Pros**: No cookie management, works everywhere, easier for non-browser clients
- **Cons**: Need to manually manage tokens
- **Usage**: Extract access token, send in `Authorization: Bearer <token>` header

## Authentication Flow

```
┌──────────┐      1. Sign Up       ┌──────────────┐
│  Client  │──────────────────────►│ SuperTokens  │
│  (curl)  │                       │     /auth    │
└──────────┘                       └──────────────┘
     │                                     │
     │         2. Get Cookies              │
     │◄────────────────────────────────────┘
     │
     │         3. Call Protected APIs
     │         (send cookies)
     ▼
┌──────────────┐
│  Backend API │
│  /api/v1/*   │
└──────────────┘
```

## Sign Up

### Create a New Account

```bash
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{
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
  }'
```

**Parameters:**
- `-H "rid: emailpassword"` - Recipe ID for email/password auth
- `-c cookies.txt` - Save cookies to file
- `formFields` - Array of email and password

**Response (Success):**
```json
{
  "status": "OK",
  "user": {
    "id": "04413f25-fdfa-42a0-a046-c3ad67d135fe",
    "email": "user@example.com",
    "timeJoined": 1732217234567
  }
}
```

**Response (Email Already Exists):**
```json
{
  "status": "FIELD_ERROR",
  "formFields": [
    {
      "id": "email",
      "error": "This email already exists. Please sign in instead."
    }
  ]
}
```

## Sign In

### Authenticate and Get Session Cookies

```bash
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{
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
  }'
```

**Response (Success):**
```json
{
  "status": "OK",
  "user": {
    "id": "04413f25-fdfa-42a0-a046-c3ad67d135fe",
    "email": "user@example.com",
    "timeJoined": 1732217234567
  }
}
```

**Cookies Received:**
After successful sign in, you'll receive these cookies (saved in `cookies.txt`):
- `sAccessToken` - Access token (short-lived)
- `sRefreshToken` - Refresh token (long-lived)
- `sFrontToken` - Frontend token metadata
- `st-last-access-token-update` - Timestamp

**Response (Invalid Credentials):**
```json
{
  "status": "WRONG_CREDENTIALS_ERROR"
}
```

## Method 1: Cookie-Based Authentication (Recommended)

### How It Works

1. Sign in with SuperTokens
2. Receive session cookies (`sAccessToken`, `sRefreshToken`, etc.)
3. Browser/curl automatically sends cookies with each request
4. Backend validates cookies and grants access

### Using Cookie File (Recommended for curl)

Save cookies during sign in with `-c cookies.txt`, then use them with `-b cookies.txt`:

```bash
# Sign in and save cookies
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{ ... }'

# Use saved cookies for API calls
curl -X GET http://localhost:3000/api/v1/tenants \
  -b cookies.txt \
  -H "Content-Type: application/json"
```

### Manual Cookie Header

Extract cookies from response and use them manually:

```bash
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Cookie: sAccessToken=eyJraWQ...; sRefreshToken=...; sFrontToken=..."
```

### Using jq to Extract Cookies

```bash
# Sign in and extract cookies
COOKIES=$(curl -i -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -d '{...}' | grep -i "set-cookie" | awk '{print $2}' | tr '\n' ';')

# Use extracted cookies
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Cookie: $COOKIES" \
  -H "Content-Type: application/json"
```

---

## Method 2: Header-Based Authentication (Bearer Token)

### How It Works

1. Sign in with `st-auth-mode: header` header
2. Receive access token in response body (not cookies)
3. Extract the access token from `accessToken` field
4. Send token in `Authorization: Bearer <token>` header
5. Backend validates token and grants access

### Why Use Header-Based Auth?

**Perfect for:**
- ✅ API testing tools (Postman, Insomnia)
- ✅ Mobile applications
- ✅ Server-to-server communication
- ✅ CI/CD pipelines
- ✅ Scripts and automation
- ✅ Non-browser clients

**Benefits:**
- No cookie management needed
- Works in any environment
- Easier to debug (token is visible)
- Standard OAuth 2.0 Bearer token pattern

### Sign In with Header Mode

**Important**: Add `st-auth-mode: header` to request headers!

```bash
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d '{
    "formFields": [
      {"id": "email", "value": "user@example.com"},
      {"id": "password", "value": "SecurePassword123!"}
    ]
  }'
```

**Response Headers:**
The tokens are returned in response headers:
```
St-Access-Token: eyJraWQiOiJkLTE3NjM3MDg4Nzg0NjYiLCJ0eXAiOiJKV1QiLCJ2ZXJzaW9uIjoiNCIsImFsZyI6IlJTMjU2In0...
St-Refresh-Token: uLa9MXtZes1+6jxtedyQjCCfDm8IPswi48B2ttQhy61HoLiDLej5pb...
Front-Token: eyJ1aWQiOiIwNDQxM2YyNS1mZGZhLTQyYTAtYTA0Ni1jM2FkNjdkMTM1ZmUi...
```

**Response Body:**
```json
{
  "status": "OK",
  "user": {
    "id": "04413f25-fdfa-42a0-a046-c3ad67d135fe",
    "email": "user@example.com",
    "timeJoined": 1732217234567,
    "tenantIds": ["public"]
  }
}
```

**Important**: With header-based auth, tokens are in response **headers**, not the body!

### Extract and Use Access Token

**Important**: Tokens are in response **headers**, not the body! Use `-i` flag to get headers.

```bash
# Sign in and extract access token from HEADERS
RESPONSE=$(curl -i -s -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d '{
    "formFields": [
      {"id": "email", "value": "user@example.com"},
      {"id": "password", "value": "SecurePassword123!"}
    ]
  }')

# Extract access token from St-Access-Token header
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Access-Token:" | sed 's/St-Access-Token: //' | tr -d '\r')

# Extract refresh token from St-Refresh-Token header
REFRESH_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Refresh-Token:" | sed 's/St-Refresh-Token: //' | tr -d '\r')

echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."

# Use token in API requests
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

### Complete Header-Based Authentication Script

```bash
#!/bin/bash
# header_auth.sh - Complete header-based authentication example

set -e

BASE_URL="http://localhost:3000"
EMAIL="user@example.com"
PASSWORD="SecurePassword123!"

echo "🔐 Signing in with header-based auth..."

# Sign in and get tokens (use -i to get headers!)
RESPONSE=$(curl -i -s -X POST $BASE_URL/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d "{
    \"formFields\": [
      {\"id\": \"email\", \"value\": \"$EMAIL\"},
      {\"id\": \"password\", \"value\": \"$PASSWORD\"}
    ]
  }")

# Extract body (after blank line)
BODY=$(echo "$RESPONSE" | sed -n '/^$/,$p' | tail -n +2)

# Check if sign in was successful
STATUS=$(echo $BODY | jq -r '.status')
if [ "$STATUS" != "OK" ]; then
  echo "❌ Sign in failed!"
  echo $BODY | jq '.'
  exit 1
fi

# Extract tokens from HEADERS (not body!)
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Access-Token:" | sed 's/St-Access-Token: //' | tr -d '\r')
REFRESH_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Refresh-Token:" | sed 's/St-Refresh-Token: //' | tr -d '\r')
USER_ID=$(echo $BODY | jq -r '.user.id')

echo "✅ Sign in successful!"
echo "User ID: $USER_ID"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo ""

# Save tokens to file for reuse
cat > tokens.json << EOF
{
  "access_token": "$ACCESS_TOKEN",
  "refresh_token": "$REFRESH_TOKEN",
  "user_id": "$USER_ID"
}
EOF

echo "💾 Tokens saved to tokens.json"
echo ""

# Test API call with Bearer token
echo "📋 Testing API call with Bearer token..."
curl -s -X GET $BASE_URL/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'

echo ""
echo "✅ Header-based authentication working!"
```

Make it executable and run:
```bash
chmod +x header_auth.sh
./header_auth.sh
```

### Reusing Saved Tokens

```bash
# Load token from file
ACCESS_TOKEN=$(cat tokens.json | jq -r '.access_token')

# Use in API calls
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'

# Create tenant
curl -X POST http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "slug": "my-company"
  }' | jq '.'

# Check platform admin status
curl -X GET http://localhost:3000/api/v1/platform/admins/check \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
```

### Token Refresh (Header Mode)

When your access token expires (after 1 hour), use the refresh token:

```bash
# Load refresh token
REFRESH_TOKEN=$(cat tokens.json | jq -r '.refresh_token')

# Refresh the access token
RESPONSE=$(curl -s -X POST http://localhost:3000/auth/session/refresh \
  -H "Content-Type: application/json" \
  -H "rid: session" \
  -H "st-auth-mode: header" \
  -d "{
    \"refreshToken\": \"$REFRESH_TOKEN\"
  }")

# Extract new access token
NEW_ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken.token')

echo "New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."

# Update tokens.json
ACCESS_TOKEN=$NEW_ACCESS_TOKEN
cat > tokens.json << EOF
{
  "access_token": "$NEW_ACCESS_TOKEN",
  "refresh_token": "$REFRESH_TOKEN"
}
EOF
```

### Postman Configuration for Header-Based Auth

1. **Create Sign In Request**
   - Method: `POST`
   - URL: `http://localhost:3000/auth/signin`
   - Headers:
     - `Content-Type: application/json`
     - `rid: emailpassword`
     - `st-auth-mode: header` ⭐ Important!
   - Body (raw JSON):
     ```json
     {
       "formFields": [
         {"id": "email", "value": "your@email.com"},
         {"id": "password", "value": "YourPassword"}
       ]
     }
     ```

2. **Extract Token (Tests tab in Postman)**
   ```javascript
   // Save access token to environment variable
   const response = pm.response.json();
   if (response.status === "OK") {
     pm.environment.set("access_token", response.accessToken.token);
     pm.environment.set("refresh_token", response.refreshToken.token);
   }
   ```

3. **Use Token in Other Requests**
   - Add to Headers:
     - `Authorization: Bearer {{access_token}}`
   - Postman will automatically substitute the variable

### Comparison: Cookie vs Header Auth

| Feature | Cookie-Based | Header-Based |
|---------|-------------|--------------|
| **Best for** | Web browsers | API clients, mobile |
| **Setup complexity** | Simple | Slightly more setup |
| **Token management** | Automatic | Manual |
| **CSRF protection** | Required | Not needed |
| **Cross-domain** | Limited | Works everywhere |
| **Debugging** | Harder | Easier (visible token) |
| **Postman** | Works | Works better |
| **Mobile apps** | Difficult | Easy |
| **Security** | Very secure | Secure if HTTPS |

### Which Method Should You Use?

**Use Cookie-Based Auth:**
- ✅ Building a web application
- ✅ Frontend is React/Vue/Angular
- ✅ Same domain for frontend and backend
- ✅ Want automatic session management

**Use Header-Based Auth:**
- ✅ Building a mobile app
- ✅ Testing APIs with Postman/Insomnia
- ✅ Server-to-server communication
- ✅ CI/CD pipelines
- ✅ Need to inspect/debug tokens
- ✅ Cross-domain API calls

## Understanding SuperTokens Tokens

SuperTokens uses multiple tokens for different purposes. Understanding what each token does helps you debug authentication issues and use the API effectively.

### The Three Token Types

When you authenticate with SuperTokens, you receive **three different tokens**:

#### 1. **Access Token** (`sAccessToken` or `St-Access-Token`)
**Purpose**: Authentication credential for API requests

**Contains**:
- Session proof and validation data
- User ID
- Session handle
- Expiry timestamp (short-lived, ~1 hour)

**Security**: HIGH - This is your actual authentication credential

**Usage**:
- **Cookie mode**: Automatically sent by browser as `sAccessToken` cookie
- **Header mode**: Send in `Authorization: Bearer <token>` header

**Example**:
```bash
# Cookie mode - automatic
curl -X GET http://localhost:3000/api/v1/tenants -b cookies.txt

# Header mode - manual
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer eyJraWQiOiJkLTE3NjM..."
```

#### 2. **Refresh Token** (`sRefreshToken` or `St-Refresh-Token`)
**Purpose**: Get new access tokens when they expire

**Contains**:
- Refresh token hash
- Long-lived credential (days/weeks)

**Security**: HIGHEST - Most sensitive token, only used for token refresh

**Usage**:
- **Cookie mode**: Automatically used by SuperTokens SDK
- **Header mode**: Call `/auth/session/refresh` with this token

**Example** (Header mode):
```bash
# When access token expires, refresh it
curl -X POST http://localhost:3000/auth/session/refresh \
  -H "Content-Type: application/json" \
  -H "rid: session" \
  -H "st-auth-mode: header" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"
```

#### 3. **Front Token** (`sFrontToken` or `Front-Token`)
**Purpose**: Session metadata for frontend (informational only)

**Contains**:
- User ID
- Access token expiry timestamp
- Session metadata
- **NO sensitive authentication data**

**Security**: LOW - Public information, safe to inspect

**Usage**: Frontend checks session status without API calls

**Example**:
```bash
# Decode front token (it's base64 encoded)
FRONT_TOKEN="eyJ1aWQiOiIwNDQxM2YyNS1mZGZhLTQyYTAtYTA0Ni1jM2FkNjdkMTM1ZmUiLCJhdGUi..."
echo $FRONT_TOKEN | base64 -d | jq '.'
```

**Output**:
```json
{
  "uid": "04413f25-fdfa-42a0-a046-c3ad67d135fe",
  "ate": 1763759365000,
  "up": {
    "sessionHandle": "...",
    "exp": 1763759365,
    "iat": 1763755765,
    "iss": "http://localhost:3000/auth",
    "sub": "04413f25-fdfa-42a0-a046-c3ad67d135fe",
    "tId": "public"
  }
}
```

### Front Token Deep Dive

The front token is often misunderstood, so let's clarify exactly what it is and isn't.

#### What the Front Token IS

✅ **Session existence indicator**
- Tells the frontend "a session exists" without making an API call
- Contains user ID and expiry time

✅ **Performance optimization**
- Prevents unnecessary API calls on every page load
- Frontend can check session status locally

✅ **Multi-tab synchronization**
- When you log out in one tab, other tabs detect it via front token changes
- SuperTokens uses this for cross-tab session management

✅ **Public metadata**
- Contains only non-sensitive information
- Safe to decode and inspect in browser DevTools

#### What the Front Token IS NOT

❌ **NOT an authentication credential**
- Cannot be used to make authenticated API requests
- Does NOT prove you are who you claim to be

❌ **NOT sensitive**
- Contains no passwords or secrets
- Exposure is not a security risk

❌ **NOT required for API calls**
- Backend doesn't validate this token
- Only `sAccessToken` (cookie) or `St-Access-Token` (header) are checked

### Token Comparison Table

| Token | Purpose | Security | Used For | Location |
|-------|---------|----------|----------|----------|
| **Access Token** | Authentication | HIGH (secret) | API requests | Cookie: `sAccessToken`<br>Header: `Authorization: Bearer` |
| **Refresh Token** | Get new tokens | HIGHEST (secret) | Token refresh only | Cookie: `sRefreshToken`<br>Header: `St-Refresh-Token` |
| **Front Token** | Session info | LOW (public) | Frontend checks | Cookie: `sFrontToken`<br>Header: `Front-Token` |

### Why SuperTokens Uses This Design

#### Problem: Traditional Approach
```javascript
// Every page load needs an API call
await fetch('/api/check-session'); // ❌ Extra API call!
if (response.ok) {
  showDashboard();
} else {
  redirectToLogin();
}
```

#### Solution: Front Token
```javascript
// Check locally first
const frontToken = getFrontToken(); // From cookie/storage
if (frontToken && !isExpired(frontToken)) {
  // ✅ Already know session exists!
  showDashboard();
} else {
  redirectToLogin();
}
```

**Benefits**:
- ✅ Instant session status check (no network delay)
- ✅ Reduced backend load (fewer verification calls)
- ✅ Better UX (no loading spinner on every page)
- ✅ Multi-tab logout synchronization

### Practical Examples

#### Example 1: Checking Session Status (Frontend)

**JavaScript** (Frontend):
```javascript
// Get front token from cookie
function getFrontToken() {
  const cookies = document.cookie.split('; ');
  const token = cookies.find(c => c.startsWith('sFrontToken='));
  return token ? token.split('=')[1] : null;
}

// Parse front token (base64 decode)
function parseFrontToken(token) {
  try {
    const decoded = atob(token); // base64 decode
    return JSON.parse(decoded);
  } catch {
    return null;
  }
}

// Check if user is logged in (NO API CALL NEEDED!)
const frontToken = getFrontToken();
if (frontToken) {
  const data = parseFrontToken(frontToken);
  const isExpired = Date.now() > data.ate;
  
  if (!isExpired) {
    console.log("User is logged in:", data.uid);
    console.log("Session expires:", new Date(data.ate));
  } else {
    console.log("Session expired");
    redirectToLogin();
  }
}
```

#### Example 2: CANNOT Authenticate with Front Token

**❌ WRONG** - This will fail:
```bash
# Trying to use front token for authentication
FRONT_TOKEN="eyJ1aWQiOiI..."
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $FRONT_TOKEN"

# Result: 401 Unauthorized ❌
```

**✅ CORRECT** - Use access token:
```bash
# Use the actual access token
ACCESS_TOKEN="eyJraWQiOiJkLTE3NjM..."
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Result: 200 OK ✅
```

#### Example 3: Inspecting Front Token (Debugging)

```bash
# Extract front token from cookie file
FRONT_TOKEN=$(grep sFrontToken cookies.txt | awk '{print $NF}')

# Decode and view (it's base64 encoded JSON)
echo "$FRONT_TOKEN" | base64 -d | jq '.'

# Output shows:
# {
#   "uid": "04413f25-fdfa-42a0-a046-c3ad67d135fe",  ← User ID
#   "ate": 1763759365000,                          ← Access Token Expiry (timestamp)
#   "up": {
#     "exp": 1763759365,                           ← Same expiry in Unix time
#     "iat": 1763755765,                           ← Issued At time
#     "iss": "http://localhost:3000/auth",         ← Issuer (your backend)
#     "sub": "04413f25-...",                       ← Subject (user ID)
#     "tId": "public"                              ← Tenant ID (SuperTokens multi-tenancy)
#   }
# }
```

### When to Use Each Token

| Scenario | Token to Use |
|----------|--------------|
| Making API request | **Access Token** (sAccessToken cookie or Authorization Bearer) |
| Token expired, need new one | **Refresh Token** (call `/auth/session/refresh`) |
| Check if user is logged in (frontend) | **Front Token** (decode and check expiry) |
| Debugging session issues | **Front Token** (inspect metadata) + **Access Token** (for actual auth) |
| Multi-tab logout detection | **Front Token** (watch for changes) |

### Common Misconceptions

#### Misconception 1: "I need to send all three tokens"
**Reality**: No! The backend only needs the **Access Token** (cookie or header). The other tokens have specific purposes:
- Refresh token → Only for `/auth/session/refresh`
- Front token → Only for frontend checks (not sent to backend APIs)

#### Misconception 2: "Front token contains my password"
**Reality**: No sensitive data at all! It only contains:
- User ID (already public to the user)
- Expiry time (public information)
- Session metadata (non-sensitive)

#### Misconception 3: "If someone steals my front token, they can access my account"
**Reality**: No! The front token is **not** an authentication credential. Even if exposed, it cannot be used to make authenticated requests. You need the **Access Token** for that.

#### Misconception 4: "I should validate the front token on the backend"
**Reality**: No need! The backend only validates the **Access Token** (cookie or Bearer header). The front token is purely for frontend optimization.

### Debugging with Tokens

When debugging authentication issues, check each token:

```bash
# 1. Check if you have an access token
grep sAccessToken cookies.txt
# If missing → Sign in again

# 2. Check front token for expiry
FRONT_TOKEN=$(grep sFrontToken cookies.txt | awk '{print $NF}')
echo "$FRONT_TOKEN" | base64 -d | jq '.ate'
# Compare with current time: date +%s%3N

# 3. If expired, refresh using refresh token
REFRESH_TOKEN=$(grep sRefreshToken cookies.txt | awk '{print $NF}')
# Call /auth/session/refresh with this token

# 4. For header-based auth, check token in headers
curl -i -X POST http://localhost:3000/auth/signin \
  -H "st-auth-mode: header" \
  -H "rid: emailpassword" \
  -d '{...}' | grep -i "St-Access-Token"
```

### Key Takeaways

1. **Three tokens, three purposes**:
   - Access Token = Authentication (API requests)
   - Refresh Token = Get new tokens (token refresh)
   - Front Token = Session info (frontend checks)

2. **Only Access Token is used for API authentication**
   - Cookie mode: `sAccessToken` cookie (automatic)
   - Header mode: `Authorization: Bearer <token>` (manual)

3. **Front Token is safe to inspect**
   - Contains only public metadata
   - Cannot be used to authenticate
   - Used for performance optimization

4. **Think of it like a badge system**:
   - **Front Token** = Badge that says "I have access" (informational)
   - **Access Token** = The actual key that opens doors (functional)
   - **Refresh Token** = Master key to get new door keys (renewal)

## Testing Protected APIs

All examples below assume you've signed in and saved cookies to `cookies.txt`.

### Check Authentication Status

```bash
curl -X GET http://localhost:3000/api/v1/platform/admins/check \
  -b cookies.txt \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "success": true,
  "message": "Platform admin status checked",
  "data": {
    "is_platform_admin": false
  }
}
```

## Complete Examples

### 1. Create Account and Create a Tenant (Cookie-Based)

```bash
#!/bin/bash

# Step 1: Sign up
echo "📝 Signing up..."
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{
    "formFields": [
      {"id": "email", "value": "john@example.com"},
      {"id": "password", "value": "SecurePass123!"}
    ]
  }' | jq '.'

# Step 2: Create a tenant
echo -e "\n🏢 Creating tenant..."
curl -X POST http://localhost:3000/api/v1/tenants \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "metadata": {
      "industry": "Technology"
    }
  }' | jq '.'

# Step 3: List tenants
echo -e "\n📋 Listing tenants..."
curl -X GET http://localhost:3000/api/v1/tenants \
  -b cookies.txt \
  -H "Content-Type: application/json" | jq '.'
```

### 1b. Create Account and Create a Tenant (Header-Based)

```bash
#!/bin/bash

# Step 1: Sign up with header mode
echo "📝 Signing up..."
SIGNUP_RESPONSE=$(curl -s -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d '{
    "formFields": [
      {"id": "email", "value": "john@example.com"},
      {"id": "password", "value": "SecurePass123!"}
    ]
  }')

echo $SIGNUP_RESPONSE | jq '.'

# Extract access token
ACCESS_TOKEN=$(echo $SIGNUP_RESPONSE | jq -r '.accessToken.token')
echo "Access Token: ${ACCESS_TOKEN:0:50}..."

# Step 2: Create a tenant using Bearer token
echo -e "\n🏢 Creating tenant..."
curl -X POST http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "metadata": {
      "industry": "Technology"
    }
  }' | jq '.'

# Step 3: List tenants
echo -e "\n📋 Listing tenants..."
curl -X GET http://localhost:3000/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
```

### 2. Sign In Existing User

```bash
#!/bin/bash

# Sign in
echo "🔐 Signing in..."
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{
    "formFields": [
      {"id": "email", "value": "john@example.com"},
      {"id": "password", "value": "SecurePass123!"}
    ]
  }' | jq '.'

echo "✅ Cookies saved to cookies.txt"
```

### 3. Tenant Management

```bash
#!/bin/bash

# List tenants
echo "📋 Listing all tenants..."
curl -X GET http://localhost:3000/api/v1/tenants \
  -b cookies.txt | jq '.'

# Create tenant
echo -e "\n🏢 Creating new tenant..."
TENANT_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/tenants \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Startup Inc",
    "slug": "tech-startup",
    "metadata": {"industry": "SaaS"}
  }')

echo $TENANT_RESPONSE | jq '.'

# Extract tenant ID
TENANT_ID=$(echo $TENANT_RESPONSE | jq -r '.data.id')
echo "Tenant ID: $TENANT_ID"

# Get tenant details
echo -e "\n🔍 Getting tenant details..."
curl -X GET http://localhost:3000/api/v1/tenants/$TENANT_ID \
  -b cookies.txt | jq '.'

# Update tenant
echo -e "\n✏️ Updating tenant..."
curl -X PATCH http://localhost:3000/api/v1/tenants/$TENANT_ID \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Startup LLC"
  }' | jq '.'
```

### 4. Member Management

```bash
#!/bin/bash

TENANT_ID="your-tenant-id-here"

# List members
echo "👥 Listing members..."
curl -X GET http://localhost:3000/api/v1/tenants/$TENANT_ID/members \
  -b cookies.txt | jq '.'

# Invite a member
echo -e "\n📧 Inviting member..."
curl -X POST http://localhost:3000/api/v1/tenants/$TENANT_ID/invitations \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "relation_id": "relation-id-here"
  }' | jq '.'

# Add existing user as member
echo -e "\n➕ Adding member..."
curl -X POST http://localhost:3000/api/v1/tenants/$TENANT_ID/members \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-id-here",
    "relation_id": "relation-id-here"
  }' | jq '.'
```

### 5. Platform Admin Operations (Cookie-Based, Requires Platform Admin)

```bash
#!/bin/bash

# Check if you're a platform admin
echo "🔐 Checking platform admin status..."
curl -X GET http://localhost:3000/api/v1/platform/admins/check \
  -b cookies.txt | jq '.'

# List relations
echo -e "\n📊 Listing relations..."
curl -X GET http://localhost:3000/api/v1/platform/relations \
  -b cookies.txt | jq '.'

# Create a role
echo -e "\n🔐 Creating role..."
ROLE_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/platform/roles \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Content Manager",
    "description": "Can manage all content",
    "is_system": false
  }')

echo $ROLE_RESPONSE | jq '.'
ROLE_ID=$(echo $ROLE_RESPONSE | jq -r '.data.id')

# List permissions
echo -e "\n🔑 Listing permissions..."
curl -X GET http://localhost:3000/api/v1/platform/permissions \
  -b cookies.txt | jq '.'

# Assign permissions to role
echo -e "\n✅ Assigning permissions to role..."
curl -X POST http://localhost:3000/api/v1/platform/roles/$ROLE_ID/permissions \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": ["perm-id-1", "perm-id-2"]
  }' | jq '.'
```

### 5b. Platform Admin Operations (Header-Based, Requires Platform Admin)

```bash
#!/bin/bash

# Load access token
ACCESS_TOKEN=$(cat tokens.json | jq -r '.access_token')

# Check if you're a platform admin
echo "🔐 Checking platform admin status..."
curl -X GET http://localhost:3000/api/v1/platform/admins/check \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'

# List relations
echo -e "\n📊 Listing relations..."
curl -X GET http://localhost:3000/api/v1/platform/relations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'

# Create a role
echo -e "\n🔐 Creating role..."
ROLE_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/platform/roles \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Content Manager",
    "description": "Can manage all content",
    "is_system": false
  }')

echo $ROLE_RESPONSE | jq '.'
ROLE_ID=$(echo $ROLE_RESPONSE | jq -r '.data.id')

# List permissions
echo -e "\n🔑 Listing permissions..."
curl -X GET http://localhost:3000/api/v1/platform/permissions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.'

# Assign permissions to role
echo -e "\n✅ Assigning permissions to role..."
curl -X POST http://localhost:3000/api/v1/platform/roles/$ROLE_ID/permissions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": ["perm-id-1", "perm-id-2"]
  }' | jq '.'
```

### 6. Relation-to-Role Mapping (Platform Admin)

```bash
#!/bin/bash

RELATION_ID="relation-id-here"

# Get roles for a relation
echo "🔍 Getting roles for relation..."
curl -X GET http://localhost:3000/api/v1/platform/relations/$RELATION_ID/roles \
  -b cookies.txt | jq '.'

# Assign roles to relation
echo -e "\n➕ Assigning roles to relation..."
curl -X POST http://localhost:3000/api/v1/platform/relations/$RELATION_ID/roles \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [
      "role-id-1",
      "role-id-2",
      "role-id-3"
    ]
  }' | jq '.'

# Verify roles assigned
echo -e "\n✅ Verifying roles..."
curl -X GET http://localhost:3000/api/v1/platform/relations/$RELATION_ID/roles \
  -b cookies.txt | jq '.'
```

### 7. Complete Workflow Script

```bash
#!/bin/bash
# complete_workflow.sh - Complete API testing workflow

set -e

BASE_URL="http://localhost:3000"
COOKIE_FILE="cookies.txt"

echo "╔════════════════════════════════════════════════════════╗"
echo "║        UTM Backend - Complete API Workflow            ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Sign Up
echo -e "${BLUE}Step 1: Sign Up${NC}"
SIGNUP_RESPONSE=$(curl -s -X POST $BASE_URL/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c $COOKIE_FILE \
  -d '{
    "formFields": [
      {"id": "email", "value": "demo@example.com"},
      {"id": "password", "value": "DemoPass123!"}
    ]
  }')

STATUS=$(echo $SIGNUP_RESPONSE | jq -r '.status')
if [ "$STATUS" == "OK" ]; then
  USER_ID=$(echo $SIGNUP_RESPONSE | jq -r '.user.id')
  echo -e "${GREEN}✅ User created: $USER_ID${NC}"
elif [ "$STATUS" == "FIELD_ERROR" ]; then
  echo "⚠️  User already exists, signing in instead..."
  
  # Sign In
  curl -s -X POST $BASE_URL/auth/signin \
    -H "Content-Type: application/json" \
    -H "rid: emailpassword" \
    -c $COOKIE_FILE \
    -d '{
      "formFields": [
        {"id": "email", "value": "demo@example.com"},
        {"id": "password", "value": "DemoPass123!"}
      ]
    }' > /dev/null
  echo -e "${GREEN}✅ Signed in successfully${NC}"
fi
echo ""

# Step 2: Create Tenant
echo -e "${BLUE}Step 2: Create Tenant${NC}"
TENANT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/tenants \
  -b $COOKIE_FILE \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Demo Company",
    "slug": "demo-company",
    "metadata": {"industry": "Technology"}
  }')

TENANT_ID=$(echo $TENANT_RESPONSE | jq -r '.data.id')
echo -e "${GREEN}✅ Tenant created: $TENANT_ID${NC}"
echo ""

# Step 3: List Tenants
echo -e "${BLUE}Step 3: List Tenants${NC}"
curl -s -X GET $BASE_URL/api/v1/tenants \
  -b $COOKIE_FILE | jq '.data.data[] | {id, name, slug, status}'
echo ""

# Step 4: Get Tenant Details
echo -e "${BLUE}Step 4: Get Tenant Details${NC}"
curl -s -X GET $BASE_URL/api/v1/tenants/$TENANT_ID \
  -b $COOKIE_FILE | jq '.data'
echo ""

# Step 5: List Members
echo -e "${BLUE}Step 5: List Tenant Members${NC}"
curl -s -X GET $BASE_URL/api/v1/tenants/$TENANT_ID/members \
  -b $COOKIE_FILE | jq '.data.data[] | {user_id, status}'
echo ""

# Step 6: Check Platform Admin Status
echo -e "${BLUE}Step 6: Check Platform Admin Status${NC}"
curl -s -X GET $BASE_URL/api/v1/platform/admins/check \
  -b $COOKIE_FILE | jq '.data'
echo ""

echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║           ✅ Workflow Completed Successfully!          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "📝 Cookies saved to: $COOKIE_FILE"
echo "🔑 You can now use these cookies for additional API calls"
echo ""
echo "Example:"
echo "  curl -X GET $BASE_URL/api/v1/tenants -b $COOKIE_FILE"
```

Make it executable:
```bash
chmod +x complete_workflow.sh
./complete_workflow.sh
```

## Error Handling

### Common Errors and Solutions

**401 Unauthorized**
```json
{
  "success": false,
  "error": "User not authenticated"
}
```
**Solution**: Sign in again and get fresh cookies

**403 Forbidden**
```json
{
  "success": false,
  "error": "Platform admin access required"
}
```
**Solution**: You need platform admin privileges. Contact an admin or use the script to make yourself one:
```bash
./scripts/create_platform_admin.sh <your-user-id>
```

**400 Bad Request**
```json
{
  "success": false,
  "error": "validation error: name is required"
}
```
**Solution**: Check request body matches the required schema

**404 Not Found**
```json
{
  "success": false,
  "error": "Tenant not found"
}
```
**Solution**: Verify the resource ID exists

## Session Management

### Refresh Token

Sessions automatically refresh. If you get a 401 error, sign in again:

```bash
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{
    "formFields": [
      {"id": "email", "value": "your@email.com"},
      {"id": "password", "value": "YourPassword123!"}
    ]
  }'
```

### Sign Out

```bash
curl -X POST http://localhost:3000/auth/signout \
  -b cookies.txt \
  -H "rid: emailpassword"
```

**Response:**
```json
{
  "status": "OK"
}
```

## Testing with Postman

### Import into Postman

1. **Create a new request**
2. **Set URL**: `http://localhost:3000/auth/signin`
3. **Method**: POST
4. **Headers**:
   - `Content-Type: application/json`
   - `rid: emailpassword`
5. **Body** (raw JSON):
   ```json
   {
     "formFields": [
       {"id": "email", "value": "your@email.com"},
       {"id": "password", "value": "YourPassword"}
     ]
   }
   ```
6. **Settings**: Enable "Automatically follow redirects" and "Send cookies with requests"

### Using Cookies in Postman

Postman automatically manages cookies after sign in. Just:
1. Sign in once
2. Cookies are stored automatically
3. All subsequent requests include cookies

## Advanced Usage

### Get Your User ID

After signing in, extract your user ID from the response or check:

```bash
# From backend logs
docker-compose logs api | grep "Session verified" | tail -1

# Or from sign in response
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d '{...}' | jq -r '.user.id'
```

### Become a Platform Admin

```bash
# Get your user ID first
USER_ID=$(docker-compose logs api | grep "Session verified" | tail -1 | grep -o '[a-f0-9-]\{36\}')

# Create platform admin
./scripts/create_platform_admin.sh $USER_ID

# Verify
curl -X GET http://localhost:3000/api/v1/platform/admins/check \
  -b cookies.txt | jq '.data.is_platform_admin'
```

### List All Relations, Roles, and Permissions

```bash
#!/bin/bash

echo "📊 Relations:"
curl -s http://localhost:3000/api/v1/relations -b cookies.txt | jq '.data[] | {id, name}'

echo -e "\n🔐 Roles:"
curl -s http://localhost:3000/api/v1/roles -b cookies.txt | jq '.data[] | {id, name, is_system}'

echo -e "\n🔑 Permissions:"
curl -s http://localhost:3000/api/v1/permissions -b cookies.txt | jq '.data[] | {id, service, entity, action}'
```

## Reference

### All API Endpoints

**Authentication**
- `POST /auth/signup` - Create account
- `POST /auth/signin` - Sign in
- `POST /auth/signout` - Sign out

**Tenants**
- `GET /api/v1/tenants` - List tenants
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/:id` - Get tenant
- `PATCH /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant

**Members**
- `GET /api/v1/tenants/:id/members` - List members
- `POST /api/v1/tenants/:id/members` - Add member
- `GET /api/v1/tenants/:id/members/:user_id` - Get member
- `PATCH /api/v1/tenants/:id/members/:user_id` - Update member
- `DELETE /api/v1/tenants/:id/members/:user_id` - Remove member

**Invitations**
- `POST /api/v1/tenants/:id/invitations` - Create invitation
- `GET /api/v1/tenants/:id/invitations` - List invitations
- `POST /api/v1/invitations/:token/accept` - Accept invitation

**Platform Admin** (requires admin)
- `GET /api/v1/platform/admins/check` - Check admin status
- `GET /api/v1/platform/admins` - List admins
- `POST /api/v1/platform/admins` - Create admin
- `DELETE /api/v1/platform/admins/:user_id` - Remove admin

**Relations** (read: all users, write: admin only)
- `GET /api/v1/relations` - List relations
- `POST /api/v1/platform/relations` - Create relation (admin)
- `PATCH /api/v1/platform/relations/:id` - Update relation (admin)
- `DELETE /api/v1/platform/relations/:id` - Delete relation (admin)
- `POST /api/v1/platform/relations/:id/roles` - Assign roles (admin)
- `GET /api/v1/platform/relations/:id/roles` - Get roles (admin)
- `DELETE /api/v1/platform/relations/:id/roles/:role_id` - Revoke role (admin)

**Roles** (read: all users, write: admin only)
- `GET /api/v1/roles` - List roles
- `POST /api/v1/platform/roles` - Create role (admin)
- `PATCH /api/v1/platform/roles/:id` - Update role (admin)
- `DELETE /api/v1/platform/roles/:id` - Delete role (admin)
- `POST /api/v1/platform/roles/:id/permissions` - Assign permissions (admin)
- `DELETE /api/v1/platform/roles/:id/permissions/:perm_id` - Revoke permission (admin)

**Permissions** (read: all users, write: admin only)
- `GET /api/v1/permissions` - List permissions
- `POST /api/v1/platform/permissions` - Create permission (admin)
- `DELETE /api/v1/platform/permissions/:id` - Delete permission (admin)

## Troubleshooting

### Cookies Not Working

**Problem**: Getting 401 errors even with cookies

**Solutions**:
1. Check cookie file exists: `cat cookies.txt`
2. Sign in again to get fresh cookies
3. Use `-b cookies.txt` (not `-H "Cookie: ..."`)
4. Ensure cookies aren't expired (1 hour default)

### Cannot Access Platform Admin APIs

**Problem**: Getting 403 Forbidden on `/api/v1/platform/*` endpoints

**Solution**: You need to be a platform admin
```bash
# Check status
curl -X GET http://localhost:3000/api/v1/platform/admins/check -b cookies.txt

# If not admin, create yourself as one
./scripts/create_platform_admin.sh <your-user-id>
```

### CORS Errors

**Problem**: Getting CORS errors in browser

**Solution**: Use the Vite proxy (port 3000), not direct API (port 8080)
- ✅ `http://localhost:3000/api/v1/...`
- ❌ `http://localhost:8080/api/v1/...`

## Additional Resources

- **API Examples**: `docs/API_EXAMPLES.md`
- **Quick Start**: `docs/QUICKSTART.md`
- **Platform Admin Guide**: `docs/changedoc/07-PLATFORM_ADMIN_COMPLETE.md`
- **Design Docs**: `docs/changedoc/06-PLATFORM_ADMIN_DESIGN.md`

---

**Last Updated**: November 21, 2025  
**Version**: 1.0

