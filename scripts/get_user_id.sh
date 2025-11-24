#!/bin/bash
# Helper script to get a user's ID by email address

set -e

echo "=========================================="
echo "Get User ID by Email"
echo "=========================================="
echo ""

EMAIL=${1:-""}

if [ -z "$EMAIL" ]; then
    echo "Usage: $0 <email>"
    echo ""
    echo "Example: $0 user@example.com"
    echo ""
    exit 1
fi

echo "Searching for user: $EMAIL"
echo ""

# SuperTokens uses its own database
ST_DB_USER=${ST_DB_USER:-supertokens}
ST_DB_NAME=${ST_DB_NAME:-supertokens}

# Query SuperTokens database for user
docker-compose exec -T supertokens-db psql -U "$ST_DB_USER" -d "$ST_DB_NAME" <<EOF
-- Query emailpassword_users table in SuperTokens database
SELECT 
    user_id,
    email,
    to_timestamp(time_joined / 1000) as created_at
FROM emailpassword_users 
WHERE email = '$EMAIL';
EOF

echo ""
echo "Checking if user is a platform admin..."

# Check platform admin status in main database
DB_USER=${DB_USER:-utmuser}
DB_NAME=${DB_NAME:-utm_backend}

# Get user_id from previous query
USER_ID=\$(docker-compose exec -T supertokens-db psql -U "$ST_DB_USER" -d "$ST_DB_NAME" -t -A -c \
  "SELECT user_id FROM emailpassword_users WHERE email = '$EMAIL';" 2>/dev/null | head -1)

if [ ! -z "\$USER_ID" ]; then
    docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" <<ADMINEOF
SELECT 
    CASE 
        WHEN EXISTS(SELECT 1 FROM platform_admins WHERE user_id = '\$USER_ID') 
        THEN '👑 Platform Admin'
        ELSE 'Regular User'
    END as status;
ADMINEOF
fi

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Error querying database"
    echo ""
    echo "Alternative method:"
    echo "1. Log in as the user"
    echo "2. Open browser console (F12)"
    echo "3. Run: await Session.getUserId()"
    echo "4. Copy the ID that appears"
fi

