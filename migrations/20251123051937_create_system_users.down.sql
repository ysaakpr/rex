-- Drop system_users table
DROP INDEX IF EXISTS idx_system_users_created_by;
DROP INDEX IF EXISTS idx_system_users_service_type;
DROP INDEX IF EXISTS idx_system_users_active;
DROP INDEX IF EXISTS idx_system_users_user_id;
DROP TABLE IF EXISTS system_users;

