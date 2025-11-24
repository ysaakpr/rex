#!/bin/bash
# Simple script to list platform admins (without email lookup)
# Use this if SuperTokens database access has issues

set -e

echo "=========================================="
echo "Platform Admins List (Simple)"
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
-- List all platform admins
\echo 'Platform Admins:'
SELECT 
    user_id,
    created_by,
    created_at,
    CASE 
        WHEN created_at > NOW() - INTERVAL '1 day' THEN '🆕 New'
        WHEN created_at > NOW() - INTERVAL '7 days' THEN 'Recent'
        ELSE 'Active'
    END as status
FROM platform_admins 
ORDER BY created_at DESC;

\echo ''
\echo 'Summary:'
SELECT 
    COUNT(*) as total_platform_admins,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 day') as new_today,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_this_week
FROM platform_admins;
EOF

echo ""
echo "Note: To see email addresses, use: ./scripts/list_platform_admins.sh"
echo "      (requires SuperTokens database access)"
echo ""

