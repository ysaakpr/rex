# Platform Admin API

Complete API reference for Platform Administrator management.

## Overview

Platform Admins have full system access across all tenants. They can:
- Manage all tenants
- Create/manage System Users
- Configure RBAC (Roles, Policies, Permissions)
- Manage other Platform Admins
- Access platform-wide analytics

**Base URL**: `/api/v1/platform-admins`

**Authentication**: Platform Admin required for all operations

:::danger Super User Powers
Platform Admins bypass all tenant-level authorization. Use this role sparingly and audit regularly.
:::

## Create Platform Admin

Promote an existing user to Platform Admin.

**Request**:
```http
POST /api/v1/platform-admins
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "user_id": "auth0|user123",
  "notes": "Engineering lead - full platform access"
}
```

**Fields**:
- `user_id` (required): SuperTokens user ID
- `notes` (optional): Administrative notes (max 500 chars)

**Response** (201):
```json
{
  "success": true,
  "message": "Platform Admin created successfully",
  "data": {
    "id": "admin-uuid",
    "user_id": "auth0|user123",
    "notes": "Engineering lead - full platform access",
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

:::tip First Admin Bootstrap
The first Platform Admin must be created via database seeding or migration. See [Production Admin Management](/deployment/production-setup#bootstrap-first-admin).
:::

## List Platform Admins

Get all Platform Admins with user details.

**Request**:
```http
GET /api/v1/platform-admins
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "admin-uuid-1",
      "user_id": "auth0|user123",
      "email": "admin@example.com",
      "notes": "Engineering lead",
      "created_at": "2024-11-01T10:00:00Z"
    },
    {
      "id": "admin-uuid-2",
      "user_id": "auth0|user456",
      "email": "ops@example.com",
      "notes": "Operations manager",
      "created_at": "2024-11-15T14:30:00Z"
    }
  ]
}
```

:::info User Details
The API automatically enriches admin records with user details from SuperTokens (email, name, etc.).
:::

## Get Platform Admin

Get details about a specific Platform Admin.

**Request**:
```http
GET /api/v1/platform-admins/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "admin-uuid",
    "user_id": "auth0|user123",
    "email": "admin@example.com",
    "notes": "Engineering lead - full platform access",
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

## Delete Platform Admin

Remove Platform Admin privileges from a user.

**Request**:
```http
DELETE /api/v1/platform-admins/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Platform Admin deleted successfully"
}
```

:::warning Effect
The user immediately loses platform-wide access but retains any tenant-specific roles they have.
:::

## Check Platform Admin Status

Check if the current authenticated user is a Platform Admin.

**Request**:
```http
GET /api/v1/platform-admins/check
Authorization: Bearer <token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "is_platform_admin": true,
    "admin_id": "admin-uuid"
  }
}
```

**Non-Admin Response** (200):
```json
{
  "success": true,
  "data": {
    "is_platform_admin": false,
    "admin_id": null
  }
}
```

## Complete Workflow Example

### Promoting a User to Platform Admin

```javascript
// 1. User must first be registered via SuperTokens
// (e.g., via signup flow at /auth)

// 2. Get user ID from SuperTokens session
const session = await Session.getSessionInfo(sessionHandle);
const userId = session.userId; // e.g., "auth0|user123"

// 3. Promote to Platform Admin
const response = await fetch('/api/v1/platform-admins', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    user_id: userId,
    notes: 'Promoted on 2024-11-20 by existing admin'
  })
});

const result = await response.json();
console.log('New admin ID:', result.data.id);
```

### Frontend: Conditional Admin UI

```jsx
import {useEffect, useState} from 'react';

function AdminPanel() {
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if current user is platform admin
    fetch('/api/v1/platform-admins/check', {
      credentials: 'include'
    })
      .then(res => res.json())
      .then(data => {
        setIsPlatformAdmin(data.data.is_platform_admin);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) return <div>Loading...</div>;
  if (!isPlatformAdmin) return <div>Access Denied</div>;

  return (
    <div>
      <h1>Platform Admin Panel</h1>
      {/* Platform-wide controls */}
    </div>
  );
}
```

### Backend: Protecting Platform Admin Routes

```go
// Middleware: internal/api/middleware/platform_admin.go
func RequirePlatformAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        // Check if user is platform admin
        var admin models.PlatformAdmin
        err := db.Where("user_id = ?", userID).First(&admin).Error
        if err != nil {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error": "Platform admin access required",
            })
            c.Abort()
            return
        }
        
        c.Set("platform_admin_id", admin.ID)
        c.Next()
    }
}

// Usage in router
platformAdmin := router.Group("/api/v1/platform")
platformAdmin.Use(middleware.RequirePlatformAdmin())
{
    platformAdmin.POST("/roles", handlers.CreateRole)
    platformAdmin.GET("/tenants", handlers.ListAllTenants)
    // ... other platform admin routes
}
```

