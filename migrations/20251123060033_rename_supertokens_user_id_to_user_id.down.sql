-- Revert: Rename user_id back to supertokens_user_id in system_users table
ALTER TABLE system_users 
RENAME COLUMN user_id TO supertokens_user_id;

-- Rename the index back
DROP INDEX IF EXISTS idx_system_users_user_id;
CREATE INDEX idx_system_users_supertokens_id ON system_users(supertokens_user_id);

-- Update comment
COMMENT ON COLUMN system_users.supertokens_user_id IS 'Link to SuperTokens user record';

