-- Create system_users table for machine-to-machine authentication
CREATE TABLE system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    service_type VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_system_users_user_id ON system_users(user_id);
CREATE INDEX idx_system_users_active ON system_users(is_active);
CREATE INDEX idx_system_users_service_type ON system_users(service_type);
CREATE INDEX idx_system_users_created_by ON system_users(created_by);

-- Add comment to table
COMMENT ON TABLE system_users IS 'System users for machine-to-machine authentication and service accounts';
COMMENT ON COLUMN system_users.name IS 'Unique service name identifier';
COMMENT ON COLUMN system_users.email IS 'Email format: service-name@system.internal';
COMMENT ON COLUMN system_users.user_id IS 'Link to SuperTokens user record';
COMMENT ON COLUMN system_users.service_type IS 'Type of service: worker, integration, cron, api';
COMMENT ON COLUMN system_users.metadata IS 'Additional configuration: ip_whitelist, rate_limit, allowed_apis, etc.';

