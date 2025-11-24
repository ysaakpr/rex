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

# Get database credentials
DB_USER=${DB_USER:-utmuser}
DB_NAME=${DB_NAME:-utm_backend}

# Query SuperTokens database for user
docker-compose exec -T postgres psql -U "$DB_USER" -d utm_backend <<EOF
-- Query emailpassword_users table in SuperTokens schema
SELECT 
    user_id,
    email,
    time_joined as created_at,
    CASE 
        WHEN EXISTS(SELECT 1 FROM platform_admins WHERE platform_admins.user_id = emailpassword_users.user_id) 
        THEN '👑 Platform Admin'
        ELSE 'Regular User'
    END as status
FROM emailpassword_users 
WHERE email = '$EMAIL';
EOF

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

