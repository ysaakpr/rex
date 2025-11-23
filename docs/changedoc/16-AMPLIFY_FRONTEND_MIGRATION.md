# Change Documentation #16: AWS Amplify Frontend Migration

**Date**: November 23, 2025  
**Author**: Infrastructure Team  
**Status**: Completed

## Purpose

Migrate the frontend deployment from ECS Fargate containers to AWS Amplify for improved developer experience, cost efficiency, and simplified deployment pipeline.

## Summary

The frontend React application has been migrated from containerized ECS Fargate deployment to AWS Amplify hosting. This change provides automatic CI/CD from GitHub, global CDN distribution, and reduces infrastructure costs while maintaining the fully managed Fargate approach for backend services.

## Key Changes

### 1. Infrastructure Changes

#### Added Files

**`infra/amplify.go`**
- New Amplify app configuration
- GitHub repository integration
- Automatic build configuration for Vite
- Environment variable management
- Branch deployment setup (main branch)
- Custom routing rules for SPA

#### Modified Files

**`infra/main.go`**
- Added GitHub repository configuration parameters
- Replaced frontend ECS service with Amplify app creation
- Updated exports to include Amplify URLs
- Added configuration for:
  - `githubRepo`: Repository URL
  - `githubBranch`: Branch to deploy (default: main)
  - `githubToken`: Optional GitHub token for private repos

**`infra/ecs_services.go`**
- Removed frontend ECS service creation
- Removed frontend task definition
- Updated struct to remove frontend-related fields
- Added comments explaining Amplify migration

**`infra/alb.go`**
- Removed frontend target group
- Changed default ALB action to fixed response
- ALB now only handles backend API and auth routes
- Updated routing to focus on `/api/*` and `/auth/*` paths

**`infra/ecr.go`**
- Removed frontend ECR repository
- Kept only API and Worker repositories
- Updated struct to remove frontend-related fields

**`infra/logs.go`**
- Removed frontend CloudWatch log group
- Frontend logs now accessible via Amplify Console
- Kept backend service log groups

**`infra/Pulumi.dev.yaml`**
- Added GitHub configuration parameters
- Added comments for required GitHub settings

**`infra/README.md`**
- Comprehensive updates for Amplify deployment
- Removed frontend Docker build instructions
- Added Amplify-specific deployment steps
- Updated cost estimates
- Updated monitoring and logging sections

### 2. Architecture Changes

#### Before (ECS Fargate for Everything)
```
┌─────────────────────────────────────────────┐
│         Application Load Balancer           │
│  - Frontend: /*                             │
│  - API: /api/*                              │
│  - Auth: /auth/*                            │
└─────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   ┌─────────┐ ┌─────────┐ ┌──────────┐
   │Frontend │ │   API   │ │SuperTokens│
   │  ECS    │ │   ECS   │ │   ECS    │
   └─────────┘ └─────────┘ └──────────┘
```

#### After (Amplify for Frontend)
```
┌──────────────────┐     ┌────────────────────┐
│  AWS Amplify     │     │Application Load    │
│  (Frontend CDN)  │     │Balancer (Backend)  │
│  - Global CDN    │     │  - API: /api/*     │
│  - Auto SSL      │     │  - Auth: /auth/*   │
└──────────────────┘     └────────────────────┘
         │                        │
         │                ┌───────┼────────┐
         │                ▼       ▼        ▼
    ┌────────┐        ┌──────┐ ┌──────┐ ┌────┐
    │ GitHub │        │ API  │ │Worker│ │ ST │
    │  Repo  │        │ ECS  │ │ ECS  │ │ECS │
    └────────┘        └──────┘ └──────┘ └────┘
```

### 3. Deployment Process Changes

#### Backend Services (No Change)
- Still use Docker + ECR + ECS Fargate
- Manual build and push to ECR
- Force new deployment via ECS

#### Frontend (New Process)
1. Push code to GitHub
2. Amplify automatically detects changes
3. Amplify builds the Vite app
4. Amplify deploys to global CDN
5. No manual intervention required

### 4. Configuration Requirements

