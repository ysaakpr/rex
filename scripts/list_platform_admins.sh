#!/bin/bash
# Script to list all platform admins

set -e

echo "=========================================="
echo "Platform Admins List"
echo "=========================================="
echo ""

# Get database credentials
DB_USER=${DB_USER:-utmuser}
DB_NAME=${DB_NAME:-utm_backend}
ST_DB_USER=${ST_DB_USER:-supertokens}
ST_DB_NAME=${ST_DB_NAME:-supertokens}

# Check if containers are running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ Error: PostgreSQL container is not running"
    exit 1
fi

if ! docker-compose ps supertokens-db | grep -q "Up"; then
    echo "❌ Error: SuperTokens database container is not running"
    exit 1
fi

# First, get all platform admins
echo "Fetching platform admins..."
ADMIN_IDS=$(docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -t -A -c \
  "SELECT user_id FROM platform_admins ORDER BY created_at DESC;")

echo ""
printf "%-38s %-30s %-15s %-25s %s\n" "USER_ID" "EMAIL" "CREATED_BY" "CREATED_AT" "STATUS"
echo "--------------------------------------------------------------------------------------------------------"

# For each admin, get their email from SuperTokens
for USER_ID in $ADMIN_IDS; do
    # Get email from SuperTokens database
    EMAIL=$(docker-compose exec -T supertokens-db psql -U "$ST_DB_USER" -d "$ST_DB_NAME" -t -A -c \
      "SELECT email FROM emailpassword_users WHERE user_id = '$USER_ID';" 2>/dev/null | head -1)
    
    # Get created info from platform_admins
    ADMIN_INFO=$(docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -t -A -c \
      "SELECT created_by, created_at, CASE WHEN created_at > NOW() - INTERVAL '1 day' THEN '🆕' ELSE '✓' END \
       FROM platform_admins WHERE user_id = '$USER_ID';" 2>/dev/null | head -1)
    
    CREATED_BY=$(echo "$ADMIN_INFO" | cut -d'|' -f1)
    CREATED_AT=$(echo "$ADMIN_INFO" | cut -d'|' -f2)
    STATUS=$(echo "$ADMIN_INFO" | cut -d'|' -f3)
    
    printf "%-38s %-30s %-15s %-25s %s\n" "$USER_ID" "${EMAIL:-Unknown}" "$CREATED_BY" "$CREATED_AT" "$STATUS"
done

echo ""
docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Show count
SELECT COUNT(*) as total_platform_admins FROM platform_admins;
EOF

echo ""
echo "To add a new platform admin:"
echo "  ./scripts/create_platform_admin_production.sh <user_id>"
echo ""
echo "To get a user's ID:"
echo "  ./scripts/get_user_id.sh <email>"
echo ""

