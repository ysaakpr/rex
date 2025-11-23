#!/bin/bash

# Script to clean up orphaned SuperTokens users
# Usage: ./scripts/cleanup_orphaned_supertoken_user.sh <email>

EMAIL="$1"

if [ -z "$EMAIL" ]; then
  echo "Usage: $0 <email>"
  echo "Example: $0 test-worker@system.internal"
  exit 1
fi

echo "🔍 Looking for SuperTokens user with email: $EMAIL"

# Get the user_id from SuperTokens
USER_ID=$(docker-compose exec -T supertokens-db psql -U supertokens -d supertokens -t -c "
  SELECT user_id FROM emailpassword_users WHERE email = '$EMAIL';
" | xargs)

if [ -z "$USER_ID" ]; then
  echo "❌ No SuperTokens user found with email: $EMAIL"
  exit 0
fi

echo "✅ Found SuperTokens user: $USER_ID"
echo "🗑️  Deleting user from SuperTokens..."

# Delete from all SuperTokens tables
docker-compose exec -T supertokens-db psql -U supertokens -d supertokens << EOSQL
-- Delete from session tables
DELETE FROM session_info WHERE user_id = '$USER_ID';
DELETE FROM session_access_token_signing_keys WHERE user_id = '$USER_ID';

-- Delete from user metadata
DELETE FROM user_metadata WHERE user_id = '$USER_ID';

-- Delete from email verification
DELETE FROM emailverification_verified_emails WHERE user_id = '$USER_ID';
DELETE FROM emailverification_tokens WHERE user_id = '$USER_ID';

-- Delete from password reset
DELETE FROM emailpassword_pswd_reset_tokens WHERE user_id = '$USER_ID';

-- Delete the main user record
DELETE FROM emailpassword_users WHERE user_id = '$USER_ID';
DELETE FROM emailpassword_user_to_tenant WHERE user_id = '$USER_ID';

-- Delete from all user roles (if exists)
DELETE FROM all_auth_recipe_users WHERE user_id = '$USER_ID';
EOSQL

echo "✅ Successfully deleted orphaned SuperTokens user: $EMAIL"
echo ""
echo "You can now try creating the application again!"
