# Authentication Testing Guide

UTM Backend supports **both cookie-based and header-based authentication** with SuperTokens.

## 🚀 Quick Start

### Frontend (Cookie-Based Auth) - **RECOMMENDED**

The easiest way to test the complete flow:

1. **Open the React frontend**: http://localhost:3000
2. **Sign Up** with your email and password
3. **Create tenants** using the UI
4. **View your tenants** - all authentication is handled automatically via cookies

**Status**: ✅ Fully functional

### API Testing (Header-Based Auth)

For direct API testing with tools like Postman, curl, or other API clients.

**Status**: ⚠️ SuperTokens Go SDK has limited header-based auth support by default. Cookie mode is recommended.

## 🍪 Cookie-Based Authentication (Frontend)

### How it Works

1. User signs up/signs in via SuperTokens UI
2. SuperTokens sets HTTP-only cookies automatically
3. Browser sends cookies with every request
4. Backend validates session from cookies

### Testing with Browser

```bash
# 1. Open frontend
open http://localhost:3000

# 2. Sign up with:
#    Email: yourname@example.com
#    Password: YourSecurePassword123!

# 3. Create a tenant using the form
# 4. View your tenants in the dashboard
```

### Frontend Features

- ✅ SuperTokens pre-built authentication UI
- ✅ Protected routes (requires login)
- ✅ Automatic session management
- ✅ Tenant creation and listing
- ✅ Logout functionality
- ✅ Responsive design (light/dark mode)

## 📡 Header-Based Authentication (API Clients)

### Current Status

SuperTokens Go SDK primarily uses cookie-based sessions. Header-based auth requires additional configuration and is not fully supported in the current setup.

### Cookie-Based API Testing with Postman

Even for API testing, using cookies is the recommended approach:

#### Step 1: Sign Up

```http
POST http://localhost:8080/auth/signup
Content-Type: application/json
rid: emailpassword

{
  "formFields": [
    {"id": "email", "value": "test@example.com"},
    {"id": "password", "value": "SecurePassword123!"}
  ]
}
```

**Important**: Enable "Automatically follow redirects" and "Send cookies with requests" in Postman settings.

#### Step 2: Sign In

```http
POST http://localhost:8080/auth/signin
Content-Type: application/json
rid: emailpassword

{
  "formFields": [
    {"id": "email", "value": "test@example.com"},
    {"id": "password", "value": "SecurePassword123!"}
  ]
}
```

Postman will automatically store the session cookies.

#### Step 3: Create Tenant

```http
POST http://localhost:8080/api/v1/tenants
Content-Type: application/json

{
  "name": "My Company",
  "slug": "my-company",
  "metadata": {
    "industry": "technology"
  }
}
```

Cookies are sent automatically by Postman!

#### Step 4: List Tenants

```http
GET http://localhost:8080/api/v1/tenants
Content-Type: application/json
```

## 🔧 Testing with cURL (Cookie Mode)

```bash
# Create a cookie jar
COOKIE_JAR="/tmp/utm_cookies.txt"

# 1. Sign up
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c "$COOKIE_JAR" \
  -d '{"formFields":[{"id":"email","value":"test@example.com"},{"id":"password","value":"Test123!"}]}'

# 2. Sign in
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -b "$COOKIE_JAR" \
  -c "$COOKIE_JAR" \
  -d '{"formFields":[{"id":"email","value":"test@example.com"},{"id":"password","value":"Test123!"}]}'

# 3. Create tenant (with cookies)
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{"name":"Test Company","slug":"test-company","metadata":{"industry":"tech"}}'

# 4. List tenants
curl -X GET http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR"
```

## 🎯 Available API Endpoints

### Public Endpoints

- `GET /health` - Health check
- `POST /auth/signup` - Create account
- `POST /auth/signin` - Sign in
- `POST /auth/signout` - Sign out

### Protected Endpoints (Require Authentication)

