# RBAC System Refactoring - Terminology Clarification

**Date**: November 23, 2025  
**Status**: ✅ Complete  
**Impact**: Major architectural change (backward incompatible)

## Overview

Completed a comprehensive refactoring of the RBAC (Role-Based Access Control) system to clarify terminology and improve semantic accuracy. The previous naming was confusing - what we called "Roles" were actually permission groups, and "Relations" were the actual user roles.

## Terminology Changes

### Before → After

| Old Name | New Name | Description |
|----------|----------|-------------|
| **Relations** | **Roles** | User's role in a tenant (e.g., Admin, Writer, Viewer) |
| **Roles** | **Policies** | Groups of permissions |
| **Permissions** | **Permissions** | _(Unchanged)_ Individual access rights |

### New Structure

```
User in Tenant
    └── Has a Role (Admin, Writer, Viewer, Basic)
            └── Role has multiple Policies
                    └── Policy has multiple Permissions
```

**Example**:
- A user with the **Admin** role in a tenant
- The Admin role has the **"Tenant Admin Policy"**
- That policy contains permissions like `tenant:create`, `member:add`, etc.

## Database Changes

### Tables Renamed

| Old Table | New Table |
|-----------|-----------|
| `relations` | `roles` |
| `roles` | `policies` |
| `role_permissions` | `policy_permissions` |
| `relation_roles` | `role_policies` |

### Migration Details

**File**: `migrations/20251123140756_refactor_rbac_structure.up.sql`

**Key Operations**:
1. Dropped all old RBAC tables
2. Recreated with new names and structure
3. Re-seeded default data:
   - 4 Roles: Admin, Writer, Viewer, Basic
   - 4 Policies: Tenant Admin Policy, Content Writer Policy, Content Viewer Policy, Basic Member Policy
   - 17 Permissions (tenant-api and platform-api)
4. Updated foreign keys in `tenant_members` and `user_invitations`

**⚠️ Note**: This is a destructive migration suitable for development. Production would need data migration scripts.

## Backend Code Changes

### 1. Models (7 files)

**New Files**:
- `internal/models/policy.go` (was `role.go`)
- `internal/models/role.go` (was `relation.go`)
- `internal/models/role_policy.go` (was `relation_role.go`)

**Updated Files**:
- `internal/models/tenant_member.go`: `relation_id` → `role_id`
- `internal/models/invitation.go`: `relation_id` → `role_id`

**Key Changes**:
- `Role` model now has `Type` field (tenant, platform)
- `Policy` model is the new permission group
- All JSON/GORM tags updated
- Response models updated

### 2. Repositories (4 files)

**File**: `internal/repository/rbac_repository.go` (completely rewritten)

**New Interface**:
```go
// Roles (user's role in tenant)
CreateRole(role *models.Role) error
GetRoleByID(id uuid.UUID) (*models.Role, error)
GetRoleWithPolicies(id uuid.UUID) (*models.Role, error)
ListRoles(tenantID *uuid.UUID) ([]*models.Role, error)
UpdateRole(role *models.Role) error
DeleteRole(id uuid.UUID) error

// Policies (group of permissions)
CreatePolicy(policy *models.Policy) error
GetPolicyByID(id uuid.UUID) (*models.Policy, error)
GetPolicyWithPermissions(id uuid.UUID) (*models.Policy, error)
ListPolicies(tenantID *uuid.UUID) ([]*models.Policy, error)
UpdatePolicy(policy *models.Policy) error
DeletePolicy(id uuid.UUID) error

// Role-Policy assignments
AssignPoliciesToRole(roleID uuid.UUID, policyIDs []uuid.UUID) error
RevokePolicyFromRole(roleID uuid.UUID, policyID uuid.UUID) error
GetRolePolicies(roleID uuid.UUID) ([]*models.Policy, error)
```

**Also Updated**:
- `member_repository.go`: Preload("Role") instead of Preload("Relation")
- `invitation_repository.go`: Preload("Role") instead of Preload("Relation")

### 3. Services (5 files)

**File**: `internal/services/rbac_service.go` (completely rewritten)

**Service Interface**: Mirrors repository with business logic
- Validation logic updated
- Error messages clarified
- All method names updated to reflect new terminology

**Also Updated**:
- `member_service.go`: GetRoleByID, role validation
- `invitation_service.go`: GetRoleByID, role validation
- `tenant_service.go`: GetRoleByName("Admin") instead of GetRelationByName
- `internal/jobs/tasks/invitation.go`: Preload("Role"), email template updated

