#!/bin/bash

echo "🧪 Testing Platform Admin Access"
echo "=================================="
echo ""

# Check if user is signed in
echo "1. Testing authentication..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/v1/platform/admins/check -b /tmp/test_cookies.txt 2>/dev/null || echo "401")

if [ "$RESPONSE" = "401" ]; then
    echo "❌ Not authenticated. Please sign in first."
    exit 1
fi

echo "✅ Authenticated"
echo ""

# Check platform admin status
echo "2. Checking platform admin status..."
ADMIN_CHECK=$(curl -s http://localhost:3000/api/v1/platform/admins/check -b /tmp/test_cookies.txt 2>/dev/null || echo "{}")
echo "$ADMIN_CHECK" | jq '.' 2>/dev/null || echo "$ADMIN_CHECK"
echo ""

# Test relations endpoint
echo "3. Testing relations endpoint..."
RELATIONS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/v1/relations -b /tmp/test_cookies.txt 2>/dev/null || echo "000")
if [ "$RELATIONS_RESPONSE" = "200" ]; then
    echo "✅ Relations endpoint accessible"
else
    echo "❌ Relations endpoint returned: $RELATIONS_RESPONSE"
fi
echo ""

# Test platform relations endpoint (requires admin)
echo "4. Testing platform relations endpoint..."
PLATFORM_RELATIONS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/v1/platform/relations -b /tmp/test_cookies.txt 2>/dev/null || echo "000")
if [ "$PLATFORM_RELATIONS" = "200" ]; then
    echo "✅ Platform relations endpoint accessible (you are a platform admin!)"
else
    echo "⚠️  Platform relations endpoint returned: $PLATFORM_RELATIONS (requires platform admin)"
fi
echo ""

echo "=================================="
echo "Test complete!"
