# RBAC Authorization Guide

**Last Updated**: November 24, 2025  
**Version**: 1.0

## Overview

This guide explains how to use the RBAC (Role-Based Access Control) authorization system to check user permissions. The `/api/v1/authorize` endpoint is the core of the permission checking system, allowing both backend services and frontend applications to verify if a user has permission to perform specific actions.

## Table of Contents

- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
- [API Reference](#api-reference)
- [Backend Implementation](#backend-implementation)
- [Frontend Implementation](#frontend-implementation)
- [Permission Format](#permission-format)
- [Best Practices](#best-practices)
- [Performance Optimization](#performance-optimization)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Check if User Can Create Members

```bash
curl -X POST 'http://localhost:8080/api/v1/authorize?tenant_id=123e4567-e89b-12d3-a456-426614174000&user_id=user_abc123&service=tenant-api&entity=member&action=create' \
  -H 'Cookie: sAccessToken=eyJhbG...'
```

**Response:**

```json
{
  "success": true,
  "data": {
    "authorized": true
  }
}
```

---

## How It Works

### Authorization Flow

The authorize endpoint checks permissions by traversing a chain of relationships in the database:

```
User → Tenant Member → Role → Policy → Permission
```

### Database Query Path

When you call the authorize endpoint, it performs the following SQL query:

```sql
SELECT COUNT(DISTINCT p.id)
FROM permissions p
INNER JOIN policy_permissions pp ON pp.permission_id = p.id
INNER JOIN role_policies rp ON rp.policy_id = pp.policy_id
INNER JOIN tenant_members tm ON tm.role_id = rp.role_id
WHERE tm.tenant_id = ?           -- Tenant context
  AND tm.user_id = ?             -- User being checked
  AND tm.status = 'active'       -- Only active members
  AND p.service = ?              -- Service name
  AND p.entity = ?               -- Resource type
  AND p.action = ?               -- Action to perform
```

### Step-by-Step Example

**Scenario**: Check if Alice can create members in Tenant ABC

```
┌─────────────────────────────────────────────────────────────┐
│ Request                                                      │
│ Can Alice create members in Tenant ABC?                     │
│                                                              │
│ tenant_id = abc                                              │
│ user_id = alice                                              │
│ service = tenant-api                                         │
│ entity = member                                              │
│ action = create                                              │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Find User's Tenant Membership                       │
│                                                              │
│ Query: tenant_members table                                  │
│ Result: Alice is an "Admin" in Tenant ABC                   │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Find Role's Policies                                │
│                                                              │
│ Query: role_policies table                                   │
│ Result: Admin role has "FullAccess" policy                  │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Find Policy's Permissions                           │
│                                                              │
│ Query: policy_permissions table                              │
│ Result: FullAccess includes tenant-api:member:create        │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Verify Permission Match                             │
│                                                              │
│ Check: Does permission match requested action?              │
│ Result: ✅ MATCH FOUND - User is authorized                 │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Response                                                     │
│ {"success": true, "data": {"authorized": true}}             │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Permissions**: Atomic actions (e.g., `tenant-api:member:create`)
2. **Policies**: Groups of permissions (e.g., "FullAccess" contains multiple permissions)
3. **Roles**: User types in a tenant (e.g., "Admin", "Writer", "Viewer")
4. **Tenant Members**: Links users to tenants with a specific role

---

## API Reference

### Endpoint

```
POST /api/v1/authorize
```

### Authentication

**Required**: Valid SuperTokens session (cookie or header-based)

### Query Parameters

| Parameter   | Type   | Required | Description                                    |
|-------------|--------|----------|------------------------------------------------|
| `tenant_id` | UUID   | Yes      | Tenant context for permission check            |
| `user_id`   | String | Yes      | SuperTokens user ID                            |
| `service`   | String | Yes      | Service name (e.g., `tenant-api`)              |
| `entity`    | String | Yes      | Resource type (e.g., `member`, `tenant`)       |
| `action`    | String | Yes      | Action to perform (e.g., `create`, `read`)     |

### Response Format

#### Success (User has permission)

```json
{
  "success": true,
  "data": {
    "authorized": true
  }
}
```

**HTTP Status**: `200 OK`

#### Success (User lacks permission)

```json
{
  "success": true,
  "data": {
    "authorized": false
  }
}
```

**HTTP Status**: `200 OK`

**Note**: The endpoint always returns `200 OK`. Check the `authorized` field to determine permission status.

#### Error (Missing parameters)

```json
{
  "success": false,
  "error": "missing required query parameters"
}
```

**HTTP Status**: `400 Bad Request`

#### Error (Unauthorized)

```json
{
  "success": false,
  "error": "Unauthorized"
}
```

**HTTP Status**: `401 Unauthorized`

---

## Backend Implementation

### 1. Using the Authorize Endpoint Directly

```go
import (
    "fmt"
    "net/http"
    "encoding/json"
)

func checkUserPermission(tenantID, userID, service, entity, action string) (bool, error) {
    url := fmt.Sprintf("http://localhost:8080/api/v1/authorize?tenant_id=%s&user_id=%s&service=%s&entity=%s&action=%s",
        tenantID, userID, service, entity, action)
    
    resp, err := http.Post(url, "application/json", nil)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Success bool `json:"success"`
        Data    struct {
            Authorized bool `json:"authorized"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, err
    }
    
    return result.Data.Authorized, nil
}

// Usage
hasPermission, err := checkUserPermission(
    "123e4567-e89b-12d3-a456-426614174000",
    "user_abc123",
    "tenant-api",
    "member",
    "create",
)

if !hasPermission {
    return c.JSON(http.StatusForbidden, gin.H{
        "error": "Permission denied",
    })
}
```

### 2. Using RBAC Service (Internal Backend)

If you're within the backend application, use the service layer directly:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/ysaakpr/rex/internal/services"
)

func (h *MemberHandler) AddMember(c *gin.Context) {
    // Get context
    userID, _ := c.Get("userID")
    tenantID, _ := c.Get("tenantID")
    
    // Check permission using service
    hasPermission, err := h.rbacService.CheckUserPermission(
        tenantID.(uuid.UUID),
        userID.(string),
        "tenant-api",
        "member",
        "create",
    )
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check permission"})
        return
    }
    
    if !hasPermission {
        c.JSON(403, gin.H{"error": "Permission denied: tenant-api:member:create"})
        return
    }
    
    // Proceed with member creation
    // ...
}
```

### 3. Using RBAC Middleware (Recommended)

The cleanest approach for backend routes:

```go
import (
    "github.com/ysaakpr/rex/internal/api/middleware"
    "github.com/ysaakpr/rex/internal/api/router"
)

// In router setup
func SetupRouter(deps *router.RouterDeps) *gin.Engine {
    // ...
    
    tenantScoped := tenants.Group("/:id")
    tenantScoped.Use(middleware.TenantAccessMiddleware(deps.MemberRepo))
    {
        // This route requires tenant-api:member:create permission
        tenantScoped.POST("/members",
            middleware.RequirePermission(deps.RBACService, "tenant-api", "member", "create"),
            deps.MemberHandler.AddMember,
        )
        
        // This route requires tenant-api:member:read permission
        tenantScoped.GET("/members",
            middleware.RequirePermission(deps.RBACService, "tenant-api", "member", "read"),
            deps.MemberHandler.ListMembers,
        )
        
        // This route requires tenant-api:member:delete permission
        tenantScoped.DELETE("/members/:user_id",
            middleware.RequirePermission(deps.RBACService, "tenant-api", "member", "delete"),
            deps.MemberHandler.RemoveMember,
        )
    }
    
    // ...
}
```

**Middleware automatically:**
- Extracts `userID` and `tenantID` from context
- Checks the specified permission
- Returns `403 Forbidden` if user lacks permission
- Continues to handler if user has permission

### 4. Checking Multiple Permissions

```go
func (h *Handler) ComplexOperation(c *gin.Context) {
    userID, _ := c.Get("userID")
    tenantID, _ := c.Get("tenantID")
    
    // Check if user has ANY of these permissions
    permissions := []struct {
        service string
        entity  string
        action  string
    }{
        {"tenant-api", "member", "update"},
        {"tenant-api", "admin", "all"},
    }
    
    hasAnyPermission := false
    for _, perm := range permissions {
        has, _ := h.rbacService.CheckUserPermission(
            tenantID.(uuid.UUID),
            userID.(string),
            perm.service,
            perm.entity,
            perm.action,
        )
        if has {
            hasAnyPermission = true
            break
        }
    }
    
    if !hasAnyPermission {
        c.JSON(403, gin.H{"error": "Insufficient permissions"})
        return
    }
    
    // Proceed with operation
    // ...
}
```

---

## Frontend Implementation

### 1. Simple Permission Check

```javascript
/**
 * Check if user has permission to perform an action
 */
async function checkPermission(tenantId, userId, service, entity, action) {
  try {
    const response = await fetch(
      `/api/v1/authorize?tenant_id=${tenantId}&user_id=${userId}&service=${service}&entity=${entity}&action=${action}`,
      {
        method: 'POST',
        credentials: 'include', // Send cookies
      }
    );
    
    const result = await response.json();
    return result.data?.authorized || false;
  } catch (error) {
    console.error('Permission check failed:', error);
    return false;
  }
}

// Usage
const canCreateMembers = await checkPermission(
  tenantId,
  userId,
  'tenant-api',
  'member',
  'create'
);

if (canCreateMembers) {
  showAddMemberButton();
}
```

### 2. React Hook for Permission Checking

```jsx
import { useState, useEffect } from 'react';

/**
 * Custom hook to check user permissions
 */
function usePermission(tenantId, userId, service, entity, action) {
  const [authorized, setAuthorized] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    async function checkPermission() {
      try {
        setLoading(true);
        const response = await fetch(
          `/api/v1/authorize?tenant_id=${tenantId}&user_id=${userId}&service=${service}&entity=${entity}&action=${action}`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );
        
        const result = await response.json();
        setAuthorized(result.data?.authorized || false);
      } catch (err) {
        console.error('Permission check failed:', err);
        setError(err);
        setAuthorized(false);
      } finally {
        setLoading(false);
      }
    }
    
    if (tenantId && userId) {
      checkPermission();
    }
  }, [tenantId, userId, service, entity, action]);
  
  return { authorized, loading, error };
}

// Usage in component
function MembersList({ tenantId, userId }) {
  const { authorized: canCreate, loading } = usePermission(
    tenantId,
    userId,
    'tenant-api',
    'member',
    'create'
  );
  
  if (loading) {
    return <div>Loading permissions...</div>;
  }
  
  return (
    <div>
      <h2>Members</h2>
      {canCreate && (
        <button onClick={handleAddMember}>Add Member</button>
      )}
      {/* Member list */}
    </div>
  );
}
```

### 3. Permission Context Provider (Advanced)

```jsx
import React, { createContext, useContext, useState, useEffect } from 'react';

const PermissionContext = createContext({});

/**
 * Provider that fetches and caches all user permissions for a tenant
 */
export function PermissionProvider({ tenantId, userId, children }) {
  const [permissions, setPermissions] = useState({});
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    async function fetchPermissions() {
      try {
        // Fetch all permissions for this user in this tenant
        const response = await fetch(
          `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
          { credentials: 'include' }
        );
        
        const result = await response.json();
        
        // Convert array to map for O(1) lookup
        const permMap = {};
        result.data.forEach(perm => {
          const key = `${perm.service}:${perm.entity}:${perm.action}`;
          permMap[key] = true;
        });
        
        setPermissions(permMap);
      } catch (error) {
        console.error('Failed to fetch permissions:', error);
      } finally {
        setLoading(false);
      }
    }
    
    if (tenantId && userId) {
      fetchPermissions();
    }
  }, [tenantId, userId]);
  
  const hasPermission = (service, entity, action) => {
    const key = `${service}:${entity}:${action}`;
    return permissions[key] || false;
  };
  
  return (
    <PermissionContext.Provider value={{ hasPermission, loading }}>
      {children}
    </PermissionContext.Provider>
  );
}

/**
 * Hook to check permission from context
 */
export function useHasPermission(service, entity, action) {
  const { hasPermission, loading } = useContext(PermissionContext);
  return {
    authorized: hasPermission(service, entity, action),
    loading,
  };
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
  const { authorized: canCreateMembers } = useHasPermission('tenant-api', 'member', 'create');
  const { authorized: canDeleteMembers } = useHasPermission('tenant-api', 'member', 'delete');
  
  return (
    <div>
      {canCreateMembers && <button>Add Member</button>}
      {canDeleteMembers && <button>Delete Member</button>}
    </div>
  );
}
```

### 4. Conditional Rendering Component

```jsx
/**
 * Component that only renders children if user has permission
 */
function ProtectedAction({ tenantId, userId, service, entity, action, fallback, children }) {
  const { authorized, loading } = usePermission(tenantId, userId, service, entity, action);
  
  if (loading) {
    return null; // or loading spinner
  }
  
  if (!authorized) {
    return fallback || null;
  }
  
  return <>{children}</>;
}

// Usage
function MemberActions({ tenantId, userId, member }) {
  return (
    <div>
      <ProtectedAction
        tenantId={tenantId}
        userId={userId}
        service="tenant-api"
        entity="member"
        action="update"
      >
        <button>Edit Member</button>
      </ProtectedAction>
      
      <ProtectedAction
        tenantId={tenantId}
        userId={userId}
        service="tenant-api"
        entity="member"
        action="delete"
        fallback={<span>No delete permission</span>}
      >
        <button>Delete Member</button>
      </ProtectedAction>
    </div>
  );
}
```

### 5. Button with Built-in Permission Check

```jsx
/**
 * Button that automatically checks permission and disables if user lacks access
 */
function PermissionButton({ 
  tenantId, 
  userId, 
  service, 
  entity, 
  action, 
  onClick, 
  children,
  ...props 
}) {
  const { authorized, loading } = usePermission(tenantId, userId, service, entity, action);
  
  return (
    <button
      onClick={onClick}
      disabled={loading || !authorized}
      title={!authorized ? 'You do not have permission for this action' : ''}
      {...props}
    >
      {loading ? 'Checking...' : children}
    </button>
  );
}

// Usage
<PermissionButton
  tenantId={tenantId}
  userId={userId}
  service="tenant-api"
  entity="member"
  action="create"
  onClick={handleAddMember}
>
  Add Member
</PermissionButton>
```

---

## Permission Format

Permissions use the format: **`service:entity:action`**

### Service

The API or system component (kebab-case):

- `tenant-api` - Tenant management API
- `billing-api` - Billing and invoicing
- `analytics-api` - Analytics and reporting
- `user-api` - User management

### Entity

The resource type being accessed (singular, kebab-case):

- `tenant` - Tenant resources
- `member` - Team members
- `role` - User roles
- `permission` - Permissions
- `invoice` - Billing invoices
- `report` - Analytics reports

### Action

The operation being performed:

- `create` - Create new resource
- `read` - View/list resources
- `update` - Modify existing resource
- `delete` - Remove resource
- `all` - All operations (admin-level)

### Common Permission Examples

```
tenant-api:tenant:create       # Create tenants
tenant-api:tenant:read         # View tenant details
tenant-api:tenant:update       # Edit tenant settings
tenant-api:tenant:delete       # Delete tenants

tenant-api:member:create       # Add members to tenant
tenant-api:member:read         # View member list
tenant-api:member:update       # Edit member roles
tenant-api:member:delete       # Remove members

tenant-api:invitation:create   # Send invitations
tenant-api:invitation:read     # View invitations
tenant-api:invitation:delete   # Cancel invitations

billing-api:invoice:create     # Create invoices
billing-api:invoice:read       # View invoices
billing-api:invoice:update     # Edit invoices
billing-api:invoice:delete     # Delete invoices

analytics-api:report:read      # View analytics reports
analytics-api:report:create    # Generate custom reports
```

---

## Best Practices

### 1. Frontend: UI vs API Security

**❌ BAD**: Rely only on UI permission checks

```javascript
// Only hiding the button - API is still accessible!
if (canDelete) {
  showDeleteButton();
}
```

**✅ GOOD**: Use both UI and API checks

```javascript
// Hide button in UI
if (canDelete) {
  showDeleteButton();
}

// Also check permission on API request
async function deleteMember(memberId) {
  const canDelete = await checkPermission(..., 'delete');
  if (!canDelete) {
    alert('Permission denied');
    return;
  }
  
  // Make API call (backend will also verify)
  await fetch(`/api/v1/tenants/${tenantId}/members/${memberId}`, {
    method: 'DELETE',
    credentials: 'include',
  });
}
```

**Remember**: Frontend checks improve UX, but backend checks enforce security.

### 2. Backend: Always Verify at API Level

**❌ BAD**: Trust client-side permission checks

```go
// No permission check - assumes frontend checked
func (h *Handler) DeleteMember(c *gin.Context) {
    // Directly delete without verification
    h.service.DeleteMember(...)
}
```

**✅ GOOD**: Verify permission in middleware or handler

```go
// Middleware checks permission before handler executes
tenantScoped.DELETE("/members/:user_id",
    middleware.RequirePermission(rbacService, "tenant-api", "member", "delete"),
    h.DeleteMember,
)
```

### 3. Cache Permissions (Frontend)

**❌ BAD**: Check permission on every render

```jsx
function MyComponent() {
  const [canCreate, setCanCreate] = useState(false);
  
  // This runs on EVERY render!
  useEffect(() => {
    checkPermission(...).then(setCanCreate);
  });
  
  return canCreate ? <button>Add</button> : null;
}
```

**✅ GOOD**: Cache permissions and check dependencies

```jsx
function MyComponent() {
  const [canCreate, setCanCreate] = useState(false);
  
  // Only runs when tenantId or userId changes
  useEffect(() => {
    checkPermission(...).then(setCanCreate);
  }, [tenantId, userId]);
  
  return canCreate ? <button>Add</button> : null;
}
```

### 4. Handle Permission Errors Gracefully

**❌ BAD**: Silent failures

```javascript
async function checkPermission(...) {
  try {
    const response = await fetch(...);
    return response.json().data.authorized;
  } catch (error) {
    return false; // Silent failure - no logging
  }
}
```

**✅ GOOD**: Log errors and provide fallback

```javascript
async function checkPermission(...) {
  try {
    const response = await fetch(...);
    if (!response.ok) {
      console.error('Permission check failed:', response.status);
      return false;
    }
    const result = await response.json();
    return result.data?.authorized || false;
  } catch (error) {
    console.error('Permission check error:', error);
    // Could also send to error tracking service
    return false; // Fail closed (deny by default)
  }
}
```

### 5. Use Descriptive Permission Names

**❌ BAD**: Generic or ambiguous names

```
api:thing:do
service:item:action
app:data:modify
```

**✅ GOOD**: Clear, specific names

```
tenant-api:member:create
billing-api:invoice:delete
analytics-api:report:export
```

---

## Performance Optimization

### 1. Batch Permission Checks (Frontend)

Instead of checking permissions individually:

```javascript
// ❌ BAD: Multiple individual checks
const canCreate = await checkPermission(..., 'create');
const canUpdate = await checkPermission(..., 'update');
const canDelete = await checkPermission(..., 'delete');
```

Fetch all user permissions once:

```javascript
// ✅ GOOD: Single request for all permissions
const response = await fetch(
  `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
  { credentials: 'include' }
);
const permissions = await response.json();

// Cache in context/state
const permissionMap = {};
permissions.data.forEach(p => {
  permissionMap[`${p.service}:${p.entity}:${p.action}`] = true;
});

// Fast O(1) lookup
const canCreate = permissionMap['tenant-api:member:create'];
const canUpdate = permissionMap['tenant-api:member:update'];
const canDelete = permissionMap['tenant-api:member:delete'];
```

### 2. Implement Caching (Backend)

Add Redis caching to reduce database queries:

```go
func (s *rbacService) CheckUserPermission(tenantID uuid.UUID, userID string, service, entity, action string) (bool, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("perm:%s:%s:%s:%s:%s", tenantID, userID, service, entity, action)
    
    if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
        return cached == "1", nil
    }
    
    // Cache miss - query database
    hasPermission, err := s.rbacRepo.CheckUserPermission(tenantID, userID, service, entity, action)
    if err != nil {
        return false, err
    }
    
    // Cache result for 5 minutes
    s.redis.Set(ctx, cacheKey, boolToString(hasPermission), 5*time.Minute)
    
    return hasPermission, nil
}
```

### 3. Preload Permissions at Login

When user authenticates, fetch and cache all their permissions:

```javascript
async function onLogin(tenantId, userId) {
  // Fetch all permissions immediately
  const response = await fetch(
    `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
    { credentials: 'include' }
  );
  
  const result = await response.json();
  
  // Store in localStorage or context
  localStorage.setItem('user-permissions', JSON.stringify(result.data));
  
  // Later, check permissions from cache
  function hasPermission(service, entity, action) {
    const permissions = JSON.parse(localStorage.getItem('user-permissions') || '[]');
    return permissions.some(p => 
      p.service === service && p.entity === entity && p.action === action
    );
  }
}
```

### 4. Database Indexing

Ensure proper indexes exist on permission-related tables:

```sql
-- Index on tenant_members for faster lookups
CREATE INDEX idx_tenant_members_tenant_user ON tenant_members(tenant_id, user_id, status);

-- Index on role_policies for faster joins
CREATE INDEX idx_role_policies_role ON role_policies(role_id);
CREATE INDEX idx_role_policies_policy ON role_policies(policy_id);

-- Index on policy_permissions for faster joins
CREATE INDEX idx_policy_permissions_policy ON policy_permissions(policy_id);
CREATE INDEX idx_policy_permissions_permission ON policy_permissions(permission_id);

-- Index on permissions for faster lookups
CREATE INDEX idx_permissions_lookup ON permissions(service, entity, action);
```

---

## Security Considerations

### 1. Always Fail Closed

If permission check fails or throws an error, **deny access** by default:

```go
hasPermission, err := rbacService.CheckUserPermission(...)
if err != nil {
    // Log error but deny access
    logger.Error("Permission check failed", zap.Error(err))
    return c.JSON(403, gin.H{"error": "Permission denied"})
}

if !hasPermission {
    return c.JSON(403, gin.H{"error": "Permission denied"})
}
```

### 2. Validate Tenant Context

Never trust client-provided tenant IDs without verification:

```go
func (h *Handler) DeleteMember(c *gin.Context) {
    // Get tenant from authenticated context (set by middleware)
    tenantID, exists := c.Get("tenantID")
    if !exists {
        return c.JSON(401, gin.H{"error": "No tenant context"})
    }
    
    // Don't use tenant ID from request body/params without verification
    // ...
}
```

### 3. Active Members Only

The permission check query includes:

```sql
WHERE tm.status = 'active'
```

This ensures suspended or pending members cannot perform actions.

### 4. Audit Permission Checks

Log important permission checks for security auditing:

```go
result, err := rbacService.CheckUserPermission(tenantID, userID, service, entity, action)

logger.Info("Permission check",
    zap.String("tenant_id", tenantID.String()),
    zap.String("user_id", userID),
    zap.String("permission", fmt.Sprintf("%s:%s:%s", service, entity, action)),
    zap.Bool("authorized", result),
)

return result, err
```

### 5. Rate Limiting

Consider rate limiting the authorize endpoint to prevent abuse:

```go
// Apply rate limiting middleware
auth.POST("/authorize",
    middleware.RateLimitByUser(100, time.Minute), // 100 checks per minute per user
    deps.RBACHandler.Authorize,
)
```

---

## Troubleshooting

### Issue: Always Returns `authorized: false`

**Possible causes:**

1. **User not a member of tenant**
   ```sql
   -- Check membership
   SELECT * FROM tenant_members 
   WHERE tenant_id = 'xxx' AND user_id = 'yyy';
   ```

2. **Member status not active**
   ```sql
   -- Check status
   SELECT status FROM tenant_members 
   WHERE tenant_id = 'xxx' AND user_id = 'yyy';
   -- Should be 'active', not 'pending' or 'suspended'
   ```

3. **Role has no policies assigned**
   ```sql
   -- Check role policies
   SELECT * FROM role_policies WHERE role_id = 'role_id';
   ```

4. **Policy has no permissions**
   ```sql
   -- Check policy permissions
   SELECT * FROM policy_permissions WHERE policy_id = 'policy_id';
   ```

5. **Permission doesn't exist**
   ```sql
   -- Check permission exists
   SELECT * FROM permissions 
   WHERE service = 'tenant-api' 
     AND entity = 'member' 
     AND action = 'create';
   ```

### Issue: Frontend Permission Check Fails

**Check these:**

1. **Cookies not being sent**
   ```javascript
   // Ensure credentials: 'include' is set
   fetch('/api/v1/authorize?...', {
     credentials: 'include', // Required!
   });
   ```

2. **CORS issues**
   ```
   Access to fetch at '...' from origin '...' has been blocked by CORS policy
   ```
   
   Solution: Ensure backend CORS middleware allows credentials:
   ```go
   router.Use(middleware.CORS()) // Should include credentials: true
   ```

3. **Session expired**
   ```javascript
   // Check if session is valid
   const session = await Session.doesSessionExist();
   if (!session) {
     // Redirect to login
   }
   ```

### Issue: Permission Check is Slow

**Optimization steps:**

1. **Add database indexes** (see Performance Optimization section)

2. **Enable query logging**
   ```go
   db.LogMode(true) // GORM
   ```
   
   Check if the query is using indexes.

3. **Implement caching** (Redis or in-memory)

4. **Batch permission checks** instead of individual calls

### Issue: 400 Bad Request - Missing Parameters

**Ensure all 5 parameters are provided:**

```javascript
// ❌ BAD: Missing action parameter
fetch('/api/v1/authorize?tenant_id=xxx&user_id=yyy&service=tenant-api&entity=member');

// ✅ GOOD: All parameters present
fetch('/api/v1/authorize?tenant_id=xxx&user_id=yyy&service=tenant-api&entity=member&action=create');
```

### Issue: 401 Unauthorized

**Session is not valid or missing:**

1. Check cookies are being sent
2. Verify session hasn't expired
3. Ensure user is logged in
4. Check `AuthMiddleware` is applied

---

## Complete Example: Member Management

### Backend Handler

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/ysaakpr/rex/internal/api/middleware"
    "github.com/ysaakpr/rex/internal/pkg/response"
)

type MemberHandler struct {
    memberService services.MemberService
    rbacService   services.RBACService
}

func (h *MemberHandler) SetupRoutes(r *gin.RouterGroup, rbacService services.RBACService) {
    members := r.Group("/members")
    {
        // Public: list members (read permission)
        members.GET("",
            middleware.RequirePermission(rbacService, "tenant-api", "member", "read"),
            h.ListMembers,
        )
        
        // Protected: add member (create permission)
        members.POST("",
            middleware.RequirePermission(rbacService, "tenant-api", "member", "create"),
            h.AddMember,
        )
        
        // Protected: update member (update permission)
        members.PATCH("/:user_id",
            middleware.RequirePermission(rbacService, "tenant-api", "member", "update"),
            h.UpdateMember,
        )
        
        // Protected: delete member (delete permission)
        members.DELETE("/:user_id",
            middleware.RequirePermission(rbacService, "tenant-api", "member", "delete"),
            h.DeleteMember,
        )
    }
}

func (h *MemberHandler) ListMembers(c *gin.Context) {
    // Permission already checked by middleware
    tenantID, _ := c.Get("tenantID")
    
    members, err := h.memberService.ListMembers(tenantID)
    if err != nil {
        response.InternalServerError(c, err)
        return
    }
    
    response.OK(c, members)
}
```

### Frontend Component

```jsx
import React, { useState, useEffect } from 'react';
import { usePermission } from './hooks/usePermission';

function MembersPage({ tenantId, userId }) {
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  
  // Check permissions
  const { authorized: canCreate } = usePermission(tenantId, userId, 'tenant-api', 'member', 'create');
  const { authorized: canUpdate } = usePermission(tenantId, userId, 'tenant-api', 'member', 'update');
  const { authorized: canDelete } = usePermission(tenantId, userId, 'tenant-api', 'member', 'delete');
  
  useEffect(() => {
    fetchMembers();
  }, [tenantId]);
  
  async function fetchMembers() {
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members`, {
        credentials: 'include',
      });
      const result = await response.json();
      setMembers(result.data);
    } catch (error) {
      console.error('Failed to fetch members:', error);
    } finally {
      setLoading(false);
    }
  }
  
  async function handleAddMember() {
    // Double-check permission before API call
    if (!canCreate) {
      alert('You do not have permission to add members');
      return;
    }
    
    // Show add member modal
    // ...
  }
  
  async function handleDeleteMember(memberId) {
    // Double-check permission before API call
    if (!canDelete) {
      alert('You do not have permission to delete members');
      return;
    }
    
    if (!confirm('Are you sure?')) return;
    
    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members/${memberId}`, {
        method: 'DELETE',
        credentials: 'include',
      });
      
      if (response.ok) {
        fetchMembers(); // Refresh list
      } else {
        const error = await response.json();
        alert(error.error || 'Delete failed');
      }
    } catch (error) {
      console.error('Failed to delete member:', error);
      alert('Delete failed');
    }
  }
  
  if (loading) {
    return <div>Loading members...</div>;
  }
  
  return (
    <div className="members-page">
      <div className="header">
        <h1>Team Members</h1>
        {canCreate && (
          <button onClick={handleAddMember} className="btn-primary">
            Add Member
          </button>
        )}
      </div>
      
      <table className="members-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
            {(canUpdate || canDelete) && <th>Actions</th>}
          </tr>
        </thead>
        <tbody>
          {members.map(member => (
            <tr key={member.id}>
              <td>{member.name}</td>
              <td>{member.email}</td>
              <td>{member.role}</td>
              <td>{member.status}</td>
              {(canUpdate || canDelete) && (
                <td>
                  {canUpdate && (
                    <button onClick={() => handleEditMember(member.id)}>
                      Edit
                    </button>
                  )}
                  {canDelete && (
                    <button onClick={() => handleDeleteMember(member.user_id)}>
                      Delete
                    </button>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default MembersPage;
```

---

## Summary

The RBAC authorize endpoint provides a flexible, database-driven permission system that works seamlessly across backend and frontend applications.

**Key Points:**

✅ **Single Source of Truth**: Permissions are stored in the database  
✅ **Flexible**: Works with any service, entity, and action combination  
✅ **Secure**: Always verify permissions on the backend  
✅ **Performance**: Optimize with caching and batch checks  
✅ **User-Friendly**: Check permissions in frontend for better UX  

**Remember:**

- Frontend checks = Better user experience
- Backend checks = Actual security
- Always use both!

---

**Related Documentation:**

- [API Examples](./API_EXAMPLES.md)
- [RBAC Setup Guide](./changedoc/08-RBAC_REFACTORING.md)
- [Authentication Guide](./API_AUTHENTICATION_GUIDE.md)

**Need Help?**

- Check [Troubleshooting](#troubleshooting) section
- Review [Common Patterns](#best-practices)
- See [Complete Example](#complete-example-member-management)

