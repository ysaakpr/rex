# AWS Amplify Frontend Migration - Summary

**Date**: November 23, 2025  
**Status**: ✅ Complete and Ready for Deployment

## Executive Summary

Your infrastructure has been successfully updated to use **AWS Amplify** for frontend deployment instead of ECS Fargate containers. Additionally, I've confirmed that your backend services are already using **fully managed AWS Fargate** - no self-managed EC2 instances!

## Key Findings

### ✅ Backend Already Fully Managed (Fargate)

Your original infrastructure **was already using fully managed Fargate**! All services are configured with:
- `LaunchType: "FARGATE"`
- `RequiresCompatibilities: ["FARGATE"]`
- No EC2 instances or capacity providers

**Backend Services on Fargate:**
- API Service: 2 tasks
- Worker Service: 1 task  
- SuperTokens: 1 task

### ✅ Frontend Migrated to AWS Amplify

Frontend has been moved from ECS Fargate to AWS Amplify for better efficiency.

## Changes Made

### 1. New Files Created

**`infra/amplify.go`** (119 lines)
- Complete AWS Amplify configuration
- GitHub repository integration
- Automatic build specification for Vite
- Environment variable management
- SPA routing rules

### 2. Files Modified

| File | Changes | Impact |
|------|---------|--------|
| `infra/main.go` | Added Amplify app creation, GitHub config | New config parameters required |
| `infra/ecs_services.go` | Removed frontend ECS service | Frontend no longer in ECS |
| `infra/alb.go` | Removed frontend target group | ALB only handles backend |
| `infra/ecr.go` | Removed frontend ECR repository | No frontend Docker images needed |
| `infra/logs.go` | Removed frontend CloudWatch logs | Logs now in Amplify Console |
| `infra/Pulumi.dev.yaml` | Added GitHub configuration | New config parameters |
| `infra/README.md` | Comprehensive updates | New deployment instructions |
| `infra/security_groups.go` | API compatibility fixes | Fixed deprecated fields |
| `infra/networking.go` | API compatibility fixes | Fixed import statements |
| `infra/redis.go` | API compatibility fixes | Fixed field names |

### 3. Documentation Created

**`docs/changedoc/16-AMPLIFY_FRONTEND_MIGRATION.md`**
- Complete migration guide
- Architecture diagrams
- Testing procedures
- Rollback plans
- Cost analysis

**`docs/changedoc/README.md`** (Updated)
- Added new documentation entry
- Updated reading order
- Added milestone

## Configuration Required

Before deployment, you need to set these Pulumi config values:

```bash
cd infra

# Required: Your GitHub repository URL
pulumi config set rex-backend:githubRepo "https://github.com/yourusername/rex-backend"

# Required: Branch to deploy
pulumi config set rex-backend:githubBranch "main"

# Optional: GitHub token for better rate limits (not required for public repos)
pulumi config set rex-backend:githubToken "ghp_your_token_here"
```

## Architecture Comparison

### Before
```
ALB → Frontend (ECS Fargate) + API (ECS Fargate) + Worker (ECS Fargate)
```
- 4 ECS Fargate tasks total
- All services behind single ALB
- Manual Docker builds and deployments for frontend

### After
```
Amplify (GitHub → CDN) ← Frontend
ALB → API (ECS Fargate) + Worker (ECS Fargate) + SuperTokens (ECS Fargate)
```
- 3 ECS Fargate tasks (backend only)
- Frontend automatically deployed from GitHub
- No frontend Docker builds needed

## Benefits

### 💰 Cost Savings
- **Before**: ~$136-176/month
- **After**: ~$126-161/month
- **Savings**: ~$10-15/month (frontend hosting)

### 🚀 Developer Experience
- **Automatic Deployments**: Push to GitHub → Auto build → Auto deploy
- **No Docker for Frontend**: Simplified workflow
- **Build Logs**: Clear logs in Amplify Console
- **Easy Rollback**: One-click rollback in console

### ⚡ Performance
- **Global CDN**: Faster load times worldwide
- **Auto HTTPS**: Free SSL certificates
- **Zero Cold Starts**: No ECS task startup delays
- **Edge Caching**: Static assets cached globally

### 🛠️ Operations
- **Fewer Services**: One less ECS service to manage
- **Auto Scaling**: Amplify handles traffic automatically
- **Zero Downtime**: Built-in blue-green deployments

## Deployment Steps

### 1. Configure Pulumi (Required)

```bash
cd /Users/vyshakhp/work/utm-backend/infra

# Set your GitHub repository
pulumi config set rex-backend:githubRepo "https://github.com/yourusername/rex-backend"

# Set branch (default: main)
pulumi config set rex-backend:githubBranch "main"
```

### 2. Verify Configuration

```bash
pulumi config
```

Should show:
```
KEY                              VALUE
aws:region                       us-east-1
rex-backend:dbMasterPassword     [secret]
rex-backend:environment          dev
rex-backend:githubBranch         main
rex-backend:githubRepo           https://github.com/yourusername/rex-backend
rex-backend:projectName          rex-backend
rex-backend:supertokensApiKey    [secret]
rex-backend:vpcCidr              10.0.0.0/16
```

### 3. Preview Changes

```bash
pulumi preview
```

This will show:
- ✅ Resources to create (Amplify app, branch)
- ❌ Resources to delete (frontend ECS service, frontend target group, etc.)
- 🔄 Resources to modify (ALB listener default action)

### 4. Deploy

```bash
pulumi up
```

Review and confirm the changes. Deployment takes ~10-15 minutes.

### 5. Get Frontend URL

```bash
pulumi stack output frontendUrl
```

