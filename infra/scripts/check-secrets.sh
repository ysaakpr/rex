#!/bin/bash
# Check what's actually in the AWS secrets

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Check AWS Secrets Manager${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

export AWS_REGION=ap-south-1

# Check database secret
echo -e "${YELLOW}Database Secret (rex-backend-dev-db-secret-v2):${NC}"
aws secretsmanager get-secret-value \
    --secret-id rex-backend-dev-db-secret-v2 \
    --region ap-south-1 \
    --query 'SecretString' \
    --output text 2>/dev/null | jq '.' 2>/dev/null || echo "❌ Secret not found or invalid JSON"

echo ""
echo -e "${YELLOW}SuperTokens Secret (rex-backend-dev-supertokens-secret-v2):${NC}"
aws secretsmanager get-secret-value \
    --secret-id rex-backend-dev-supertokens-secret-v2 \
    --region ap-south-1 \
    --query 'SecretString' \
    --output text 2>/dev/null | jq '.' 2>/dev/null || echo "❌ Secret not found or invalid JSON"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Expected Structure${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Database secret should contain:${NC}"
cat <<EOF
{
  "host": "...",
  "port": "5432",
  "username": "rexadmin",
  "password": "...",
  "dbname": "rex_backend",
  "supertokens_dbname": "supertokens",
  "sslmode": "require",
  "redis_host": "...",
  "redis_port": "6379"
}
EOF

echo ""
echo -e "${GREEN}SuperTokens secret should contain:${NC}"
cat <<EOF
{
  "api_key": "your-actual-api-key"
}
EOF

echo ""

