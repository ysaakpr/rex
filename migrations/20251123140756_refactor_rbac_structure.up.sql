-- Refactor RBAC structure for clearer terminology
-- OLD: relations (user's role in tenant), roles (group of permissions)
-- NEW: roles (user's role in tenant), policies (group of permissions)

-- Step 1: Drop dependent tables first
DROP TABLE IF EXISTS relation_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS member_roles CASCADE;

-- Step 2: Drop main tables
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS relations CASCADE;

-- Step 3: Create new ROLES table (was relations - represents user's role in a tenant)
CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) NOT NULL UNIQUE,
  type VARCHAR(20) NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_type ON roles(type);
CREATE INDEX idx_roles_created_at ON roles(created_at);

-- Step 4: Create new POLICIES table (was roles - represents group of permissions)
CREATE TABLE policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policies_name ON policies(name);
CREATE INDEX idx_policies_created_at ON policies(created_at);

-- Step 5: Create ROLE_POLICIES junction table (was relation_roles)
CREATE TABLE role_policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(role_id, policy_id)
);

CREATE INDEX idx_role_policies_role_id ON role_policies(role_id);
CREATE INDEX idx_role_policies_policy_id ON role_policies(policy_id);

-- Step 6: Create POLICY_PERMISSIONS junction table (was role_permissions)
CREATE TABLE policy_permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(policy_id, permission_id)
);

CREATE INDEX idx_policy_permissions_policy_id ON policy_permissions(policy_id);
CREATE INDEX idx_policy_permissions_permission_id ON policy_permissions(permission_id);

-- Step 7: Seed default ROLES (was relations)
INSERT INTO roles (name, type, description) VALUES
  ('Admin', 'tenant', 'Full administrative access to the tenant'),
  ('Writer', 'tenant', 'Can create and modify content'),
  ('Viewer', 'tenant', 'Read-only access to tenant resources'),
  ('Basic', 'tenant', 'Basic tenant membership with minimal permissions');

-- Step 8: Seed default POLICIES (was roles)
INSERT INTO policies (name, description) VALUES
  ('Tenant Admin Policy', 'Full administrative permissions for tenant management'),
  ('Content Writer Policy', 'Permissions for creating and editing content'),
  ('Content Viewer Policy', 'Read-only permissions for viewing content'),
  ('Basic Member Policy', 'Minimal permissions for basic tenant members');

-- Step 9: Get policy and role IDs for mapping
DO $$
DECLARE
  admin_role_id UUID;
  writer_role_id UUID;
  viewer_role_id UUID;
  basic_role_id UUID;
  
  tenant_admin_policy_id UUID;
  writer_policy_id UUID;
  viewer_policy_id UUID;
  basic_policy_id UUID;
BEGIN
  -- Get role IDs
  SELECT id INTO admin_role_id FROM roles WHERE name = 'Admin';
  SELECT id INTO writer_role_id FROM roles WHERE name = 'Writer';
  SELECT id INTO viewer_role_id FROM roles WHERE name = 'Viewer';
  SELECT id INTO basic_role_id FROM roles WHERE name = 'Basic';
  
  -- Get policy IDs
  SELECT id INTO tenant_admin_policy_id FROM policies WHERE name = 'Tenant Admin Policy';
  SELECT id INTO writer_policy_id FROM policies WHERE name = 'Content Writer Policy';
  SELECT id INTO viewer_policy_id FROM policies WHERE name = 'Content Viewer Policy';
  SELECT id INTO basic_policy_id FROM policies WHERE name = 'Basic Member Policy';
  
  -- Map roles to policies
  INSERT INTO role_policies (role_id, policy_id) VALUES
    (admin_role_id, tenant_admin_policy_id),
    (writer_role_id, writer_policy_id),
    (viewer_role_id, viewer_policy_id),
    (basic_role_id, basic_policy_id);
END $$;

-- Step 10: Map policies to permissions
DO $$
DECLARE
  tenant_admin_policy_id UUID;
  writer_policy_id UUID;
  viewer_policy_id UUID;
  basic_policy_id UUID;
  perm_id UUID;
BEGIN
  -- Get policy IDs
  SELECT id INTO tenant_admin_policy_id FROM policies WHERE name = 'Tenant Admin Policy';
  SELECT id INTO writer_policy_id FROM policies WHERE name = 'Content Writer Policy';
  SELECT id INTO viewer_policy_id FROM policies WHERE name = 'Content Viewer Policy';
  SELECT id INTO basic_policy_id FROM policies WHERE name = 'Basic Member Policy';
  
  -- Tenant Admin Policy: All tenant-api permissions
  FOR perm_id IN SELECT id FROM permissions WHERE service = 'tenant-api' LOOP
    INSERT INTO policy_permissions (policy_id, permission_id) VALUES (tenant_admin_policy_id, perm_id)
    ON CONFLICT DO NOTHING;
  END LOOP;
  
  -- Writer Policy: Create, read, update permissions
  FOR perm_id IN SELECT id FROM permissions 
    WHERE service = 'tenant-api' 
    AND action IN ('create', 'read', 'update') LOOP
    INSERT INTO policy_permissions (policy_id, permission_id) VALUES (writer_policy_id, perm_id)
    ON CONFLICT DO NOTHING;
  END LOOP;
  
  -- Viewer Policy: Read-only permissions
  FOR perm_id IN SELECT id FROM permissions 
    WHERE service = 'tenant-api' 
    AND action = 'read' LOOP
    INSERT INTO policy_permissions (policy_id, permission_id) VALUES (viewer_policy_id, perm_id)
    ON CONFLICT DO NOTHING;
  END LOOP;
  
  -- Basic Policy: Minimal read permissions
  FOR perm_id IN SELECT id FROM permissions 
    WHERE service = 'tenant-api' 
    AND entity IN ('tenant', 'member') 
    AND action = 'read' LOOP
    INSERT INTO policy_permissions (policy_id, permission_id) VALUES (basic_policy_id, perm_id)
    ON CONFLICT DO NOTHING;
  END LOOP;
END $$;

-- Step 11: Update tenant_members to use new role_id column
ALTER TABLE tenant_members DROP CONSTRAINT IF EXISTS tenant_members_relation_id_fkey;
ALTER TABLE tenant_members DROP COLUMN IF EXISTS relation_id;
ALTER TABLE tenant_members ADD COLUMN role_id UUID REFERENCES roles(id);

-- Update existing tenant_members to use "Basic" role by default
UPDATE tenant_members 
SET role_id = (SELECT id FROM roles WHERE name = 'Basic' LIMIT 1)
WHERE role_id IS NULL;

ALTER TABLE tenant_members ALTER COLUMN role_id SET NOT NULL;

CREATE INDEX idx_tenant_members_role_id ON tenant_members(role_id);

-- Step 12: Update user_invitations to use new role_id column
ALTER TABLE user_invitations DROP CONSTRAINT IF EXISTS user_invitations_relation_id_fkey;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS relation_id;
ALTER TABLE user_invitations ADD COLUMN role_id UUID REFERENCES roles(id);

-- Set default role for invitations
UPDATE user_invitations 
SET role_id = (SELECT id FROM roles WHERE name = 'Basic' LIMIT 1)
WHERE role_id IS NULL AND status = 'pending';

CREATE INDEX IF NOT EXISTS idx_user_invitations_role_id ON user_invitations(role_id);

