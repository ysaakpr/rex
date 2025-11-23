-- Drop indexes
DROP INDEX IF EXISTS idx_policies_is_system;
DROP INDEX IF EXISTS idx_policies_tenant_id;
DROP INDEX IF EXISTS idx_roles_is_system;
DROP INDEX IF EXISTS idx_roles_tenant_id;

-- Remove columns from policies table
ALTER TABLE policies 
  DROP COLUMN IF EXISTS is_system,
  DROP COLUMN IF EXISTS tenant_id;

-- Remove columns from roles table
ALTER TABLE roles 
  DROP COLUMN IF EXISTS is_system,
  DROP COLUMN IF EXISTS tenant_id;
