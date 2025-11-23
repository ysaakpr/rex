-- Revert grace period rotation support

-- Drop indexes
DROP INDEX IF EXISTS idx_system_users_is_primary;
DROP INDEX IF EXISTS idx_system_users_expires_at;
DROP INDEX IF EXISTS idx_system_users_application_name;

-- Drop columns
ALTER TABLE system_users DROP COLUMN IF EXISTS is_primary;
ALTER TABLE system_users DROP COLUMN IF EXISTS expires_at;
ALTER TABLE system_users DROP COLUMN IF EXISTS application_name;

