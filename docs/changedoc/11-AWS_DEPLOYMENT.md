# Change Doc 11: AWS Deployment with Pulumi

**Date**: November 23, 2025  
**Status**: Complete  
**Type**: Infrastructure

## Purpose

Implement complete AWS infrastructure deployment for UTM Backend using Pulumi (Go), enabling production-ready deployment on AWS Fargate with Aurora RDS Serverless and ElastiCache Redis.

## Overview

Created a comprehensive infrastructure-as-code solution using Pulumi in Go to deploy the entire UTM Backend stack to AWS. The infrastructure uses modern, scalable AWS services with a focus on cost optimization and high availability.

## Changes

### 1. Infrastructure Code (`infra/`)

Created complete Pulumi project in Go with the following modules:

#### Core Files
- **`main.go`**: Main Pulumi program orchestrating all resources
- **`Pulumi.yaml`**: Project configuration
- **`Pulumi.dev.yaml`**: Development environment configuration
- **`go.mod`**: Go module dependencies
- **`.gitignore`**: Ignore file for infrastructure directory

#### Infrastructure Modules
- **`networking.go`**: VPC, subnets, NAT gateway, route tables
- **`security_groups.go`**: Security groups for ALB, ECS, RDS, Redis
- **`database.go`**: Aurora RDS Serverless v2 cluster with PostgreSQL
- **`redis.go`**: ElastiCache Redis for caching and job queues
- **`ecr.go`**: ECR repositories for container images
- **`ecs_cluster.go`**: ECS cluster configuration
- **`alb.go`**: Application Load Balancer with routing rules
- **`iam.go`**: IAM roles and policies for ECS tasks
- **`secrets.go`**: AWS Secrets Manager for sensitive configuration
- **`logs.go`**: CloudWatch log groups for all services
- **`ecs_services.go`**: ECS task definitions and services for:
  - API service (2 tasks, 512 CPU, 1024 MB)
  - Worker service (1 task, 512 CPU, 1024 MB)
  - Frontend service (2 tasks, 256 CPU, 512 MB)
  - SuperTokens service (1 task, 512 CPU, 1024 MB)
- **`migration_task.go`**: ECS task definition for database migrations

### 2. Production Dockerfiles

#### API & Worker (`Dockerfile.prod`)
- Multi-stage build with Alpine Linux
- Separate targets for API and Worker
- Includes migration binary
- Non-root user for security
- Optimized binary size

#### Frontend (`frontend/Dockerfile.prod`)
- Build stage with Vite production build
- Nginx Alpine for serving
- Custom nginx configuration
- Gzip compression enabled
- Security headers configured

#### Nginx Configuration (`frontend/nginx.conf`)
- SPA routing support
- Static asset caching
- Security headers
- Health check endpoint

### 3. Deployment Scripts (`infra/scripts/`)

Created automated scripts for common operations:

#### `setup-pulumi.sh`
- Interactive Pulumi setup
- S3 bucket creation for state
- Stack initialization
- Configuration and secrets setup
- Domain configuration (optional)

#### `build-and-push.sh`
- Build all Docker images
- Login to ECR
- Tag and push images to repositories
- Automated with Pulumi outputs

#### `run-migration.sh`
- Run database migrations as ECS task
- Wait for completion
- Check exit status
- Display logs on failure

#### `force-deploy.sh`
- Force new deployment of all services
- Pull latest container images
- Rolling update without downtime

#### `create-supertokens-db.sh`
- Create SuperTokens database in Aurora
- Support for docker and local psql
- Handles network access requirements

### 4. Documentation

#### `infra/README.md`
- Complete infrastructure documentation
- Architecture overview with diagram
- Prerequisites and setup instructions
- Deployment process
- Scaling and monitoring
- Troubleshooting guide
- Cost estimates

#### `docs/AWS_DEPLOYMENT_GUIDE.md`
- Step-by-step deployment guide
- Post-deployment configuration
- Custom domain setup
- Auto-scaling configuration
- Monitoring and maintenance
- Troubleshooting common issues

### 5. Architecture Highlights

#### Networking
- VPC: 10.0.0.0/16
- 2 Public subnets across 2 AZs
- 2 Private subnets across 2 AZs
- Internet Gateway for public access
- NAT Gateway for private subnet internet access
- Proper route table configuration

#### Database
- Aurora RDS Serverless v2 (PostgreSQL 15.4)
- Auto-scaling: 0.5 - 2.0 ACUs
- Two databases in single cluster:
  - `utm_backend` (main application)
  - `supertokens` (authentication)
- 7-day backup retention
- Automated backups

#### Cache & Queue
- ElastiCache Redis 7.0
- cache.t4g.micro for development
- At-rest encryption enabled
- Automated snapshots

#### Container Registry
- Separate ECR repositories for each service
- Automated image scanning
- Lifecycle policy (keep last 10 images)

#### Compute (ECS Fargate)
- Container Insights enabled
- Service discovery for internal communication
- Health checks for all services
- Auto-restart on failure
- Multiple tasks for high availability

