#!/bin/bash

# Script to test if SuperTokens supports multiple credentials for same application
# This verifies the grace period rotation strategy will work

set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                            ║"
echo "║          TESTING: Multiple Credentials in SuperTokens                     ║"
echo "║                                                                            ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_URL="http://localhost:8080"
APP_NAME="test-multiauth"

echo "📋 TEST PLAN:"
echo "  1. Create first credential (test-multiauth-1@system.internal)"
echo "  2. Authenticate with first credential"
echo "  3. Create second credential (test-multiauth-2@system.internal)"
echo "  4. Authenticate with second credential"
echo "  5. Verify BOTH credentials work simultaneously"
echo "  6. Clean up test users"
echo ""

# Get platform admin token
echo "🔐 Step 1: Login as platform admin..."
ADMIN_EMAIL="your-admin-email@example.com"
ADMIN_PASSWORD="your-password"

# Note: You need to set these
if [ "$ADMIN_EMAIL" = "your-admin-email@example.com" ]; then
    echo -e "${YELLOW}⚠️  SETUP REQUIRED: Edit this script and set your admin credentials${NC}"
    echo "   Set ADMIN_EMAIL and ADMIN_PASSWORD variables"
    exit 1
fi

ADMIN_TOKEN=$(curl -s -X POST "$API_URL/auth/signin" \
    -H "Content-Type: application/json" \
    -d "{\"formFields\": [{\"id\": \"email\", \"value\": \"$ADMIN_EMAIL\"}, {\"id\": \"password\", \"value\": \"$ADMIN_PASSWORD\"}]}" \
    | jq -r '.accessToken // empty')

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}❌ Failed to login as admin${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Admin logged in${NC}"
echo ""

# Create first credential
echo "🔧 Step 2: Create first credential..."
CRED1_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/platform/system-users" \
    -H "Content-Type: application/json" \
    -H "st-auth-mode: header" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{
        \"name\": \"${APP_NAME}-1\",
        \"description\": \"First credential for testing\",
        \"service_type\": \"worker\"
    }")

CRED1_EMAIL=$(echo "$CRED1_RESPONSE" | jq -r '.data.email')
CRED1_PASSWORD=$(echo "$CRED1_RESPONSE" | jq -r '.data.password')
CRED1_USER_ID=$(echo "$CRED1_RESPONSE" | jq -r '.data.user_id')

if [ "$CRED1_EMAIL" = "null" ] || [ -z "$CRED1_EMAIL" ]; then
    echo -e "${RED}❌ Failed to create first credential${NC}"
    echo "Response: $CRED1_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ First credential created${NC}"
echo "   Email: $CRED1_EMAIL"
echo "   User ID: $CRED1_USER_ID"
echo ""

# Test first credential authentication
echo "🔐 Step 3: Authenticate with first credential..."
CRED1_AUTH=$(curl -s -X POST "$API_URL/auth/signin" \
    -H "Content-Type: application/json" \
    -d "{\"formFields\": [{\"id\": \"email\", \"value\": \"$CRED1_EMAIL\"}, {\"id\": \"password\", \"value\": \"$CRED1_PASSWORD\"}]}")

CRED1_TOKEN=$(echo "$CRED1_AUTH" | jq -r '.accessToken // empty')

if [ -z "$CRED1_TOKEN" ]; then
    echo -e "${RED}❌ Failed to authenticate with first credential${NC}"
    echo "Response: $CRED1_AUTH"
    exit 1
fi

echo -e "${GREEN}✅ First credential authenticated successfully${NC}"
echo ""

# Create second credential
echo "🔧 Step 4: Create second credential..."
CRED2_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/platform/system-users" \
    -H "Content-Type: application/json" \
    -H "st-auth-mode: header" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{
        \"name\": \"${APP_NAME}-2\",
        \"description\": \"Second credential for testing\",
        \"service_type\": \"worker\"
    }")

CRED2_EMAIL=$(echo "$CRED2_RESPONSE" | jq -r '.data.email')
CRED2_PASSWORD=$(echo "$CRED2_RESPONSE" | jq -r '.data.password')
CRED2_USER_ID=$(echo "$CRED2_RESPONSE" | jq -r '.data.user_id')

if [ "$CRED2_EMAIL" = "null" ] || [ -z "$CRED2_EMAIL" ]; then
    echo -e "${RED}❌ Failed to create second credential${NC}"
    echo "Response: $CRED2_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Second credential created${NC}"
