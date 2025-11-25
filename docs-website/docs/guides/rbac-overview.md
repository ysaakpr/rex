# RBAC Overview

Complete guide to the Role-Based Access Control (RBAC) system in Rex.

## What is RBAC?

RBAC (Role-Based Access Control) is a 3-tier authorization system that controls what users can do within tenants:

```
User → Role → Policy → Permission
```

- **Role**: User's position in tenant (Admin, Writer, Viewer)
- **Policy**: Group of related permissions
- **Permission**: Specific action that can be performed

## 3-Tier Architecture

### Tier 1: Roles

**What**: User's role or position within a tenant

**Examples**:
- Admin - Full control
- Writer - Can create/edit content
- Viewer - Read-only access
- Basic - Minimal access

**Previously Called**: "Relations" (renamed for clarity)

### Tier 2: Policies

**What**: Groups of related permissions

**Examples**:
- Tenant Admin Policy - All tenant management permissions
- Content Writer Policy - Create, edit, view content
- Content Viewer Policy - View content only

**Previously Called**: "Roles" (renamed to avoid confusion)

### Tier 3: Permissions

**What**: Atomic actions that can be performed

**Format**: `service:entity:action`

**Examples**:
- `tenant-api:member:invite`
- `tenant-api:member:remove`
- `blog-api:post:create`
- `blog-api:post:delete`

## How It Works

### Permission Check Flow

```
1. Request to protected endpoint
   ↓
2. Get user's role in tenant
   ↓
3. Get role's policies
   ↓
4. Get policies' permissions
   ↓
5. Check if required permission exists
   ↓
6. Allow or deny request
```

### Example

```
User: john@example.com
  ↓ member of
Tenant: Acme Corp
  ↓ has role
Role: Writer
  ↓ contains
Policy: Content Writer Policy
  ↓ grants
Permissions:
  - blog-api:post:create
  - blog-api:post:update
  - blog-api:post:read
```

**Result**: John can create, update, and read blog posts in Acme Corp.

## Default Roles

Rex ships with 4 default roles:

### 1. Admin

**Description**: Full control over tenant

**Policies**:
- Tenant Admin Policy (all tenant permissions)

**Can Do**:
- Manage members
- Invite users
- Update tenant settings
- Delete tenant
- All content operations

### 2. Writer

**Description**: Can create and edit content

**Policies**:
- Content Writer Policy

**Can Do**:
- Create content
- Edit own content
- View all content
- Cannot manage members

### 3. Viewer

**Description**: Read-only access

**Policies**:
- Content Viewer Policy

**Can Do**:
- View content
- Cannot create or edit
- Cannot manage members

### 4. Basic

**Description**: Minimal access

**Policies**:
- Basic Access Policy

**Can Do**:
- View own profile
- Limited read access

## Permission Format

### Structure

```
service:entity:action
```

**Service**: The API or microservice
- Examples: `tenant-api`, `blog-api`, `media-api`

**Entity**: The resource type
- Examples: `member`, `post`, `image`, `tenant`

**Action**: The operation
- Common: `create`, `read`, `update`, `delete`
- Custom: `invite`, `publish`, `archive`

### Examples

```
tenant-api:member:invite      # Invite member to tenant
tenant-api:member:remove      # Remove member from tenant
tenant-api:tenant:update      # Update tenant settings
tenant-api:tenant:delete      # Delete tenant

blog-api:post:create          # Create blog post
blog-api:post:update          # Update blog post
blog-api:post:publish         # Publish blog post
blog-api:post:delete          # Delete blog post

media-api:image:upload        # Upload image
media-api:image:delete        # Delete image
```

### Wildcard Permissions (Future)

```
tenant-api:*:*                # All actions on all entities
tenant-api:member:*           # All member operations
*:*:read                      # Read-only across all services
```

::: warning
Wildcard permissions are not currently implemented but planned for future releases.
:::

## RBAC Database Schema

```sql
-- Roles (user's position in tenant)
roles
├── id (UUID)
├── name (Admin, Writer, etc.)
├── type (tenant, platform)
├── description
├── tenant_id (NULL for system roles)
└── is_system (true for built-in roles)

-- Policies (groups of permissions)
policies
├── id (UUID)
├── name (Tenant Admin Policy, etc.)
├── description
├── tenant_id (NULL for system policies)
└── is_system

-- Permissions (atomic actions)
permissions
├── id (UUID)
├── service (tenant-api, blog-api)
├── entity (member, post)
├── action (create, read, update, delete)
└── description

-- Role ← → Policy (many-to-many)
role_policies
├── role_id
└── policy_id

-- Policy ← → Permission (many-to-many)
policy_permissions
├── policy_id
└── permission_id

-- User ← → Role (via tenant membership)
tenant_members
├── user_id
├── tenant_id
└── role_id
```

## Platform vs Tenant Roles

### Platform Roles

**Scope**: System-wide

**Used For**: Platform administration

**Examples**:
- Platform Admin - Super user
- Support Agent - Customer support

**Characteristics**:
- `tenant_id` is NULL
- `type` is "platform"
- `is_system` is true

### Tenant Roles

**Scope**: Specific tenant

**Used For**: Tenant-specific access

**Examples**:
- Admin - Tenant administrator
- Writer - Content creator
- Viewer - Read-only user

