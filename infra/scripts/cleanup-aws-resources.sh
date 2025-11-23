#!/bin/bash
# Manually clean up conflicting AWS resources

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

REGION="ap-south-1"
PROJECT="rex-backend"
ENV="dev"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Clean Up Conflicting AWS Resources${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}This will delete the following AWS resources:${NC}"
echo "  - Load Balancer: ${PROJECT}-${ENV}-alb"
echo "  - Target Groups: ${PROJECT}-${ENV}-api-tg, ${PROJECT}-${ENV}-st-tg"
echo "  - ECS Services: api, worker, supertokens"
echo "  - ECS Task Definitions"
echo ""
echo -e "${RED}WARNING: This will disrupt running services!${NC}"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

# Get ALB ARN
echo ""
echo -e "${YELLOW}Finding ALB...${NC}"
ALB_ARN=$(aws elbv2 describe-load-balancers \
    --region $REGION \
    --query "LoadBalancers[?LoadBalancerName=='${PROJECT}-${ENV}-alb'].LoadBalancerArn" \
    --output text 2>/dev/null || echo "")

if [ -n "$ALB_ARN" ]; then
    echo "Found ALB: $ALB_ARN"
    
    # Get listeners
    LISTENER_ARNS=$(aws elbv2 describe-listeners \
        --load-balancer-arn "$ALB_ARN" \
        --region $REGION \
        --query "Listeners[].ListenerArn" \
        --output text 2>/dev/null || echo "")
    
    # Delete listener rules
    for LISTENER_ARN in $LISTENER_ARNS; do
        echo -e "${YELLOW}Deleting rules for listener: $LISTENER_ARN${NC}"
        RULE_ARNS=$(aws elbv2 describe-rules \
            --listener-arn "$LISTENER_ARN" \
            --region $REGION \
            --query "Rules[?!IsDefault].RuleArn" \
            --output text 2>/dev/null || echo "")
        
        for RULE_ARN in $RULE_ARNS; do
            echo "  Deleting rule: $RULE_ARN"
            aws elbv2 delete-rule --rule-arn "$RULE_ARN" --region $REGION 2>/dev/null || echo "    (failed)"
        done
        
        # Delete listener
        echo "  Deleting listener: $LISTENER_ARN"
        aws elbv2 delete-listener --listener-arn "$LISTENER_ARN" --region $REGION 2>/dev/null || echo "    (failed)"
    done
    
    # Delete ALB
    echo -e "${YELLOW}Deleting ALB...${NC}"
    aws elbv2 delete-load-balancer --load-balancer-arn "$ALB_ARN" --region $REGION 2>/dev/null || echo "  (failed)"
    echo -e "${GREEN}✓ ALB deleted (will take ~2 minutes to complete)${NC}"
else
    echo "ALB not found (already deleted?)"
fi

# Wait for ALB to be fully deleted before deleting target groups
if [ -n "$ALB_ARN" ]; then
    echo ""
    echo -e "${YELLOW}Waiting for ALB deletion to complete...${NC}"
    sleep 10
fi

# Delete target groups
echo ""
echo -e "${YELLOW}Deleting target groups...${NC}"

for TG_NAME in "${PROJECT}-${ENV}-api-tg" "${PROJECT}-${ENV}-st-tg" "${PROJECT}-${ENV}-supertokens-tg"; do
    TG_ARN=$(aws elbv2 describe-target-groups \
        --region $REGION \
        --query "TargetGroups[?TargetGroupName=='${TG_NAME}'].TargetGroupArn" \
        --output text 2>/dev/null || echo "")
    
    if [ -n "$TG_ARN" ]; then
        echo "Deleting: $TG_NAME ($TG_ARN)"
        aws elbv2 delete-target-group --target-group-arn "$TG_ARN" --region $REGION 2>/dev/null || echo "  (failed - may still be attached)"
    else
        echo "Not found: $TG_NAME"
    fi
done

# Delete ECS services
echo ""
echo -e "${YELLOW}Deleting ECS services...${NC}"

CLUSTER_NAME="${PROJECT}-${ENV}-cluster"

for SERVICE_NAME in "${PROJECT}-${ENV}-api-service" "${PROJECT}-${ENV}-worker-service" "${PROJECT}-${ENV}-supertokens-service"; do
    echo "Checking service: $SERVICE_NAME"
    SERVICE_EXISTS=$(aws ecs describe-services \
        --cluster $CLUSTER_NAME \
        --services $SERVICE_NAME \
        --region $REGION \
        --query "services[0].serviceName" \
        --output text 2>/dev/null || echo "")
    
    if [ "$SERVICE_EXISTS" = "$SERVICE_NAME" ]; then
        echo "  Scaling down to 0..."
        aws ecs update-service \
            --cluster $CLUSTER_NAME \
            --service $SERVICE_NAME \
            --desired-count 0 \
            --region $REGION >/dev/null 2>&1 || echo "    (failed)"
        
        echo "  Deleting service..."
        aws ecs delete-service \
            --cluster $CLUSTER_NAME \
            --service $SERVICE_NAME \
            --force \
            --region $REGION >/dev/null 2>&1 || echo "    (failed)"
        echo "  ✓ Deleted"
    else
        echo "  Not found (already deleted?)"
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Cleanup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Note: Some resources may take a few minutes to fully delete.${NC}"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Wait 2-3 minutes for deletions to complete"
echo "2. Run: cd /Users/vyshakhp/work/utm-backend/infra"
echo "3. Run: pulumi up --yes"
echo ""

