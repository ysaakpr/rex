# RBAC API

Complete API reference for Role-Based Access Control management.

## Overview

The RBAC API allows platform admins to manage the 3-tier authorization system:
- **Roles** - User positions in tenants (Admin, Writer, Viewer)
- **Policies** - Groups of permissions
- **Permissions** - Atomic actions (service:entity:action)

## Base URLs

```
Roles:       /api/v1/platform/roles
Policies:    /api/v1/platform/policies
Permissions: /api/v1/platform/permissions
```

**Authentication**: Platform Admin required for all write operations

## Roles API

### Create Role

Create a new role that can be assigned to tenant members.

**Request**:
```http
POST /api/v1/platform/roles
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "name": "Content Editor",
  "type": "tenant",
  "description": "Can create and edit content",
  "tenant_id": null
}
```

**Fields**:
- `name` (required): Role name, 2-100 characters
- `type` (required): "tenant" or "platform"
- `description` (optional): Role description, max 500 characters
- `tenant_id` (optional): For tenant-specific roles, null for system roles

**Response** (201):
```json
{
  "success": true,
  "message": "Role created successfully",
  "data": {
    "id": "role-uuid",
    "name": "Content Editor",
    "type": "tenant",
    "description": "Can create and edit content",
    "tenant_id": null,
    "is_system": true,
    "policies": [],
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

### List Roles

Get all roles, optionally filtered by tenant.

**Request**:
```http
GET /api/v1/platform/roles?tenant_id=<tenant-uuid>
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `tenant_id` (optional): Filter by tenant

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "admin-role-uuid",
      "name": "Admin",
      "type": "tenant",
      "description": "Full tenant control",
      "is_system": true,
      "created_at": "2024-11-01T10:00:00Z"
    },
    {
      "id": "writer-role-uuid",
      "name": "Writer",
      "type": "tenant",
      "description": "Can create content",
      "is_system": true,
      "created_at": "2024-11-01T10:00:00Z"
    }
  ]
}
```

### Get Role

Get detailed information about a specific role including its policies.

**Request**:
```http
GET /api/v1/platform/roles/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "role-uuid",
    "name": "Admin",
    "type": "tenant",
    "description": "Full tenant control",
    "tenant_id": null,
    "is_system": true,
    "policies": [
      {
        "id": "policy-uuid",
        "name": "Tenant Admin Policy",
        "description": "All tenant management permissions"
      }
    ],
    "created_at": "2024-11-01T10:00:00Z",
    "updated_at": "2024-11-01T10:00:00Z"
  }
}
```

### Update Role

Update role details.

**Request**:
```http
PATCH /api/v1/platform/roles/:id
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body** (all fields optional):
```json
{
  "name": "Senior Editor",
  "description": "Can edit and publish content"
}
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "role-uuid",
    "name": "Senior Editor",
    "description": "Can edit and publish content",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

### Delete Role

Delete a role (cannot delete if assigned to users).

**Request**:
```http
DELETE /api/v1/platform/roles/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Role deleted successfully"
}
```

### Assign Policies to Role

Assign one or more policies to a role.

**Request**:
```http
POST /api/v1/platform/roles/:id/policies
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "policy_ids": [
    "policy-uuid-1",
    "policy-uuid-2"
  ]
}
```

**Response** (200):
```json
{
  "success": true,
  "message": "Policies assigned successfully"
}
```

### Get Role Policies

Get all policies assigned to a role.

**Request**:
```http
GET /api/v1/platform/roles/:id/policies
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "policy-uuid-1",
      "name": "Content Writer Policy",
      "description": "Can create and edit content"
    },
    {
      "id": "policy-uuid-2",
      "name": "Media Manager Policy",
      "description": "Can upload and manage media"
    }
  ]
}
```

### Revoke Policy from Role

Remove a policy from a role.

**Request**:
```http
DELETE /api/v1/platform/roles/:id/policies/:policy_id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Policy revoked successfully"
}
```

## Policies API

### Create Policy

Create a new policy (group of permissions).

**Request**:
```http
POST /api/v1/platform/policies
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "name": "Blog Writer Policy",
  "description": "Can manage blog posts",
  "tenant_id": null
}
```

**Response** (201):
```json
{
  "success": true,
  "message": "Policy created successfully",
  "data": {
    "id": "policy-uuid",
    "name": "Blog Writer Policy",
    "description": "Can manage blog posts",
    "tenant_id": null,
    "is_system": true,
    "permissions": [],
    "created_at": "2024-11-20T10:00:00Z"
  }
}
```

### List Policies

Get all policies.

**Request**:
```http
GET /api/v1/platform/policies?tenant_id=<tenant-uuid>
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "policy-uuid",
      "name": "Tenant Admin Policy",
      "description": "All tenant management permissions",
      "is_system": true
    }
  ]
}
```

### Get Policy

Get detailed policy information including permissions.

**Request**:
```http
GET /api/v1/platform/policies/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "policy-uuid",
    "name": "Content Writer Policy",
    "description": "Can create and edit content",
    "permissions": [
      {
        "id": "perm-uuid-1",
        "service": "blog-api",
        "entity": "post",
        "action": "create",
        "key": "blog-api:post:create"
      },
      {
        "id": "perm-uuid-2",
        "service": "blog-api",
        "entity": "post",
        "action": "update",
        "key": "blog-api:post:update"
      }
    ]
  }
}
```

### Update Policy

Update policy details.

**Request**:
```http
PATCH /api/v1/platform/policies/:id
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "name": "Content Management Policy",
  "description": "Full content management access"
}
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "policy-uuid",
    "name": "Content Management Policy",
    "description": "Full content management access",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

