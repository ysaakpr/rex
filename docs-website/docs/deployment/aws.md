# AWS Deployment

Detailed guide for deploying to Amazon Web Services using Pulumi.

## Overview

This guide covers AWS-specific deployment using Pulumi Infrastructure as Code (IaC). The infrastructure includes:

- **Compute**: ECS Fargate or EC2
- **Database**: RDS PostgreSQL
- **Cache**: ElastiCache Redis  
- **Load Balancer**: Application Load Balancer (ALB)
- **Secrets**: AWS Secrets Manager
- **Monitoring**: CloudWatch Logs and Metrics
- **CDN**: CloudFront (optional)

## Prerequisites

```bash
# Install required tools
brew install pulumi awscli

# Configure AWS CLI
aws configure

# Verify access
aws sts get-caller-identity
```

## Infrastructure Modes

### Mode 1: All-in-One EC2 (Recommended)

**Use case**: Small to medium workloads, cost-effective

**Architecture**:
- Single EC2 instance (t3.medium)
- Docker Compose running all services
- Managed PostgreSQL (RDS)
- Managed Redis (ElastiCache)  
- Application Load Balancer for HTTPS

**Cost**: ~$50-100/month

**Deployment**:
```bash
cd infra/
pulumi config set deploymentMode allinone
pulumi up
```

### Mode 2: ECS with Separate Services

**Use case**: Production at scale, high availability

**Architecture**:
- ECS Fargate tasks (API, Worker, Frontend, SuperTokens)
- RDS PostgreSQL (Multi-AZ)
- ElastiCache Redis (Cluster mode)
- Application Load Balancer
- Auto-scaling enabled

**Cost**: ~$200-500/month

**Deployment**:
```bash
pulumi config set deploymentMode ecs
pulumi up
```

### Mode 3: Low-Cost

**Use case**: Development, staging, testing

**Architecture**:
- Minimal EC2 instance (t3.micro)
- Single availability zone
- No redundancy

**Cost**: ~$20-30/month

**Deployment**:
```bash
pulumi config set deploymentMode lowcost
pulumi up
```

## Step-by-Step Deployment

### 1. Initialize Pulumi Project

```bash
cd infra/

# Create new stack
pulumi stack init production

# Set AWS region
pulumi config set aws:region us-east-1

# Set deployment mode
pulumi config set rex-backend-infra:deploymentMode allinone

# Set domain
pulumi config set rex-backend-infra:domain yourdomain.com
```

### 2. Configure Secrets

```bash
# SuperTokens API Key (generate random string)
pulumi config set --secret supertokensApiKey "$(openssl rand -base64 32)"

# Database password
pulumi config set --secret dbPassword "$(openssl rand -base64 24)"

# JWT secret
pulumi config set --secret jwtSecret "$(openssl rand -base64 32)"
```

### 3. Request SSL Certificate

```bash
# Via AWS Console or CLI
aws acm request-certificate \
  --domain-name yourdomain.com \
  --subject-alternative-names "*.yourdomain.com" \
  --validation-method DNS \
  --region us-east-1

# Get certificate ARN
aws acm list-certificates --region us-east-1
```

**Add DNS validation records** to your domain provider, then:

```bash
# Set certificate ARN in Pulumi
pulumi config set rex-backend-infra:certificateArn arn:aws:acm:us-east-1:ACCOUNT:certificate/ID
```

### 4. Deploy Infrastructure

```bash
# Preview changes
pulumi preview

# Deploy
pulumi up

# Save outputs
pulumi stack output --json > outputs.json
```

**Pulumi will create**:
- VPC with public/private subnets
- Security groups
- RDS PostgreSQL instance
- ElastiCache Redis cluster
- EC2 instance or ECS cluster
- Application Load Balancer
- CloudWatch log groups
- IAM roles and policies

### 5. Configure DNS

After deployment, set up DNS records:

