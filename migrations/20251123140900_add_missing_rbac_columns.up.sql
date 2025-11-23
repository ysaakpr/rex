-- Add missing columns to roles table
ALTER TABLE roles 
  ADD COLUMN tenant_id UUID,
  ADD COLUMN is_system BOOLEAN DEFAULT false;

-- Add missing columns to policies table
ALTER TABLE policies 
  ADD COLUMN tenant_id UUID,
  ADD COLUMN is_system BOOLEAN DEFAULT false;

-- Set is_system = true for all existing roles (they are system-level)
UPDATE roles SET is_system = true WHERE tenant_id IS NULL;

-- Set is_system = true for all existing policies (they are system-level)
UPDATE policies SET is_system = true WHERE tenant_id IS NULL;

-- Add index for faster queries
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_policies_tenant_id ON policies(tenant_id);
CREATE INDEX idx_policies_is_system ON policies(is_system);
