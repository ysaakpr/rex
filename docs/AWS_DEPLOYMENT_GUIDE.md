# AWS Deployment Guide

Complete guide for deploying UTM Backend to AWS using Pulumi and Fargate.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Prerequisites](#prerequisites)
4. [Initial Setup](#initial-setup)
5. [Deployment Steps](#deployment-steps)
6. [Post-Deployment Configuration](#post-deployment-configuration)
7. [Monitoring & Maintenance](#monitoring--maintenance)
8. [Troubleshooting](#troubleshooting)

## Overview

The UTM Backend is deployed to AWS using:
- **Infrastructure as Code**: Pulumi (Go)
- **Compute**: ECS Fargate
- **Database**: Aurora RDS Serverless v2 (PostgreSQL)
- **Cache/Queue**: ElastiCache Redis
- **Load Balancer**: Application Load Balancer (ALB)
- **Container Registry**: ECR
- **Secrets**: AWS Secrets Manager
- **Logs**: CloudWatch

## Architecture

```
                    Internet
                       |
                    [ALB]
                   /   |   \
                  /    |    \
           Frontend   API   SuperTokens
                |      |      |
                +------+------+
                       |
           [Service Discovery: Private DNS]
                       |
        +-------------+-------------+
        |             |             |
    [API Task]   [Worker Task] [SuperTokens Task]
        |             |             |
        +-------------+-------------+
                      |
        +-------------+-------------+
        |                           |
    [Aurora RDS]               [Redis]
    - utm_backend              - Cache
    - supertokens              - Queue
```

### Components

- **VPC**: 10.0.0.0/16 CIDR
  - Public Subnets: 10.0.0.0/24, 10.0.1.0/24 (2 AZs)
  - Private Subnets: 10.0.10.0/24, 10.0.11.0/24 (2 AZs)
  
- **Database**: Aurora Serverless v2
  - Engine: PostgreSQL 15.4
  - Scaling: 0.5 - 2.0 ACUs
  - Databases: utm_backend, supertokens
  
- **Cache**: ElastiCache Redis 7.0
  - Node Type: cache.t4g.micro (dev)
  - Single node (dev), multi-node (prod)
  
- **Compute**: ECS Fargate
  - API: 2 tasks (512 CPU, 1024 MB)
  - Worker: 1 task (512 CPU, 1024 MB)
  - Frontend: 2 tasks (256 CPU, 512 MB)
  - SuperTokens: 1 task (512 CPU, 1024 MB)

## Prerequisites

### Local Tools

1. **Pulumi CLI** (latest)
```bash
# macOS
brew install pulumi/tap/pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Verify
pulumi version
```

2. **AWS CLI** (v2)
```bash
# macOS
brew install awscli

# Linux
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Verify
aws --version
```

3. **Docker**
```bash
# Install Docker Desktop from https://www.docker.com/products/docker-desktop
```

4. **Go 1.23+**
```bash
# macOS
brew install go

# Verify
go version
```

5. **jq** (for scripts)
```bash
# macOS
brew install jq

# Linux
sudo apt-get install jq
```

### AWS Setup

1. **AWS Account** with appropriate permissions
2. **IAM User** with permissions for:
   - EC2, VPC, Subnets, Security Groups
   - RDS, ElastiCache
   - ECS, ECR
   - ALB, Target Groups
   - Secrets Manager, IAM
   - CloudWatch Logs
   - S3 (for Pulumi state)

3. **Configure AWS CLI**
```bash
aws configure
# Enter: Access Key ID, Secret Access Key, Region (us-east-1), Output format (json)
```

## Initial Setup

### Step 1: Run Setup Script

The easiest way to get started:

```bash
cd infra/scripts
./setup-pulumi.sh
```

This interactive script will:
1. Create S3 bucket for Pulumi state
2. Initialize Pulumi stack
3. Configure environment and region
4. Set secrets (database password, SuperTokens API key)
5. Optional: Configure custom domain

### Step 2: Manual Setup (Alternative)

If you prefer manual setup:

```bash
cd infra

# Create S3 bucket
aws s3 mb s3://utm-backend-pulumi-state --region us-east-1
aws s3api put-bucket-versioning \
  --bucket utm-backend-pulumi-state \
  --versioning-configuration Status=Enabled

# Login to Pulumi
pulumi login s3://utm-backend-pulumi-state

# Initialize stack
pulumi stack init dev

# Set configuration
pulumi config set aws:region us-east-1
pulumi config set utm-backend:environment dev
pulumi config set utm-backend:projectName utm-backend

# Set secrets
pulumi config set --secret utm-backend:dbMasterPassword "YourStrongPassword123!"
pulumi config set --secret utm-backend:supertokensApiKey "your-supertokens-api-key"

# Install dependencies
go mod download
```

## Deployment Steps

### Step 1: Preview Infrastructure

```bash
cd infra
pulumi preview
```

Review the resources that will be created (~50-60 resources).

### Step 2: Deploy Infrastructure

```bash
pulumi up
```

Confirm and wait for deployment (15-20 minutes).

### Step 3: Save Outputs

```bash
# Get important outputs
pulumi stack output > outputs.txt

# Or individually
pulumi stack output albDnsName
pulumi stack output rdsClusterEndpoint
pulumi stack output apiRepositoryUrl
```

### Step 4: Build and Push Docker Images

Use the provided script:

```bash
cd infra/scripts
./build-and-push.sh
```

Or manually:

```bash
# Get ECR URLs
API_REPO=$(pulumi stack output apiRepositoryUrl)
WORKER_REPO=$(pulumi stack output workerRepositoryUrl)
FRONTEND_REPO=$(pulumi stack output frontendRepositoryUrl)

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build images
cd ../..
docker build -f Dockerfile.prod --target api -t utm-backend-api:latest .
docker build -f Dockerfile.prod --target worker -t utm-backend-worker:latest .
docker build -f frontend/Dockerfile.prod -t utm-backend-frontend:latest frontend/

# Tag and push
docker tag utm-backend-api:latest $API_REPO:latest
docker tag utm-backend-worker:latest $WORKER_REPO:latest
docker tag utm-backend-frontend:latest $FRONTEND_REPO:latest

docker push $API_REPO:latest
docker push $WORKER_REPO:latest
docker push $FRONTEND_REPO:latest
```

### Step 5: Wait for Services to Start

```bash
# Check service status
CLUSTER=$(pulumi stack output ecsClusterName)
aws ecs describe-services --cluster $CLUSTER --services utm-backend-dev-api

# Services should show RUNNING state
```

### Step 6: Create SuperTokens Database

```bash
cd infra/scripts
./create-supertokens-db.sh
```

Or manually:
```sql
-- Connect to database
psql -h <rds-endpoint> -U utmadmin -d utm_backend

-- Create database
CREATE DATABASE supertokens;
GRANT ALL PRIVILEGES ON DATABASE supertokens TO utmadmin;
```

### Step 7: Run Database Migrations

```bash
cd infra/scripts
./run-migration.sh
```

This runs the migration task in Fargate and waits for completion.

### Step 8: Verify Deployment

```bash
ALB_DNS=$(pulumi stack output albDnsName)

# Test API health
curl http://$ALB_DNS/api/health

# Test SuperTokens
curl http://$ALB_DNS/auth/hello

# Test Frontend
curl http://$ALB_DNS/
```

## Post-Deployment Configuration

### Custom Domain Setup

1. **Request ACM Certificate**
```bash
aws acm request-certificate \
  --domain-name utm.example.com \
  --validation-method DNS \
  --region us-east-1
```

2. **Validate Certificate** (add DNS records)

3. **Update Pulumi Configuration**
```bash
cd infra
pulumi config set utm-backend:domainName utm.example.com
pulumi config set utm-backend:certificateArn arn:aws:acm:us-east-1:123456789:certificate/abc-123
pulumi up
```

4. **Create Route53 Record**
```bash
# Point your domain to ALB DNS name
aws route53 change-resource-record-sets \
  --hosted-zone-id YOUR_ZONE_ID \
  --change-batch file://route53-change.json
```

### Enable Auto Scaling

```bash
# Register scalable target for API
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/utm-backend-dev-cluster/utm-backend-dev-api \
  --min-capacity 2 \
  --max-capacity 10

# Create scaling policy
aws application-autoscaling put-scaling-policy \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/utm-backend-dev-cluster/utm-backend-dev-api \
  --policy-name cpu-scaling \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration '{
    "TargetValue": 70.0,
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
    }
  }'
```

## Monitoring & Maintenance

### View Logs

```bash
# API logs
aws logs tail /ecs/utm-backend-dev-api --follow

# Worker logs
aws logs tail /ecs/utm-backend-dev-worker --follow

# Frontend logs
aws logs tail /ecs/utm-backend-dev-frontend --follow

# SuperTokens logs
aws logs tail /ecs/utm-backend-dev-supertokens --follow
```

### CloudWatch Metrics

- Navigate to AWS Console → CloudWatch → Container Insights
- Select your cluster to view metrics

### Update Application

```bash
# Build and push new images
cd infra/scripts
./build-and-push.sh

# Force new deployment
./force-deploy.sh
```

### Database Backups

Aurora automatically creates daily backups. To create a manual snapshot:

```bash
aws rds create-db-cluster-snapshot \
  --db-cluster-identifier utm-backend-dev-aurora-cluster \
  --db-cluster-snapshot-identifier utm-backend-dev-snapshot-$(date +%Y%m%d)
```

## Troubleshooting

### Services Not Starting

**Check Task Logs:**
```bash
# Get task ARN
CLUSTER=$(pulumi stack output ecsClusterName)
TASK_ARN=$(aws ecs list-tasks --cluster $CLUSTER --service-name utm-backend-dev-api --query 'taskArns[0]' --output text)

# Describe task
aws ecs describe-tasks --cluster $CLUSTER --tasks $TASK_ARN

# Check logs
aws logs tail /ecs/utm-backend-dev-api --follow
```

**Common Issues:**
1. Image not found in ECR → Push images first
2. Secrets not accessible → Check IAM permissions
3. Database connection failed → Check security groups

### Database Connection Issues

```bash
# Test database connectivity from ECS task
aws ecs run-task \
  --cluster utm-backend-dev-cluster \
  --task-definition utm-backend-dev-api \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[...],securityGroups=[...],assignPublicIp=DISABLED}" \
  --overrides '{
    "containerOverrides": [{
      "name": "api",
      "command": ["sh", "-c", "nc -zv <db-endpoint> 5432"]
    }]
  }'
```

### ALB Health Checks Failing

```bash
# Check target group health
aws elbv2 describe-target-health \
  --target-group-arn $(pulumi stack output apiTargetGroupArn)

# Check task health endpoint
curl http://<task-private-ip>:8080/health
```

### High Costs

1. **Stop unused environments:**
```bash
pulumi destroy
```

2. **Scale down tasks:**
```bash
aws ecs update-service --cluster <cluster> --service <service> --desired-count 1
```

3. **Pause Aurora cluster:**
```bash
aws rds stop-db-cluster --db-cluster-identifier utm-backend-dev-aurora-cluster
```

## Cost Estimates

### Development Environment (~$150-200/month)

- Aurora Serverless v2: $40-60
- ElastiCache (t4g.micro): $12
- ECS Fargate: $50-70
- NAT Gateway: $33
- ALB: $18
- Data Transfer: $10-20
- CloudWatch Logs: $5

### Production Environment (estimate)

- Scale up tasks, add auto-scaling
- Use larger database instances
- Add CloudFront CDN
- Enable additional monitoring
- Estimated: $500-800/month

## Next Steps

1. ✅ Infrastructure deployed
2. ✅ Services running
3. ✅ Database migrated
4. ⬜ Set up CI/CD pipeline
5. ⬜ Configure monitoring and alerts
6. ⬜ Set up custom domain
7. ⬜ Enable auto-scaling
8. ⬜ Configure backup procedures

## Resources

- [Pulumi AWS Documentation](https://www.pulumi.com/docs/clouds/aws/)
- [ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html)
- [Main Infrastructure README](../infra/README.md)

---

**Need Help?** Check the troubleshooting section or review CloudWatch logs for detailed error messages.

