# Core Concepts

Understanding these key concepts will help you work effectively with Rex.

## Tenant

A **Tenant** is an isolated workspace or organization in the system. Think of it as a company, team, or any group that needs its own separate space.

**Key Properties**:
- **ID**: Unique identifier (UUID)
- **Name**: Display name (e.g., "Acme Corporation")
- **Slug**: URL-friendly identifier (e.g., "acme-corp")
- **Status**: `pending`, `active`, `suspended`, or `deleted`
- **Metadata**: Custom JSON data

**Example**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "status": "active",
  "created_by": "user@example.com"
}
```

**Use Cases**:
- **Companies**: Each business gets a tenant
- **Teams**: Departments or project teams
- **Customers**: Each customer organization
- **Workspaces**: User-created workspaces

## Member

A **Member** represents a user's membership in a specific tenant with an assigned role.

**Key Properties**:
- User ID (from SuperTokens)
- Tenant ID
- Role ID (Admin, Writer, etc.)
- Status: `active`, `inactive`, or `pending`

**Important**: One user can be a member of multiple tenants with different roles in each.

**Example**:
```
User: john@example.com
  ├─ Acme Corp → Admin
  ├─ Startup Inc → Writer
  └─ Beta Co → Viewer
```

## Role

A **Role** defines a user's position or job function within a tenant. It determines what permissions they have.

**Previously Called**: "Relation" (renamed for clarity)

**Built-in Roles**:
- **Admin**: Full control over tenant
- **Writer**: Can create and edit content
- **Viewer**: Read-only access
- **Basic**: Minimal permissions

**Properties**:
- Name (e.g., "Admin", "Support Agent")
- Type: `tenant` or `platform`
- Associated Policies
- System flag (for built-in roles)

**Custom Roles**: You can create your own roles via API.

## Policy

A **Policy** is a group of related permissions. Policies make it easier to manage permissions by grouping them logically.

**Previously Called**: "Role" (renamed to avoid confusion)

**Example Policies**:
- **Tenant Admin Policy**: All tenant management permissions
- **Content Writer Policy**: Create, edit, view content
- **Content Viewer Policy**: View content only
- **Support Policy**: View tickets, update ticket status

**Why Policies?**
- **Reusability**: One policy can be used by multiple roles
- **Clarity**: Group related permissions together
- **Maintenance**: Update permissions in one place

**Example**:
```
"Content Writer Policy" contains:
  - blog-api:post:create
  - blog-api:post:update
  - blog-api:post:read
  - media-api:image:upload
```

## Permission

A **Permission** is the atomic unit of authorization - a specific action that can be performed.

**Format**: `service:entity:action`

**Components**:
- **Service**: The API or service (e.g., `tenant-api`, `blog-api`)
- **Entity**: The resource type (e.g., `member`, `post`)
- **Action**: The operation (e.g., `create`, `read`, `update`, `delete`)

**Examples**:
```
tenant-api:member:invite      # Can invite members to tenant
tenant-api:member:remove      # Can remove members
tenant-api:tenant:update      # Can update tenant settings
blog-api:post:publish         # Can publish blog posts
blog-api:post:delete          # Can delete blog posts
media-api:image:upload        # Can upload images
```

**Wildcard Support** (implementation dependent):
```
tenant-api:*:*                # All actions on all entities in tenant-api
tenant-api:member:*           # All actions on members
*:*:read                      # Read-only across all services
```

## RBAC Hierarchy

The complete authorization hierarchy:

```
User (SuperTokens)
  ↓ belongs to
Tenant (with membership)
  ↓ assigned
Role (Admin, Writer, etc.)
  ↓ contains
Policies (groups of permissions)
  ↓ contains
