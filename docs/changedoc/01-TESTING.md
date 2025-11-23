# Testing Guide for UTM Backend

## Current Status

✅ **Backend Services Running**
- API Server: `http://localhost:8080`
- SuperTokens: `http://localhost:3567`
- PostgreSQL: Healthy with all migrations applied
- Worker: Background jobs processing
- MailHog: `http://localhost:8025` (email testing)

✅ **Test User Created**
- Email: `testuser@example.com`
- Password: `TestPassword123!`
- User ID: `04d23859-f61d-4d6e-88ed-3bcc2c1f7ae7`

## Testing with Postman (Recommended)

### 1. Import Collection

Create a new Postman collection with these requests:

#### A. Sign Up
```
POST http://localhost:8080/auth/signup
Headers:
  Content-Type: application/json
  rid: emailpassword

Body (JSON):
{
  "formFields": [
    {
      "id": "email",
      "value": "your-email@example.com"
    },
    {
      "id": "password",
      "value": "YourPassword123!"
    }
  ]
}
```

#### B. Sign In
```
POST http://localhost:8080/auth/signin
Headers:
  Content-Type: application/json
  rid: emailpassword

Body (JSON):
{
  "formFields": [
    {
      "id": "email",
      "value": "your-email@example.com"
    },
    {
      "id": "password",
      "value": "YourPassword123!"
    }
  ]
}
```

**Important:** Make sure "Send cookies" is enabled in Postman settings!

#### C. Create Tenant
```
POST http://localhost:8080/api/v1/tenants
Headers:
  Content-Type: application/json

Body (JSON):
{
  "name": "My Company",
  "slug": "my-company",
  "metadata": {
    "industry": "technology",
    "size": "10-50"
  }
}
```

#### D. List Tenants
```
GET http://localhost:8080/api/v1/tenants
Headers:
  Content-Type: application/json
```

#### E. Get Tenant Details
```
GET http://localhost:8080/api/v1/tenants/{tenant_id}
Headers:
  Content-Type: application/json
```

### 2. Testing Flow

1. **Sign Up** → Creates a new user
2. **Sign In** → Establishes session (cookies will be saved automatically in Postman)
3. **Create Tenant** → Should work with the session cookies
4. **List Tenants** → See all your tenants
5. **Check tenant status** → Monitor initialization

## Testing with Frontend

### React Example

```javascript
import SuperTokens from "supertokens-auth-react";
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";
import Session from "supertokens-auth-react/recipe/session";

SuperTokens.init({
  appInfo: {
    appName: "UTM Backend",
    apiDomain: "http://localhost:8080",
    websiteDomain: "http://localhost:3000",
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    Session.init()
  ]
});

// After sign in, make API calls:
const createTenant = async () => {
  const response = await fetch('http://localhost:8080/api/v1/tenants', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Important: Include cookies
    body: JSON.stringify({
      name: "My Company",
      slug: "my-company",
      metadata: {
        industry: "technology"
      }
    })
  });
  
  return response.json();
};
```

## Current Known Issue

### Cookie vs Header Mode

The current setup has SuperTokens configured but cookies are not being set properly for curl/script-based testing. This is normal for SuperTokens as it's designed to work with browser-based clients or proper HTTP clients that handle cookies.

**Solutions:**

1. **Use Postman/Insomnia** (Recommended) - These tools properly handle Super Tokens sessions
2. **Use a Frontend** - Integrate with React/Vue/Angular using SuperTokens SDK
3. **Use Thunder Client** in VS Code - Similar to Postman, handles cookies properly

## API Endpoints Reference

### Authentication
- `POST /auth/signup` - Create new user
- `POST /auth/signin` - Sign in user  
- `POST /auth/signout` - Sign out user

### Tenants
- `POST /api/v1/tenants` - Create tenant (self-onboarding)
- `POST /api/v1/tenants/managed` - Create managed tenant
- `GET /api/v1/tenants` - List user's tenants
- `GET /api/v1/tenants/:id` - Get tenant details
- `PATCH /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant
- `GET /api/v1/tenants/:id/status` - Check initialization status

### Members
- `POST /api/v1/tenants/:id/members` - Add member
- `GET /api/v1/tenants/:id/members` - List members
- `GET /api/v1/tenants/:id/members/:user_id` - Get member
- `PATCH /api/v1/tenants/:id/members/:user_id` - Update member
- `DELETE /api/v1/tenants/:id/members/:user_id` - Remove member

### Invitations
- `POST /api/v1/tenants/:id/invitations` - Invite user
- `GET /api/v1/tenants/:id/invitations` - List invitations
- `POST /api/v1/invitations/:token/accept` - Accept invitation
- `DELETE /api/v1/invitations/:id` - Cancel invitation

### RBAC
- `GET /api/v1/relations` - List relations
- `POST /api/v1/relations` - Create relation
- `GET /api/v1/roles` - List roles
- `POST /api/v1/roles` - Create role
- `GET /api/v1/permissions` - List permissions
- `POST /api/v1/authorize` - Check permissions

## Database Inspection

```bash
# Access PostgreSQL
make shell-db

# View tenants
SELECT * FROM tenants;

# View members
SELECT * FROM tenant_members;

# View relations
SELECT * FROM relations;

# View roles and permissions
SELECT r.name as role, p.service, p.entity, p.action
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id;
```

## Monitoring

### View Logs
```bash
make logs-api      # API logs
make logs-worker   # Worker logs
make logs          # All logs
```

### Check Background Jobs

Worker logs will show tenant initialization jobs being processed.

### View Emails

All invitation emails are captured in MailHog: http://localhost:8025

## Troubleshooting

### 401 Unauthorized
- Make sure you're signed in
- Check that cookies are being sent (Postman: enable "Send cookies")
- Verify session hasn't expired

### Tenant Status "Pending"
- Check worker logs: `make logs-worker`
- Background job processes tenant initialization
- If `TENANT_INIT_SERVICES` is not configured, tenant will be marked active immediately

### Migration Issues
```bash
make migrate-status  # Check current version
make migrate-down    # Rollback if needed
make migrate-up      # Apply migrations
```

## Next Steps

1. **Build a Frontend** - Integrate with your React/Vue/Angular app
2. **Configure Services** - Set `TENANT_INIT_SERVICES` in `.env` to call your other backend services
3. **Customize RBAC** - Add custom relations, roles, and permissions for your use case
4. **Email Setup** - Configure SMTP/SendGrid for production emails

## Need Help?

- Check API logs: `make logs-api`
- Check health: `curl http://localhost:8080/health`
- View all services: `docker-compose ps`
- Restart services: `make stop && make run`

