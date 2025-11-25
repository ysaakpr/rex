# Permissions

Complete guide to managing permissions in the RBAC system.

## Overview

Permissions are the atomic units of authorization in the system. Each permission represents a single action that can be performed.

**Format**: `service:entity:action`

**Examples**:
- `blog-api:post:create` - Create blog posts
- `tenant-api:member:invite` - Invite tenant members
- `media-api:image:delete` - Delete images

## Permission Structure

### Components

1. **Service**: The API or service providing the resource
   - `tenant-api` - Tenant management
   - `blog-api` - Blog/content management
   - `media-api` - Media management
   - `analytics-api` - Analytics and reporting

2. **Entity**: The resource type
   - `member` - Tenant members
   - `post` - Blog posts
   - `comment` - Comments
   - `image` - Images

3. **Action**: The operation
   - `create` - Create new resource
   - `read` - View resource
   - `update` - Modify resource
   - `delete` - Remove resource
   - `publish` - Publish content (custom)
   - `approve` - Approve submissions (custom)

### Permission Key

The full permission is automatically generated: `{service}:{entity}:{action}`

## Creating Permissions

### Basic Permission

```javascript
const response = await fetch('/api/v1/platform/permissions', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    service: 'blog-api',
    entity: 'post',
    action: 'create',
    description: 'Create new blog posts'
  })
});

const {data} = await response.json();
console.log('Permission key:', data.key); // blog-api:post:create
```

### Batch Create Permissions

```javascript
async function createServicePermissions(service, entities, actions) {
  const permissions = [];
  
  for (const entity of entities) {
    for (const action of actions) {
      const resp = await fetch('/api/v1/platform/permissions', {
        method: 'POST',
        credentials: 'include',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
          service,
          entity,
          action,
          description: `${action} ${entity} in ${service}`
        })
      });
      const {data} = await resp.json();
      permissions.push(data);
    }
  }
  
  return permissions;
}

// Usage: Create CRUD permissions for blog entities
const blogPerms = await createServicePermissions(
  'blog-api',
  ['post', 'comment', 'category'],
  ['create', 'read', 'update', 'delete']
);
```

## Listing Permissions

### Get All Permissions

```javascript
const response = await fetch('/api/v1/platform/permissions', {
  credentials: 'include'
});

const {data} = await response.json();
data.forEach(perm => {
  console.log(`${perm.key}: ${perm.description}`);
});
```

### Filter by Service

```javascript
const response = await fetch('/api/v1/platform/permissions?service=blog-api', {
  credentials: 'include'
});

const {data} = await response.json();
console.log(`Found ${data.length} blog-api permissions`);
```

### Get Permission Details

```javascript
const response = await fetch(`/api/v1/platform/permissions/${permissionId}`, {
  credentials: 'include'
});

const {data} = await response.json();
console.log('Permission:', {
  key: data.key,
  service: data.service,
  entity: data.entity,
  action: data.action,
  description: data.description
});
```

## Assigning Permissions

Permissions are assigned to **Policies**, not directly to roles or users.

```
User → Role → Policy → Permissions
```

### Assign to Policy

```javascript
await fetch(`/api/v1/platform/policies/${policyId}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    permission_ids: [
      'perm-uuid-1',
      'perm-uuid-2',
      'perm-uuid-3'
    ]
  })
});
```

### Remove from Policy

```javascript
await fetch(`/api/v1/platform/policies/${policyId}/permissions/${permissionId}`, {
  method: 'DELETE',
  credentials: 'include'
});
```

## Checking Permissions

### Check Single Permission

```javascript
const response = await fetch(
  `/api/v1/authorize?` +
  `tenant_id=${tenantId}&` +
  `user_id=${userId}&` +
  `service=blog-api&` +
  `entity=post&` +
  `action=create`,
  {credentials: 'include'}
);

const {data} = await response.json();

