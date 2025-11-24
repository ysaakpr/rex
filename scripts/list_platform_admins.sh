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

# Check if postgres container is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ Error: PostgreSQL container is not running"
    exit 1
fi

docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" <<EOF
-- List all platform admins with their email addresses
SELECT 
    pa.user_id,
    COALESCE(ep.email, 'Unknown') as email,
    pa.created_by,
    pa.created_at,
    CASE 
        WHEN pa.created_at > NOW() - INTERVAL '1 day' THEN '🆕'
        ELSE '✓'
    END as status
FROM platform_admins pa
LEFT JOIN emailpassword_users ep ON pa.user_id = ep.user_id
ORDER BY pa.created_at DESC;

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

