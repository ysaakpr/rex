#!/bin/bash
# Manually fix the SuperTokens secret with correct api_key value

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Fix SuperTokens Secret${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

export AWS_REGION=ap-south-1

# Get the API key from Pulumi config
cd "$(dirname "$0")/.."
echo -e "${YELLOW}Getting SuperTokens API key from Pulumi config...${NC}"

# Try to get from Pulumi
API_KEY=$(pulumi config get rex-backend:supertokensApiKey 2>/dev/null || echo "")

if [ -z "$API_KEY" ]; then
    echo -e "${RED}Could not get API key from Pulumi config${NC}"
    echo ""
    echo "Please enter the SuperTokens API key manually:"
    read -s API_KEY
    echo ""
fi

if [ -z "$API_KEY" ]; then
    echo -e "${RED}No API key provided. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ API key retrieved${NC}"
echo ""

# Create the JSON structure
SECRET_JSON=$(cat <<EOF
{
  "api_key": "$API_KEY"
}
EOF
)

echo -e "${YELLOW}Updating secret in AWS Secrets Manager...${NC}"

# Update the secret
aws secretsmanager put-secret-value \
    --region ap-south-1 \
    --secret-id rex-backend-dev-supertokens-secret-v2 \
    --secret-string "$SECRET_JSON" \
    2>&1 > /dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Secret updated successfully!${NC}"
else
    echo -e "${RED}✗ Failed to update secret${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Verify${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}Secret value:${NC}"
aws secretsmanager get-secret-value \
    --secret-id rex-backend-dev-supertokens-secret-v2 \
    --region ap-south-1 \
    --query 'SecretString' \
    --output text | jq '.'

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Next Steps${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "1. Force ECS services to restart:"
echo "   ./infra/scripts/force-deploy.sh"
echo ""
echo "2. Check task logs:"
echo "   aws logs tail /ecs/rex-backend-dev-supertokens --follow --region ap-south-1"
echo ""