#### New Pulumi Config Parameters

```bash
# Required: GitHub repository URL
pulumi config set rex-backend:githubRepo "https://github.com/yourusername/utm-backend"

# Required: Branch to deploy
pulumi config set rex-backend:githubBranch "main"

# Optional: GitHub token for better rate limits
pulumi config set rex-backend:githubToken "ghp_your_token_here"
```

#### Amplify Build Specification

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - cd frontend
        - npm ci
    build:
      commands:
        - npm run build
  artifacts:
    baseDirectory: frontend/dist
    files:
      - '**/*'
  cache:
    paths:
      - frontend/node_modules/**/*
```

### 5. Benefits

#### Cost Savings
- **Before**: Frontend on ECS Fargate = ~$10-15/month
- **After**: Frontend on Amplify = ~$0-5/month
- **Savings**: ~$10/month (67% reduction for frontend hosting)

#### Developer Experience
- **Automatic Deployments**: Push to GitHub triggers automatic build and deploy
- **No Docker Required**: No need to build/push frontend Docker images
- **Build Logs**: Clear build logs in Amplify Console
- **Instant Rollback**: Easy rollback to previous deployments

#### Performance
- **Global CDN**: Content delivered from edge locations worldwide
- **Automatic HTTPS**: SSL/TLS certificates managed automatically
- **Faster Load Times**: Static assets cached at edge
- **No Cold Starts**: Unlike ECS tasks, no startup delays

#### Infrastructure
- **Simplified**: One less ECS service to manage
- **Fully Managed**: AWS handles all infrastructure
- **Auto Scaling**: Amplify automatically scales with traffic
- **Zero Downtime**: Blue-green deployments built-in

### 6. Backend Architecture (Confirmed Fully Managed)

All backend services remain on **AWS Fargate** (fully managed):

```go
// All services explicitly use Fargate launch type
LaunchType: pulumi.String("FARGATE"),
RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
NetworkMode: pulumi.String("awsvpc"),
```

**Services on Fargate**:
- API Service: 2 tasks (512 CPU, 1024 MB)
- Worker Service: 1 task (512 CPU, 1024 MB)
- SuperTokens: 1 task (512 CPU, 1024 MB)

**No EC2 instances** are created or managed. This is a fully serverless architecture.

## Deployment Steps

### Initial Setup

1. **Update Pulumi Configuration**:
```bash
cd infra
pulumi config set rex-backend:githubRepo "https://github.com/yourusername/utm-backend"
pulumi config set rex-backend:githubBranch "main"
```

2. **Deploy Infrastructure**:
```bash
pulumi preview  # Review changes
pulumi up       # Deploy
```

3. **Get Frontend URL**:
```bash
pulumi stack output frontendUrl
# Output: https://main.xxxxx.amplifyapp.com
```

### Updating Frontend

Simply push to GitHub:
```bash
git add frontend/
git commit -m "Update frontend"
git push origin main
```

Amplify will automatically:
1. Detect the push
2. Run the build
3. Deploy the new version
4. Update the live site

### Monitoring Deployments

**AWS Console**:
- Go to AWS Console → Amplify → Your App
- View build status, logs, and deployment history

**CLI**:
```bash
# List apps
aws amplify list-apps

# Get app details
aws amplify get-app --app-id <app-id>

# List branches
aws amplify list-branches --app-id <app-id>
```

## Testing

### 1. Verify Infrastructure

```bash
# Check all outputs
pulumi stack output

