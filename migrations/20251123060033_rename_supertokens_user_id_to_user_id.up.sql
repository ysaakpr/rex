-- Rename supertokens_user_id to user_id in system_users table
ALTER TABLE system_users 
RENAME COLUMN supertokens_user_id TO user_id;

-- Rename the index as well
DROP INDEX IF EXISTS idx_system_users_supertokens_id;
CREATE INDEX idx_system_users_user_id ON system_users(user_id);

-- Update comment
COMMENT ON COLUMN system_users.user_id IS 'Link to SuperTokens user record';

