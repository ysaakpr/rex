-- Add grace period rotation support to system_users
-- This allows multiple credentials per application with expiry dates

-- Add application_name to group related system users
ALTER TABLE system_users ADD COLUMN application_name VARCHAR(100);

-- Add expiry support for grace period rotation
ALTER TABLE system_users ADD COLUMN expires_at TIMESTAMP;

-- Add primary flag to mark the current active credential
ALTER TABLE system_users ADD COLUMN is_primary BOOLEAN DEFAULT true NOT NULL;

-- Update existing records to use name as application_name
UPDATE system_users SET application_name = name WHERE application_name IS NULL;

-- Make application_name required going forward
ALTER TABLE system_users ALTER COLUMN application_name SET NOT NULL;

-- Add index for querying by application
CREATE INDEX idx_system_users_application_name ON system_users(application_name);
CREATE INDEX idx_system_users_expires_at ON system_users(expires_at);
CREATE INDEX idx_system_users_is_primary ON system_users(is_primary);

-- Add comments
COMMENT ON COLUMN system_users.application_name IS 'Logical application name - multiple system users can belong to same application';
COMMENT ON COLUMN system_users.expires_at IS 'Grace period expiry - NULL means never expires, used for credential rotation';
COMMENT ON COLUMN system_users.is_primary IS 'Marks the current/recommended credential for the application';