if (data.authorized) {
  console.log('User can create posts');
} else {
  console.log('Access denied');
}
```

### Get All User Permissions

```javascript
const response = await fetch(
  `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
  {credentials: 'include'}
);

const {data} = await response.json();

// Create permission lookup
const userPermissions = new Set(data.map(p => p.key));

// Check permission
if (userPermissions.has('blog-api:post:delete')) {
  console.log('User can delete posts');
}
```

## Standard Permission Sets

### Tenant Management Permissions

```javascript
const tenantPermissions = {
  service: 'tenant-api',
  entities: {
    tenant: ['create', 'read', 'update', 'delete'],
    member: ['invite', 'read', 'update', 'remove'],
    settings: ['read', 'update']
  }
};
```

### Content Management Permissions

```javascript
const contentPermissions = {
  service: 'blog-api',
  entities: {
    post: ['create', 'read', 'update', 'delete', 'publish'],
    comment: ['create', 'read', 'update', 'delete', 'approve'],
    category: ['create', 'read', 'update', 'delete']
  }
};
```

### Media Management Permissions

```javascript
const mediaPermissions = {
  service: 'media-api',
  entities: {
    image: ['upload', 'read', 'update', 'delete'],
    video: ['upload', 'read', 'delete'],
    file: ['upload', 'read', 'delete']
  }
};
```

## Frontend Permission Checks

### React Hook: usePermission

```jsx
import {useState, useEffect} from 'react';

function usePermission(tenantId, service, entity, action) {
  const [hasPermission, setHasPermission] = useState(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    fetch(
      `/api/v1/authorize?` +
      `tenant_id=${tenantId}&` +
      `service=${service}&` +
      `entity=${entity}&` +
      `action=${action}`,
      {credentials: 'include'}
    )
      .then(res => res.json())
      .then(data => {
        setHasPermission(data.data.authorized);
        setLoading(false);
      })
      .catch(() => {
        setHasPermission(false);
        setLoading(false);
      });
  }, [tenantId, service, entity, action]);
  
  return {hasPermission, loading};
}

// Usage
function CreatePostButton({tenantId}) {
  const {hasPermission, loading} = usePermission(
    tenantId,
    'blog-api',
    'post',
    'create'
  );
  
  if (loading) return <button disabled>Loading...</button>;
  if (!hasPermission) return null;
  
  return <button onClick={createPost}>Create Post</button>;
}
```

### React Context: PermissionProvider

```jsx
import {createContext, useContext, useEffect, useState} from 'react';

const PermissionContext = createContext(null);

export function PermissionProvider({children, tenantId, userId}) {
  const [permissions, setPermissions] = useState(new Set());
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    fetch(
      `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
      {credentials: 'include'}
    )
      .then(res => res.json())
      .then(data => {
        const permSet = new Set(data.data.map(p => p.key));
        setPermissions(permSet);
        setLoading(false);
      });
  }, [tenantId, userId]);
  
  const hasPermission = (service, entity, action) => {
    return permissions.has(`${service}:${entity}:${action}`);
  };
  
  return (
    <PermissionContext.Provider value={{hasPermission, loading}}>
      {children}
    </PermissionContext.Provider>
  );
}

export function usePermissions() {
  return useContext(PermissionContext);
}

// Usage
function App() {
  return (
    <PermissionProvider tenantId={tenantId} userId={userId}>
      <Dashboard />
    </PermissionProvider>
  );
}

function Dashboard() {
  const {hasPermission} = usePermissions();
  
  return (
    <div>
      {hasPermission('blog-api', 'post', 'create') && (
        <button>Create Post</button>
      )}
      {hasPermission('tenant-api', 'member', 'invite') && (
        <button>Invite Member</button>
      )}
    </div>
  );
}
```

### Conditional Rendering Component

```jsx
function Can({service, entity, action, children, fallback = null}) {
  const {hasPermission, loading} = usePermissions();
  
  if (loading) return fallback;
  if (!hasPermission(service, entity, action)) return fallback;
  
  return children;
}

// Usage
<Can service="blog-api" entity="post" action="delete">
  <button onClick={deletePost}>Delete</button>
</Can>

<Can
  service="tenant-api"
  entity="member"
  action="invite"
  fallback={<div>You cannot invite members</div>}