# Should see:
# - frontendUrl: https://main.xxxxx.amplifyapp.com
# - apiUrl: http://xxx.elb.amazonaws.com/api
# - albDnsName: xxx.elb.amazonaws.com
```

### 2. Test Backend API

```bash
ALB_DNS=$(pulumi stack output albDnsName)
curl http://$ALB_DNS/api/health
# Expected: {"success": true, "message": "API is healthy"}
```

### 3. Test Frontend

```bash
FRONTEND_URL=$(pulumi stack output frontendUrl)
curl -I $FRONTEND_URL
# Expected: HTTP/2 200
```

Open in browser:
```bash
open $(pulumi stack output frontendUrl)
```

### 4. Test Frontend-Backend Integration

1. Open frontend URL in browser
2. Try to log in / sign up
3. Verify API calls work (check browser network tab)
4. Environment variables should be correctly set:
   - `VITE_API_URL`: Points to ALB `/api` path
   - `VITE_AUTH_URL`: Points to ALB `/auth` path

## Rollback Plan

### If Issues Occur During Migration

1. **Revert Pulumi Changes**:
```bash
cd infra
pulumi stack export --file backup.json
git checkout HEAD~1 -- .
pulumi up
```

2. **Or Use Previous Pulumi State**:
```bash
pulumi stack history
pulumi stack select <previous-version>
```

### If Frontend Not Working

1. **Check Amplify Build Logs**:
   - AWS Console → Amplify → Your App → Build details

2. **Common Issues**:
   - Wrong build command: Check `buildSpec` in `amplify.go`
   - Missing environment variables: Check `EnvironmentVariables` in `amplify.go`
   - Build directory wrong: Should be `frontend/dist` for Vite

3. **Manual Trigger Rebuild**:
```bash
aws amplify start-job --app-id <app-id> --branch-name main --job-type RELEASE
```

## Migration Checklist

- [x] Create `infra/amplify.go` with Amplify configuration
- [x] Update `infra/main.go` to create Amplify app
- [x] Remove frontend ECS service from `infra/ecs_services.go`
- [x] Update `infra/alb.go` to remove frontend target group
- [x] Remove frontend ECR repository from `infra/ecr.go`
- [x] Remove frontend log group from `infra/logs.go`
- [x] Update `infra/Pulumi.dev.yaml` with GitHub config
- [x] Update `infra/README.md` with new deployment instructions
- [x] Fix compilation errors in Go code
- [x] Verify code builds successfully
- [x] Create this change documentation

## Post-Migration Tasks

### Immediate
- [ ] Deploy infrastructure with `pulumi up`
- [ ] Verify frontend URL is accessible
- [ ] Test frontend-backend integration
- [ ] Update team documentation
- [ ] Update CI/CD pipelines if any

### Optional Enhancements
- [ ] Set up custom domain for frontend via Amplify
- [ ] Configure Amplify email notifications for builds
- [ ] Set up preview environments for pull requests
- [ ] Add performance monitoring to Amplify
- [ ] Configure Amplify access control (if needed)

## Troubleshooting

### Frontend Not Building

**Symptom**: Amplify build fails

**Solutions**:
1. Check build logs in Amplify Console
2. Verify `frontend/` directory structure is correct
3. Ensure `package.json` has `build` script
4. Check Node.js version compatibility

### API Calls Failing from Frontend

**Symptom**: Frontend loads but API calls fail

**Solutions**:
1. Check CORS settings on API
2. Verify `VITE_API_URL` environment variable
3. Check browser console for errors
4. Verify ALB security group allows traffic

### Frontend Shows 404

**Symptom**: Routes not working in SPA

**Solutions**:
1. Check custom rules in `amplify.go`
2. Verify Amplify rewrites are configured correctly
3. Check browser console for errors

## References

- [AWS Amplify Hosting Documentation](https://docs.aws.amazon.com/amplify/latest/userguide/welcome.html)
- [Pulumi AWS Amplify](https://www.pulumi.com/registry/packages/aws/api-docs/amplify/)
- [Vite Build Configuration](https://vitejs.dev/guide/build.html)
- [React Router with Amplify](https://docs.amplify.aws/guides/hosting/react-router/q/platform/js/)

## Conclusion

This migration successfully moves the frontend from ECS Fargate to AWS Amplify, providing:
- ✅ Cost savings (~$10/month)
- ✅ Simplified deployment (automatic from GitHub)
- ✅ Better performance (global CDN)
- ✅ Improved developer experience
- ✅ Maintained fully managed infrastructure for backend (Fargate)

The backend services continue to run on fully managed AWS Fargate with no EC2 instances to manage, maintaining a complete serverless architecture.