**Characteristics**:
- `tenant_id` references tenant
- `type` is "tenant"
- Can be system or custom

## Creating Custom RBAC

### 1. Define Permission

```json
POST /api/v1/platform/permissions
{
  "service": "blog-api",
  "entity": "post",
  "action": "publish",
  "description": "Publish blog posts"
}
```

### 2. Create Policy

```json
POST /api/v1/platform/policies
{
  "name": "Blog Publisher Policy",
  "description": "Can publish blog posts"
}
```

### 3. Assign Permissions to Policy

```json
POST /api/v1/platform/policies/:policy_id/permissions
{
  "permission_ids": ["permission-uuid-1", "permission-uuid-2"]
}
```

### 4. Create Role

```json
POST /api/v1/platform/roles
{
  "name": "Publisher",
  "type": "tenant",
  "description": "Can publish content"
}
```

### 5. Assign Policies to Role

```json
POST /api/v1/platform/roles/:role_id/policies
{
  "policy_ids": ["policy-uuid-1", "policy-uuid-2"]
}
```

### 6. Assign Role to User

```json
PATCH /api/v1/tenants/:tenant_id/members/:user_id
{
  "role_id": "publisher-role-uuid"
}
```

## Checking Permissions

### Backend (Middleware)

```go
// In route definition
router.POST("/posts",
    middleware.RequirePermission(rbacService, "blog-api", "post", "create"),
    handler.CreatePost,
)
```

### Backend (Service Layer)

```go
// In business logic
hasPermission, err := rbacService.CheckUserPermission(
    tenantID, 
    userID, 
    "blog-api", 
    "post", 
    "publish",
)

if !hasPermission {
    return errors.New("permission denied")
}
```

### Frontend

```javascript
// Get user permissions
const response = await fetch(
  `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
  {credentials: 'include'}
);
const permissions = await response.json();

// Check permission
const canPublish = permissions.data.some(
  p => p.service === 'blog-api' && 
       p.entity === 'post' && 
       p.action === 'publish'
);

if (canPublish) {
  // Show publish button
}
```

## Best Practices

### 1. Principle of Least Privilege

Grant minimum necessary permissions:

```javascript
// ❌ Bad: Too broad
Role: Admin (all permissions)

// ✅ Good: Specific permissions
Role: Content Editor
Policies: Content Writer Policy
Permissions:
  - blog-api:post:create
  - blog-api:post:update
  - blog-api:post:read
```

### 2. Use Policies for Reusability

Group related permissions:

```javascript
// ❌ Bad: Assign permissions directly to roles
Role → Permission (tight coupling)

// ✅ Good: Use policies
Role → Policy → Permission (flexible)
```

### 3. Descriptive Names

Use clear, descriptive names:

```javascript
// ❌ Bad
Role: "R1"
Policy: "P1"
Permission: "tenant-api:m:c"

// ✅ Good
Role: "Content Editor"
Policy: "Content Management Policy"
Permission: "blog-api:post:create"
```

### 4. Document Custom Permissions

Keep a registry of custom permissions:

```markdown
# Custom Permissions

## Blog API
- `blog-api:post:publish` - Publish blog posts
- `blog-api:post:archive` - Archive old posts
- `blog-api:comment:moderate` - Moderate comments
```

### 5. Test Permission Changes

Always test after modifying RBAC:

```bash
# Test permission check
POST /api/v1/authorize?tenant_id=xxx&user_id=yyy&service=blog-api&entity=post&action=create
```

## Common Patterns

### Read-Only Role

```json
{
  "name": "Auditor",
  "policies": [
    {
      "name": "Read-Only Policy",
      "permissions": [
        "*:*:read",  // Future: wildcard support
        "*:*:list"
      ]
    }
  ]
}
```

### Department-Specific Role

```json
{
  "name": "Marketing Team",
  "policies": [
    {
      "name": "Marketing Content Policy",
      "permissions": [
        "blog-api:post:create",
        "blog-api:post:update",
        "media-api:image:upload",
        "analytics-api:report:view"
      ]
    }
  ]
}
```

### Temporary Access Role

```json
{
  "name": "Contractor",
  "policies": [
    {
      "name": "Limited Access Policy",
      "permissions": [
        "project-api:task:read",
        "project-api:task:update",
        // No delete or admin permissions
      ]
    }
  ]
}
```

## Troubleshooting

### Permission Denied

**Check**:
1. User is member of tenant
2. Member status is "active"
3. User's role has required policy
4. Policy has required permission

**Debug**:
```bash
# Get user permissions
GET /api/v1/permissions/user?tenant_id=xxx&user_id=yyy

# Check authorization
POST /api/v1/authorize?tenant_id=xxx&user_id=yyy&service=blog-api&entity=post&action=create
```

### Role Not Found

**Cause**: Role doesn't exist or deleted

**Solution**: Create role or use existing role

### Policy Not Working

**Check**:
1. Policy assigned to role
2. Permissions assigned to policy
3. Permission format correct

## Next Steps

- **[Roles & Policies](/guides/roles-policies)** - Managing roles and policies
- **[Permissions](/guides/permissions)** - Permission details
- **[Managing RBAC](/guides/managing-rbac)** - Best practices
- **[RBAC API](/x-api/rbac)** - API reference
