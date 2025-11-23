-- Remove platform-api permissions
DELETE FROM permissions WHERE service = 'platform-api';

-- Drop relation_roles table
DROP TABLE IF EXISTS relation_roles;

-- Drop platform_admins table
DROP TABLE IF EXISTS platform_admins;
