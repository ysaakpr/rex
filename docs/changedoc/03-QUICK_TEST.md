# 🚀 Quick Test Guide - Get Started in 2 Minutes

## ✅ Prerequisites Check

```bash
# Make sure all services are running
docker-compose ps

# You should see all services as "Up" or "healthy"
```

## 🎨 Test 1: Frontend (Cookie-Based Auth) - **EASIEST**

### Step 1: Open the Frontend

```bash
open http://localhost:3000
# Or visit: http://localhost:3000 in your browser
```

### Step 2: Create an Account

1. Click **"Sign Up"** (usually the default view)
2. Enter:
   - **Email**: `demo@example.com` (or any email)
   - **Password**: `DemoPassword123!` (must be strong)
3. Click **"SIGN UP"**

### Step 3: Create a Tenant

You'll be redirected to the dashboard. Fill in the form:
- **Tenant Name**: `My First Company`
- **Slug**: `my-first-company` (auto-generated)
- **Industry**: `Technology` (optional)

Click **"Create Tenant"**

### Step 4: View Your Tenants

Your newly created tenant appears below the form with:
- Tenant name
- Status badge (PENDING → ACTIVE)
- Creation date
- Metadata

**🎉 Success! Cookie-based authentication works perfectly!**

---

## 📡 Test 2: API with Postman (Cookie-Based)

### Setup Postman

1. Open Postman
2. Go to Settings → Enable **"Send cookies with requests"**
3. Go to Settings → Enable **"Automatically follow redirects"**

### Create Requests

#### Request 1: Sign In

```
Method: POST
URL: http://localhost:8080/auth/signin
Headers:
  Content-Type: application/json
  rid: emailpassword
Body (JSON):
{
  "formFields": [
    {"id": "email", "value": "demo@example.com"},
    {"id": "password", "value": "DemoPassword123!"}
  ]
}
```

**Expected**: Status 200, user object returned, cookies stored

#### Request 2: Create Tenant

```
Method: POST
URL: http://localhost:8080/api/v1/tenants
Headers:
  Content-Type: application/json
Body (JSON):
{
  "name": "API Test Company",
  "slug": "api-test-company",
  "metadata": {
    "industry": "Technology",
    "employees": "10-50"
  }
}
```

**Expected**: Status 201, tenant created

#### Request 3: List Tenants

```
Method: GET
URL: http://localhost:8080/api/v1/tenants
Headers:
  Content-Type: application/json
```

**Expected**: Status 200, array of your tenants

**🎉 Success! Postman cookie-based authentication works!**

---

## 🔧 Test 3: Command Line (Using Existing Test User)

```bash
# Use the pre-created test user
EMAIL="testuser@example.com"
PASSWORD="TestPassword123!"

# Create tenant using cookie-based auth
cd /Users/vyshakhp/work/utm-backend
bash scripts/test_tenant_api.sh
```

This script will:
1. Sign in with the test user
2. Create a unique tenant
3. Check tenant status
4. List all tenants

---

## 📊 Verify Everything Works

### 1. Check Database

```bash
make shell-db

# In psql:
SELECT id, name, slug, status FROM tenants;
SELECT user_id, relation_id FROM tenant_members;
\q
```

### 2. Check Background Jobs

```bash
docker-compose logs worker | tail -20
```

You should see tenant initialization jobs being processed.

### 3. Check Email Capture

Visit: http://localhost:8025

Any invitation emails will appear here.

---

## 🎯 What Just Happened?

### Cookie-Based Auth Flow

1. **Sign In** → SuperTokens creates session → Sets HTTP-only cookies
2. **Protected Request** → Browser/Postman sends cookies automatically
3. **Middleware** → Validates session from cookies
4. **Success** → Request processed

### Why Cookies?

- ✅ More secure (HTTP-only, SameSite)
- ✅ Automatic handling by browsers
- ✅ No manual token management
- ✅ Built-in CSRF protection
- ✅ Works perfectly with SuperTokens

### What About Headers?

SuperTokens Go SDK primarily uses cookies. For pure API clients that need headers:
- Use API keys (separate implementation needed)
- Use OAuth2/JWT (separate implementation needed)
- Or use cookies (works with all HTTP clients)

---

## 🐛 Troubleshooting

### "unauthorised" Error

```bash
# 1. Make sure you signed in first
# 2. Check cookies are being sent
# 3. In Postman: Settings → "Send cookies with requests"
# 4. With curl: use -b and -c flags
```

### Frontend Not Loading

```bash
docker-compose restart frontend
docker-compose logs frontend
```

### Database Issues

```bash
make migrate-status
make migrate-down  # if needed
make migrate-up
```

---

## 🎊 Next Steps

1. **Explore the Frontend**: http://localhost:3000
   - Create multiple tenants
   - View tenant status updates
   - Test the UI/UX

2. **Try Member Management**:
   - Invite users to your tenant
   - Assign roles and permissions
   - Check emails in MailHog: http://localhost:8025

3. **Test RBAC**:
   - Create custom roles
   - Define permissions
   - Test authorization

4. **Build Your App**:
   - Use the frontend as a reference
   - Integrate with your own UI
   - Customize for your use case

---

## 📚 Resources

- **Full API Documentation**: [API_EXAMPLES.md](./API_EXAMPLES.md)
- **Authentication Guide**: [AUTH_TESTING.md](./AUTH_TESTING.md)
- **Frontend Guide**: [frontend/README.md](./frontend/README.md)
- **SuperTokens Docs**: https://supertokens.com/docs

---

## ✨ Summary

✅ **Cookie-Based Auth**: Fully functional with frontend and Postman  
✅ **Frontend UI**: Complete tenant management dashboard  
✅ **API Testing**: Works with any HTTP client that supports cookies  
✅ **Background Jobs**: Tenant initialization processing  
✅ **RBAC**: 4 relations, 4 roles, 27 permissions seeded  
✅ **Email Testing**: MailHog captures all emails  

**Everything is working! Start building your multi-tenant application! 🚀**