```bash
# Get ALB DNS name
pulumi stack output albDnsName

# Or from AWS CLI
aws elbv2 describe-load-balancers \
  --names utm-alb \
  --query 'LoadBalancers[0].DNSName' \
  --output text
```

**Add DNS records**:
```
CNAME  api.yourdomain.com   →  utm-alb-xxxxx.us-east-1.elb.amazonaws.com
CNAME  app.yourdomain.com   →  utm-alb-xxxxx.us-east-1.elb.amazonaws.com
CNAME  www.yourdomain.com   →  utm-alb-xxxxx.us-east-1.elb.amazonaws.com
```

### 6. Build and Push Docker Images

```bash
# Get ECR repository URLs
pulumi stack output ecrApiRepo
pulumi stack output ecrWorkerRepo
pulumi stack output ecrFrontendRepo

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-1.amazonaws.com

# Build and push (multi-architecture for compatibility)
./scripts/build-and-push-multiarch.sh

# Or single architecture
./scripts/build-and-push.sh
```

### 7. Deploy Application

**For All-in-One EC2**:
```bash
./scripts/allinone-deploy.sh
```

**For ECS**:
```bash
./scripts/force-deploy.sh
```

### 8. Run Database Migrations

**All-in-One EC2**:
```bash
# SSH into instance
EC2_IP=$(pulumi stack output ec2PublicIp)
ssh -i ~/.ssh/your-key.pem ec2-user@$EC2_IP

# Run migrations
docker exec utm-api ./migrate up
```

**ECS**:
```bash
# Run migration task
aws ecs run-task \
  --cluster utm-cluster \
  --task-definition utm-migration \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}"
```

### 9. Bootstrap First Admin

```bash
# Get RDS endpoint
RDS_ENDPOINT=$(pulumi stack output rdsEndpoint)

# Connect to database
psql -h $RDS_ENDPOINT -U postgres -d utm_backend

# Insert first admin (replace with your SuperTokens user ID)
INSERT INTO platform_admins (id, user_id, notes, created_at, updated_at)
VALUES (gen_random_uuid(), 'YOUR_USER_ID', 'Bootstrap admin', NOW(), NOW());
```

### 10. Verify Deployment

```bash
# Test API health
curl https://api.yourdomain.com/health

# Test authentication
curl https://api.yourdomain.com/auth/hello

# Test frontend
curl https://app.yourdomain.com
```

## Infrastructure Components

### VPC and Networking

```typescript
// Pulumi code creates:
const vpc = new aws.ec2.Vpc("utm-vpc", {
  cidrBlock: "10.0.0.0/16",
  enableDnsHostnames: true,
  enableDnsSupport: true,
  tags: {Name: "utm-vpc"}
});

const publicSubnet = new aws.ec2.Subnet("utm-public-subnet", {
  vpcId: vpc.id,
  cidrBlock: "10.0.1.0/24",
  availabilityZone: "us-east-1a",
  mapPublicIpOnLaunch: true
});

const privateSubnet = new aws.ec2.Subnet("utm-private-subnet", {
  vpcId: vpc.id,
  cidrBlock: "10.0.2.0/24",
  availabilityZone: "us-east-1b"
});
```

### Security Groups

**ALB Security Group**:
- Inbound: 80 (HTTP), 443 (HTTPS) from 0.0.0.0/0
- Outbound: All traffic

**Application Security Group**:
- Inbound: 8080 from ALB security group
- Outbound: All traffic

**RDS Security Group**:
- Inbound: 5432 from application security group
- Outbound: None

**ElastiCache Security Group**:
- Inbound: 6379 from application security group
- Outbound: None

### RDS Configuration

