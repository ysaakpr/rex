#!/bin/bash

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║           🔍 SUPERTOKENS AUTHENTICATION DIAGNOSTICS           ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Check configuration
echo "1️⃣  Checking Backend Configuration..."
echo "   API_DOMAIN: $(grep API_DOMAIN .env | cut -d'=' -f2)"
echo "   WEBSITE_DOMAIN: $(grep WEBSITE_DOMAIN .env | cut -d'=' -f2)"
echo ""

# Check services
echo "2️⃣  Checking Services Status..."
docker-compose ps api frontend supertokens | grep -E "NAME|utm"
echo ""

# Test auth flow
echo "3️⃣  Testing Authentication Flow..."
COOKIES=$(mktemp)
EMAIL="diagnose-$(date +%s)@example.com"

echo "   Creating test user: $EMAIL"
SIGNUP=$(curl -s -c "$COOKIES" -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -H "rid: emailpassword" \
  -d "{\"formFields\":[{\"id\":\"email\",\"value\":\"${EMAIL}\"},{\"id\":\"password\",\"value\":\"Test123!@#\"}]}")

if echo "$SIGNUP" | jq -e '.status == "OK"' > /dev/null 2>&1; then
  echo "   ✅ Signup successful"
  USER_ID=$(echo "$SIGNUP" | jq -r '.user.id')
  echo "   User ID: $USER_ID"
else
  echo "   ❌ Signup failed"
  echo "$SIGNUP" | jq '.'
fi

echo ""
echo "   Cookies after signup:"
cat "$COOKIES" | grep -v "^#" | awk '{print "     " $6 "=" substr($7,1,40) "..."}'

echo ""
echo "4️⃣  Testing Protected Endpoint..."
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -b "$COOKIES" http://localhost:3000/api/v1/tenants)
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
  echo "   ✅ SUCCESS! Authentication working!"
  echo "$BODY" | jq '.'
else
  echo "   ❌ FAILED! Still getting 401"
  echo "   Response: $BODY"
fi

echo ""
echo "5️⃣  Checking Backend Logs..."
echo "   Recent auth-related logs:"
docker-compose logs --tail=20 api | grep -E "(DEBUG|401|Session)" | tail -10

rm "$COOKIES"
echo ""
echo "═══════════════════════════════════════════════════════════════"

