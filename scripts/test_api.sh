#!/bin/bash

# Test API Script for UTM Backend
# This script creates a user, signs in, and tests tenant creation

set -e

BASE_URL="http://localhost:8080"
API_DOMAIN="http://localhost:8080"

echo "=== UTM Backend API Test ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test data
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="TestPassword123!"
TENANT_NAME="Test Company"
TENANT_SLUG="test-company"

echo -e "${YELLOW}1. Creating test user...${NC}"
SIGNUP_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/signup" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -d "{
    \"formFields\": [
      {
        \"id\": \"email\",
        \"value\": \"${TEST_EMAIL}\"
      },
      {
        \"id\": \"password\",
        \"value\": \"${TEST_PASSWORD}\"
      }
    ]
  }")

echo "$SIGNUP_RESPONSE" | jq '.'

if echo "$SIGNUP_RESPONSE" | grep -q '"status":"OK"'; then
  echo -e "${GREEN}✓ User created successfully${NC}"
  USER_ID=$(echo "$SIGNUP_RESPONSE" | jq -r '.user.id')
  echo "User ID: $USER_ID"
elif echo "$SIGNUP_RESPONSE" | grep -q '"status":"EMAIL_ALREADY_EXISTS_ERROR"'; then
  echo -e "${YELLOW}! User already exists, proceeding with sign in...${NC}"
else
  echo -e "${RED}✗ Failed to create user${NC}"
  exit 1
fi

echo ""
echo -e "${YELLOW}2. Signing in...${NC}"
SIGNIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/signin" \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -c cookies.txt \
  -b cookies.txt \
  -d "{
    \"formFields\": [
      {
        \"id\": \"email\",
        \"value\": \"${TEST_EMAIL}\"
      },
      {
        \"id\": \"password\",
        \"value\": \"${TEST_PASSWORD}\"
      }
    ]
  }")

echo "$SIGNIN_RESPONSE" | jq '.'

if echo "$SIGNIN_RESPONSE" | grep -q '"status":"OK"'; then
  echo -e "${GREEN}✓ Signed in successfully${NC}"
  USER_ID=$(echo "$SIGNIN_RESPONSE" | jq -r '.user.id')
  echo "User ID: $USER_ID"
else
  echo -e "${RED}✗ Failed to sign in${NC}"
  exit 1
fi

echo ""
echo -e "${YELLOW}3. Creating tenant...${NC}"
TENANT_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{
    \"name\": \"${TENANT_NAME}\",
    \"slug\": \"${TENANT_SLUG}\",
    \"metadata\": {
      \"industry\": \"technology\",
      \"size\": \"10-50\",
      \"test\": true
    }
  }")

echo "$TENANT_RESPONSE" | jq '.'

if echo "$TENANT_RESPONSE" | grep -q '"success":true'; then
  echo -e "${GREEN}✓ Tenant created successfully${NC}"
  TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.data.id')
  echo "Tenant ID: $TENANT_ID"
  
  echo ""
  echo -e "${YELLOW}4. Checking tenant status...${NC}"
  sleep 2
  STATUS_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/tenants/${TENANT_ID}/status" \
    -H "Content-Type: application/json" \
    -b cookies.txt)
  
  echo "$STATUS_RESPONSE" | jq '.'
  
  echo ""
  echo -e "${YELLOW}5. Listing user's tenants...${NC}"
  LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -b cookies.txt)
  
  echo "$LIST_RESPONSE" | jq '.'
  
  echo ""
  echo -e "${YELLOW}6. Getting tenant details...${NC}"
  DETAIL_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/tenants/${TENANT_ID}" \
    -H "Content-Type: application/json" \
    -b cookies.txt)
  
  echo "$DETAIL_RESPONSE" | jq '.'
  
  echo ""
  echo -e "${GREEN}=== All tests passed! ===${NC}"
  echo ""
  echo "Summary:"
  echo "  User Email: ${TEST_EMAIL}"
  echo "  User ID: ${USER_ID}"
  echo "  Tenant ID: ${TENANT_ID}"
  echo "  Tenant Name: ${TENANT_NAME}"
  echo "  Tenant Slug: ${TENANT_SLUG}"
else
  echo -e "${RED}✗ Failed to create tenant${NC}"
  echo "Error response:"
  echo "$TENANT_RESPONSE" | jq '.'
  exit 1
fi

# Cleanup
rm -f cookies.txt

echo ""
echo -e "${GREEN}Test completed successfully!${NC}"

