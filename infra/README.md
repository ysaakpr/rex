# Rex Backend - AWS Infrastructure with Pulumi

This directory contains the Pulumi infrastructure-as-code (IaC) for deploying the Rex Backend to AWS using Fargate.

## Architecture Overview

### Components

1. **Networking**
   - VPC with public and private subnets across 2 availability zones
   - Internet Gateway for public subnet access
   - NAT Gateway for private subnet internet access
   - Route tables and associations

2. **Database**
   - Aurora RDS Serverless v2 (PostgreSQL 15.4)
   - Two databases in a single cluster:
     - `rex_backend` - Main application database
     - `supertokens` - SuperTokens authentication database
   - Auto-scaling from 0.5 to 2 ACUs
   - Automated backups with 7-day retention

3. **Cache & Queue**
   - ElastiCache Redis (7.0) for caching and job queues
   - At-rest encryption enabled
   - Automated snapshots

4. **Container Registry**
   - ECR repositories for:
     - API service
     - Worker service
   - Automated image scanning on push
   - Lifecycle policy to keep last 10 images

5. **Compute (ECS Fargate)**
   - **Fully managed Fargate** - No EC2 instances to manage
   - ECS Cluster with Container Insights enabled
   - Services:
     - **API**: 2 tasks (512 CPU, 1024 MB) - Go application
     - **Worker**: 1 task (512 CPU, 1024 MB) - Background jobs
     - **SuperTokens**: 1 task (512 CPU, 1024 MB) - Authentication service
   - Service discovery for internal communication
   - Auto-restart on failure

6. **Frontend Hosting (AWS Amplify)**
   - **AWS Amplify** for React SPA hosting
   - Connected to GitHub repository for CI/CD
   - Automatic builds on push
   - Global CDN distribution
   - Custom domain support
   - Automatic SSL/TLS certificates

7. **Load Balancer**
   - Application Load Balancer (ALB) for backend services only
   - HTTP/HTTPS listeners
   - Path-based routing:
     - `/api/*` → API service
     - `/auth/*` → SuperTokens service
   - Health checks for all backend services

8. **Security**
   - AWS Secrets Manager for sensitive configuration
   - IAM roles with least privilege
   - Security groups with minimal required access
   - SSL/TLS for data in transit

9. **Monitoring**
   - CloudWatch Log Groups for backend services
   - Container Insights enabled
   - Amplify Console for frontend logs and build status
   - 7-day log retention (configurable)

## Deployment Modes

### Standard Mode (Default)
Uses fully managed AWS services for production-ready deployments:
- **RDS Aurora Serverless v2**: Managed PostgreSQL with auto-scaling
- **ElastiCache**: Managed Redis cluster
- **High Availability**: Multi-AZ support, automated backups, automatic failover
- **Cost**: ~$65-130/month for database services

### Low-Cost Mode 🆕
Perfect for development, testing, and staging environments:
- **Single EC2 Spot Instance**: Self-hosted PostgreSQL + Redis on one instance
- **Instance Type**: t3a.small (2 vCPU, 2 GB RAM)
- **Ultra-low Cost**: ~$5-10/month for database services (85-90% savings!)
- **Automatic Setup**: Fully automated installation and configuration
- **Trade-offs**: No high availability, potential spot interruptions

**Enable low-cost mode:**
```bash
pulumi config set rex-backend:lowcost true
```

**📖 See [LOWCOST_MODE.md](./LOWCOST_MODE.md) for complete documentation.**

## Prerequisites

### Local Tools

1. **Pulumi CLI** (latest version)
```bash
# macOS
brew install pulumi/tap/pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Windows
choco install pulumi
```

2. **AWS CLI** (configured with credentials)
```bash
# Install
brew install awscli  # macOS
# or follow: https://aws.amazon.com/cli/

# Configure
aws configure
```

3. **Go 1.23+**
```bash
brew install go  # macOS
```

4. **Docker** (for building and pushing images)
```bash
# Install from https://www.docker.com/products/docker-desktop
```

### AWS Requirements

1. **AWS Account** with appropriate permissions
2. **IAM User/Role** with permissions for:
   - VPC, Subnets, Internet Gateway, NAT Gateway
   - RDS, ElastiCache
   - ECS, ECR
   - ALB, Target Groups
   - Secrets Manager
   - CloudWatch Logs
   - IAM Roles and Policies

3. **S3 Bucket** for Pulumi state (created in setup)

## Initial Setup

### 1. Configure Pulumi Backend

Create an S3 bucket for Pulumi state:

```bash
# Create S3 bucket for state
aws s3 mb s3://rex-backend-pulumi-state --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket rex-backend-pulumi-state \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket rex-backend-pulumi-state \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'
```

