#!/bin/bash
# Create SuperTokens database in Aurora RDS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Create SuperTokens Database${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check prerequisites
command -v aws >/dev/null 2>&1 || { echo -e "${RED}Error: aws CLI is not installed${NC}" >&2; exit 1; }
command -v pulumi >/dev/null 2>&1 || { echo -e "${RED}Error: pulumi is not installed${NC}" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}Error: jq is not installed${NC}" >&2; exit 1; }

# Get configuration from Pulumi
cd "$(dirname "$0")/.."

echo -e "${YELLOW}Getting database configuration from Pulumi...${NC}"
DB_ENDPOINT=$(pulumi stack output rdsClusterEndpoint)
DB_SECRET_ARN=$(aws secretsmanager list-secrets \
    --query "SecretList[?contains(Name, '$(pulumi config get utm-backend:projectName)-$(pulumi config get utm-backend:environment)-db-secret')].ARN | [0]" \
    --output text)

if [ -z "$DB_ENDPOINT" ] || [ -z "$DB_SECRET_ARN" ]; then
    echo -e "${RED}Error: Could not get database configuration${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Configuration retrieved${NC}"
echo "  Endpoint: $DB_ENDPOINT"
echo ""

# Get database credentials from Secrets Manager
echo -e "${YELLOW}Retrieving database credentials...${NC}"
DB_SECRET=$(aws secretsmanager get-secret-value --secret-id "$DB_SECRET_ARN" --query SecretString --output text)
DB_USERNAME=$(echo "$DB_SECRET" | jq -r '.username')
DB_PASSWORD=$(echo "$DB_SECRET" | jq -r '.password')
DB_NAME=$(echo "$DB_SECRET" | jq -r '.dbname')
SUPERTOKENS_DB_NAME=$(echo "$DB_SECRET" | jq -r '.supertokens_dbname')

echo -e "${GREEN}✓ Credentials retrieved${NC}"
echo ""

echo -e "${YELLOW}Creating SuperTokens database...${NC}"
echo "Note: This requires network access to the RDS cluster."
echo "You may need to:"
echo "  1. Run this from an EC2 instance in the VPC"
echo "  2. Use AWS Systems Manager Session Manager"
echo "  3. Set up a bastion host"
echo "  4. Temporarily allow your IP in the RDS security group"
echo ""

read -p "Do you have network access to the database? (y/N): " HAS_ACCESS
if [[ ! $HAS_ACCESS =~ ^[Yy]$ ]]; then
    echo ""
    echo "To create the database manually:"
    echo "  1. Connect to the database: psql -h $DB_ENDPOINT -U $DB_USERNAME -d $DB_NAME"
    echo "  2. Run: CREATE DATABASE $SUPERTOKENS_DB_NAME;"
    echo "  3. Run: GRANT ALL PRIVILEGES ON DATABASE $SUPERTOKENS_DB_NAME TO $DB_USERNAME;"
    echo ""
    exit 0
fi

# Try to create the database using docker with psql
if command -v docker >/dev/null 2>&1; then
    echo -e "${YELLOW}Using docker to run psql...${NC}"
    docker run --rm -it \
        -e PGPASSWORD="$DB_PASSWORD" \
        postgres:16-alpine \
        psql -h "$DB_ENDPOINT" -U "$DB_USERNAME" -d "$DB_NAME" \
        -c "CREATE DATABASE $SUPERTOKENS_DB_NAME;" \
        -c "GRANT ALL PRIVILEGES ON DATABASE $SUPERTOKENS_DB_NAME TO $DB_USERNAME;"
    
    echo -e "${GREEN}✓ SuperTokens database created${NC}"
elif command -v psql >/dev/null 2>&1; then
    echo -e "${YELLOW}Using local psql...${NC}"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_ENDPOINT" -U "$DB_USERNAME" -d "$DB_NAME" \
        -c "CREATE DATABASE $SUPERTOKENS_DB_NAME;" \
        -c "GRANT ALL PRIVILEGES ON DATABASE $SUPERTOKENS_DB_NAME TO $DB_USERNAME;"
    
    echo -e "${GREEN}✓ SuperTokens database created${NC}"
else
    echo -e "${RED}Error: Neither docker nor psql is available${NC}"
    echo "Please install one of them to create the database."
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Database Created!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "SuperTokens will automatically create its schema on first run."
echo ""

