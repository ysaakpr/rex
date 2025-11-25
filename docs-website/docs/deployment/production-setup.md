# Production Setup

Complete guide to deploying the Rex to production.

## Overview

This guide covers:
- Production environment setup
- Infrastructure provisioning (AWS)
- SSL/HTTPS configuration
- Database and Redis setup
- Application deployment
- Monitoring and logging
- Security hardening

## Prerequisites

- AWS Account with appropriate IAM permissions
- Domain name configured with DNS access
- Docker and Docker Compose installed locally
- Pulumi CLI installed
- AWS CLI configured

## Architecture Overview

### Production Stack

```
Internet
  ↓
ALB (HTTPS) → ECS Tasks
              ├── API (Go)
              ├── Worker (Go)
              ├── Frontend (React)
              └── SuperTokens
  ↓
RDS PostgreSQL
Redis ElastiCache
Secrets Manager
CloudWatch Logs
```

### Deployment Modes

1. **All-in-One EC2** (Recommended for small to medium workloads)
   - Single EC2 instance running all services via Docker Compose
   - Cost-effective (~$30-50/month)
   - Easy to manage
   - Good for up to 1000 users

2. **ECS with ALB** (Recommended for production at scale)
   - Managed containers with auto-scaling
   - Separate RDS and ElastiCache
   - High availability
   - Production-grade (~$200+/month)

3. **Low-Cost Mode** (Development/staging)
   - Minimal resources
   - Single availability zone
   - ~$20-30/month

## Quick Deployment (All-in-One)

### 1. Clone and Setup

```bash
cd infra/
cp Pulumi.dev.yaml Pulumi.prod.yaml
```

### 2. Configure Pulumi Stack

```yaml
# Pulumi.prod.yaml
config:
  aws:region: us-east-1
  rex-backend-infra:environment: production
  rex-backend-infra:deploymentMode: allinone
  rex-backend-infra:domain: yourdomain.com
  rex-backend-infra:certificateArn: arn:aws:acm:region:account:certificate/id
```

### 3. Set Secrets

```bash
# SuperTokens API Key
pulumi config set --secret supertokensApiKey "your-secure-api-key-here"

# Database Password
pulumi config set --secret dbPassword "your-secure-db-password"

# JWT Secret
pulumi config set --secret jwtSecret "your-jwt-secret-key"
```

### 4. Deploy

```bash
# Preview changes
pulumi preview

# Deploy infrastructure
pulumi up
```

### 5. Build and Push Docker Images

```bash
# Build multi-architecture images
./scripts/build-and-push-multiarch.sh

# Or standard build
./scripts/build-and-push.sh
```

### 6. Deploy Application

```bash
# For all-in-one
./scripts/allinone-deploy.sh

# Update existing deployment
./scripts/allinone-update.sh
```

## Detailed Setup

### Step 1: Domain and SSL Certificate

#### Request ACM Certificate

```bash
# Via AWS Console
1. Go to AWS Certificate Manager
2. Request a public certificate
3. Add domain: yourdomain.com
4. Add subdomain: *.yourdomain.com
5. Choose DNS validation
6. Add CNAME records to your DNS provider
7. Wait for certificate to be issued
8. Copy the ARN
```

#### Configure DNS

```
# DNS Records
A     yourdomain.com          → ALB/EC2 IP
CNAME api.yourdomain.com      → ALB DNS
CNAME app.yourdomain.com      → ALB DNS
```

### Step 2: Environment Configuration

Create production environment file:

