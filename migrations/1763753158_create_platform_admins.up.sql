-- Create platform_admins table
CREATE TABLE IF NOT EXISTS platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_platform_admins_user_id ON platform_admins(user_id);

-- Create relation_roles table for relation-to-role mapping
CREATE TABLE IF NOT EXISTS relation_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relation_id UUID NOT NULL REFERENCES relations(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(relation_id, role_id)
);

CREATE INDEX idx_relation_roles_relation_id ON relation_roles(relation_id);
CREATE INDEX idx_relation_roles_role_id ON relation_roles(role_id);

-- Add platform-api permissions
INSERT INTO permissions (service, entity, action, description) VALUES
    ('platform-api', 'permission', 'create', 'Create platform permissions'),
    ('platform-api', 'permission', 'read', 'Read platform permissions'),
    ('platform-api', 'permission', 'update', 'Update platform permissions'),
    ('platform-api', 'permission', 'delete', 'Delete platform permissions'),
    ('platform-api', 'role', 'create', 'Create platform roles'),
    ('platform-api', 'role', 'read', 'Read platform roles'),
    ('platform-api', 'role', 'update', 'Update platform roles'),
    ('platform-api', 'role', 'delete', 'Delete platform roles'),
    ('platform-api', 'relation', 'create', 'Create platform relations'),
    ('platform-api', 'relation', 'read', 'Read platform relations'),
    ('platform-api', 'relation', 'update', 'Update platform relations'),
    ('platform-api', 'relation', 'delete', 'Delete platform relations'),
    ('platform-api', 'admin', 'create', 'Create platform admins'),
    ('platform-api', 'admin', 'read', 'Read platform admins'),
    ('platform-api', 'admin', 'delete', 'Delete platform admins')
ON CONFLICT (service, entity, action) DO NOTHING;
