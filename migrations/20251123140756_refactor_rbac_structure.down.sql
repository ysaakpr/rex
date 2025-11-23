-- Revert RBAC structure refactor
-- NEW: roles (user's role in tenant), policies (group of permissions)
-- OLD: relations (user's role in tenant), roles (group of permissions)

-- Step 1: Drop indexes
DROP INDEX IF EXISTS idx_user_invitations_role_id;
DROP INDEX IF EXISTS idx_tenant_members_role_id;

-- Step 2: Update tenant_members back to relation_id
ALTER TABLE tenant_members DROP CONSTRAINT IF EXISTS tenant_members_role_id_fkey;
ALTER TABLE tenant_members DROP COLUMN IF EXISTS role_id;
ALTER TABLE tenant_members ADD COLUMN relation_id UUID;

-- Step 3: Update user_invitations back to relation_id
ALTER TABLE user_invitations DROP COLUMN IF EXISTS role_id;
ALTER TABLE user_invitations ADD COLUMN relation_id UUID;

-- Step 4: Drop new tables
DROP TABLE IF EXISTS role_policies CASCADE;
DROP TABLE IF EXISTS policy_permissions CASCADE;
DROP TABLE IF EXISTS policies CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

-- Step 5: Recreate old RELATIONS table
CREATE TABLE relations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) NOT NULL UNIQUE,
  type VARCHAR(20) NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_relations_name ON relations(name);
CREATE INDEX idx_relations_type ON relations(type);
CREATE INDEX idx_relations_created_at ON relations(created_at);

-- Step 6: Recreate old ROLES table
CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_created_at ON roles(created_at);

-- Step 7: Recreate RELATION_ROLES junction table
CREATE TABLE relation_roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  relation_id UUID NOT NULL REFERENCES relations(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(relation_id, role_id)
);

CREATE INDEX idx_relation_roles_relation_id ON relation_roles(relation_id);
CREATE INDEX idx_relation_roles_role_id ON relation_roles(role_id);

-- Step 8: Recreate ROLE_PERMISSIONS junction table
CREATE TABLE role_permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Step 9: Re-seed original data
INSERT INTO relations (name, type, description) VALUES
  ('Admin', 'tenant', 'Full administrative access to the tenant'),
  ('Writer', 'tenant', 'Can create and modify content'),
  ('Viewer', 'tenant', 'Read-only access to tenant resources'),
  ('Basic', 'tenant', 'Basic tenant membership with minimal permissions');

INSERT INTO roles (name, description) VALUES
  ('Tenant Admin Role', 'Full administrative permissions for tenant management'),
  ('Content Writer Role', 'Permissions for creating and editing content'),
  ('Content Viewer Role', 'Read-only permissions for viewing content'),
  ('Basic Member Role', 'Minimal permissions for basic tenant members');

-- Re-map data
DO $$
DECLARE
  admin_rel_id UUID;
  writer_rel_id UUID;
  viewer_rel_id UUID;
  basic_rel_id UUID;
  
  tenant_admin_role_id UUID;
  writer_role_id UUID;
  viewer_role_id UUID;
  basic_role_id UUID;
BEGIN
  SELECT id INTO admin_rel_id FROM relations WHERE name = 'Admin';
  SELECT id INTO writer_rel_id FROM relations WHERE name = 'Writer';
  SELECT id INTO viewer_rel_id FROM relations WHERE name = 'Viewer';
  SELECT id INTO basic_rel_id FROM relations WHERE name = 'Basic';
  
  SELECT id INTO tenant_admin_role_id FROM roles WHERE name = 'Tenant Admin Role';
  SELECT id INTO writer_role_id FROM roles WHERE name = 'Content Writer Role';
  SELECT id INTO viewer_role_id FROM roles WHERE name = 'Content Viewer Role';
  SELECT id INTO basic_role_id FROM roles WHERE name = 'Basic Member Role';
  
  INSERT INTO relation_roles (relation_id, role_id) VALUES
    (admin_rel_id, tenant_admin_role_id),
    (writer_rel_id, writer_role_id),
    (viewer_rel_id, viewer_role_id),
    (basic_rel_id, basic_role_id);
END $$;

-- Re-map permissions
DO $$
DECLARE
  tenant_admin_role_id UUID;
  writer_role_id UUID;
  viewer_role_id UUID;
  basic_role_id UUID;
  perm_id UUID;
BEGIN
  SELECT id INTO tenant_admin_role_id FROM roles WHERE name = 'Tenant Admin Role';
  SELECT id INTO writer_role_id FROM roles WHERE name = 'Content Writer Role';
  SELECT id INTO viewer_role_id FROM roles WHERE name = 'Content Viewer Role';
  SELECT id INTO basic_role_id FROM roles WHERE name = 'Basic Member Role';
  
  FOR perm_id IN SELECT id FROM permissions WHERE app = 'tenant-api' LOOP
    INSERT INTO role_permissions (role_id, permission_id) VALUES (tenant_admin_role_id, perm_id);
  END LOOP;
  
  FOR perm_id IN SELECT id FROM permissions 
    WHERE app = 'tenant-api' 
    AND action IN ('create', 'read', 'update') LOOP
    INSERT INTO role_permissions (role_id, permission_id) VALUES (writer_role_id, perm_id);
  END LOOP;
  
  FOR perm_id IN SELECT id FROM permissions 
    WHERE app = 'tenant-api' 
    AND action = 'read' LOOP
    INSERT INTO role_permissions (role_id, permission_id) VALUES (viewer_role_id, perm_id);
  END LOOP;
  
  FOR perm_id IN SELECT id FROM permissions 
    WHERE app = 'tenant-api' 
    AND entity IN ('tenant', 'member') 
    AND action = 'read' LOOP
    INSERT INTO role_permissions (role_id, permission_id) VALUES (basic_role_id, perm_id);
  END LOOP;
END $$;

-- Restore foreign keys
ALTER TABLE tenant_members 
  ADD CONSTRAINT tenant_members_relation_id_fkey 
  FOREIGN KEY (relation_id) REFERENCES relations(id);

ALTER TABLE user_invitations 
  ADD CONSTRAINT user_invitations_relation_id_fkey 
  FOREIGN KEY (relation_id) REFERENCES relations(id);

-- Update existing records to use basic relation
UPDATE tenant_members 
SET relation_id = (SELECT id FROM relations WHERE name = 'Basic' LIMIT 1)
WHERE relation_id IS NULL;

CREATE INDEX idx_tenant_members_relation_id ON tenant_members(relation_id);

