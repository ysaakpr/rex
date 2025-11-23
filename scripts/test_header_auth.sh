#!/bin/bash

# Test header-based authentication with SuperTokens

set -e

BASE_URL="http://localhost:8080"

echo "=== Testing Header-Based Authentication ==="
echo ""

TIMESTAMP=$(date +%s)
EMAIL="api-test-${TIMESTAMP}@example.com"
PASSWORD="TestPassword123!"

echo "1. Creating user: $EMAIL"

# Sign up with header mode
SIGNUP_RESPONSE=$(curl -s -D /tmp/signup_headers.txt -X POST "${BASE_URL}/auth/signup" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -H "st-auth-mode: header" \
  -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")

echo "$SIGNUP_RESPONSE" | jq '.'

# Extract tokens
ACCESS_TOKEN=$(grep -i "st-access-token:" /tmp/signup_headers.txt | cut -d' ' -f2- | tr -d '\r\n' || echo "")
REFRESH_TOKEN=$(grep -i "st-refresh-token:" /tmp/signup_headers.txt | cut -d' ' -f2- | tr -d '\r\n' || echo "")
FRONT_TOKEN=$(grep -i "front-token:" /tmp/signup_headers.txt | cut -d' ' -f2- | tr -d '\r\n' || echo "")

if [ -z "$ACCESS_TOKEN" ]; then
  echo "✗ No access token received. Trying sign in..."
  
  SIGNIN_RESPONSE=$(curl -s -D /tmp/signin_headers.txt -X POST "${BASE_URL}/auth/signin" \
    -H "Content-Type: application/json" \
    -H "rid: emailpassword" \
    -H "st-auth-mode: header" \
    -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")
  
  echo "$SIGNIN_RESPONSE" | jq '.'
  
  ACCESS_TOKEN=$(grep -i "st-access-token:" /tmp/signin_headers.txt | cut -d' ' -f2- | tr -d '\r\n' || echo "")
  FRONT_TOKEN=$(grep -i "front-token:" /tmp/signin_headers.txt | cut -d' ' -f2- | tr -d '\r\n' || echo "")
fi

if [ -z "$ACCESS_TOKEN" ]; then
  echo "✗ Failed to get access token"
  exit 1
fi

echo ""
echo "✓ Tokens received:"
echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
echo "  Front Token: ${FRONT_TOKEN:0:50}..."

echo ""
echo "2. Creating tenant with header auth..."

TENANT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -H "st-auth-mode: header" \
  -H "st-access-token: ${ACCESS_TOKEN}" \
  -H "front-token: ${FRONT_TOKEN}" \
  -d "{\"name\":\"API Test Company ${TIMESTAMP}\",\"slug\":\"api-test-${TIMESTAMP}\",\"metadata\":{\"test\":true}}")

HTTP_STATUS=$(echo "$TENANT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
BODY=$(echo "$TENANT_RESPONSE" | sed '$d')

echo "$BODY" | jq '.'
echo "HTTP Status: $HTTP_STATUS"

if [ "$HTTP_STATUS" = "201" ]; then
  TENANT_ID=$(echo "$BODY" | jq -r '.data.id')
  echo ""
  echo "✓ Tenant created successfully!"
  echo "  Tenant ID: $TENANT_ID"
  
  echo ""
  echo "3. Listing tenants..."
  curl -s -X GET "${BASE_URL}/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -H "st-auth-mode: header" \
    -H "st-access-token: ${ACCESS_TOKEN}" \
    -H "front-token: ${FRONT_TOKEN}" | jq '.'
  
  echo ""
  echo "=== SUCCESS! Header-based auth works! ==="
else
  echo "✗ Failed to create tenant"
  exit 1
fi

# Cleanup
rm -f /tmp/signup_headers.txt /tmp/signin_headers.txt

