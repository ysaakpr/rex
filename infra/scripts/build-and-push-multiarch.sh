#!/bin/bash
# Build and push Docker images to ECR with multi-architecture support
# Builds for both ARM64 (M-series Mac) and AMD64 (x86_64)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Multi-Arch Build and Push Docker Images${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Error: docker is not installed${NC}" >&2; exit 1; }
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
INFRA_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${YELLOW}Getting ECR repository URLs from Pulumi...${NC}"
cd "$INFRA_DIR"
API_REPO=$(pulumi stack output apiRepositoryUrl 2>/dev/null)
WORKER_REPO=$(pulumi stack output workerRepositoryUrl 2>/dev/null)

if [ -z "$API_REPO" ] || [ -z "$WORKER_REPO" ]; then
    echo -e "${RED}Error: Could not get repository URLs. Make sure infrastructure is deployed.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Repository URLs retrieved${NC}"
echo "  API: $API_REPO"
echo "  Worker: $WORKER_REPO"
echo ""

# Get AWS account and region
AWS_REGION=$(pulumi config get aws:region 2>/dev/null || echo "ap-south-1")
AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)

# Login to ECR
echo -e "${YELLOW}Logging in to ECR...${NC}"
aws ecr get-login-password --region "$AWS_REGION" | \
    docker login --username AWS --password-stdin "${AWS_ACCOUNT}.dkr.ecr.${AWS_REGION}.amazonaws.com"
echo -e "${GREEN}✓ Logged in to ECR${NC}"
echo ""

# Setup buildx for multi-arch
echo -e "${YELLOW}Setting up Docker buildx for multi-architecture builds...${NC}"
docker buildx create --name multiarch-builder --use --bootstrap 2>/dev/null || docker buildx use multiarch-builder
echo -e "${GREEN}✓ Buildx configured${NC}"
echo ""

# Navigate to project root
cd "$PROJECT_ROOT"
echo "Building from: $(pwd)"
echo ""

# Check if Dockerfile.prod exists
if [ ! -f "Dockerfile.prod" ]; then
    echo -e "${RED}Error: Dockerfile.prod not found in $(pwd)${NC}"
    exit 1
fi

# Build and push API image for multiple architectures
echo -e "${YELLOW}Building and pushing API image (multi-arch: amd64, arm64)...${NC}"
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -f Dockerfile.prod \
    --target api \
    -t "${API_REPO}:latest" \
    --push \
    .
echo -e "${GREEN}✓ API image built and pushed (multi-arch)${NC}"

# Build and push Worker image for multiple architectures
echo -e "${YELLOW}Building and pushing Worker image (multi-arch: amd64, arm64)...${NC}"
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -f Dockerfile.prod \
    --target worker \
    -t "${WORKER_REPO}:latest" \
    --push \
    .
echo -e "${GREEN}✓ Worker image built and pushed (multi-arch)${NC}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Multi-arch images built and pushed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Images now support both:"
echo "  • linux/amd64 (x86_64) - for t3a, t3, c5 instances"
echo "  • linux/arm64 (ARM64)  - for t4g, c6g Graviton instances"
echo ""
echo "Next steps:"
echo "1. Deploy: cd infra && pulumi up"
echo ""

