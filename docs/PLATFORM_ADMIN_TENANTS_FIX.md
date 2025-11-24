# Platform Admin Tenants List Fix

## Issue

The Tenants page in the admin dashboard was showing only the tenants that the logged-in user was a member of, not ALL tenants in the platform. This was incorrect behavior for platform administrators who should see all tenants.

## Solution

Created a separate endpoint for platform admins to list ALL tenants in the system, and updated the frontend to use the appropriate endpoint based on the user's role.

## Changes Made

### Backend Changes

#### 1. Service Layer (`internal/services/tenant_service.go`)

**Added method to interface:**
```go
type TenantService interface {
    // ... existing methods ...
    GetAllTenants(pagination *models.PaginationParams) ([]*models.Tenant, int64, error)
}
```

**Implemented method:**
```go
// GetAllTenants returns all tenants in the system (for platform admins)
func (s *tenantService) GetAllTenants(pagination *models.PaginationParams) ([]*models.Tenant, int64, error) {
    return s.tenantRepo.List(pagination)
}
```

#### 2. Handler Layer (`internal/api/handlers/tenant_handler.go`)

**Added new handler:**
```go
// ListAllTenants - Platform admins only endpoint
// Returns ALL tenants in the system with pagination
func (h *TenantHandler) ListAllTenants(c *gin.Context)
```

This handler:
- Accepts pagination parameters
- Calls `GetAllTenants` service method
- Returns all tenants with member counts
- Returns paginated response

#### 3. Router (`internal/api/router/router.go`)

**Added new route:**
```go
platform.GET("/tenants", deps.TenantHandler.ListAllTenants)
```

This route:
- Path: `/api/v1/platform/tenants`
- Protected by `PlatformAdminMiddleware`
- Only accessible to platform administrators

### Frontend Changes

#### TenantsPage Component (`frontend/src/components/pages/TenantsPage.jsx`)

**Added platform admin check:**
```javascript
const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
const [checkingAdmin, setCheckingAdmin] = useState(true);

const checkPlatformAdmin = async () => {
  const response = await fetch('/api/v1/platform/admins/check', {
    credentials: 'include'
  });
  
  if (response.ok) {
    const data = await response.json();
    setIsPlatformAdmin(data.data?.is_platform_admin || false);
  }
};
```

**Updated tenant loading logic:**
```javascript
const loadTenants = async () => {
  // Use platform endpoint if user is platform admin
  const endpoint = isPlatformAdmin 
    ? '/api/v1/platform/tenants'  // All tenants
    : '/api/v1/tenants';            // User's tenants only
  
  const response = await fetch(endpoint, { credentials: 'include' });
  // ... process response
};
```

**Updated page description:**
```javascript
<p className="text-muted-foreground mt-2">
  {isPlatformAdmin 
    ? 'Manage and monitor all tenants in the platform'
    : 'Manage and monitor your tenants'
  }
</p>
```

## API Endpoints

### For Regular Users
```
GET /api/v1/tenants
```
- Returns only tenants where the user is a member
- Requires authentication

### For Platform Admins
```
GET /api/v1/platform/tenants
```
- Returns ALL tenants in the system
- Requires platform admin privileges
- Same response format as regular endpoint

## Response Format

Both endpoints return the same format:

```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "uuid",
        "name": "Tenant Name",
        "slug": "tenant-slug",
        "status": "active",
        "member_count": 5,
        "created_at": "2025-11-24T10:00:00Z",
        "updated_at": "2025-11-24T10:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5
  }
}
```

## Testing

### Test as Regular User

1. Log in as a non-admin user
2. Navigate to `/tenants`
3. Should see only your tenants

### Test as Platform Admin

1. Log in as platform admin
2. Navigate to `/tenants`
3. Should see ALL tenants in the platform
4. Page description should say "all tenants in the platform"

### Verify Backend

```bash
# As platform admin
curl -X GET https://rex.stage.fauda.dream11.in/api/v1/platform/tenants \
  -H "Cookie: YOUR_SESSION_COOKIE" | jq .

# Should return all tenants in the system
```

## Benefits

✅ **Correct Admin Behavior**: Platform admins now see all tenants as expected  
✅ **Proper Separation**: Regular users still only see their own tenants  
✅ **No Breaking Changes**: Existing `/api/v1/tenants` endpoint unchanged  
✅ **Consistent UI**: Same page, different data based on user role  
✅ **Security**: Platform endpoint protected by middleware  

## Security Considerations

- Platform admin endpoint is protected by `PlatformAdminMiddleware`
- Only users in `platform_admins` table can access `/api/v1/platform/tenants`
- Regular users get 403 Forbidden if they try to access platform endpoints
- No data leakage: Frontend checks admin status before showing data

## Database Queries

### Regular User Endpoint
```sql
SELECT * FROM tenants 
WHERE created_by = 'user_id'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;
```

### Platform Admin Endpoint
```sql
SELECT * FROM tenants
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;
```

## Migration Notes

No database migrations needed - this is a pure logic change.

## Performance Impact

Minimal. Platform admins may see slower response times if there are many tenants (1000+), but pagination keeps it manageable.

Consider adding indexes if needed:
```sql
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at DESC);
```

## Related Files

- **Service**: `internal/services/tenant_service.go`
- **Handler**: `internal/api/handlers/tenant_handler.go`
- **Router**: `internal/api/router/router.go`
- **Repository**: `internal/repository/tenant_repository.go` (no changes - uses existing `List` method)
- **Frontend**: `frontend/src/components/pages/TenantsPage.jsx`

## Related Documentation

- [Platform Admin Guide](PRODUCTION_ADMIN_MANAGEMENT.md)
- [RBAC Authorization Guide](RBAC_AUTHORIZATION_GUIDE.md)
- [API Examples](API_EXAMPLES.md)

---

**Date**: November 24, 2025  
**Issue**: Platform admins saw only their tenants  
**Status**: ✅ Fixed  
**Impact**: Backend + Frontend changes