### 4. API Handlers & Routes (2 files)

**File**: `internal/api/handlers/rbac_handler.go` (completely rewritten)

**Handler Methods**:
- `CreateRole`, `ListRoles`, `GetRole`, `UpdateRole`, `DeleteRole`
- `CreatePolicy`, `ListPolicies`, `GetPolicy`, `UpdatePolicy`, `DeletePolicy`
- `CreatePermission`, `ListPermissions`, `GetPermission`, `DeletePermission`
- `AssignPermissionsToPolicy`, `RevokePermissionFromPolicy`
- `AssignPoliciesToRole`, `RevokePolicyFromRole`, `GetRolePolicies`
- `Authorize`, `GetUserPermissions`

**File**: `internal/api/router/router.go`

**New Routes**:
```go
// Platform-level (admin only)
/api/v1/platform/roles              // User roles: Admin, Writer, etc.
/api/v1/platform/roles/:id/policies // Assign policies to a role
/api/v1/platform/policies           // Permission groups
/api/v1/platform/policies/:id/permissions
/api/v1/platform/permissions

// Legacy (for backward compatibility)
/api/v1/roles                       // Read-only
/api/v1/policies                    // Read-only
/api/v1/permissions
```

## Frontend Changes

### 1. Navigation (Sidebar)

**File**: `frontend/src/components/layout/Sidebar.jsx`

**Changes**:
- Removed "Roles" navigation item (now under Policies)
- Renamed "Permissions" → "Policies"
- Renamed "Tenant Relations" → "Roles"
- Reordered: Tenants → Users → Applications → Policies → Roles

### 2. Page Structure

**Files Renamed**:
- `RelationsPage.jsx` → `RolesPage.jsx`
- `RelationDetailsPage.jsx` → `RoleDetailsPage.jsx`
- `RolesPage.jsx` → `PoliciesListTab.jsx`
- `RoleDetailsPage.jsx` → `PolicyDetailsPage.jsx`

**New File**: `PoliciesPage.jsx`
- Tabs component with two tabs:
  - **Policies Tab**: Lists policy groups (PoliciesListTab)
  - **Permissions Tab**: Lists individual permissions (PermissionsPage)

### 3. Routes

**File**: `frontend/src/App.jsx`

**Updated Routes**:
```jsx
/permissions        → PoliciesPage (with tabs)
/policies/:id       → PolicyDetailsPage
/roles              → RolesPage
/roles/:id          → RoleDetailsPage
```

### 4. Component Updates (8+ files)

**API Endpoints Updated**:
- `/api/v1/relations` → `/api/v1/roles`
- `/api/v1/roles` → `/api/v1/platform/policies`
- `/api/v1/platform/relations` → `/api/v1/platform/roles`

**Fields Updated**:
- `relation_id` → `role_id`
- `relationId` → `roleId`
- All variable names and labels updated

**Files**:
- `Members.jsx`
- `ManagedTenantOnboarding.jsx`
- `TenantUserManagement.jsx`
- `RoleDetailsPage.jsx`
- `RolesPage.jsx`
- `PoliciesListTab.jsx`
- `PolicyDetailsPage.jsx`
- And more...

## Testing & Verification

### Database Verification

```sql
-- Check tables exist
\dt

-- Verify default roles
SELECT name, type FROM roles ORDER BY name;
-- Result: Admin, Basic, Viewer, Writer (all type 'tenant')

-- Verify default policies
SELECT name, description FROM policies ORDER BY name;
-- Result: 4 policies (Admin, Writer, Viewer, Basic)

-- Verify permissions
SELECT COUNT(*) FROM permissions;
-- Result: 17 permissions

-- Verify role-policy mapping
SELECT r.name as role, p.name as policy 
FROM roles r
JOIN role_policies rp ON r.id = rp.role_id
JOIN policies p ON p.id = rp.policy_id
ORDER BY r.name;
```

### API Testing

```bash
# Test new endpoints (requires auth)
curl http://localhost:8080/api/v1/platform/roles
curl http://localhost:8080/api/v1/platform/policies
curl http://localhost:8080/api/v1/platform/permissions

# Test legacy endpoints
curl http://localhost:8080/api/v1/roles
curl http://localhost:8080/api/v1/policies
```

### Frontend Testing

1. Navigate to http://localhost:3000
2. Check navigation sidebar:
   - ✅ "Policies" menu item (was "Permissions")
   - ✅ "Roles" menu item (was "Tenant Relations")
