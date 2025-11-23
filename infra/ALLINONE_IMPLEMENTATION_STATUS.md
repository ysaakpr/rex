# All-in-One Implementation Status

## ✅ What's Been Created

### 1. Core Infrastructure Code
- **`ec2_allinone.go`** - Creates single EC2 instance with Docker Compose
  - Installs Docker + Docker Compose
  - Creates docker-compose.yml with all services
  - Pulls images from ECR
  - Auto-starts everything
  - Includes update script

### 2. ALB Integration
- **`alb_ec2.go`** - ALB target groups for EC2 instance
  - API target group (port 8080)
  - SuperTokens target group (port 3567)
  - Listener rules for routing

### 3. Documentation
- **`LOWCOST_ALLINONE.md`** - Complete guide
  - Architecture diagrams
  - Cost comparison (75-80% savings!)
  - Deployment instructions
  - Management and troubleshooting

## 🔨 What Needs Integration

To complete the implementation, `main.go` needs to be updated:

### Option 1: Separate Mode (Recommended)
Add a new config flag `allinone` separate from `lowcost`:

```bash
pulumi config set rex-backend:allinone true
```

Benefits:
- Clean separation
- `lowcost` = DB on EC2, Fargate for apps
- `allinone` = Everything on one EC2
- Easy to maintain

### Option 2: Extend lowcost Mode
Use the existing `lowcost` flag to enable all-in-one:

```bash
pulumi config set rex-backend:lowcost true
```

Benefits:
- Simpler user experience
- One flag for "cheap mode"
- Less configuration

## 🎯 Recommended Approach

**I recommend Option 1 (separate `allinone` flag)** because:

1. **Flexibility**: Users can choose:
   - Standard: RDS + ElastiCache + Fargate
   - Lowcost: EC2 DB + Fargate
   - Allinone: Everything on one EC2

2. **Clear Intent**: Flag name clearly indicates what it does

3. **Migration Path**: Easy to move between modes

## 📋 Integration Checklist

To complete this:

- [ ] Add `allinone` config to `Pulumi.yaml`
- [ ] Update `main.go` to handle `allinone` mode
- [ ] Skip ECS/Fargate creation when `allinone=true`
- [ ] Use simpler ALB configuration
- [ ] Test end-to-end deployment
- [ ] Update root README

## 🚀 Quick Start (Once Integrated)

```bash
cd infra

# Configure all-in-one mode
pulumi config set rex-backend:allinone true

# Set secrets
pulumi config set --secret rex-backend:dbMasterPassword "your-password"
pulumi config set --secret rex-backend:supertokensApiKey "your-api-key"

# Deploy
pulumi up

# Build and push images
./scripts/build-and-push.sh

# Access instance
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID
```

## 💡 Implementation Notes

### Key Changes Needed in main.go

```go
// Add config
allInOne := cfg.GetBool("allinone")

if allInOne {
    // Create all-in-one EC2 instance
    allInOneRes, err := createAllInOneEC2(ctx, projectName, environment, 
        network, securityGroups.ALBSG, dbMasterUsername, dbMasterPassword,
        supertokensApiKey, repositories, tags)
    
    // Skip ECS, Fargate, separate DB setup
    // Configure ALB to target EC2 instance
    err = createEC2TargetGroups(ctx, projectName, environment, network,
        alb.LoadBalancer, allInOneRes.Instance, tags)
    
} else if lowCost {
    // Existing lowcost mode (EC2 DB + Fargate)
    ...
} else {
    // Standard mode (RDS + ElastiCache + Fargate)
    ...
}
```

### Security Groups
- EC2 security group allows ALB traffic on ports 8080, 3567
- ALB security group unchanged

### Health Checks
- API: GET /health on port 8080
- SuperTokens: GET /hello on port 3567

### Monitoring
- CloudWatch agent installed on EC2
- Docker logs rotated automatically
- Can stream logs to CloudWatch Logs

## 🎨 Architecture Comparison

### Standard Mode (~$100/mo)
```
ALB → Fargate (API, Worker, SuperTokens)
       ↓
      RDS Aurora + ElastiCache
```

### Lowcost Mode (~$50/mo)
```
ALB → Fargate (API, Worker, SuperTokens)
       ↓
      EC2 Spot (PostgreSQL, Redis)
```

### Allinone Mode (~$12/mo)
```
ALB → EC2 Spot (Docker Compose)
       ├─ API container
       ├─ Worker container
       ├─ SuperTokens container
       ├─ PostgreSQL container
       └─ Redis container
```

## ⚡ Performance

**t3a.medium Capacity:**
- 2 vCPU
- 4 GB RAM

**Expected Load Handling:**
- ~50-100 req/sec
- ~500 concurrent connections
- Good for dev/test/small prod

**If you need more:**
- t3a.large: 2 vCPU, 8 GB RAM (~$24/mo)
- t3a.xlarge: 4 vCPU, 16 GB RAM (~$48/mo)

## 📝 Next Steps

Would you like me to:

1. **Complete the integration** - Update `main.go` to support all-in-one mode
2. **Test deployment** - Deploy and verify everything works
3. **Create migration scripts** - Scripts to move between modes
4. **Add monitoring** - CloudWatch dashboards and alarms

Let me know and I'll continue with the implementation!

---

**Status**: Infrastructure code ready, needs integration into main.go
**Estimated Time**: 30-60 minutes to complete integration
**Risk**: Low (clean separation from existing code)

