#!/bin/bash
# Switch from existing deployment to all-in-one mode

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Switch to All-in-One Mode${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$(dirname "$0")/.."

# Check current mode
CURRENT_ALLINONE=$(pulumi config get rex-backend:allinone 2>/dev/null || echo "false")
CURRENT_LOWCOST=$(pulumi config get rex-backend:lowcost 2>/dev/null || echo "false")

echo -e "${YELLOW}Current configuration:${NC}"
echo "  allinone: $CURRENT_ALLINONE"
echo "  lowcost: $CURRENT_LOWCOST"
echo ""

if [ "$CURRENT_ALLINONE" = "true" ]; then
    echo -e "${GREEN}Already in all-in-one mode!${NC}"
    exit 0
fi

echo -e "${YELLOW}This will:${NC}"
echo "  1. Destroy existing infrastructure (ECS, RDS, etc.)"
echo "  2. Enable all-in-one mode"
echo "  3. Deploy new all-in-one EC2 architecture"
echo ""
echo -e "${RED}WARNING: This will delete your databases and Redis!${NC}"
echo -e "${RED}Make sure you have backups if needed.${NC}"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

# Step 1: Destroy existing infrastructure
echo ""
echo -e "${YELLOW}Step 1: Destroying existing infrastructure...${NC}"
pulumi destroy --yes

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Destroy failed. Check errors above.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Old infrastructure destroyed${NC}"
echo ""

# Step 2: Enable all-in-one mode
echo -e "${YELLOW}Step 2: Enabling all-in-one mode...${NC}"
pulumi config set rex-backend:allinone true
pulumi config set rex-backend:lowcost false

echo -e "${GREEN}✓ All-in-one mode enabled${NC}"
echo ""

# Step 3: Deploy new infrastructure
echo -e "${YELLOW}Step 3: Deploying all-in-one infrastructure...${NC}"
pulumi up --yes

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Deployment failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All-in-one infrastructure deployed${NC}"
echo ""

# Step 4: Build and push images
echo -e "${YELLOW}Step 4: Building and pushing Docker images...${NC}"
cd ..
./infra/scripts/build-and-push.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Build/push failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Images pushed${NC}"
echo ""

# Get outputs
cd infra
INSTANCE_ID=$(pulumi stack output allInOneInstanceId 2>/dev/null)
ALB_DNS=$(pulumi stack output albDnsName 2>/dev/null)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Migration Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Instance ID:${NC} $INSTANCE_ID"
echo -e "${GREEN}ALB DNS:${NC} $ALB_DNS"
echo ""
echo -e "${YELLOW}Wait ~5 minutes for services to start, then test:${NC}"
echo "curl http://$ALB_DNS/api/health"
echo ""

