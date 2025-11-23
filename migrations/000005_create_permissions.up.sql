CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service VARCHAR(100) NOT NULL,
    entity VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service, entity, action)
);

CREATE INDEX idx_permissions_service ON permissions(service);
CREATE INDEX idx_permissions_entity ON permissions(entity);

-- Insert default permissions for tenant management
INSERT INTO permissions (service, entity, action, description) VALUES
-- Tenant service permissions
('tenant-api', 'tenant', 'create', 'Create new tenant'),
('tenant-api', 'tenant', 'read', 'View tenant details'),
('tenant-api', 'tenant', 'update', 'Update tenant information'),
('tenant-api', 'tenant', 'delete', 'Delete tenant'),
('tenant-api', 'member', 'create', 'Add member to tenant'),
('tenant-api', 'member', 'read', 'View tenant members'),
('tenant-api', 'member', 'update', 'Update member details'),
('tenant-api', 'member', 'delete', 'Remove member from tenant'),
('tenant-api', 'invitation', 'create', 'Invite users to tenant'),
('tenant-api', 'invitation', 'read', 'View invitations'),
('tenant-api', 'invitation', 'delete', 'Cancel invitations'),
('tenant-api', 'role', 'create', 'Create roles'),
('tenant-api', 'role', 'read', 'View roles'),
('tenant-api', 'role', 'update', 'Update roles'),
('tenant-api', 'role', 'delete', 'Delete roles'),
('tenant-api', 'permission', 'assign', 'Assign permissions to roles'),
('tenant-api', 'permission', 'revoke', 'Revoke permissions from roles'),

-- Analytics service permissions (example for other services)
('analytics-api', 'report', 'create', 'Create analytics reports'),
('analytics-api', 'report', 'read', 'View analytics reports'),
('analytics-api', 'report', 'update', 'Update analytics reports'),
('analytics-api', 'report', 'delete', 'Delete analytics reports'),
('analytics-api', 'dashboard', 'read', 'View analytics dashboard'),

-- Content service permissions (example)
('content-api', 'content', 'create', 'Create content'),
('content-api', 'content', 'read', 'View content'),
('content-api', 'content', 'update', 'Update content'),
('content-api', 'content', 'delete', 'Delete content'),
('content-api', 'content', 'publish', 'Publish content');