### 2. Initialize Pulumi Project

```bash
cd infra

# Login to S3 backend
pulumi login s3://rex-backend-pulumi-state

# Install Go dependencies
go mod download

# Create a new stack (e.g., dev, staging, production)
pulumi stack init dev

# Set AWS region
pulumi config set aws:region us-east-1
```

### 3. Configure Required Settings

Set required configuration:

```bash
# Database master password (generate a strong password)
pulumi config set --secret rex-backend:dbMasterPassword "YourStrongPasswordHere123!"

# SuperTokens API key (generate a strong random string)
pulumi config set --secret rex-backend:supertokensApiKey "your-supertokens-api-key-here"

# GitHub repository for Amplify frontend deployment
pulumi config set rex-backend:githubRepo "https://github.com/yourusername/rex-backend"
pulumi config set rex-backend:githubBranch "main"

# Optional: GitHub token (ONLY needed for private repos or better rate limits)
# For public repos, you can skip this - leave it unset
# For private repos, you MUST set this:
# pulumi config set rex-backend:githubToken "ghp_your_github_token_here"

# Optional: Set custom domain and certificate for backend API
# pulumi config set rex-backend:domainName "api.example.com"
# pulumi config set rex-backend:certificateArn "arn:aws:acm:us-east-1:123456789:certificate/abc-123"
```

### 4. Review Configuration

Check your configuration:

```bash
pulumi config

# Should show:
# KEY                                    VALUE
# aws:region                             us-east-1
# rex-backend:dbMasterPassword           [secret]
# rex-backend:environment                dev
# rex-backend:projectName                rex-backend
# rex-backend:supertokensApiKey          [secret]
# rex-backend:vpcCidr                    10.0.0.0/16
```

## Deployment Process

### Step 1: Initial Infrastructure Deployment

First, deploy the infrastructure to create ECR repositories and other resources.

**Note**: ECS services will initially fail to start because Docker images don't exist yet. This is expected!

```bash
cd infra

# Preview what will be created
pulumi preview

# Deploy infrastructure (creates VPC, RDS, Redis, ECR, ALB, etc.)
pulumi up
```

Review the changes and confirm. Initial deployment takes ~15-20 minutes.

**Expected**: ECS tasks (API, Worker, SuperTokens) will fail to start because images don't exist in ECR yet. This is normal for first deployment!

### Step 2: Build and Push Docker Images

Now that ECR repositories exist, build and push your Docker images.

**Note**: Frontend is deployed via AWS Amplify from GitHub, so no frontend Docker image is needed!

**Option A: Using the helper script (recommended)**

```bash
# Run the automated build and push script
./infra/scripts/build-and-push.sh
```

**Option B: Manual build and push**

```bash
# Get ECR repository URLs from Pulumi
cd infra
API_REPO=$(pulumi stack output apiRepositoryUrl)
WORKER_REPO=$(pulumi stack output workerRepositoryUrl)
AWS_REGION=$(pulumi config get aws:region)
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "API Repository: $API_REPO"
echo "Worker Repository: $WORKER_REPO"

# Go back to project root
cd ..

# Build production images for backend services
docker build -f Dockerfile.prod --target api -t rex-backend-api:latest .
docker build -f Dockerfile.prod --target worker -t rex-backend-worker:latest .

# Login to ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Tag images with ECR repository URLs
docker tag rex-backend-api:latest $API_REPO:latest
docker tag rex-backend-worker:latest $WORKER_REPO:latest

# Push images to ECR
docker push $API_REPO:latest
docker push $WORKER_REPO:latest

echo "✅ Images pushed successfully!"
```

### Step 3: Update ECS Services

Force ECS services to redeploy with the newly pushed images.

**Option A: Using the helper script (recommended)**

```bash
# Run the automated deployment script
./infra/scripts/force-deploy.sh
```

**Option B: Manual deployment**

```bash
cd infra

# Get cluster and service names
CLUSTER_NAME=$(pulumi stack output ecsClusterName)
API_SERVICE=$(pulumi stack output apiServiceName)
WORKER_SERVICE=$(pulumi stack output workerServiceName)
SUPERTOKENS_SERVICE=$(pulumi stack output supertokensServiceName)

# Force new deployment for all services
aws ecs update-service --cluster $CLUSTER_NAME --service $API_SERVICE --force-new-deployment
aws ecs update-service --cluster $CLUSTER_NAME --service $WORKER_SERVICE --force-new-deployment
aws ecs update-service --cluster $CLUSTER_NAME --service $SUPERTOKENS_SERVICE --force-new-deployment

echo "✅ ECS services are redeploying with new images..."
echo "This will take 2-3 minutes. Monitor status in AWS Console."
```