```bash
# infra/.env.production
APP_ENV=production
APP_PORT=8080

# Database (will be created by Pulumi)
DB_HOST=<rds-endpoint>
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<from-secrets-manager>
DB_NAME=utm_backend
DB_SSL_MODE=require

# SuperTokens
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=<from-secrets-manager>

# Redis
REDIS_HOST=<elasticache-endpoint>
REDIS_PORT=6379

# JWT
JWT_SECRET_KEY=<from-secrets-manager>

# CORS
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Step 3: Infrastructure Deployment

#### All-in-One EC2 Mode

```typescript
// infra/main.go - Key settings
const config = {
  deploymentMode: 'allinone',
  instanceType: 't3.medium',  // 2 vCPU, 4 GB RAM
  volumeSize: 50,  // GB
  enableHttps: true,
  domain: 'yourdomain.com'
};
```

Deploy:

```bash
cd infra/
pulumi stack init prod
pulumi config set aws:region us-east-1
pulumi config set rex-backend-infra:deploymentMode allinone
pulumi config set rex-backend-infra:domain yourdomain.com
pulumi config set --secret supertokensApiKey "your-key"
pulumi up
```

#### ECS with ALB Mode

```typescript
const config = {
  deploymentMode: 'ecs',
  apiTaskCount: 2,
  workerTaskCount: 1,
  apiCpu: 512,
  apiMemory: 1024,
  enableAutoScaling: true,
  minCapacity: 2,
  maxCapacity: 10
};
```

Deploy:

```bash
pulumi config set rex-backend-infra:deploymentMode ecs
pulumi up
```

### Step 4: Database Setup

#### Run Migrations

```bash
# SSH into EC2 (all-in-one)
ssh -i ~/.ssh/your-key.pem ec2-user@<ec2-ip>

# Or use ECS Exec (ECS mode)
aws ecs execute-command \
  --cluster utm-cluster \
  --task <task-id> \
  --container api \
  --interactive \
  --command "/bin/sh"

# Run migrations
cd /app
./migrate up
```

#### Bootstrap First Platform Admin

```sql
-- Connect to database
psql -h <rds-endpoint> -U postgres -d utm_backend

-- Insert first admin (replace USER_ID with SuperTokens user ID)
INSERT INTO platform_admins (id, user_id, notes, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'YOUR_SUPERTOKENS_USER_ID',
  'Bootstrap admin - production',
  NOW(),
  NOW()
);
```

### Step 5: Application Configuration

#### SuperTokens Production Settings

```javascript
// Backend configuration
SuperTokens.init({
  appInfo: {
    appName: "Rex",
    apiDomain: "https://api.yourdomain.com",
    websiteDomain: "https://app.yourdomain.com",
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  supertokens: {
    connectionURI: process.env.SUPERTOKENS_CONNECTION_URI,
    apiKey: process.env.SUPERTOKENS_API_KEY
  },
  recipeList: [
    Session.init({
      cookieSecure: true,  // HTTPS only
      cookieSameSite: "lax",
      cookieDomain: ".yourdomain.com",  // Subdomain support
      sessionExpiredStatusCode: 401
    })
  ]
});
```

#### Frontend Production Config

```javascript
// Frontend SuperTokens init
SuperTokens.init({
  appInfo: {
    appName: "Rex",
    apiDomain: "https://api.yourdomain.com",
    websiteDomain: "https://app.yourdomain.com",
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    ThirdParty.init({
      signInAndUpFeature: {
        providers: [
          ThirdParty.Google.init({
            clientId: "YOUR_GOOGLE_CLIENT_ID"
          })
        ]
      }
    }),
    Session.init()
  ]
});
```

### Step 6: Monitoring and Logging

#### CloudWatch Logs

Logs are automatically sent to CloudWatch:

```bash
# View API logs
aws logs tail /aws/ecs/utm-api --follow

# View worker logs
aws logs tail /aws/ecs/utm-worker --follow

# Filter errors
aws logs filter-pattern '{ $.level = "error" }' /aws/ecs/utm-api
```

#### Metrics and Alarms

Create CloudWatch alarms:

```bash
# API health check alarm
aws cloudwatch put-metric-alarm \
  --alarm-name utm-api-health \
  --alarm-description "API health check failed" \
  --metric-name UnhealthyHostCount \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 300 \
  --threshold 1 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2

# Database connections
aws cloudwatch put-metric-alarm \
  --alarm-name utm-db-connections \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold
```

## Security Hardening

### 1. Database Security

```hcl
# RDS security group
resource "aws_security_group" "rds" {
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    security_groups = [aws_security_group.app.id]  # Only from app
  }
}
```

**Enable encryption**:
- Storage encryption at rest
- SSL/TLS for connections
- Automated backups

### 2. Secrets Management

Use AWS Secrets Manager:

```bash
# Store secret
aws secretsmanager create-secret \
  --name utm/production/supertokens-api-key \
  --secret-string "your-secret-key"

# Rotate secret
aws secretsmanager rotate-secret \
  --secret-id utm/production/db-password \
  --rotation-lambda-arn <lambda-arn>
