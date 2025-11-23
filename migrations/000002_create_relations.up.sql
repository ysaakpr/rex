CREATE TABLE IF NOT EXISTS relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, tenant_id)
);

CREATE INDEX idx_relations_tenant_id ON relations(tenant_id);
CREATE INDEX idx_relations_is_system ON relations(is_system);

-- Insert default system relations
INSERT INTO relations (name, description, tenant_id, is_system) VALUES
('Admin', 'Full administrative access to tenant', NULL, true),
('Writer', 'Can create and edit content', NULL, true),
('Viewer', 'Read-only access', NULL, true),
('Basic', 'Basic tenant member access', NULL, true);

