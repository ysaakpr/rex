CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, tenant_id)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);

-- Insert default system roles
INSERT INTO roles (name, description, tenant_id, is_system) VALUES
('Tenant Admin', 'Full tenant management capabilities', NULL, true),
('Content Manager', 'Manage content across services', NULL, true),
('Analytics Viewer', 'View analytics and reports', NULL, true),
('User Manager', 'Manage users and invitations', NULL, true);