```

### 3. Network Security

```hcl
# Application security group
resource "aws_security_group" "app" {
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # HTTPS from anywhere
  }
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # HTTP (redirect to HTTPS)
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

### 4. IAM Roles

Minimal permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:utm/*"
    }
  ]
}
```

### 5. Rate Limiting

Configure at ALB:

```hcl
resource "aws_lb_listener_rule" "rate_limit" {
  listener_arn = aws_lb_listener.https.arn
  
  action {
    type = "fixed-response"
    fixed_response {
      status_code  = 429
      content_type = "application/json"
      message_body = "{\"error\": \"Rate limit exceeded\"}"
    }
  }
  
  condition {
    path_pattern {
      values = ["/api/*"]
    }
  }
}
```

## Backup and Recovery

### Database Backups

```bash
# Enable automated backups
aws rds modify-db-instance \
  --db-instance-identifier utm-production \
  --backup-retention-period 7 \
  --preferred-backup-window "03:00-04:00"

# Manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier utm-production \
  --db-snapshot-identifier utm-manual-$(date +%Y%m%d)
```

### Application Backups

```bash
# Backup volumes (all-in-one)
aws ec2 create-snapshot \
  --volume-id vol-xxxxx \
  --description "UTM backup $(date +%Y%m%d)"
```

### Restore Procedure

```bash
# 1. Create new RDS instance from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier utm-restore \
  --db-snapshot-identifier utm-snapshot-xxxxx

# 2. Update application config to point to new database
# 3. Run migrations if needed
# 4. Verify data integrity
# 5. Switch DNS to new instance
```

## Scaling

### Horizontal Scaling (ECS)

```bash
# Update desired count
aws ecs update-service \
  --cluster utm-cluster \
  --service utm-api \
  --desired-count 4
```

### Vertical Scaling (All-in-One)

```bash
# 1. Stop instance
aws ec2 stop-instances --instance-ids i-xxxxx

# 2. Change instance type
aws ec2 modify-instance-attribute \
  --instance-id i-xxxxx \
  --instance-type t3.large

# 3. Start instance
aws ec2 start-instances --instance-ids i-xxxxx
```

### Database Scaling

```bash
# Increase instance size
aws rds modify-db-instance \
  --db-instance-identifier utm-production \
  --db-instance-class db.t3.large \
  --apply-immediately
```

## Troubleshooting

### Application Not Starting

```bash
# Check logs
docker logs utm-api
docker logs utm-worker

# Check environment variables
docker exec utm-api env | grep DB_HOST

# Test database connection
docker exec utm-api psql -h $DB_HOST -U $DB_USER -d $DB_NAME
```

### SSL Certificate Issues

```bash
# Verify certificate
openssl s_client -connect api.yourdomain.com:443

# Check ALB listener
aws elbv2 describe-listeners --load-balancer-arn <alb-arn>
```

### Database Connection Issues

```bash
# Test from EC2/ECS
telnet <rds-endpoint> 5432

# Check security groups
aws ec2 describe-security-groups --group-ids sg-xxxxx
```

### High CPU/Memory Usage

```bash
# Check metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=utm-api \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 3600 \
  --statistics Average

# Scale up
aws ecs update-service \
  --cluster utm-cluster \
  --service utm-api \
  --desired-count 4
```

## Cost Optimization

### All-in-One Mode (~$30-50/month)

- EC2 t3.medium: ~$30/month
- EBS volume (50 GB): ~$5/month
- Data transfer: ~$5-15/month

### ECS Mode (~$200+/month)

- ALB: ~$20/month
- ECS tasks (3x t3.small): ~$45/month
- RDS (db.t3.small): ~$30/month
- ElastiCache (cache.t3.micro): ~$15/month
- Data transfer: ~$20-50/month
- CloudWatch logs: ~$10/month
- Secrets Manager: ~$1/month

### Savings Tips

1. **Use Reserved Instances** (save 30-60%)
2. **Stop non-production environments** overnight
3. **Use S3 lifecycle policies** for log archival
4. **Enable AWS Savings Plans**
5. **Monitor with AWS Cost Explorer**

## Next Steps

- [Docker Deployment](/deployment/docker) - Docker compose setup
- [AWS Deployment](/deployment/aws) - Detailed AWS guide
- [Monitoring](/guides/monitoring) - Monitoring and alerting
- [Security](/guides/security) - Security best practices

