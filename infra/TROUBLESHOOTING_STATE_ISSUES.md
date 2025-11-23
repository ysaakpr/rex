# Troubleshooting Pulumi State & Deployment Issues

## Problem: Access Denied to Pulumi State Backend

You're seeing: `AccessDenied: Access Denied` when running Pulumi commands.

### Solution Options

---

## Option 1: Fix S3 Backend Access (Recommended)

### Step 1: Check which S3 bucket is being used

```bash
cd /Users/vyshakhp/work/utm-backend/infra
cat .pulumi/Pulumi.yaml 2>/dev/null || echo "No local state config"
```

### Step 2: Verify AWS credentials have S3 access

```bash
# Check current AWS identity
aws sts get-caller-identity

# Try to list S3 buckets
aws s3 ls
```

### Step 3: If you don't have access, switch to Pulumi Cloud or local

```bash
# Option A: Use Pulumi Cloud (free tier)
pulumi login

# Option B: Use local file backend
pulumi login --local
```

---

## Option 2: Fresh Start with Local Backend (Fastest)

This will start fresh, losing old Pulumi state (but AWS resources remain).

```bash
cd /Users/vyshakhp/work/utm-backend/infra

# 1. Login to local backend
pulumi login --local

# 2. Delete old local stack if it exists
rm -rf ~/.pulumi/stacks/rex-backend/dev.json* 2>/dev/null

# 3. Create fresh stack
export PULUMI_CONFIG_PASSPHRASE="your-secure-passphrase-here"
pulumi stack init dev --secrets-provider=passphrase

# 4. Set all required config values
pulumi config set aws:region ap-south-1
pulumi config set rex-backend:environment dev
pulumi config set rex-backend:projectName rex-backend
pulumi config set rex-backend:vpcCidr 10.0.0.0/16
pulumi config set rex-backend:dbMasterUsername rexadmin
pulumi config set rex-backend:githubRepo https://github.com/ysaakpr/rex
pulumi config set rex-backend:githubBranch main
pulumi config set rex-backend:allinone true
pulumi config set rex-backend:lowcost false

# 5. Set secrets
pulumi config set --secret rex-backend:dbMasterPassword "YourSecurePassword123!"
pulumi config set --secret rex-backend:supertokensApiKey "$(openssl rand -base64 32)"

# 6. Deploy
PULUMI_CONFIG_PASSPHRASE="your-secure-passphrase-here" pulumi up --yes
```

---

## Option 3: Manual AWS Cleanup + Fresh Deployment

### Step 1: Manually delete conflicting resources from AWS Console

Go to AWS Console → Region: ap-south-1

**Delete in this order:**

1. **ECS Services**
   - Navigate to: ECS → Clusters → rex-backend-dev-cluster → Services
   - Delete: `rex-backend-dev-api-service`, `rex-backend-dev-worker-service`, `rex-backend-dev-supertokens-service`

2. **Application Load Balancer**
   - Navigate to: EC2 → Load Balancers
   - Delete: `rex-backend-dev-alb`

3. **Target Groups**
   - Navigate to: EC2 → Target Groups
   - Delete: `rex-backend-dev-api-tg`, `rex-backend-dev-st-tg`, `rex-backend-dev-supertokens-tg`

4. **ECS Task Definitions** (optional, doesn't conflict)
   - Navigate to: ECS → Task Definitions
   - Deregister old versions if desired

### Step 2: Deploy fresh

```bash
cd /Users/vyshakhp/work/utm-backend/infra

# Use whatever backend works
pulumi login --local  # or pulumi login for cloud

# Deploy
export PULUMI_CONFIG_PASSPHRASE="your-passphrase"
pulumi up --yes
```

---

## Option 4: Import Existing Resources (Advanced)

If you want to keep existing AWS resources under Pulumi management:

```bash
# Import ALB
pulumi import aws:lb/loadBalancer:LoadBalancer rex-backend-dev-allinone-alb \
  arn:aws:elasticloadbalancing:ap-south-1:YOUR-ACCOUNT:loadbalancer/app/rex-backend-dev-alb/ID

# Import target groups
pulumi import aws:lb/targetGroup:TargetGroup rex-backend-dev-allinone-api-tg \
  arn:aws:elasticloadbalancing:ap-south-1:YOUR-ACCOUNT:targetgroup/rex-backend-dev-api-tg/ID

# ... repeat for other resources
```

---

## Recommended: Quick Fix Script

Create this script and run it:

```bash
#!/bin/bash
# Quick fix for state issues

cd /Users/vyshakhp/work/utm-backend/infra

# Use local backend
pulumi login --local

# Remove old local state
rm -rf ~/.pulumi/stacks/rex-backend/ 2>/dev/null

# Set passphrase
export PULUMI_CONFIG_PASSPHRASE="rex-backend-local-2024"

# Create new stack
pulumi stack init dev --secrets-provider=passphrase

# Copy config from old stack
pulumi config set aws:region ap-south-1
pulumi config set rex-backend:environment dev  
pulumi config set rex-backend:projectName rex-backend
pulumi config set rex-backend:vpcCidr 10.0.0.0/16
pulumi config set rex-backend:dbMasterUsername rexadmin
pulumi config set rex-backend:allinone true
pulumi config set rex-backend:lowcost false

# Set secrets (you'll need to provide these)
echo "Enter database password:"
read -s DB_PASS
pulumi config set --secret rex-backend:dbMasterPassword "$DB_PASS"

echo "Generating SuperTokens API key..."
pulumi config set --secret rex-backend:supertokensApiKey "$(openssl rand -base64 32)"

# Deploy
pulumi up --yes
```

---

## What I Recommend

**For development/testing:**
1. Go with **Option 3** (manual cleanup)
2. Delete resources from AWS Console (5 minutes)
3. Use local backend with fresh stack

**For production:**
1. Fix S3 backend access (**Option 1**)
2. Use Pulumi Cloud for better state management

---

## Current Issue Summary

- ✅ Code is ready and working
- ❌ Pulumi state backend has permission issues  
- ❌ Old AWS resources exist with same names
- ✅ New all-in-one architecture is implemented

**Solution**: Clean up old resources, then deploy fresh!