Permissions (service:entity:action)
```

**Example Flow**:
```
1. User "john@example.com"
2. Is member of tenant "Acme Corp"
3. Has role "Content Manager"
4. Role has policy "Content Writer Policy"
5. Policy has permission "blog-api:post:create"
6. Therefore, John can create blog posts in Acme Corp
```

## System User

A **System User** is a service account for machine-to-machine (M2M) authentication.

**Use Cases**:
- Background workers
- Scheduled jobs (cron)
- API integrations
- Automated scripts

**Characteristics**:
- Authenticates via email/password (like regular users)
- Email format: `service-name@system.internal`
- Special JWT payload flags:
  ```json
  {
    "is_system_user": true,
    "service_name": "background-worker",
    "service_type": "worker"
  }
  ```
- Longer token expiry (24 hours vs 1 hour)
- Can be rotated with grace periods

**Example**:
```json
{
  "name": "background-worker",
  "email": "background-worker@system.internal",
  "service_type": "worker",
  "is_active": true
}
```

## Platform Admin

A **Platform Admin** is a super user with system-wide privileges.

**Capabilities**:
- Access any tenant (without being a member)
- Create and manage tenants
- Create system users
- Manage RBAC configuration (roles, policies, permissions)
- View all users and platform statistics

**Security**: Platform admins should be carefully managed and limited.

## Invitation

An **Invitation** is an email-based invite to join a tenant as a member.

**Properties**:
- Email address
- Tenant
- Role to assign
- Unique token (UUID)
- Expiry time (default: 72 hours)
- Status: `pending`, `accepted`, `expired`, or `cancelled`

**Flow**:
```
1. Admin creates invitation for user@example.com
2. System sends email with link containing token
3. User clicks link, signs up/logs in
4. User accepts invitation
5. User becomes member of tenant with assigned role
```

## Session

A **Session** represents an authenticated user's active login.

**SuperTokens Manages**:
- Session creation on login
- Session verification on requests
- Token refresh before expiry
- Session revocation on logout

**Authentication Modes**:
- **Cookie-based**: For web applications (default)
- **Header-based**: For APIs and mobile apps

**Token Types**:
- **Access Token**: Short-lived (1 hour), used for API requests
- **Refresh Token**: Longer-lived, used to get new access tokens

## Background Job

An **Background Job** is an asynchronous task processed outside the request/response cycle.

**Built-in Jobs**:
- **Tenant Initialization**: Set up new tenant resources
- **User Invitation Email**: Send invitation emails
- **System User Expiry**: Clean up expired credentials

**Job Properties**:
- Type (e.g., `tenant:init`)
- Payload (JSON data)
- Priority queue: `critical`, `default`, or `low`
- Retry policy
- Timeout

## Middleware

**Middleware** are functions that process requests before they reach handlers.

**Common Middleware**:
```
1. Logger          # Log all requests
2. CORS            # Handle cross-origin requests
3. SuperTokens     # Handle SuperTokens auth routes
4. Auth            # Verify user is authenticated
5. TenantAccess    # Verify user has access to tenant
6. RBAC            # Verify user has required permission
```

**Middleware Chain**:
```
Request → Logger → CORS → SuperTokens → Auth → TenantAccess → RBAC → Handler
```

## API Response Format

All API responses follow a consistent structure:

**Success Response**:
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { /* response data */ }
}
```

**Error Response**:
```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

**Paginated Response**:
```json
{
  "success": true,
  "data": {
    "data": [ /* array of items */ ],
    "page": 1,
    "page_size": 20,
    "total_count": 150,
    "total_pages": 8
  }
}
```

## Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET, PATCH, POST (non-creation) |
| 201 | Created | Successful resource creation |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid input, validation error |
| 401 | Unauthorized | Not authenticated (no/invalid session) |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Unexpected server error |

## Request Authentication

**Cookie-based** (Web):
```javascript
fetch('/api/v1/tenants', {
  credentials: 'include'  // Include cookies
})
```

**Header-based** (API):
```bash
Authorization: Bearer <access-token>
st-auth-mode: header
```

## Tenant Isolation

**Data Isolation**: Each tenant's data is completely separate.

**How It Works**:
```sql
-- All queries include tenant_id filter
SELECT * FROM content 
WHERE tenant_id = 'xxx' AND ...

-- Prevents cross-tenant data access
```

**Middleware Enforcement**:
```
TenantAccessMiddleware checks:
1. Is user a member of this tenant? OR
2. Is user a platform admin?
```

## Permission Checking

**At API Level**:
```go
// In route definition
router.POST("/members",
    middleware.RequirePermission("tenant-api", "member", "create"),
    handler.AddMember,
)
```

**In Service Layer**:
```go
// In business logic
hasPermission, err := rbacService.CheckUserPermission(
    tenantID, userID, "tenant-api", "member", "invite"
)
if !hasPermission {
    return errors.New("permission denied")
}
```

**In Frontend**:
```javascript
// Get user permissions
const permissions = await fetchUserPermissions(tenantId);

// Check permission
if (permissions.includes('tenant-api:member:invite')) {
    // Show "Invite Member" button
}
```

## Configuration

**Environment Variables**: All configuration via `.env` file.

**Key Variables**:
- `APP_ENV`: `development` or `production`
- `DB_*`: Database connection
- `SUPERTOKENS_*`: Authentication settings
- `REDIS_*`: Redis connection
- `SMTP_*`: Email configuration

**Loading Order**:
1. `.env` file (development)
2. Environment variables (production)
3. Default values (if safe)

## Next Steps

- **[Quick Start](/getting-started/quick-start)** - Get the system running
- **[Authentication Guide](/guides/authentication)** - Learn about auth
- **[RBAC Guide](/guides/rbac-overview)** - Master authorization
- **[API Reference](/x-api/overview)** - Explore the API

