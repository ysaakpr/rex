#!/bin/bash
# Fresh all-in-one deployment with local state

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

PASSPHRASE="rex-backend-local-2024"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Fresh All-in-One Deployment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$(dirname "$0")/.."

echo -e "${YELLOW}Step 1: Setting up local Pulumi backend...${NC}"
pulumi login --local
echo -e "${GREEN}✓ Using local backend${NC}"
echo ""

echo -e "${YELLOW}Step 2: Removing old local state...${NC}"
rm -rf ~/.pulumi/stacks/rex-backend/dev.json* 2>/dev/null || true
echo -e "${GREEN}✓ Old state cleared${NC}"
echo ""

echo -e "${YELLOW}Step 3: Creating fresh stack...${NC}"
export PULUMI_CONFIG_PASSPHRASE="$PASSPHRASE"
pulumi stack init dev --secrets-provider=passphrase 2>/dev/null || pulumi stack select dev
echo -e "${GREEN}✓ Stack ready${NC}"
echo ""

echo -e "${YELLOW}Step 4: Configuring stack...${NC}"
pulumi config set aws:region ap-south-1
pulumi config set rex-backend:environment dev
pulumi config set rex-backend:projectName rex-backend
pulumi config set rex-backend:vpcCidr 10.0.0.0/16
pulumi config set rex-backend:dbMasterUsername rexadmin
pulumi config set rex-backend:githubRepo https://github.com/ysaakpr/rex
pulumi config set rex-backend:githubBranch main
pulumi config set rex-backend:allinone true
pulumi config set rex-backend:lowcost false
echo -e "${GREEN}✓ Configuration set${NC}"
echo ""

echo -e "${YELLOW}Step 5: Setting secrets...${NC}"

# Check if secrets already exist
EXISTING_DB_PASS=$(pulumi config get rex-backend:dbMasterPassword 2>/dev/null || echo "")
EXISTING_ST_KEY=$(pulumi config get rex-backend:supertokensApiKey 2>/dev/null || echo "")

if [ -z "$EXISTING_DB_PASS" ]; then
    echo "Enter database master password (or press Enter for auto-generated):"
    read -s DB_PASS
    if [ -z "$DB_PASS" ]; then
        DB_PASS="RexBackend$(openssl rand -base64 16 | tr -d '/+=' | head -c 16)!"
        echo "  Auto-generated password"
    fi
    pulumi config set --secret rex-backend:dbMasterPassword "$DB_PASS"
else
    echo "  Database password already set"
fi

if [ -z "$EXISTING_ST_KEY" ]; then
    ST_KEY=$(openssl rand -base64 32)
    pulumi config set --secret rex-backend:supertokensApiKey "$ST_KEY"
    echo "  Generated SuperTokens API key"
else
    echo "  SuperTokens API key already set"
fi

echo -e "${GREEN}✓ Secrets configured${NC}"
echo ""

echo -e "${YELLOW}Step 6: Deploying infrastructure...${NC}"
echo ""
PULUMI_CONFIG_PASSPHRASE="$PASSPHRASE" pulumi up --yes

if [ $? -ne 0 ]; then
    echo ""
    echo -e "${RED}✗ Deployment failed!${NC}"
    echo ""
    echo -e "${YELLOW}Common issues:${NC}"
    echo "1. AWS resources with same names already exist"
    echo "   Solution: Delete them from AWS Console first"
    echo ""
    echo "2. AWS credentials not configured"
    echo "   Solution: Run 'aws configure'"
    echo ""
    echo "3. Region mismatch"
    echo "   Solution: Check AWS_REGION=ap-south-1"
    echo ""
    exit 1
fi

echo ""
echo -e "${GREEN}✓ Infrastructure deployed${NC}"
echo ""

echo -e "${YELLOW}Step 7: Building and pushing Docker images...${NC}"
cd ..
./infra/scripts/build-and-push.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Docker build/push failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Images pushed${NC}"
echo ""

# Get outputs
cd infra
INSTANCE_ID=$(PULUMI_CONFIG_PASSPHRASE="$PASSPHRASE" pulumi stack output allInOneInstanceId 2>/dev/null || echo "pending")
ALB_DNS=$(PULUMI_CONFIG_PASSPHRASE="$PASSPHRASE" pulumi stack output albDnsName 2>/dev/null || echo "pending")
API_URL=$(PULUMI_CONFIG_PASSPHRASE="$PASSPHRASE" pulumi stack output apiUrl 2>/dev/null || echo "pending")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Deployment Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Instance ID:${NC} $INSTANCE_ID"
echo -e "${GREEN}ALB DNS:${NC} $ALB_DNS"
echo -e "${GREEN}API URL:${NC} $API_URL"
echo ""
echo -e "${YELLOW}IMPORTANT: Save this passphrase for future Pulumi operations:${NC}"
echo -e "${GREEN}$PASSPHRASE${NC}"
echo ""
echo -e "${YELLOW}To use Pulumi commands later:${NC}"
echo "export PULUMI_CONFIG_PASSPHRASE=\"$PASSPHRASE\""
echo ""
echo -e "${YELLOW}Wait ~5 minutes for Docker Compose to start services, then test:${NC}"
echo "curl http://$ALB_DNS/api/health"
echo ""
echo -e "${YELLOW}View logs:${NC}"
echo "aws ssm start-session --target $INSTANCE_ID"
echo "cd /app && docker-compose logs -f"
echo ""