### Step 4: Run Database Migrations

After ECS services are running, run database migrations:

```bash
# Get outputs
CLUSTER_NAME=$(pulumi stack output ecsClusterName)
TASK_DEF=$(pulumi stack output migrationTaskDefinitionArn)
SUBNET_1=$(pulumi stack output privateSubnetIds | jq -r '.[0]')
SUBNET_2=$(pulumi stack output privateSubnetIds | jq -r '.[1]')
SECURITY_GROUP=$(pulumi stack output ecsSecurityGroup)

# Run migration task
aws ecs run-task \
  --cluster $CLUSTER_NAME \
  --task-definition $TASK_DEF \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNET_1,$SUBNET_2],securityGroups=[$SECURITY_GROUP],assignPublicIp=DISABLED}"

# Monitor logs
aws logs tail /ecs/rex-backend-dev-migration --follow
```

### Step 5: Create SuperTokens Database

SuperTokens will auto-create its schema on first run, but you need to create the database:

```bash
# Get database endpoint
DB_ENDPOINT=$(pulumi stack output rdsClusterEndpoint)

# Connect to database (you may need to be in VPC or use bastion)
# Option 1: Use ECS task with psql
# Option 2: Use AWS Systems Manager Session Manager

# Create supertokens database
psql -h $DB_ENDPOINT -U rexadmin -d rex_backend -c "CREATE DATABASE supertokens;"
```

### Step 6: Verify Deployment

```bash
# Get output URLs
ALB_DNS=$(pulumi stack output albDnsName)
FRONTEND_URL=$(pulumi stack output frontendUrl)

# Test API health
curl http://$ALB_DNS/api/health

# Test SuperTokens
curl http://$ALB_DNS/auth/hello

# Test frontend (open in browser)
echo "Frontend URL: $FRONTEND_URL"
open $FRONTEND_URL  # macOS
# Or visit the URL in your browser
```

**Frontend Deployment**: 
- Amplify will automatically build and deploy your frontend when you push to GitHub
- Check build status in AWS Console → Amplify → Your App
- Frontend URL will be in the format: `https://main.xxxxx.amplifyapp.com`

## Quick Deployment Reference

After initial setup, use these commands for quick deployments:

### Deploy Backend Code Changes

```bash
# 1. Build and push new images
./infra/scripts/build-and-push.sh

# 2. Force ECS services to redeploy
./infra/scripts/force-deploy.sh

# 3. (Optional) Run migrations if database schema changed
./infra/scripts/run-migration.sh
```

### Deploy Frontend Changes

```bash
# Simply push to GitHub - Amplify handles the rest!
git add frontend/
git commit -m "Update frontend"
git push origin main

# Monitor build in AWS Console → Amplify → rex-backend-dev-frontend
```

### Deploy Infrastructure Changes

```bash
cd infra
pulumi preview  # Review changes
pulumi up       # Apply changes
```

## Updating the Stack

### Update Application Code

**Backend Services**:
1. Build new Docker images with updated code
2. Tag with a new version or `latest`
3. Push to ECR
4. Force new deployment:

```bash
# Force new deployment of backend services
aws ecs update-service --cluster rex-backend-dev-cluster --service rex-backend-dev-api --force-new-deployment
aws ecs update-service --cluster rex-backend-dev-cluster --service rex-backend-dev-worker --force-new-deployment
aws ecs update-service --cluster rex-backend-dev-cluster --service rex-backend-dev-supertokens --force-new-deployment
```

**Frontend**:
- Simply push to GitHub - Amplify will automatically build and deploy
- No manual steps required!

```bash
git add .
git commit -m "Update frontend"
git push origin main
# Amplify automatically builds and deploys
```

### Update Infrastructure

```bash
cd infra

# Preview changes
pulumi preview

# Apply changes
pulumi up
```

### Rollback

```bash
# Rollback to previous stack state
pulumi stack history
pulumi stack export --version <version-number> > previous-state.json
pulumi stack import --file previous-state.json
```

## Scaling

### Manual Scaling

Update task counts in `ecs_services.go`:

```go
DesiredCount: pulumi.Int(4), // Increase from 2
```

Then run:
```bash
pulumi up
```

### Auto Scaling

Add Application Auto Scaling (future enhancement):

```bash
# Create scaling target
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/rex-backend-dev-cluster/rex-backend-dev-api \
  --min-capacity 2 \
  --max-capacity 10

# Create scaling policy
aws application-autoscaling put-scaling-policy \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/rex-backend-dev-cluster/rex-backend-dev-api \
  --policy-name cpu-scaling \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration file://scaling-policy.json
```