>
  <InviteMemberForm />
</Can>
```

## Backend Permission Checks

### Go Middleware

```go
// Require specific permission
func RequirePermission(service, entity, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.Param("tenant_id")
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        // Check permission
        authorized, err := rbacService.CheckUserPermission(
            userID,
            tenantID,
            service,
            entity,
            action,
        )
        
        if err != nil || !authorized {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error": "Permission denied",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Usage in routes
router.POST("/tenants/:tenant_id/posts",
    middleware.RequirePermission("blog-api", "post", "create"),
    handlers.CreatePost,
)

router.DELETE("/tenants/:tenant_id/posts/:post_id",
    middleware.RequirePermission("blog-api", "post", "delete"),
    handlers.DeletePost,
)
```

### Go Helper Function

```go
func HasPermission(c *gin.Context, service, entity, action string) bool {
    tenantID := c.Param("tenant_id")
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    userID := sessionContainer.GetUserID()
    
    authorized, _ := rbacService.CheckUserPermission(
        userID,
        tenantID,
        service,
        entity,
        action,
    )
    
    return authorized
}

// Usage in handler
func UpdatePostHandler(c *gin.Context) {
    if !HasPermission(c, "blog-api", "post", "update") {
        c.JSON(403, gin.H{"error": "Cannot update posts"})
        return
    }
    
    // ... handle update
}
```

## Permission Naming Conventions

### Service Names

Use lowercase with hyphens:
- ✅ `blog-api`, `tenant-api`, `media-api`
- ❌ `BlogAPI`, `Tenant_API`, `mediaapi`

### Entity Names

Use singular nouns:
- ✅ `post`, `member`, `comment`
- ❌ `posts`, `Members`, `COMMENT`

### Action Names

Use standard CRUD + custom actions:
- ✅ `create`, `read`, `update`, `delete`, `publish`, `approve`
- ❌ `add`, `view`, `edit`, `remove` (use standard CRUD)

## Complete Permission Setup Example

```javascript
async function setupCompletePermissions() {
  // Define permission structure
  const permissionStructure = {
    'tenant-api': {
      'tenant': ['create', 'read', 'update', 'delete'],
      'member': ['invite', 'read', 'update', 'remove'],
      'settings': ['read', 'update']
    },
    'blog-api': {
      'post': ['create', 'read', 'update', 'delete', 'publish'],
      'comment': ['create', 'read', 'update', 'delete', 'moderate'],
      'category': ['create', 'read', 'update', 'delete']
    },
    'media-api': {
      'image': ['upload', 'read', 'delete'],
      'video': ['upload', 'read', 'delete']
    }
  };
  
  const createdPermissions = {};
  
  // Create all permissions
  for (const [service, entities] of Object.entries(permissionStructure)) {
    createdPermissions[service] = {};
    
    for (const [entity, actions] of Object.entries(entities)) {
      createdPermissions[service][entity] = {};
      
      for (const action of actions) {
        const resp = await fetch('/api/v1/platform/permissions', {
          method: 'POST',
          credentials: 'include',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({
            service,
            entity,
            action,
            description: `${action} ${entity} in ${service}`
          })
        });
        
        const {data} = await resp.json();
        createdPermissions[service][entity][action] = data.id;
        console.log(`Created: ${data.key}`);
      }
    }
  }
  
  return createdPermissions;
}
```

## Deleting Permissions

```javascript
const response = await fetch(`/api/v1/platform/permissions/${permissionId}`, {
  method: 'DELETE',
  credentials: 'include'
});

if (response.ok) {
  console.log('Permission deleted');
}
```

:::warning Cascade Effect
Deleting a permission removes it from all policies. Users with those policies will lose this permission immediately.
:::

## Permission Auditing

### Log Permission Checks

```javascript
function logPermissionCheck(userId, tenantId, permission, granted) {
  console.log('[Permission Check]', {
    user_id: userId,
    tenant_id: tenantId,
    permission,
    granted,
    timestamp: new Date().toISOString()
  });
  
  // Send to analytics/logging service
  analytics.track('permission_check', {
    userId,
    tenantId,
    permission,
    granted
  });
}
```

### Generate Permission Report

```javascript
async function generatePermissionReport(tenantId) {
  // Get all members
  const membersResp = await fetch(`/api/v1/tenants/${tenantId}/members`, {
    credentials: 'include'
  });
  const members = await membersResp.json();
  
  const report = [];
  
  for (const member of members.data.data) {
    // Get user permissions
    const permsResp = await fetch(
      `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${member.user_id}`,
      {credentials: 'include'}
    );
    const perms = await permsResp.json();
    
    report.push({
      user: member.email,
      role: member.role_name,
      permissions: perms.data.map(p => p.key)
    });
  }
  
  return report;
}

// Usage
const report = await generatePermissionReport(tenantId);
console.table(report);
```

## Best Practices

### 1. Use Standard CRUD Actions

Prefer standard actions for consistency:
- `create`, `read`, `update`, `delete`

Add custom actions only when necessary:
- `publish`, `approve`, `moderate`, `export`

### 2. Granular Permissions

Create specific permissions rather than broad ones:

✅ **Good**:
```
blog-api:post:create
blog-api:post:publish
blog-api:comment:moderate
```

❌ **Bad**:
```
blog-api:*:*
admin:all:all
```

### 3. Permission Discovery

Document all available permissions:

```javascript
// permissions.js
export const PERMISSIONS = {
  BLOG: {
    POST_CREATE: 'blog-api:post:create',
    POST_UPDATE: 'blog-api:post:update',
    POST_DELETE: 'blog-api:post:delete',
    POST_PUBLISH: 'blog-api:post:publish'
  },
  TENANT: {
    MEMBER_INVITE: 'tenant-api:member:invite',
    MEMBER_REMOVE: 'tenant-api:member:remove'
  }
};

// Usage
if (userPermissions.has(PERMISSIONS.BLOG.POST_CREATE)) {
  // ...
}
```

### 4. Permission Caching

Cache user permissions in frontend:

```javascript
class PermissionCache {
  constructor() {
    this.cache = new Map();
    this.ttl = 5 * 60 * 1000; // 5 minutes
  }
  
  async getPermissions(tenantId, userId) {
    const key = `${tenantId}:${userId}`;
    const cached = this.cache.get(key);
    
    if (cached && Date.now() - cached.timestamp < this.ttl) {
      return cached.permissions;
    }
    
    const resp = await fetch(
      `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
      {credentials: 'include'}
    );
    const {data} = await resp.json();
    
    this.cache.set(key, {
      permissions: new Set(data.map(p => p.key)),
      timestamp: Date.now()
    });
    
    return this.cache.get(key).permissions;
  }
  
  invalidate(tenantId, userId) {
    this.cache.delete(`${tenantId}:${userId}`);
  }
}
```

## Troubleshooting

### Permission Not Working

1. **Check permission exists**:
```javascript
const perms = await fetch('/api/v1/platform/permissions?service=blog-api', {
  credentials: 'include'
});
console.log(await perms.json());
```

2. **Check permission is in policy**:
```javascript
const policy = await fetch(`/api/v1/platform/policies/${policyId}`, {
  credentials: 'include'
});
console.log('Policy permissions:', policy.data.permissions);
```

3. **Check policy is assigned to role**:
```javascript
const role = await fetch(`/api/v1/platform/roles/${roleId}/policies`, {
  credentials: 'include'
});
console.log('Role policies:', role.data);
```

4. **Check user has role**:
```javascript
const member = await fetch(
  `/api/v1/tenants/${tenantId}/members/${userId}`,
  {credentials: 'include'}
);
console.log('User role:', member.data.role_name);
```

## Next Steps

- [RBAC Overview](/guides/rbac-overview) - System architecture
- [Roles & Policies](/guides/roles-policies) - Managing roles and policies
- [RBAC API](/x-api/rbac) - API reference
- [Backend Integration](/guides/backend-integration) - Implementing permission checks
