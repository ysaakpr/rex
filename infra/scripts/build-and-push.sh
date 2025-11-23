#!/bin/bash
# Build and push Docker images to ECR

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build and Push Docker Images${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Error: docker is not installed${NC}" >&2; exit 1; }
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}Error: jq is not installed${NC}" >&2; exit 1; }

# Get ECR repository URLs from Pulumi outputs
cd "$(dirname "$0")/.."

echo -e "${YELLOW}Getting ECR repository URLs from Pulumi...${NC}"
API_REPO=$(pulumi stack output apiRepositoryUrl)
WORKER_REPO=$(pulumi stack output workerRepositoryUrl)

if [ -z "$API_REPO" ] || [ -z "$WORKER_REPO" ]; then
    echo -e "${RED}Error: Could not get repository URLs. Make sure infrastructure is deployed.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Repository URLs retrieved${NC}"
echo "  API: $API_REPO"
echo "  Worker: $WORKER_REPO"
echo ""
echo -e "${YELLOW}Note: Frontend is deployed via AWS Amplify (not ECR)${NC}"
echo ""

# Get AWS account and region
AWS_REGION=$(pulumi config get aws:region)
AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)

# Login to ECR
echo -e "${YELLOW}Logging in to ECR...${NC}"
aws ecr get-login-password --region "$AWS_REGION" | \
    docker login --username AWS --password-stdin "${AWS_ACCOUNT}.dkr.ecr.${AWS_REGION}.amazonaws.com"
echo -e "${GREEN}✓ Logged in to ECR${NC}"
echo ""

# Navigate to project root
cd ../..

# Build API image
echo -e "${YELLOW}Building API image...${NC}"
docker build -f Dockerfile.prod --target api -t rex-backend-api:latest .
echo -e "${GREEN}✓ API image built${NC}"

# Build Worker image
echo -e "${YELLOW}Building Worker image...${NC}"
docker build -f Dockerfile.prod --target worker -t rex-backend-worker:latest .
echo -e "${GREEN}✓ Worker image built${NC}"
echo ""

# Tag and push API
echo -e "${YELLOW}Pushing API image to ECR...${NC}"
docker tag rex-backend-api:latest "${API_REPO}:latest"
docker push "${API_REPO}:latest"
echo -e "${GREEN}✓ API image pushed${NC}"

# Tag and push Worker
echo -e "${YELLOW}Pushing Worker image to ECR...${NC}"
docker tag rex-backend-worker:latest "${WORKER_REPO}:latest"
docker push "${WORKER_REPO}:latest"
echo -e "${GREEN}✓ Worker image pushed${NC}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Backend images built and pushed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Force new deployment: ./infra/scripts/force-deploy.sh"
echo "2. Run database migrations: ./infra/scripts/run-migration.sh"
echo ""
echo -e "${YELLOW}Frontend Deployment:${NC}"
echo "Frontend is automatically deployed via AWS Amplify when you push to GitHub."
echo "No manual build/push required for frontend!"
echo ""

