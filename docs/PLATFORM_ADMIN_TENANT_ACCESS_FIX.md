# Platform Admin Tenant Access Fix

**Date**: November 24, 2024  
**Issue**: Platform admins receiving 403 errors when viewing tenant details they don't belong to  
**Status**: ✅ Fixed

## Problem

Platform administrators were unable to view tenant details for tenants they weren't members of. When clicking on a tenant in the admin dashboard, the app would attempt to fetch:

```
GET /api/v1/tenants/{id}
```

This endpoint is protected by the `TenantAccessMiddleware`, which requires the user to be a member of the tenant. This caused a 403 Forbidden error for platform admins trying to view any tenant's details.

## Root Cause

The tenant list page was correctly updated to show all tenants for platform admins (using `/api/v1/platform/tenants`), but the tenant details page was still using the regular membership-protected endpoint.

**Two separate issues:**
1. ✅ Tenant list - Fixed (shows all tenants for platform admins)
2. ✅ Tenant details - Fixed (now uses platform admin endpoint)

## Solution

### Backend Changes

#### 1. Added New Handler (`internal/api/handlers/tenant_handler.go`)

Created `GetTenantForPlatformAdmin` method that bypasses membership checks:

```go
// GetTenantForPlatformAdmin godoc
// @Summary Get tenant by ID (platform admins only, no membership required)
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} response.Response{data=models.TenantResponse}
// @Router /platform/tenants/{id} [get]
func (h *TenantHandler) GetTenantForPlatformAdmin(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        response.BadRequest(c, err)
        return
    }

    tenant, err := h.tenantService.GetTenant(id)
    if err != nil {
        response.NotFound(c, err.Error())
        return
    }

    tenantResp := tenant.ToResponse()

    // Count active members for this tenant
    var memberCount int64
    h.db.Table("tenant_members").
        Where("tenant_id = ?", tenant.ID).
        Where("status = ?", "active").
        Count(&memberCount)

    tenantResp.MemberCount = int(memberCount)

    response.OK(c, tenantResp)
}
```

**Key Points:**
- No `TenantAccessMiddleware` - only requires `PlatformAdminMiddleware`
- Uses the same `TenantService.GetTenant` method
- Returns the same response format as the regular endpoint
- Includes member count

#### 2. Added Route (`internal/api/router/router.go`)

```go
// Platform Admin routes (require platform admin access)
platform := auth.Group("/platform")
platform.Use(middleware.PlatformAdminMiddleware(deps.DB))
{
    // ... other routes ...
    
    // Tenants management (all tenants)
    platform.GET("/tenants", deps.TenantHandler.ListAllTenants)
    platform.GET("/tenants/:id", deps.TenantHandler.GetTenantForPlatformAdmin) // NEW
}
```

### Frontend Changes

#### Updated `TenantDetailsPage.jsx`

Added platform admin detection and conditional endpoint usage:

```jsx
export function TenantDetailsPage() {
  // ... existing state ...
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    checkPlatformAdmin();
  }, []);

  useEffect(() => {
    if (isPlatformAdmin !== null) {
      loadTenantDetails();
    }
  }, [id, isPlatformAdmin]);

  const checkPlatformAdmin = async () => {
    try {
      const response = await fetch('/api/v1/platform/admins/check', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        setIsPlatformAdmin(data.data?.is_platform_admin || false);
      } else {
        setIsPlatformAdmin(false);
      }
    } catch (err) {
      console.error('Error checking platform admin status:', err);
      setIsPlatformAdmin(false);
    }
  };

  const loadTenantDetails = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Use platform admin endpoint if user is a platform admin
      const endpoint = isPlatformAdmin 
        ? `/api/v1/platform/tenants/${id}`
        : `/api/v1/tenants/${id}`;
      
      const response = await fetch(endpoint, {
        credentials: 'include'
      });

      // ... rest of the code ...
    }
  };
}
```

## API Endpoints

### Regular Users (Tenant Members)

```
GET /api/v1/tenants/{id}
```

**Requirements:**
- Authenticated user
- User must be a member of the tenant (via `TenantAccessMiddleware`)

**Response:** Tenant details with member count

### Platform Admins

```
GET /api/v1/platform/tenants/{id}
```

**Requirements:**
- Authenticated user
- User must be a platform admin (via `PlatformAdminMiddleware`)
- No tenant membership required

**Response:** Tenant details with member count (same format)

## Testing

### Test Platform Admin Access

```bash
# 1. Login as platform admin
# 2. Navigate to admin dashboard: https://localhost/tenants
# 3. Click on any tenant
# 4. Should load tenant details without 403 error
```

### Verify Network Calls

**For Platform Admins:**
```
GET /api/v1/platform/tenants/4aa73956-eb02-45e1-9800-4cbea847adba
Status: 200 OK
```

**For Regular Users (members):**
```
GET /api/v1/tenants/4aa73956-eb02-45e1-9800-4cbea847adba
Status: 200 OK (if member)
Status: 403 Forbidden (if not member)
```

## Benefits

1. **Platform admins** can now view details for any tenant without needing membership
2. **Regular users** still have proper access control (membership required)
3. **Same response format** ensures frontend works consistently
4. **Minimal code duplication** - reuses existing service methods
5. **Follows established patterns** - uses platform admin middleware group

## Related Endpoints

These endpoints also follow the platform admin pattern:

```
GET  /api/v1/platform/tenants          # List all tenants
GET  /api/v1/platform/tenants/:id      # Get tenant details (NEW)
GET  /api/v1/platform/admins           # List platform admins
POST /api/v1/platform/admins           # Add platform admin
```

## Files Modified

**Backend:**
- `internal/api/handlers/tenant_handler.go` - Added `GetTenantForPlatformAdmin`
- `internal/api/router/router.go` - Added route `GET /platform/tenants/:id`

**Frontend:**
- `frontend/src/components/pages/TenantDetailsPage.jsx` - Added platform admin detection and conditional endpoint usage

## Deployment

**Local Development:**
```bash
docker-compose restart api
# Frontend hot-reloads automatically
```

**Production:**
```bash
cd ~/rex
git pull origin main
docker-compose up -d --build api
docker-compose restart frontend  # If needed
```

---

**Status**: ✅ Complete  
**Verified**: Tested locally with platform admin user  
**Production Ready**: Yes

