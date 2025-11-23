#!/bin/bash
# Force new deployment of ECS services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Force New Deployment${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }

# Get configuration from Pulumi
cd "$(dirname "$0")/.."

echo -e "${YELLOW}Getting service names from Pulumi...${NC}"
CLUSTER_NAME=$(pulumi stack output ecsClusterName)
API_SERVICE=$(pulumi stack output apiServiceName)
WORKER_SERVICE=$(pulumi stack output workerServiceName)
SUPERTOKENS_SERVICE=$(pulumi stack output supertokensServiceName)

if [ -z "$CLUSTER_NAME" ]; then
    echo -e "${RED}Error: Could not get cluster name. Make sure infrastructure is deployed.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Service names retrieved${NC}"
echo ""
echo -e "${YELLOW}Note: Frontend is deployed via AWS Amplify (push to GitHub to deploy)${NC}"
echo ""

# Force new deployment for API
if [ -n "$API_SERVICE" ]; then
    echo -e "${YELLOW}Forcing new deployment for API service...${NC}"
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$API_SERVICE" \
        --force-new-deployment \
        --output text > /dev/null
    echo -e "${GREEN}✓ API service deployment triggered${NC}"
fi

# Force new deployment for Worker
if [ -n "$WORKER_SERVICE" ]; then
    echo -e "${YELLOW}Forcing new deployment for Worker service...${NC}"
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$WORKER_SERVICE" \
        --force-new-deployment \
        --output text > /dev/null
    echo -e "${GREEN}✓ Worker service deployment triggered${NC}"
fi

# Optional: SuperTokens (usually doesn't need redeployment)
read -p "Do you want to redeploy SuperTokens service? (y/N): " REDEPLOY_ST
if [[ $REDEPLOY_ST =~ ^[Yy]$ ]] && [ -n "$SUPERTOKENS_SERVICE" ]; then
    echo -e "${YELLOW}Forcing new deployment for SuperTokens service...${NC}"
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$SUPERTOKENS_SERVICE" \
        --force-new-deployment \
        --output text > /dev/null
    echo -e "${GREEN}✓ SuperTokens service deployment triggered${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Triggered!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Monitor deployment progress:"
echo "  aws ecs describe-services --cluster $CLUSTER_NAME --services $API_SERVICE"
echo ""
echo "View logs:"
echo "  aws logs tail /ecs/\$(pulumi config get rex-backend:projectName)-\$(pulumi config get rex-backend:environment)-api --follow"
echo ""

