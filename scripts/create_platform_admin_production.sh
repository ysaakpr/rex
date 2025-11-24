#!/bin/bash
# Script to create a platform admin on production/cloud server
# Works with docker-compose setup

set -e

echo "=========================================="
echo "Platform Admin Creation - Production"
echo "=========================================="
echo ""

# Check if user_id is provided
USER_ID=${1:-""}

if [ -z "$USER_ID" ]; then
    echo "❌ Error: User ID is required"
    echo ""
    echo "Usage: $0 <user_id>"
    echo ""
    echo "Example: $0 04413f25-fdfa-42a0-a046-c3ad67d135fe"
    echo ""
    echo "To get a user's ID:"
    echo "  1. Log in to your application"
    echo "  2. Open browser console and run:"
    echo "     Session.getUserId().then(id => console.log(id))"
    echo "  3. Or check the user details in the database"
    echo ""
    exit 1
fi

echo "User ID: $USER_ID"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: docker-compose not found"
    exit 1
fi

# Check if postgres container is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ Error: PostgreSQL container is not running"
    echo "   Start services: docker-compose up -d"
    exit 1
fi

echo "✓ Docker Compose found"
echo "✓ PostgreSQL container is running"
echo ""

# Get database credentials from environment or use defaults
DB_USER=${DB_USER:-utmuser}
DB_NAME=${DB_NAME:-utm_backend}

echo "Creating platform admin..."
echo ""

# Execute SQL to add platform admin
docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Check if user exists in platform_admins
DO \$\$
DECLARE
    admin_exists BOOLEAN;
BEGIN
    SELECT EXISTS(SELECT 1 FROM platform_admins WHERE user_id = '$USER_ID') INTO admin_exists;
    
    IF admin_exists THEN
        RAISE NOTICE '⚠️  User is already a platform admin';
    ELSE
        -- Insert new platform admin
        INSERT INTO platform_admins (user_id, created_by, created_at)
        VALUES ('$USER_ID', 'system', NOW());
        RAISE NOTICE '✅ Platform admin created successfully!';
    END IF;
END \$\$;

-- Show the admin record
SELECT 
    user_id,
    created_by,
    created_at,
    CASE 
        WHEN created_at > NOW() - INTERVAL '1 minute' THEN '🆕 Just created'
        ELSE '✓ Existing'
    END as status
FROM platform_admins 
WHERE user_id = '$USER_ID';
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✅ Success!"
    echo "=========================================="
    echo ""
    echo "User $USER_ID is now a platform admin."
    echo ""
    echo "They can now access:"
    echo "  • Platform Admin Management"
    echo "  • Roles & Policies Management"
    echo "  • Permissions Management"
    echo ""
    echo "Access URLs:"
    echo "  • https://rex.stage.fauda.dream11.in/platform/admins"
    echo "  • https://rex.stage.fauda.dream11.in/roles"
    echo "  • https://rex.stage.fauda.dream11.in/permissions"
    echo ""
else
    echo ""
    echo "❌ Failed to create platform admin"
    echo "   Check the error messages above"
    exit 1
fi

