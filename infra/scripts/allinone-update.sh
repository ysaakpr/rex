#!/bin/bash
# Update running services in all-in-one mode

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Update All-in-One Services${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$(dirname "$0")/.."

# Build and push new images
echo -e "${YELLOW}Step 1: Building and pushing Docker images...${NC}"
cd ..
./infra/scripts/build-and-push.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Docker build/push failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Images pushed to ECR${NC}"
echo ""

# Get instance ID
cd infra
INSTANCE_ID=$(pulumi stack output allInOneInstanceId 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
    echo -e "${RED}✗ Could not get instance ID${NC}"
    echo "Make sure all-in-one mode is deployed"
    exit 1
fi

echo -e "${YELLOW}Step 2: Connecting to instance and updating services...${NC}"
echo "Instance ID: $INSTANCE_ID"
echo ""

# Create update command
UPDATE_CMD='cd /app && ./update.sh && docker-compose ps'

echo -e "${YELLOW}Running update on instance...${NC}"
echo ""

# Execute via SSM
aws ssm send-command \
    --instance-ids "$INSTANCE_ID" \
    --document-name "AWS-RunShellScript" \
    --parameters "commands=[$UPDATE_CMD]" \
    --region ap-south-1 \
    --output text \
    --query "Command.CommandId" > /tmp/command-id.txt

COMMAND_ID=$(cat /tmp/command-id.txt)
echo "Command ID: $COMMAND_ID"

# Wait for command to complete
echo -e "${YELLOW}Waiting for update to complete...${NC}"
sleep 5

# Check command status
STATUS=$(aws ssm get-command-invocation \
    --command-id "$COMMAND_ID" \
    --instance-id "$INSTANCE_ID" \
    --region ap-south-1 \
    --query "Status" \
    --output text 2>/dev/null || echo "Unknown")

echo "Status: $STATUS"

if [ "$STATUS" = "Success" ]; then
    echo -e "${GREEN}✓ Services updated successfully!${NC}"
else
    echo -e "${YELLOW}⚠ Status: $STATUS${NC}"
    echo "You can check the output with:"
    echo "aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INSTANCE_ID --region ap-south-1"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Update Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}To view logs:${NC}"
echo "aws ssm start-session --target $INSTANCE_ID"
echo "cd /app && docker-compose logs -f api worker"
echo ""
echo -e "${YELLOW}To check service status:${NC}"
echo "aws ssm start-session --target $INSTANCE_ID"
echo "cd /app && docker-compose ps"
echo ""