### Delete Policy

Delete a policy.

**Request**:
```http
DELETE /api/v1/platform/policies/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Policy deleted successfully"
}
```

### Assign Permissions to Policy

Assign permissions to a policy.

**Request**:
```http
POST /api/v1/platform/policies/:id/permissions
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "permission_ids": [
    "permission-uuid-1",
    "permission-uuid-2",
    "permission-uuid-3"
  ]
}
```

**Response** (200):
```json
{
  "success": true,
  "message": "Permissions assigned successfully"
}
```

### Revoke Permission from Policy

Remove a permission from a policy.

**Request**:
```http
DELETE /api/v1/platform/policies/:id/permissions/:permission_id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Permission revoked successfully"
}
```

## Permissions API

### Create Permission

Create a new permission.

**Request**:
```http
POST /api/v1/platform/permissions
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "service": "blog-api",
  "entity": "post",
  "action": "publish",
  "description": "Publish blog posts"
}
```

**Fields**:
- `service` (required): Service name (e.g., "tenant-api", "blog-api")
- `entity` (required): Resource type (e.g., "member", "post")
- `action` (required): Operation (e.g., "create", "read", "update", "delete")
- `description` (optional): Human-readable description

**Response** (201):
```json
{
  "success": true,
  "message": "Permission created successfully",
  "data": {
    "id": "permission-uuid",
    "service": "blog-api",
    "entity": "post",
    "action": "publish",
    "description": "Publish blog posts",
    "key": "blog-api:post:publish",
    "created_at": "2024-11-20T10:00:00Z"
  }
}
```

### List Permissions

Get all permissions, optionally filtered by service.

**Request**:
```http
GET /api/v1/platform/permissions?service=blog-api
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `service` (optional): Filter by service name

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "perm-uuid-1",
      "service": "blog-api",
      "entity": "post",
      "action": "create",
      "description": "Create blog posts",
      "key": "blog-api:post:create"
    },
    {
      "id": "perm-uuid-2",
      "service": "blog-api",
      "entity": "post",
      "action": "publish",
      "description": "Publish blog posts",
      "key": "blog-api:post:publish"
    }
  ]
}
```

### Get Permission

Get detailed permission information.

**Request**:
```http
GET /api/v1/platform/permissions/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "permission-uuid",
    "service": "blog-api",
    "entity": "post",
    "action": "publish",
    "description": "Publish blog posts",
    "key": "blog-api:post:publish",
    "created_at": "2024-11-20T10:00:00Z"
  }
}
```

### Delete Permission

Delete a permission.

**Request**:
```http
DELETE /api/v1/platform/permissions/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Permission deleted successfully"
}
```

## Authorization API

### Check Authorization

Check if a user has a specific permission in a tenant.

**Request**:
```http
POST /api/v1/authorize?tenant_id=<uuid>&user_id=<id>&service=blog-api&entity=post&action=publish
Authorization: Bearer <token>
```

**Query Parameters**:
- `tenant_id` (required): Tenant UUID
- `user_id` (required): User ID
- `service` (required): Service name
- `entity` (required): Entity/resource type
- `action` (required): Action to check

**Response** (200):
```json
{
  "success": true,
  "data": {
    "authorized": true
  }
}
```

### Get User Permissions

Get all permissions for a user in a specific tenant.

**Request**:
```http
GET /api/v1/permissions/user?tenant_id=<uuid>&user_id=<id>
Authorization: Bearer <token>
```

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "perm-uuid-1",
      "service": "tenant-api",
      "entity": "member",
      "action": "invite",
      "key": "tenant-api:member:invite"
    },
    {
      "id": "perm-uuid-2",
      "service": "blog-api",
      "entity": "post",
      "action": "create",
      "key": "blog-api:post:create"
    }
  ]
}
```

## Complete Example: Custom RBAC Setup

```javascript
// 1. Create Permission
const createPermResp = await fetch('/api/v1/platform/permissions', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    service: 'blog-api',
    entity: 'post',
    action: 'publish',
    description: 'Publish blog posts'
  })
});
const permission = await createPermResp.json();

// 2. Create Policy
const createPolicyResp = await fetch('/api/v1/platform/policies', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Blog Publisher Policy',
    description: 'Can publish blog posts'
  })
});
const policy = await createPolicyResp.json();

// 3. Assign Permission to Policy
await fetch(`/api/v1/platform/policies/${policy.data.id}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    permission_ids: [permission.data.id]
  })
});

// 4. Create Role
const createRoleResp = await fetch('/api/v1/platform/roles', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Publisher',
    type: 'tenant',
    description: 'Can publish content'
  })
});
const role = await createRoleResp.json();

// 5. Assign Policy to Role
await fetch(`/api/v1/platform/roles/${role.data.id}/policies`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    policy_ids: [policy.data.id]
  })
});

// 6. Assign Role to User
await fetch(`/api/v1/tenants/${tenantId}/members/${userId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    role_id: role.data.id
  })
});
```

## Next Steps

- [RBAC Overview](/guides/rbac-overview) - Understanding RBAC
- [Roles & Policies](/guides/roles-policies) - Managing roles and policies
- [Permissions](/guides/permissions) - Permission details