Output will be something like:
```
https://main.d1234abcdefg.amplifyapp.com
```

### 6. Test

```bash
# Open frontend
open $(pulumi stack output frontendUrl)

# Test API
curl http://$(pulumi stack output albDnsName)/api/health
```

## Future Deployments

### Backend Updates (No Change)
```bash
# Build and push to ECR
docker build -f Dockerfile.prod --target api -t rex-backend-api:latest .
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/rex-backend-dev-api:latest

# Force new deployment
aws ecs update-service \
  --cluster rex-backend-dev-cluster \
  --service rex-backend-dev-api \
  --force-new-deployment
```

### Frontend Updates (NEW - Automatic!)
```bash
# Simply push to GitHub!
git add frontend/
git commit -m "Update frontend"
git push origin main

# Amplify automatically builds and deploys
# No manual steps required!
```

## Monitoring

### Backend (CloudWatch)
```bash
# API logs
aws logs tail /ecs/rex-backend-dev-api --follow

# Worker logs
aws logs tail /ecs/rex-backend-dev-worker --follow
```

### Frontend (Amplify Console)
- Go to AWS Console → Amplify → Your App
- View build logs, deployment history, and access logs

## Rollback Plan

If anything goes wrong:

### Option 1: Pulumi Rollback
```bash
cd infra
pulumi stack history
pulumi stack export --version <previous-version> > previous.json
pulumi stack import --file previous.json
pulumi up
```

### Option 2: Git Revert
```bash
cd infra
git log --oneline
git checkout <previous-commit> -- .
pulumi up
```

### Option 3: Manual Fix
- Check Amplify Console for build errors
- Verify environment variables are correct
- Check GitHub repository is accessible

## Testing Checklist

After deployment:

- [ ] Frontend URL is accessible
- [ ] Frontend loads without errors
- [ ] Can sign up / log in
- [ ] API calls work from frontend
- [ ] Backend services running
- [ ] Database connections working
- [ ] Redis connections working

## Cost Breakdown (Updated)

### Monthly Costs (Development)

| Service | Before | After | Savings |
|---------|--------|-------|---------|
| Aurora Serverless v2 | $30-50 | $30-50 | - |
| ElastiCache Redis | $12 | $12 | - |
| ECS Fargate | $40-60 | $30-40 | -$10-20 |
| AWS Amplify | - | $0-5 | New |
| NAT Gateway | $33 | $33 | - |
| ALB | $16 | $16 | - |
| CloudWatch Logs | $5 | $5 | - |
| **Total** | **$136-176** | **$126-161** | **-$10-15** |

### Cost Optimization Tips
- Amplify free tier: 1000 build minutes/month
- Amplify free tier: 15 GB served/month
- Frontend hosting cheaper than ECS
- Fewer log storage costs

## Troubleshooting

### Build Failures in Amplify

**Check**:
1. Amplify Console → Your App → Build details
2. Verify `frontend/package.json` has `build` script
3. Check Node.js version compatibility
4. Verify `frontend/dist` is output directory

**Fix**:
```bash
# Manually trigger rebuild
aws amplify start-job \
  --app-id <app-id> \
  --branch-name main \
  --job-type RELEASE
```

### API Calls Failing

**Check**:
1. Browser console for CORS errors
2. Environment variables in Amplify
3. ALB security group rules
4. Backend service health

**Fix**:
```bash
# Check backend health
curl http://$(pulumi stack output albDnsName)/api/health

# Check Amplify environment variables
aws amplify get-app --app-id <app-id>
```

## Support Resources

- **Change Documentation**: `docs/changedoc/16-AMPLIFY_FRONTEND_MIGRATION.md`
- **Infrastructure README**: `infra/README.md`
- **AWS Amplify Docs**: https://docs.aws.amazon.com/amplify/
- **Pulumi AWS Docs**: https://www.pulumi.com/registry/packages/aws/

## Next Steps

1. **Update GitHub Repository URL** in Pulumi config
2. **Deploy infrastructure** with `pulumi up`
3. **Verify deployment** works end-to-end
4. **Update team documentation** with new deployment process
5. **Optional**: Set up custom domain for frontend in Amplify

## Questions?

Common questions answered:

**Q: Do I need to build Docker images for frontend anymore?**  
A: No! Amplify builds directly from your GitHub repository.

**Q: How do I deploy frontend updates?**  
A: Just push to GitHub. Amplify automatically builds and deploys.

**Q: Is the backend still on Fargate?**  
A: Yes! Backend continues to run on fully managed AWS Fargate.

**Q: What if I want to use a custom domain?**  
A: Configure it in Amplify Console → Domain management.

**Q: Can I preview changes before deploying?**  
A: Yes! Amplify supports preview environments for pull requests.

**Q: Is this more expensive?**  
A: No, it's actually ~$10-15/month cheaper!

## Summary

✅ **Confirmed**: Backend already uses fully managed Fargate (no EC2)  
✅ **Added**: AWS Amplify for frontend deployment  
✅ **Removed**: Frontend ECS service, ECR repo, and CloudWatch logs  
✅ **Updated**: ALB to handle only backend routes  
✅ **Fixed**: Pulumi AWS SDK v6 compatibility issues  
✅ **Documented**: Complete migration guide and testing procedures  
✅ **Ready**: Code compiles and ready for deployment  

Your infrastructure is now optimized with a fully managed, serverless architecture that's easier to maintain and more cost-effective!

---

**Need Help?**
- Review: `docs/changedoc/16-AMPLIFY_FRONTEND_MIGRATION.md` for detailed guide
- Check: `infra/README.md` for deployment instructions
- Contact: Your infrastructure team for support