## Access Control Matrix

| Operation | Platform Admin | Tenant Admin | Regular User |
|-----------|---------------|--------------|--------------|
| Create tenant | ✅ | ❌ | ❌ |
| View all tenants | ✅ | ❌ | ❌ |
| Manage RBAC | ✅ | ❌ | ❌ |
| Create System Users | ✅ | ❌ | ❌ |
| Access any tenant | ✅ | Own tenant only | Own tenant only |
| Promote to Platform Admin | ✅ | ❌ | ❌ |
| View platform analytics | ✅ | ❌ | ❌ |

## Security Best Practices

### Minimal Admin Count

- Keep Platform Admin count as low as possible (recommended: 2-5)
- Use tenant-level roles for day-to-day operations
- Platform Admin should be for exceptional cases only

### Audit Logging

```go
// Log all Platform Admin actions
logger.Info("Platform Admin action",
    zap.String("admin_id", adminID),
    zap.String("action", "create_tenant"),
    zap.String("target", tenantID),
    zap.String("ip", c.ClientIP()),
)
```

### MFA Enforcement

- Require MFA for all Platform Admin accounts
- Configure in SuperTokens dashboard
- Enforce strict password policies

### Regular Reviews

- Monthly audit of active Platform Admins
- Remove admins who no longer need access
- Update notes field with access justification

### Emergency Access

```sql
-- Emergency: Manually promote user to Platform Admin
-- (Use only if locked out of platform)
INSERT INTO platform_admins (id, user_id, notes, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'auth0|emergency-user-id',
  'Emergency access - 2024-11-20',
  NOW(),
  NOW()
);
```

## Monitoring Platform Admin Activity

### Get All Admin Actions (Custom Implementation)

```go
// Example: Track admin actions
type AdminAuditLog struct {
    ID            uuid.UUID
    AdminID       uuid.UUID
    Action        string
    ResourceType  string
    ResourceID    string
    IPAddress     string
    Metadata      map[string]interface{}
    CreatedAt     time.Time
}

// Log every admin action
func LogAdminAction(adminID uuid.UUID, action, resourceType, resourceID, ip string, metadata map[string]interface{}) {
    log := AdminAuditLog{
        ID:           uuid.New(),
        AdminID:      adminID,
        Action:       action,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        IPAddress:    ip,
        Metadata:     metadata,
        CreatedAt:    time.Now(),
    }
    db.Create(&log)
}

// Query audit logs
func GetAdminAuditLogs(adminID uuid.UUID, limit int) ([]AdminAuditLog, error) {
    var logs []AdminAuditLog
    err := db.Where("admin_id = ?", adminID).
        Order("created_at DESC").
        Limit(limit).
        Find(&logs).Error
    return logs, err
}
```

## Bootstrapping First Platform Admin

### During Development

Use the migration script:

```sql
-- migrations/20241120_insert_first_admin.up.sql
INSERT INTO platform_admins (id, user_id, notes, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'YOUR_SUPERTOKENS_USER_ID',
  'Bootstrap admin - created via migration',
  NOW(),
  NOW()
);
```

### In Production

1. **Method 1: Via SuperTokens Dashboard**
   - Register first user via frontend
   - Get user ID from SuperTokens dashboard
   - Manually insert into database

2. **Method 2: Via CLI Migration Tool**
   ```bash
   ./cmd/migrate bootstrap-admin --user-id="auth0|user123"
   ```

3. **Method 3: Via Database Client**
   ```bash
   make shell-db
   # Then run INSERT statement
   ```

See [Production Admin Management](/deployment/production-setup#admin-bootstrap) for detailed instructions.

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 400 | User ID required | Missing user_id in request |
| 403 | Not authorized | Current user is not a Platform Admin |
| 404 | Platform Admin not found | Invalid admin ID |
| 409 | User is already Platform Admin | User already has Platform Admin privileges |
| 422 | User not found in SuperTokens | User ID doesn't exist |

## Related Guides

- [User Authentication](/guides/user-authentication) - SuperTokens setup
- [RBAC Overview](/guides/rbac-overview) - Authorization system
- [Production Setup](/deployment/production-setup) - Deployment guide
- [Security Best Practices](/guides/security) - Security guidelines