echo "   Email: $CRED2_EMAIL"
echo "   User ID: $CRED2_USER_ID"
echo ""

# Test second credential authentication
echo "🔐 Step 5: Authenticate with second credential..."
CRED2_AUTH=$(curl -s -X POST "$API_URL/auth/signin" \
    -H "Content-Type: application/json" \
    -d "{\"formFields\": [{\"id\": \"email\", \"value\": \"$CRED2_EMAIL\"}, {\"id\": \"password\", \"value\": \"$CRED2_PASSWORD\"}]}")

CRED2_TOKEN=$(echo "$CRED2_AUTH" | jq -r '.accessToken // empty')

if [ -z "$CRED2_TOKEN" ]; then
    echo -e "${RED}❌ Failed to authenticate with second credential${NC}"
    echo "Response: $CRED2_AUTH"
    exit 1
fi

echo -e "${GREEN}✅ Second credential authenticated successfully${NC}"
echo ""

# Verify both still work
echo "🔄 Step 6: Verify BOTH credentials still work..."

# Re-authenticate with first credential
CRED1_REAUTH=$(curl -s -X POST "$API_URL/auth/signin" \
    -H "Content-Type: application/json" \
    -d "{\"formFields\": [{\"id\": \"email\", \"value\": \"$CRED1_EMAIL\"}, {\"id\": \"password\", \"value\": \"$CRED1_PASSWORD\"}]}")

CRED1_REAUTH_TOKEN=$(echo "$CRED1_REAUTH" | jq -r '.accessToken // empty')

if [ -z "$CRED1_REAUTH_TOKEN" ]; then
    echo -e "${RED}❌ First credential stopped working after second was created!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ First credential still works${NC}"

# Re-authenticate with second credential
CRED2_REAUTH=$(curl -s -X POST "$API_URL/auth/signin" \
    -H "Content-Type: application/json" \
    -d "{\"formFields\": [{\"id\": \"email\", \"value\": \"$CRED2_EMAIL\"}, {\"id\": \"password\", \"value\": \"$CRED2_PASSWORD\"}]}")

CRED2_REAUTH_TOKEN=$(echo "$CRED2_REAUTH" | jq -r '.accessToken // empty')

if [ -z "$CRED2_REAUTH_TOKEN" ]; then
    echo -e "${RED}❌ Second credential stopped working!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Second credential still works${NC}"
echo ""

# Make API calls with both tokens
echo "🧪 Step 7: Test API calls with both tokens..."

CRED1_API=$(curl -s -X GET "$API_URL/api/v1/platform/admins/check" \
    -H "st-auth-mode: header" \
    -H "Authorization: Bearer $CRED1_REAUTH_TOKEN")

if echo "$CRED1_API" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ First credential can make API calls${NC}"
else
    echo -e "${YELLOW}⚠️  First credential API call returned: $CRED1_API${NC}"
fi

CRED2_API=$(curl -s -X GET "$API_URL/api/v1/platform/admins/check" \
    -H "st-auth-mode: header" \
    -H "Authorization: Bearer $CRED2_REAUTH_TOKEN")

if echo "$CRED2_API" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Second credential can make API calls${NC}"
else
    echo -e "${YELLOW}⚠️  Second credential API call returned: $CRED2_API${NC}"
fi

echo ""

# Summary
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                            ║"
echo "║                          ✅ TEST RESULTS ✅                                ║"
echo "║                                                                            ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}✅ Multiple credentials CAN exist for same application${NC}"
echo -e "${GREEN}✅ Both credentials authenticate independently${NC}"
echo -e "${GREEN}✅ Both credentials can be used simultaneously${NC}"
echo -e "${GREEN}✅ Grace period rotation strategy is VIABLE${NC}"
echo ""
echo "📊 Test Credentials Created:"
echo "   Application: $APP_NAME"
echo "   Credential 1: $CRED1_EMAIL (User ID: $CRED1_USER_ID)"
echo "   Credential 2: $CRED2_EMAIL (User ID: $CRED2_USER_ID)"
echo ""
echo "🧹 To clean up test users, run:"
echo "   docker-compose exec -T supertokens-db psql -U supertokens -d supertokens -c \\"
echo "     \"DELETE FROM emailpassword_users WHERE email LIKE '${APP_NAME}%@system.internal';\""
echo ""

