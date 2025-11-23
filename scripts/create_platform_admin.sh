#!/bin/bash

# Script to manually add a platform admin
# Usage: ./scripts/create_platform_admin.sh <user_id>

USER_ID=${1:-""}

if [ -z "$USER_ID" ]; then
    echo "Usage: $0 <user_id>"
    echo "Example: $0 04413f25-fdfa-42a0-a046-c3ad67d135fe"
    exit 1
fi

# Database connection details from .env
source .env

echo "Adding platform admin for user: $USER_ID"

docker-compose exec postgres psql -U utmuser -d utm_backend << EOF
INSERT INTO platform_admins (user_id, created_by)
VALUES ('$USER_ID', 'system')
ON CONFLICT (user_id) DO NOTHING;

SELECT * FROM platform_admins WHERE user_id = '$USER_ID';
EOF

echo "✅ Platform admin added successfully!"
echo ""
echo "You can now access platform admin features at:"
echo "  http://localhost:3000/platform/admins"
echo "  http://localhost:3000/roles (will check for platform admin)"
echo "  http://localhost:3000/permissions (will check for platform admin)"

