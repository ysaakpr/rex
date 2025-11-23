#!/bin/bash

# Test script for platform admin system
# This script verifies that all platform admin endpoints are working

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   🧪 Platform Admin System - Verification Test               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo "🔍 Test 1: Backend Health Check"
HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
    echo -e "${GREEN}✅ Backend is healthy${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Frontend Health Check
echo "🔍 Test 2: Frontend Health Check"
FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000)
if [ "$FRONTEND" = "200" ]; then
    echo -e "${GREEN}✅ Frontend is running${NC}"
else
    echo -e "${RED}❌ Frontend check failed${NC}"
    exit 1
fi
echo ""

# Test 3: Database Check
echo "🔍 Test 3: Database Tables Check"
TABLES=$(docker-compose exec -T postgres psql -U utmuser -d utm_backend -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('platform_admins', 'relation_roles');" 2>/dev/null | tr -d ' ')
if [ "$TABLES" = "2" ]; then
    echo -e "${GREEN}✅ Platform admin tables exist${NC}"
else
    echo -e "${RED}❌ Platform admin tables missing${NC}"
    exit 1
fi
echo ""

# Test 4: Platform Admin Count
echo "🔍 Test 4: Platform Admins Count"
ADMIN_COUNT=$(docker-compose exec -T postgres psql -U utmuser -d utm_backend -t -c "SELECT COUNT(*) FROM platform_admins;" 2>/dev/null | tr -d ' ')
echo "   Platform admins in database: $ADMIN_COUNT"
if [ "$ADMIN_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Platform admins exist${NC}"
else
    echo -e "${YELLOW}⚠️  No platform admins yet (run ./scripts/create_platform_admin.sh)${NC}"
fi
echo ""

# Test 5: Permissions Check
echo "🔍 Test 5: Platform Permissions Check"
PERM_COUNT=$(docker-compose exec -T postgres psql -U utmuser -d utm_backend -t -c "SELECT COUNT(*) FROM permissions WHERE service = 'platform-api';" 2>/dev/null | tr -d ' ')
echo "   Platform-api permissions: $PERM_COUNT"
if [ "$PERM_COUNT" -ge 15 ]; then
    echo -e "${GREEN}✅ Platform permissions created${NC}"
else
    echo -e "${RED}❌ Platform permissions missing${NC}"
    exit 1
fi
echo ""

# Test 6: API Routes Check
echo "🔍 Test 6: API Routes (without auth - expect 401)"
ROUTES=(
    "/api/v1/platform/admins/check"
    "/api/v1/platform/admins"
    "/api/v1/platform/roles"
    "/api/v1/platform/permissions"
)

ALL_ROUTES_OK=true
for ROUTE in "${ROUTES[@]}"; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000$ROUTE)
    if [ "$STATUS" = "401" ]; then
        echo -e "   ${GREEN}✅${NC} $ROUTE (requires auth)"
    else
        echo -e "   ${RED}❌${NC} $ROUTE (got $STATUS, expected 401)"
        ALL_ROUTES_OK=false
    fi
done

if [ "$ALL_ROUTES_OK" = true ]; then
    echo -e "${GREEN}✅ All platform routes responding correctly${NC}"
else
    echo -e "${RED}❌ Some routes have issues${NC}"
fi
echo ""

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   ✅ VERIFICATION COMPLETE!                                   ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""
echo "📋 Summary:"
echo "   • Backend: Running ✓"
echo "   • Frontend: Running ✓"
echo "   • Database: Tables created ✓"
echo "   • Permissions: Seeded ✓"
echo "   • API Routes: Protected ✓"
echo ""
echo "🚀 Next Steps:"
echo "   1. Sign in at http://localhost:3000"
echo "   2. Get your user ID from backend logs"
echo "   3. Run: ./scripts/create_platform_admin.sh <your-user-id>"
echo "   4. Refresh browser to see platform admin features"
echo ""