## Monitoring & Logs

### View Logs

**Backend Logs** (CloudWatch):
```bash
# API logs
aws logs tail /ecs/rex-backend-dev-api --follow

# Worker logs
aws logs tail /ecs/rex-backend-dev-worker --follow

# SuperTokens logs
aws logs tail /ecs/rex-backend-dev-supertokens --follow
```

**Frontend Logs** (Amplify Console):
- Go to AWS Console → Amplify → Your App
- View build logs and deployment history
- Access logs show HTTP requests

### CloudWatch Metrics

- Container Insights: Enabled by default
- View in AWS Console → CloudWatch → Container Insights

### Alarms

Create CloudWatch Alarms for:
- High CPU usage
- High memory usage
- HTTP 5xx errors
- Database connections

## Cost Optimization

### Ultra Low-Cost Mode 💰

**Best for dev/test/staging environments:**
```bash
pulumi config set rex-backend:lowcost true
```
- Saves 85-90% on database costs (~$5-10/month vs ~$65-130/month)
- Uses single EC2 Spot instance for PostgreSQL + Redis
- See [LOWCOST_MODE.md](./LOWCOST_MODE.md) for details

### Development Environment (Standard Mode)

- Aurora Serverless scales to 0.5 ACU when idle
- Use smallest Redis node (cache.t4g.micro)
- Reduce task counts (1 API, 1 Frontend, 1 Worker)
- Shorter log retention (7 days)

### Production Environment

- Enable Aurora auto-pause (scales to 0 when idle)
- Use Reserved Instances or Savings Plans
- Enable S3 lifecycle policies
- Use CloudFront for frontend caching

### Estimated Monthly Costs (Development)

- **Aurora Serverless v2**: ~$30-50 (minimal usage)
- **ElastiCache (t4g.micro)**: ~$12
- **ECS Fargate**: ~$30-40 (3 backend tasks, no frontend)
- **AWS Amplify**: ~$0-5 (build minutes + hosting for 1 app)
- **NAT Gateway**: ~$33 (with data transfer)
- **ALB**: ~$16
- **CloudWatch Logs**: ~$5
- **Total**: ~$126-161/month

**Cost Savings vs ECS Frontend**: ~$10-15/month by using Amplify instead of ECS Fargate for frontend

## Troubleshooting

### Tasks Won't Start

Check:
1. ECR images exist and are tagged correctly
2. Secrets Manager has correct values
3. Security groups allow necessary traffic
4. Subnets have internet access (via NAT)

```bash
# View ECS task failures
aws ecs describe-tasks --cluster rex-backend-dev-cluster --tasks <task-arn>
```

### Database Connection Issues

Check:
1. RDS security group allows traffic from ECS security group
2. Connection string is correct
3. Database exists (especially `supertokens` database)

### SuperTokens Issues

Check:
1. SuperTokens database exists
2. SuperTokens service can connect to database
3. API can reach SuperTokens via service discovery

### Cannot Access Application

Check:
1. ALB security group allows inbound on 80/443
2. Target groups show healthy targets
3. ECS tasks are running

## Security Best Practices

1. **Rotate Secrets Regularly**
```bash
# Update database password
pulumi config set --secret rex-backend:dbMasterPassword "NewPassword123!"
pulumi up
```

2. **Enable HTTPS**
- Request ACM certificate for your domain
- Configure certificate ARN in Pulumi config
- Update DNS to point to ALB

3. **Enable WAF** (optional)
- Add AWS WAF in front of ALB
- Configure rate limiting and common attack protection

4. **Enable VPC Flow Logs** (for audit)
```bash
# Add to networking.go
```

5. **Enable CloudTrail** (for API audit)
- Enable in AWS Console or add to Pulumi

## Cleanup

To destroy all resources:

```bash
cd infra

# Preview what will be deleted
pulumi destroy --preview

# Destroy all resources
pulumi destroy

# Remove stack
pulumi stack rm dev
```

**Warning**: This will delete all data including databases!

## Support & References

- [Pulumi AWS Documentation](https://www.pulumi.com/docs/clouds/aws/)
- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html)
- [SuperTokens Self-Hosted](https://supertokens.com/docs/emailpassword/introduction)

## Next Steps

1. Set up CI/CD pipeline (GitHub Actions, AWS CodePipeline)
2. Configure custom domain with Route53
3. Add CloudFront for CDN
4. Enable auto-scaling policies
5. Set up monitoring and alerts
6. Implement backup and disaster recovery procedures

