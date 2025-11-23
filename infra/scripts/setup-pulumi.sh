#!/bin/bash
# Setup script for Pulumi infrastructure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}UTM Backend - Pulumi Setup${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo -e "${RED}Error: go is not installed${NC}" >&2; exit 1; }

echo -e "${GREEN}✓ All prerequisites installed${NC}"
echo ""

# Get configuration from user
read -p "Enter environment name (dev/staging/production) [dev]: " ENVIRONMENT
ENVIRONMENT=${ENVIRONMENT:-dev}

read -p "Enter AWS region [us-east-1]: " AWS_REGION
AWS_REGION=${AWS_REGION:-us-east-1}

read -p "Enter S3 bucket name for Pulumi state [utm-backend-pulumi-state]: " S3_BUCKET
S3_BUCKET=${S3_BUCKET:-utm-backend-pulumi-state}

read -p "Enter project name [utm-backend]: " PROJECT_NAME
PROJECT_NAME=${PROJECT_NAME:-utm-backend}

echo ""
echo -e "${YELLOW}Creating S3 bucket for Pulumi state...${NC}"

# Create S3 bucket if it doesn't exist
if aws s3 ls "s3://$S3_BUCKET" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3 mb "s3://$S3_BUCKET" --region "$AWS_REGION"
    echo -e "${GREEN}✓ S3 bucket created${NC}"
    
    # Enable versioning
    aws s3api put-bucket-versioning \
        --bucket "$S3_BUCKET" \
        --versioning-configuration Status=Enabled
    echo -e "${GREEN}✓ Versioning enabled${NC}"
    
    # Enable encryption
    aws s3api put-bucket-encryption \
        --bucket "$S3_BUCKET" \
        --server-side-encryption-configuration '{
            "Rules": [{
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                }
            }]
        }'
    echo -e "${GREEN}✓ Encryption enabled${NC}"
else
    echo -e "${GREEN}✓ S3 bucket already exists${NC}"
fi

echo ""
echo -e "${YELLOW}Configuring Pulumi...${NC}"

# Navigate to infra directory
cd "$(dirname "$0")/.."

# Login to S3 backend
pulumi login "s3://$S3_BUCKET"
echo -e "${GREEN}✓ Logged in to Pulumi backend${NC}"

# Initialize stack if it doesn't exist
if pulumi stack select "$ENVIRONMENT" 2>/dev/null; then
    echo -e "${GREEN}✓ Stack '$ENVIRONMENT' already exists${NC}"
else
    pulumi stack init "$ENVIRONMENT"
    echo -e "${GREEN}✓ Stack '$ENVIRONMENT' created${NC}"
fi

# Install Go dependencies
echo ""
echo -e "${YELLOW}Installing Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"

# Set configuration
echo ""
echo -e "${YELLOW}Setting configuration...${NC}"
pulumi config set aws:region "$AWS_REGION"
pulumi config set utm-backend:environment "$ENVIRONMENT"
pulumi config set utm-backend:projectName "$PROJECT_NAME"
pulumi config set utm-backend:vpcCidr "10.0.0.0/16"
pulumi config set utm-backend:dbMasterUsername "utmadmin"
echo -e "${GREEN}✓ Basic configuration set${NC}"

# Prompt for secrets
echo ""
echo -e "${YELLOW}Setting secrets...${NC}"
echo "Please enter the database master password (will be hidden):"
read -s DB_PASSWORD
pulumi config set --secret utm-backend:dbMasterPassword "$DB_PASSWORD"
echo -e "${GREEN}✓ Database password set${NC}"

echo "Please enter the SuperTokens API key (will be hidden):"
read -s ST_API_KEY
pulumi config set --secret utm-backend:supertokensApiKey "$ST_API_KEY"
echo -e "${GREEN}✓ SuperTokens API key set${NC}"

# Optional: Domain and certificate
echo ""
read -p "Do you want to configure a custom domain? (y/N): " CONFIGURE_DOMAIN
if [[ $CONFIGURE_DOMAIN =~ ^[Yy]$ ]]; then
    read -p "Enter domain name: " DOMAIN_NAME
    pulumi config set utm-backend:domainName "$DOMAIN_NAME"
    
    read -p "Enter ACM certificate ARN: " CERT_ARN
    pulumi config set utm-backend:certificateArn "$CERT_ARN"
    echo -e "${GREEN}✓ Domain configuration set${NC}"
fi

# Display configuration
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Configuration Summary${NC}"
echo -e "${GREEN}========================================${NC}"
pulumi config
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Review the configuration above"
echo "2. Preview infrastructure: pulumi preview"
echo "3. Deploy infrastructure: pulumi up"
echo "4. Build and push Docker images (see infra/scripts/build-and-push.sh)"
echo "5. Run database migrations (see infra/scripts/run-migration.sh)"
echo ""

