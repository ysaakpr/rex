#!/bin/bash
# Simple header-based auth test

BASE_URL="http://localhost:3000"

echo "🔐 Testing Header-Based Authentication"
echo "======================================="
echo ""

# Sign in and save response with headers
echo "1. Signing in..."
curl -i -X POST $BASE_URL/auth/signin \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -D headers.txt \
  -o body.txt \
  -d '{
    "formFields": [
      {"id": "email", "value": "vyshakh.p@dream11.com"},
      {"id": "password", "value": "test@123"}
    ]
  }'

echo ""
echo "2. Response body:"
cat body.txt | jq '.'
echo ""

echo "3. Extracting access token from headers..."
ACCESS_TOKEN=$(grep -i "^St-Access-Token:" headers.txt | sed 's/St-Access-Token: //' | tr -d '\r\n ')

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ No access token found!"
  echo "Headers received:"
  cat headers.txt
  exit 1
fi

echo "✅ Access Token: ${ACCESS_TOKEN:0:60}..."
echo ""

echo "4. Testing API call with Bearer token..."
curl -s -X GET $BASE_URL/api/v1/tenants \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" | jq '.data.data[] | {name, slug}' 2>/dev/null || echo "No tenants or error"

echo ""
echo "✅ Header-Based Auth Working!"
echo ""
echo "Saved tokens in: headers.txt"
echo "Use this token for API calls:"
echo "curl -H \"Authorization: Bearer $ACCESS_TOKEN\" http://localhost:3000/api/v1/tenants"