3. Test Policies page:
   - ✅ Two tabs: "Policies" and "Permissions"
   - ✅ Can view/create/edit policies
   - ✅ Can view/create permissions
4. Test Roles page:
   - ✅ Can view/create/edit roles
   - ✅ Can assign policies to roles
   - ✅ Role details show attached policies

## Impact & Breaking Changes

### ⚠️ Breaking Changes

1. **API Endpoints Changed**:
   - `/api/v1/platform/relations/*` → `/api/v1/platform/roles/*`
   - `/api/v1/platform/roles/*` → `/api/v1/platform/policies/*`

2. **Request/Response Bodies**:
   - `relation_id` → `role_id` in all payloads
   - `relations` → `roles` in response arrays
   - `roles` → `policies` in response arrays

3. **Database Schema**:
   - All RBAC tables renamed
   - Foreign keys in `tenant_members`, `user_invitations` updated

### Migration Path (For Production)

If you have existing data in production:

1. **Before Migration**:
   - Backup database
   - Document all existing roles, relations, and mappings
   - Plan downtime window

2. **Data Migration Script** (not included, would need):
   ```sql
   -- Save existing data to temp tables
   CREATE TABLE temp_old_relations AS SELECT * FROM relations;
   CREATE TABLE temp_old_roles AS SELECT * FROM roles;
   -- etc.
   
   -- Run new migration
   
   -- Migrate data from temp tables to new structure
   INSERT INTO roles (id, name, type, ...)
   SELECT id, name, 'tenant', ... FROM temp_old_relations;
   
   INSERT INTO policies (id, name, ...)
   SELECT id, name, ... FROM temp_old_roles;
   
   -- Update foreign keys
   UPDATE tenant_members SET role_id = (SELECT ...)
   
   -- Clean up
   DROP TABLE temp_old_relations;
   DROP TABLE temp_old_roles;
   ```

3. **Testing**:
   - Verify all user-tenant assignments
   - Test permission checks
   - Verify RBAC enforcement

4. **Rollback Plan**:
   - Keep backups for at least 30 days
   - Document reverse migration steps

## Files Modified

### Backend (20+ files)
- `migrations/20251123140756_refactor_rbac_structure.{up,down}.sql`
- `internal/models/{policy,role,role_policy,tenant_member,invitation}.go`
- `internal/repository/{rbac,member,invitation}_repository.go`
- `internal/services/{rbac,member,invitation,tenant}_service.go`
- `internal/jobs/tasks/invitation.go`
- `internal/api/handlers/rbac_handler.go`
- `internal/api/router/router.go`

### Frontend (15+ files)
- `frontend/src/components/layout/Sidebar.jsx`
- `frontend/src/App.jsx`
- `frontend/src/components/pages/PoliciesPage.jsx` (new)
- `frontend/src/components/pages/{Roles,Role,Policy,Policies}*.jsx` (renamed/updated)
- `frontend/src/components/{Members,Roles}.jsx`
- `frontend/src/components/pages/{ManagedTenantOnboarding,TenantUserManagement,UserDetailsPage}.jsx`

## Files Deleted

- `internal/models/relation.go`
- `internal/models/relation_role.go`
- `frontend/src/components/pages/RelationsPage.jsx`
- `frontend/src/components/pages/RelationDetailsPage.jsx`

## Future Improvements

1. **Multi-Tenant Policies**: Currently policies are system-wide. Could add tenant-specific policy customization.
2. **Policy Templates**: Pre-built policy templates for common use cases.
3. **Permission Discovery**: Auto-discover permissions from code annotations.
4. **Audit Logging**: Track all RBAC changes (role assignments, policy modifications).
5. **UI Improvements**: Visual policy builder, permission matrix view.
6. **API Versioning**: Add `/api/v2/` with new structure, deprecate old endpoints.

## Related Documentation

- API Examples: `docs/API_EXAMPLES.md` (to be updated)
- RBAC Overview: See project README
- Migration Guide: This document

## Notes

- ✅ All backend code compiles successfully
- ✅ All services (API, Worker, Frontend) running
- ✅ Database migration applied and seeded
- ✅ Frontend components updated and tested
- ⚠️ Legacy API routes maintained for backward compatibility (read-only)
- 📝 Update API documentation to reflect new terminology
- 📝 Update client SDKs if any exist

---

**Completed By**: AI Assistant  
**Reviewed**: Pending  
**Production Deployment**: Not yet (development only)
