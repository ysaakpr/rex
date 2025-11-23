#!/bin/bash
# Run database migrations as an ECS task

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Run Database Migrations${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}Error: jq is not installed${NC}" >&2; exit 1; }

# Get configuration from Pulumi
cd "$(dirname "$0")/.."

echo -e "${YELLOW}Getting configuration from Pulumi...${NC}"
CLUSTER_NAME=$(pulumi stack output ecsClusterName)
TASK_DEF=$(pulumi stack output migrationTaskDefinitionArn)
SUBNETS=$(pulumi stack output privateSubnetIds -j | jq -r 'join(",")')
# Note: We need to export the security group ID in main.go
# For now, we'll get it from ECS service
SECURITY_GROUP=$(aws ecs describe-services \
    --cluster "$CLUSTER_NAME" \
    --services "$(pulumi stack output apiServiceName)" \
    --query 'services[0].networkConfiguration.awsvpcConfiguration.securityGroups[0]' \
    --output text)

if [ -z "$CLUSTER_NAME" ] || [ -z "$TASK_DEF" ] || [ -z "$SUBNETS" ]; then
    echo -e "${RED}Error: Could not get required configuration. Make sure infrastructure is deployed.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Configuration retrieved${NC}"
echo "  Cluster: $CLUSTER_NAME"
echo "  Task Definition: $TASK_DEF"
echo "  Security Group: $SECURITY_GROUP"
echo ""

# Run the migration task
echo -e "${YELLOW}Starting migration task...${NC}"
TASK_ARN=$(aws ecs run-task \
    --cluster "$CLUSTER_NAME" \
    --task-definition "$TASK_DEF" \
    --launch-type FARGATE \
    --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$SECURITY_GROUP],assignPublicIp=DISABLED}" \
    --query 'tasks[0].taskArn' \
    --output text)

if [ -z "$TASK_ARN" ] || [ "$TASK_ARN" == "None" ]; then
    echo -e "${RED}Error: Failed to start migration task${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Migration task started${NC}"
echo "  Task ARN: $TASK_ARN"
echo ""

# Wait for task to complete
echo -e "${YELLOW}Waiting for task to complete...${NC}"
aws ecs wait tasks-stopped --cluster "$CLUSTER_NAME" --tasks "$TASK_ARN"

# Check task status
EXIT_CODE=$(aws ecs describe-tasks \
    --cluster "$CLUSTER_NAME" \
    --tasks "$TASK_ARN" \
    --query 'tasks[0].containers[0].exitCode' \
    --output text)

if [ "$EXIT_CODE" == "0" ]; then
    echo -e "${GREEN}✓ Migration completed successfully${NC}"
else
    echo -e "${RED}Error: Migration failed with exit code $EXIT_CODE${NC}"
    echo ""
    echo "View logs with:"
    echo "  aws logs tail /ecs/$(pulumi config get utm-backend:projectName)-$(pulumi config get utm-backend:environment)-migration --follow"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Migration Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

