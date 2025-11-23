#!/bin/bash
# Manually clean up ECS/Fargate resources before switching to all-in-one

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Clean Up ECS/Fargate Resources${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$(dirname "$0")/.."

echo -e "${YELLOW}This will delete ECS resources while keeping the rest.${NC}"
echo -e "${YELLOW}Useful if you want to preserve databases/secrets.${NC}"
echo ""

# List of ECS/Fargate resources to remove
ECS_RESOURCES=(
    "aws:ecs/service:Service::rex-backend-dev-api-service"
    "aws:ecs/service:Service::rex-backend-dev-worker-service"
    "aws:ecs/service:Service::rex-backend-dev-supertokens-service"
    "aws:ecs/taskDefinition:TaskDefinition::rex-backend-dev-api-task"
    "aws:ecs/taskDefinition:TaskDefinition::rex-backend-dev-worker-task"
    "aws:ecs/taskDefinition:TaskDefinition::rex-backend-dev-supertokens-task"
    "aws:ecs/taskDefinition:TaskDefinition::rex-backend-dev-migration-task"
)

echo -e "${YELLOW}Resources to delete:${NC}"
for resource in "${ECS_RESOURCES[@]}"; do
    echo "  - $resource"
done
echo ""

read -p "Proceed with deletion? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo -e "${YELLOW}Deleting ECS resources...${NC}"

# Delete each resource
for resource in "${ECS_RESOURCES[@]}"; do
    echo -e "${YELLOW}Deleting: $resource${NC}"
    pulumi state delete "$resource" --yes 2>/dev/null || echo "  (not found, skipping)"
done

echo ""
echo -e "${GREEN}✓ ECS resources removed from state${NC}"
echo ""
echo -e "${YELLOW}Note: Resources may still exist in AWS.${NC}"
echo -e "${YELLOW}You can manually delete them from AWS Console if needed.${NC}"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Enable all-in-one mode: pulumi config set rex-backend:allinone true"
echo "2. Deploy: pulumi up"
echo ""