#### Load Balancing
- Application Load Balancer
- Path-based routing:
  - `/api/*` → API service
  - `/auth/*` → SuperTokens service
  - `/*` → Frontend
- Health checks configured
- HTTPS support (optional)

#### Security
- AWS Secrets Manager for credentials
- IAM roles with least privilege
- Security groups with minimal access
- VPC isolation
- Private subnets for compute and data
- Encryption at rest and in transit

#### Monitoring
- CloudWatch Log Groups (7-day retention)
- Container Insights for metrics
- Task-level logging
- ALB access logs support

### 6. Key Features

1. **State Management**: S3 backend for Pulumi state with versioning and encryption

2. **Service Discovery**: AWS Cloud Map for internal service communication

3. **Secrets Management**: 
   - Database credentials in Secrets Manager
   - SuperTokens API key in Secrets Manager
   - Automatic rotation support

4. **High Availability**:
   - Multi-AZ deployment
   - Multiple task instances
   - Auto-recovery on failure

5. **Cost Optimization**:
   - Aurora Serverless auto-scaling
   - Smallest viable instance types for dev
   - 7-day log retention (dev)
   - NAT Gateway optimization

6. **Deployment Strategy**:
   - Rolling updates
   - Zero-downtime deployments
   - Automated migration tasks
   - Easy rollback with Pulumi

## Deployment Process

### Initial Setup
```bash
cd infra/scripts
./setup-pulumi.sh
```

### Deploy Infrastructure
```bash
cd infra
pulumi up
```

### Build and Push Images
```bash
cd infra/scripts
./build-and-push.sh
```

### Create SuperTokens Database
```bash
./create-supertokens-db.sh
```

### Run Migrations
```bash
./run-migration.sh
```

### Verify
```bash
ALB_DNS=$(pulumi stack output albDnsName)
curl http://$ALB_DNS/api/health
```

## Cost Estimates

### Development Environment
- **Total**: ~$150-200/month
- Aurora Serverless: $40-60
- ElastiCache: $12
- ECS Fargate: $50-70
- NAT Gateway: $33
- ALB: $18
- Other: $15-25

### Production Environment
- **Estimated**: $500-800/month
- Larger instances
- More tasks
- CloudFront CDN
- Enhanced monitoring

## Benefits

1. **Infrastructure as Code**: Version-controlled, repeatable deployments
2. **Scalability**: Easy to scale up/down based on demand
3. **High Availability**: Multi-AZ with automatic failover
4. **Security**: Best practices with AWS services
5. **Cost-Effective**: Serverless components reduce costs
6. **Monitoring**: Built-in observability with CloudWatch
7. **Maintainability**: Clear structure, well-documented

## Migration from Docker Compose

| Component | Docker Compose | AWS |
|-----------|----------------|-----|
| API | Local container | ECS Fargate (2 tasks) |
| Worker | Local container | ECS Fargate (1 task) |
| Frontend | Local container | ECS Fargate (2 tasks) |
| SuperTokens | Local container | ECS Fargate (1 task) |
| PostgreSQL | Local container | Aurora RDS Serverless v2 |
| Redis | Local container | ElastiCache Redis |
| Load Balancer | None | Application Load Balancer |
| Networking | Bridge network | VPC with public/private subnets |

## Testing

Deployment tested with:
- ✅ Infrastructure creation (all resources)
- ✅ Service deployment patterns
- ✅ Database connectivity
- ✅ Service discovery
- ✅ Load balancer routing
- ✅ Migration task execution
- ✅ Log collection
- ✅ Secrets management

## Future Enhancements

1. **CI/CD Pipeline**:
   - GitHub Actions workflow
   - Automated testing and deployment
   - Multi-environment support

2. **Advanced Monitoring**:
   - CloudWatch alarms
   - SNS notifications
   - Custom dashboards

3. **CDN**:
   - CloudFront for frontend
   - Improved global performance

4. **Auto-Scaling**:
   - Application Auto Scaling policies
   - CPU/Memory-based scaling
   - Schedule-based scaling

5. **Disaster Recovery**:
   - Automated backups
   - Cross-region replication
   - Restore procedures

6. **Security Enhancements**:
   - AWS WAF for ALB
   - VPC Flow Logs
   - CloudTrail for audit
   - GuardDuty for threat detection

## References

- Infrastructure code: `infra/`
- Deployment guide: `docs/AWS_DEPLOYMENT_GUIDE.md`
- Production Dockerfiles: `Dockerfile.prod`, `frontend/Dockerfile.prod`
- Deployment scripts: `infra/scripts/`

## Notes

- All scripts are tested and production-ready
- Pulumi state stored securely in S3
- Supports multiple environments (dev, staging, production)
- Easy to customize for specific requirements
- Cost-optimized for development, ready for production scaling

---

**Next Steps**: Follow the AWS Deployment Guide to deploy your first environment!

