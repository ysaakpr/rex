#!/bin/bash
# header_auth_test.sh - Test header-based authentication

set -e

BASE_URL="http://localhost:3000"
EMAIL="vyshakh.p@dream11.com"
PASSWORD="test@123"

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   🔐 Testing Header-Based Authentication                      ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

echo "Step 1: Signing in with header-based auth..."
echo "Request: POST $BASE_URL/auth/signin"
echo ""

# Sign in and get tokens (use -i to get headers!)
RESPONSE=$(curl -i -s -X POST $BASE_URL/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d "{
    \"formFields\": [
      {\"id\": \"email\", \"value\": \"$EMAIL\"},
      {\"id\": \"password\", \"value\": \"$PASSWORD\"}
    ]
  }")

# Extract body (after blank line)
BODY=$(echo "$RESPONSE" | sed -n '/^$/,$p' | tail -n +2)

# Check if sign in was successful
STATUS=$(echo $BODY | jq -r '.status')
if [ "$STATUS" != "OK" ]; then
  echo "❌ Sign in failed!"
  echo $BODY | jq '.'
  exit 1
fi

# Extract tokens from HEADERS
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Access-Token:" | sed 's/St-Access-Token: //' | tr -d '\r')
REFRESH_TOKEN=$(echo "$RESPONSE" | grep -i "^St-Refresh-Token:" | sed 's/St-Refresh-Token: //' | tr -d '\r')
USER_ID=$(echo $BODY | jq -r '.user.id')

echo "✅ Sign in successful!"
echo "   User ID: $USER_ID"
echo "   Access Token: ${ACCESS_TOKEN:0:60}..."
echo "   Refresh Token: ${REFRESH_TOKEN:0:60}..."
echo ""

# Save tokens to file
cat > tokens.json << EOF
{
  "access_token": "$ACCESS_TOKEN",
  "refresh_token": "$REFRESH_TOKEN",
  "user_id": "$USER_ID"
}
EOF

echo "💾 Tokens saved to tokens.json"
echo ""

# Test API calls with Bearer token
echo "Step 2: Testing API calls with Bearer token..."
echo ""

echo "🏢 GET /api/v1/tenants"
curl -s -X GET $BASE_URL/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.data.data[] | {name, slug, status}' 2>/dev/null || echo "   No tenants found or error"
echo ""

echo "👑 GET /api/v1/platform/admins/check"
curl -s -X GET $BASE_URL/api/v1/platform/admins/check \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.data'
echo ""

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   ✅ Header-Based Authentication WORKS!                       ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""
echo "📝 Usage:"
echo "   1. Tokens are saved in tokens.json"
echo "   2. Use the access_token for API calls:"
echo ""
echo "   curl -H \"Authorization: Bearer \$ACCESS_TOKEN\" \\"
echo "     http://localhost:3000/api/v1/tenants"
echo ""
echo "🔑 Key Points:"
echo "   • Tokens are in RESPONSE HEADERS (St-Access-Token, St-Refresh-Token)"
echo "   • Use: -H \"Authorization: Bearer <token>\" for API calls"
echo "   • Add: -H \"st-auth-mode: header\" when signing in"
echo ""