```typescript
const dbInstance = new aws.rds.Instance("utm-db", {
  engine: "postgres",
  engineVersion: "15.4",
  instanceClass: "db.t3.micro",  // Scale up for production
  allocatedStorage: 20,
  storageType: "gp3",
  dbName: "utm_backend",
  username: "postgres",
  password: dbPassword,
  vpcSecurityGroupIds: [dbSecurityGroup.id],
  dbSubnetGroupName: dbSubnetGroup.name,
  backupRetentionPeriod: 7,
  backupWindow: "03:00-04:00",
  maintenanceWindow: "mon:04:00-mon:05:00",
  skipFinalSnapshot: false,
  finalSnapshotIdentifier: "utm-final-snapshot",
  storageEncrypted: true,
  multiAz: true,  // For production
  publiclyAccessible: false
});
```

### ElastiCache Configuration

```typescript
const redisCluster = new aws.elasticache.Cluster("utm-redis", {
  engine: "redis",
  engineVersion: "7.0",
  nodeType: "cache.t3.micro",
  numCacheNodes: 1,
  parameterGroupName: "default.redis7",
  port: 6379,
  securityGroupIds: [redisSecurityGroup.id],
  subnetGroupName: redisSubnetGroup.name
});
```

### Application Load Balancer

```typescript
const alb = new aws.lb.LoadBalancer("utm-alb", {
  internal: false,
  loadBalancerType: "application",
  securityGroups: [albSecurityGroup.id],
  subnets: [publicSubnet1.id, publicSubnet2.id],
  enableHttp2: true,
  enableCrossZoneLoadBalancing: true
});

const httpsListener = new aws.lb.Listener("utm-https-listener", {
  loadBalancerArn: alb.arn,
  port: 443,
  protocol: "HTTPS",
  certificateArn: certificateArn,
  defaultActions: [{
    type: "forward",
    targetGroupArn: apiTargetGroup.arn
  }]
});
```

## Monitoring and Logging

### CloudWatch Logs

**Log Groups**:
- `/aws/ecs/utm-api` - API application logs
- `/aws/ecs/utm-worker` - Worker logs
- `/aws/ecs/utm-frontend` - Frontend logs
- `/aws/rds/instance/utm-db/postgresql` - Database logs

**View logs**:
```bash
# Stream API logs
aws logs tail /aws/ecs/utm-api --follow

# Filter errors
aws logs filter-pattern '{ $.level = "error" }' /aws/ecs/utm-api

# Get recent errors
aws logs tail /aws/ecs/utm-api --since 1h --filter-pattern "ERROR"
```

### CloudWatch Metrics

**Key metrics to monitor**:
- `CPUUtilization` - EC2/ECS CPU usage
- `MemoryUtilization` - Memory usage
- `TargetResponseTime` - API response time
- `RequestCount` - Traffic volume
- `HealthyHostCount` - Healthy targets
- `DatabaseConnections` - RDS connections

**Create dashboard**:
```bash
aws cloudwatch put-dashboard \
  --dashboard-name UTM-Dashboard \
  --dashboard-body file://dashboard.json
```

### CloudWatch Alarms

```bash
# High CPU alarm
aws cloudwatch put-metric-alarm \
  --alarm-name utm-high-cpu \
  --alarm-description "CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions <sns-topic-arn>

# Database connections alarm
aws cloudwatch put-metric-alarm \
  --alarm-name utm-db-connections \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --dimensions Name=DBInstanceIdentifier,Value=utm-db \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold
```

## Scaling

### Auto Scaling (ECS)

```typescript
// Enable auto scaling
const scalingTarget = new aws.appautoscaling.Target("utm-api-scaling", {
  serviceNamespace: "ecs",
  resourceId: pulumi.interpolate`service/${ecsCluster.name}/${ecsService.name}`,
  scalableDimension: "ecs:service:DesiredCount",
  minCapacity: 2,
  maxCapacity: 10
});

const scalingPolicy = new aws.appautoscaling.Policy("utm-api-policy", {
  policyType: "TargetTrackingScaling",
  resourceId: scalingTarget.resourceId,
  scalableDimension: scalingTarget.scalableDimension,
  serviceNamespace: scalingTarget.serviceNamespace,
  targetTrackingScalingPolicyConfiguration: {
    targetValue: 70,  // Target CPU 70%
    predefinedMetricSpecification: {
      predefinedMetricType: "ECSServiceAverageCPUUtilization"
    },
    scaleInCooldown: 300,
    scaleOutCooldown: 60
  }
});
```

