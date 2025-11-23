#!/bin/bash

set -e

BASE_URL="http://localhost:8080"

echo "=== Testing UTM Backend with Header Auth ==="
echo ""

TIMESTAMP=$(date +%s)
EMAIL="testuser-${TIMESTAMP}@example.com"
PASSWORD="TestPassword123!"

echo "1. Creating and signing in user: $EMAIL"

# Sign in and capture headers
RESPONSE=$(curl -s -D /tmp/headers.txt -X POST "${BASE_URL}/auth/signin" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")

# If user doesn't exist, create first
if echo "$RESPONSE" | grep -q "WRONG_CREDENTIALS"; then
  echo "User doesn't exist, creating..."
  RESPONSE=$(curl -s -D /tmp/headers.txt -X POST "${BASE_URL}/auth/signup" \
    -H "Content-Type: application/json" \
    -H "rid: emailpassword" \
    -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")
  
  # Now sign in
  RESPONSE=$(curl -s -D /tmp/headers.txt -X POST "${BASE_URL}/auth/signin" \
    -H "Content-Type: application/json" \
    -H "rid: emailpassword" \
    -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")
fi

echo "$RESPONSE" | jq '.'
USER_ID=$(echo "$RESPONSE" | jq -r '.user.id')
echo "✓ User ID: $USER_ID"

# Extract tokens from headers
ACCESS_TOKEN=$(grep -i "st-access-token:" /tmp/headers.txt | cut -d' ' -f2- | tr -d '\r\n')
REFRESH_TOKEN=$(grep -i "st-refresh-token:" /tmp/headers.txt | cut -d' ' -f2- | tr -d '\r\n')
FRONT_TOKEN=$(grep -i "front-token:" /tmp/headers.txt | cut -d' ' -f2- | tr -d '\r\n')

echo ""
echo "Tokens extracted:"
echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
echo "  Front Token: ${FRONT_TOKEN:0:50}..."

echo ""
echo "2. Creating tenant..."

TENANT_RESULT=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -H "st-access-token: ${ACCESS_TOKEN}" \
  -H "front-token: ${FRONT_TOKEN}" \
  -d "{\"name\":\"Test Company ${TIMESTAMP}\",\"slug\":\"test-company-${TIMESTAMP}\",\"metadata\":{\"industry\":\"technology\"}}")

HTTP_STATUS=$(echo "$TENANT_RESULT" | grep "HTTP_STATUS" | cut -d':' -f2)
BODY=$(echo "$TENANT_RESULT" | sed '$d')

echo "$BODY" | jq '.'
echo "HTTP Status: $HTTP_STATUS"

if [ "$HTTP_STATUS" = "201" ]; then
  TENANT_ID=$(echo "$BODY" | jq -r '.data.id')
  echo "✓ Tenant created: $TENANT_ID"
  
  echo ""
  echo "3. Listing tenants..."
  curl -s -X GET "${BASE_URL}/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -H "st-access-token: ${ACCESS_TOKEN}" \
    -H "front-token: ${FRONT_TOKEN}" | jq '.'
  
  echo ""
  echo "=== SUCCESS! ==="
  echo ""
  echo "Summary:"
  echo "  User: $EMAIL"
  echo "  User ID: $USER_ID"
  echo "  Tenant ID: $TENANT_ID"
else
  echo "✗ Failed (HTTP $HTTP_STATUS)"
fi

rm -f /tmp/headers.txt
