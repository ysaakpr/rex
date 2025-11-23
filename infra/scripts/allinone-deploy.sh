#!/bin/bash
# Deploy all-in-one mode from scratch

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}All-in-One Mode Deployment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$(dirname "$0")/.."

# Check if allinone mode is enabled
ALLINONE=$(pulumi config get rex-backend:allinone 2>/dev/null || echo "false")

if [ "$ALLINONE" != "true" ]; then
    echo -e "${YELLOW}All-in-one mode not enabled. Enabling now...${NC}"
    pulumi config set rex-backend:allinone true
    echo -e "${GREEN}✓ All-in-one mode enabled${NC}"
    echo ""
fi

# Deploy infrastructure
echo -e "${YELLOW}Step 1: Deploying infrastructure with Pulumi...${NC}"
pulumi up --yes

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Pulumi deployment failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Infrastructure deployed${NC}"
echo ""

# Build and push Docker images
echo -e "${YELLOW}Step 2: Building and pushing Docker images to ECR...${NC}"
cd ..
./infra/scripts/build-and-push.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Docker build/push failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker images pushed to ECR${NC}"
echo ""

# Get instance ID
cd infra
INSTANCE_ID=$(pulumi stack output allInOneInstanceId 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
    echo -e "${RED}✗ Could not get instance ID${NC}"
    exit 1
fi

echo -e "${YELLOW}Step 3: Waiting for EC2 instance to initialize...${NC}"
echo "Instance ID: $INSTANCE_ID"

# Wait for instance to be ready (Docker setup takes ~5 minutes)
echo "This will take about 5 minutes while Docker Compose starts all services..."
sleep 300

echo -e "${GREEN}✓ Instance should be ready${NC}"
echo ""

# Get endpoints
ALB_DNS=$(pulumi stack output albDnsName)
API_URL=$(pulumi stack output apiUrl)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Deployment Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}API URL:${NC}"
echo "  $API_URL"
echo ""
echo -e "${GREEN}ALB DNS:${NC}"
echo "  $ALB_DNS"
echo ""
echo -e "${GREEN}Instance ID:${NC}"
echo "  $INSTANCE_ID"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Test API: curl http://$ALB_DNS/api/health"
echo "2. View logs: aws ssm start-session --target $INSTANCE_ID"
echo "   Then: cd /app && docker-compose logs -f"
echo "3. Update code: ./infra/scripts/allinone-update.sh"
echo ""

