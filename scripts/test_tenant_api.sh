#!/bin/bash

# Simple tenant API test with curl and cookie jar

set -e

BASE_URL="http://localhost:8080"
COOKIE_JAR="/tmp/utm_cookies.txt"

# Clean up old cookies
rm -f "$COOKIE_JAR"

echo "=== Testing UTM Backend Tenant API ==="
echo ""

# Create unique email
TIMESTAMP=$(date +%s)
EMAIL="testuser-${TIMESTAMP}@example.com"
PASSWORD="TestPassword123!"

echo "1. Creating user: $EMAIL"
SIGNUP_RESULT=$(curl -s -X POST "${BASE_URL}/auth/signup" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c "$COOKIE_JAR" \
  -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")

echo "$SIGNUP_RESULT" | jq '.'

if echo "$SIGNUP_RESULT" | jq -e '.status == "OK"' > /dev/null; then
  USER_ID=$(echo "$SIGNUP_RESULT" | jq -r '.user.id')
  echo "✓ User created: $USER_ID"
elif echo "$SIGNUP_RESULT" | jq -e '.status' | grep -q "FIELD_ERROR\|EMAIL_ALREADY_EXISTS"; then
  echo "User already exists, proceeding..."
else
  echo "✗ Failed to create user"
  exit 1
fi

echo ""
echo "2. Signing in..."
SIGNIN_RESULT=$(curl -s -X POST "${BASE_URL}/auth/signin" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -b "$COOKIE_JAR" \
  -c "$COOKIE_JAR" \
  -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"${PASSWORD}\"}]}")

echo "$SIGNIN_RESULT" | jq '.'

if echo "$SIGNIN_RESULT" | jq -e '.status == "OK"' > /dev/null; then
  USER_ID=$(echo "$SIGNIN_RESULT" | jq -r '.user.id')
  echo "✓ Signed in: $USER_ID"
  
  echo ""
  echo "Cookies stored:"
  cat "$COOKIE_JAR" | grep -v "^#" | awk '{print "  " $6 "=" substr($0, index($0,$7))}'
else
  echo "✗ Failed to sign in"
  exit 1
fi

echo ""
echo "3. Creating tenant..."
TENANT_NAME="Test Company ${TIMESTAMP}"
TENANT_SLUG="test-company-${TIMESTAMP}"

TENANT_RESULT=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d "{\"name\":\"${TENANT_NAME}\",\"slug\":\"${TENANT_SLUG}\",\"metadata\":{\"industry\":\"technology\",\"test\":true}}")

HTTP_STATUS=$(echo "$TENANT_RESULT" | grep "HTTP_STATUS" | cut -d':' -f2)
BODY=$(echo "$TENANT_RESULT" | sed '$d')

echo "$BODY" | jq '.'
echo "HTTP Status: $HTTP_STATUS"

if [ "$HTTP_STATUS" = "201" ]; then
  TENANT_ID=$(echo "$BODY" | jq -r '.data.id')
  echo "✓ Tenant created: $TENANT_ID"
  
  echo ""
  echo "4. Checking tenant status..."
  sleep 2
  STATUS_RESULT=$(curl -s -X GET "${BASE_URL}/api/v1/tenants/${TENANT_ID}/status" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR")
  
  echo "$STATUS_RESULT" | jq '.'
  
  echo ""
  echo "5. Listing tenants..."
  LIST_RESULT=$(curl -s -X GET "${BASE_URL}/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR")
  
  echo "$LIST_RESULT" | jq '.'
  
  echo ""
  echo "=== SUCCESS! ==="
  echo ""
  echo "Summary:"
  echo "  User ID: $USER_ID"
  echo "  Email: $EMAIL"
  echo "  Tenant ID: $TENANT_ID"
  echo "  Tenant Name: $TENANT_NAME"
  echo "  Tenant Slug: $TENANT_SLUG"
else
  echo "✗ Failed to create tenant (HTTP $HTTP_STATUS)"
  echo "Response: $BODY"
  echo ""
  echo "Debugging: Cookies in jar:"
  cat "$COOKIE_JAR"
  exit 1
fi

# Cleanup
rm -f "$COOKIE_JAR"