### Manual Scaling

```bash
# Scale API service
aws ecs update-service \
  --cluster utm-cluster \
  --service utm-api \
  --desired-count 5

# Scale database
aws rds modify-db-instance \
  --db-instance-identifier utm-db \
  --db-instance-class db.t3.medium \
  --apply-immediately
```

## Backup and Disaster Recovery

### Automated Backups

**RDS automated backups** (enabled by default):
- Retention: 7 days
- Backup window: 03:00-04:00 UTC
- Point-in-time recovery available

**Manual snapshots**:
```bash
aws rds create-db-snapshot \
  --db-instance-identifier utm-db \
  --db-snapshot-identifier utm-manual-$(date +%Y%m%d-%H%M)
```

### Disaster Recovery Plan

1. **Database Failure**:
```bash
# Restore from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier utm-db-restored \
  --db-snapshot-identifier utm-auto-snapshot-xxxxx

# Update application config
pulumi config set rdsEndpoint <new-endpoint>
pulumi up
```

2. **Region Failure**:
- Enable cross-region RDS snapshots
- Replicate Docker images to another region
- Have Pulumi stack ready in backup region

3. **Data Corruption**:
```bash
# Point-in-time recovery
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier utm-db \
  --target-db-instance-identifier utm-db-restored \
  --restore-time 2024-01-15T10:00:00Z
```

## Cost Optimization

### Reserved Instances

Save 30-60% by committing to 1 or 3 years:

```bash
# Purchase RDS reserved instance
aws rds purchase-reserved-db-instances-offering \
  --reserved-db-instances-offering-id xxxxx \
  --reserved-db-instance-id utm-db-reserved

# Purchase EC2 reserved instance
aws ec2 purchase-reserved-instances-offering \
  --reserved-instances-offering-id xxxxx \
  --instance-count 1
```

### Savings Plans

```bash
# Compute Savings Plan (most flexible)
aws savingsplans create-savings-plan \
  --savings-plan-offering-id xxxxx \
  --commitment 10.00 \
  --upfront-payment-amount 0.00
```

### Cost Monitoring

```bash
# Get current month costs
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=SERVICE

# Set budget alert
aws budgets create-budget \
  --account-id $(aws sts get-caller-identity --query Account --output text) \
  --budget file://budget.json
```

## Troubleshooting

### Cannot Access Application

1. Check ALB health targets:
```bash
aws elbv2 describe-target-health \
  --target-group-arn <arn>
```

2. Verify security groups:
```bash
aws ec2 describe-security-groups \
  --group-ids sg-xxxxx
```

3. Check DNS:
```bash
dig api.yourdomain.com
nslookup api.yourdomain.com
```

### Database Connection Issues

```bash
# Test from EC2/ECS
aws ssm start-session --target <instance-id>
telnet <rds-endpoint> 5432

# Check RDS status
aws rds describe-db-instances \
  --db-instance-identifier utm-db \
  --query 'DBInstances[0].DBInstanceStatus'
```

### High Costs

```bash
# Identify top resources
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity DAILY \
  --metrics UnblendedCost \
  --group-by Type=SERVICE \
  --group-by Type=USAGE_TYPE
```

**Common cost issues**:
- NAT Gateway (~$30/month) - consider removing if not needed
- Data transfer - enable CloudFront
- Over-provisioned RDS/EC2 - downsize
- Unnecessary EBS snapshots - set lifecycle policy

## Next Steps

- [Production Setup](/deployment/production-setup) - Complete production guide
- [Docker Deployment](/deployment/docker) - Docker compose setup
- [Monitoring](/guides/monitoring) - Detailed monitoring guide
- [Security](/guides/security) - Security hardening

