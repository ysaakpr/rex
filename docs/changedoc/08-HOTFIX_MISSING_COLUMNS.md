# Hotfix: Missing RBAC Columns

**Date**: November 23, 2025  
**Status**: ✅ Fixed  
**Impact**: Critical - Pages were not loading

## Issue

After the RBAC refactoring, the Roles and Policies pages were failing with errors:

```
"error": "failed to list roles: ERROR: column \"is_system\" does not exist (SQLSTATE 42703)"
"error": "failed to list policies: ERROR: column \"is_system\" does not exist (SQLSTATE 42703)"
```

## Root Cause

The initial RBAC refactoring migration (`20251123140756_refactor_rbac_structure.up.sql`) created the `roles` and `policies` tables but forgot to include the `is_system` and `tenant_id` columns that the Go models expected.

**Missing columns**:
- `tenant_id UUID` - for tenant-specific roles/policies
- `is_system BOOLEAN` - to mark system-level vs tenant-level

## Solution

Created a new migration to add the missing columns:

**File**: `migrations/20251123140900_add_missing_rbac_columns.up.sql`

**Changes**:
```sql
-- Add missing columns to roles table
ALTER TABLE roles 
  ADD COLUMN tenant_id UUID,
  ADD COLUMN is_system BOOLEAN DEFAULT false;

-- Add missing columns to policies table
ALTER TABLE policies 
  ADD COLUMN tenant_id UUID,
  ADD COLUMN is_system BOOLEAN DEFAULT false;

-- Set is_system = true for all existing roles/policies
UPDATE roles SET is_system = true WHERE tenant_id IS NULL;
UPDATE policies SET is_system = true WHERE tenant_id IS NULL;

-- Add indexes
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_policies_tenant_id ON policies(tenant_id);
CREATE INDEX idx_policies_is_system ON policies(is_system);
```

## Verification

### Database Structure
```bash
# Check roles table
\d roles
# Should show: tenant_id UUID, is_system BOOLEAN

# Check policies table
\d policies
# Should show: tenant_id UUID, is_system BOOLEAN
```

### Data Verification
```sql
-- All default roles should have is_system = true
SELECT name, is_system, tenant_id FROM roles;
--  name  | is_system | tenant_id 
-- -------+-----------+-----------
-- Admin  | t         | 
-- Basic  | t         | 
-- Viewer | t         | 
-- Writer | t         | 

-- All default policies should have is_system = true
SELECT name, is_system, tenant_id FROM policies;
--          name         | is_system | tenant_id 
-- ----------------------+-----------+-----------
-- Basic Member Policy   | t         | 
-- Content Viewer Policy | t         | 
-- Content Writer Policy | t         | 
-- Tenant Admin Policy   | t         | 
```

### API Testing
```bash
# API responds without errors (auth required for actual data)
curl http://localhost:8080/api/v1/roles
# Should return: {"message": "unauthorised"} (expected without auth)
```

### Frontend Testing
1. Open http://localhost:3000
2. Login as platform admin
3. Navigate to "Policies" page → Should load successfully ✅
4. Navigate to "Roles" page → Should load successfully ✅

## Files Modified

**New Files**:
- `migrations/20251123140900_add_missing_rbac_columns.up.sql`
- `migrations/20251123140900_add_missing_rbac_columns.down.sql`

**Services Restarted**:
- API container (to pick up database changes)
- Frontend container (to clear cache)

## Timeline

1. **9:21 AM** - User reports pages not working
2. **9:22 AM** - Identified missing columns in database
3. **9:23 AM** - Created migration to add columns
4. **9:24 AM** - Applied migration successfully
5. **9:25 AM** - Verified data and structure
6. **9:26 AM** - Restarted services
7. **9:27 AM** - Confirmed fix working ✅

## Lesson Learned

When creating destructive migrations that recreate tables:
1. Always include ALL columns that models expect
2. Test immediately after migration
3. Compare model definitions with SQL CREATE statements
4. Use a checklist for required columns

## Related Documents

- Original refactoring: `docs/changedoc/08-RBAC_REFACTORING.md`
- Main summary: `RBAC_REFACTORING_SUMMARY.md`

---

**Status**: ✅ **RESOLVED**  
**Both Roles and Policies pages now working correctly!**