#### Tenants
- `POST /api/v1/tenants` - Create tenant (self-onboarding)
- `GET /api/v1/tenants` - List user's tenants
- `GET /api/v1/tenants/:id` - Get tenant details
- `PATCH /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant
- `GET /api/v1/tenants/:id/status` - Check tenant initialization status

#### Members (Require Tenant Access)
- `POST /api/v1/tenants/:id/members` - Add member to tenant
- `GET /api/v1/tenants/:id/members` - List tenant members
- `GET /api/v1/tenants/:id/members/:user_id` - Get member details
- `PATCH /api/v1/tenants/:id/members/:user_id` - Update member
- `DELETE /api/v1/tenants/:id/members/:user_id` - Remove member
- `POST /api/v1/tenants/:id/members/:user_id/roles` - Assign roles
- `DELETE /api/v1/tenants/:id/members/:user_id/roles/:role_id` - Remove role

#### Invitations
- `POST /api/v1/tenants/:id/invitations` - Invite user to tenant
- `GET /api/v1/tenants/:id/invitations` - List invitations
- `POST /api/v1/invitations/:token/accept` - Accept invitation
- `DELETE /api/v1/invitations/:id` - Cancel invitation

#### RBAC
- `GET /api/v1/relations` - List relations
- `POST /api/v1/relations` - Create relation
- `GET /api/v1/roles` - List roles
- `POST /api/v1/roles` - Create role
- `POST /api/v1/roles/:id/permissions` - Assign permissions
- `GET /api/v1/permissions` - List permissions
- `POST /api/v1/authorize` - Check user permissions

## 🐛 Troubleshooting

### "unauthorised" Error

**Symptom**: Getting 401 status with `{"message":"unauthorised"}`

**Solutions**:
1. **Use the frontend** at http://localhost:3000 - cookies work automatically
2. **In Postman**: Enable "Send cookies with requests" in Settings
3. **With curl**: Always use `-b` (send cookies) and `-c` (save cookies) flags
4. Make sure you signed in before making protected API calls

### Cookies Not Being Set

**Check**:
1. Backend is running: `curl http://localhost:8080/health`
2. SuperTokens is healthy: `docker-compose ps`
3. You're using the correct domain (localhost, not 127.0.0.1)
4. Cookie jar file exists and is writable (for curl)

### Frontend Not Loading

```bash
# Check frontend logs
docker-compose logs frontend

# Restart frontend
docker-compose restart frontend

# Check if port 3000 is available
lsof -i :3000
```

### Session Expired

Simply sign in again. Sessions expire after a period of inactivity.

## 📊 Monitoring

### View Logs
```bash
docker-compose logs -f frontend   # Frontend logs
docker-compose logs -f api        # Backend API logs
docker-compose logs -f worker     # Background worker logs
```

### Check Services
```bash
docker-compose ps                 # All services status
curl http://localhost:8080/health # API health
curl http://localhost:3567/hello  # SuperTokens health
```

### Email Testing

All invitation emails are captured in MailHog: http://localhost:8025

## 🎨 Frontend Stack

- **React 18** - UI library
- **Vite** - Build tool (fast HMR)
- **SuperTokens React** - Pre-built auth UI
- **React Router** - Client-side routing

## 📚 Learn More

- [SuperTokens Documentation](https://supertokens.com/docs)
- [SuperTokens React SDK](https://supertokens.com/docs/emailpassword/quick-setup/frontend)
- [Backend API Documentation](./API_EXAMPLES.md)
- [Frontend README](./frontend/README.md)

## ✅ Recommended Testing Flow

1. **Start**: Use the React frontend (http://localhost:3000)
2. **Sign up**: Create an account
3. **Create tenant**: Use the UI form
4. **Explore**: View tenant list, check status
5. **API Testing**: Use Postman with cookie support enabled
6. **Advanced**: Test invitations, RBAC, member management

The frontend provides the best developer experience with automatic session management! 🚀

