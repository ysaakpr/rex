# All-in-One Low-Cost Mode

## 🎯 New Simplified Architecture

Instead of using Fargate with separate EC2 for database, everything runs on **one single EC2 Spot instance** using Docker Compose.

### What Runs on the Instance

**Single t3a.medium EC2 Spot Instance ($6-12/month):**
```
┌─────────────────────────────────────────────┐
│  EC2 Spot Instance (t3a.medium)            │
│  2 vCPU, 4 GB RAM + Public IP              │
│                                            │
│  Docker Compose Stack:                     │
│  ├─ PostgreSQL (container)                │
│  │  ├─ rex_backend DB                     │
│  │  └─ supertokens DB                     │
│  ├─ Redis (container)                     │
│  ├─ SuperTokens (port 3567)               │
│  ├─ API (port 8080 from ECR)              │
│  └─ Worker (container from ECR)           │
└─────────────────────────────────────────────┘
         ▲
         │ Direct Port Access
         │
    Internet (No ALB needed!)
    - API: http://PUBLIC_DNS:8080
    - SuperTokens: http://PUBLIC_DNS:3567
```

## 💰 Cost Comparison

| Component | Old (Fargate+ALB) | New (All-in-One) | Savings |
|-----------|--------------|------------------|---------|
| Database | Included in EC2 | Included | - |
| Redis | Included in EC2 | Included | - |
| API | Fargate (~$15/mo) | Docker container | ~$15 |
| Worker | Fargate (~$15/mo) | Docker container | ~$15 |
| SuperTokens | Fargate (~$15/mo) | Docker container | ~$15 |
| ALB | ~$16/mo | Not needed! | ~$16 |
| NAT Gateway | ~$32/mo | Not needed! | ~$32 |
| EC2 Instance | t3a.small (~$6) | t3a.medium (~$10) | -$4 |
| **Total** | **~$99-114/mo** | **~$10/mo** | **~$89-104/mo saved (90%+)!** |

**Key Architecture Difference:**
- **Old Setup**: Private subnet → NAT Gateway required for outbound internet
- **New Setup**: Public subnet with public IP → Internet Gateway (free!)

**Why No NAT Gateway Needed:**
- Instance is in **public subnet** with **public IP**
- Uses **Internet Gateway** for outbound traffic (included with VPC, no cost)
- NAT Gateway only needed for instances in private subnets

## ✅ Benefits

1. **Much Simpler**
   - One instance to manage
   - Standard Docker Compose (familiar)
   - Easy to debug and monitor
   - Direct access to all logs
   - No ALB or NAT Gateway complexity

2. **Much Cheaper**
   - Single spot instance vs multiple Fargate tasks
   - No ALB: Save $16/month
   - No NAT Gateway: Save $32/month
   - **90%+ cost savings** vs standard setup
   - Perfect for dev/test/staging

3. **Direct Internet Access**
   - Public IP with Internet Gateway (free!)
   - No NAT Gateway charges
   - Simple security group rules
   - Direct port access (no ALB routing)

4. **Faster Development**
   - Quick restarts (docker-compose restart)
   - Easy to update (./update.sh script)
   - Direct shell access via SSM
   - Can run migrations easily

5. **Better Resource Utilization**
   - Shared CPU and memory
   - Docker handles resource allocation
   - No cold starts
   - Better for low-traffic scenarios

## 🌐 Networking Architecture Explained

### Why No NAT Gateway Needed?

**NAT Gateway** is ONLY required for instances in **private subnets** that need outbound internet access.

**Our All-in-One Setup:**
```
✅ Instance in PUBLIC subnet
✅ Has PUBLIC IP address
✅ Uses Internet Gateway (FREE)
✅ Direct inbound/outbound access
❌ NO NAT Gateway needed
```

**Standard/Low-Cost Fargate Setup:**
```
❌ Instances in PRIVATE subnet
❌ Need NAT Gateway for outbound ($32/mo)
❌ Need ALB for inbound ($16/mo)
❌ More complex routing
```

### Cost Breakdown

| Service | All-in-One | Standard | Why? |
|---------|------------|----------|------|
| Internet Gateway | FREE | FREE | Included with VPC |
| NAT Gateway | Not needed | $32/mo | Only for private subnets |
| ALB | Not needed | $16/mo | Direct port access instead |
| Public IP | FREE | N/A | EIP allocation included |

**Total Networking Savings: ~$48/month**

## 📋 How It Works

### 1. Infrastructure Setup (Pulumi)

When `lowcost=true`:
- Creates EC2 Spot instance (t3a.medium)
- Installs Docker + Docker Compose
- Creates docker-compose.yml
- Pulls images from ECR
- Starts all services
- Configures ALB to target the instance

### 2. Docker Compose Stack

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: rexadmin
      POSTGRES_PASSWORD: (from Pulumi)
      POSTGRES_DB: rex_backend
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rexadmin"]
    
  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 512mb
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
    
  supertokens:
    image: registry.supertokens.io/supertokens/supertokens-postgresql:7.0
    depends_on:
      - postgres
    ports:
      - "3567:3567"
    
  api:
    image: (your-ecr-repo)/api:latest
    depends_on:
      - postgres
      - redis
      - supertokens
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
      SUPERTOKENS_CONNECTION_URI: http://supertokens:3567
    
  worker:
    image: (your-ecr-repo)/worker:latest
    depends_on:
      - postgres
      - redis
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
```

### 3. Network Architecture

```
Internet
    ↓
Internet Gateway (FREE - included with VPC)
    ↓
Public Subnet (10.0.0.0/24)
    ↓
EC2 Instance with Public IP
    ├─ Port 8080 → API
    ├─ Port 3567 → SuperTokens
    └─ Port 80 → Frontend (optional)
```

**Key Benefits:**
- No ALB needed (~$16/mo saved)
- No NAT Gateway needed (~$32/mo saved) 
- Direct internet access via Internet Gateway (free)
- Simple security group rules control access

## 🚀 Deployment

### Enable All-in-One Mode

```bash
cd infra
pulumi config set rex-backend:lowcost true
pulumi up
```

This will:
1. ✅ Create EC2 spot instance
2. ✅ Install Docker + Docker Compose
3. ✅ Create docker-compose.yml with all services
4. ✅ Pull images from ECR
5. ✅ Start all containers
6. ✅ Configure ALB to point to instance

### Build and Push Images

```bash
# Build and push to ECR
./infra/scripts/build-and-push.sh
```

### Update Running Services

SSH into the instance:
```bash
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID

# On the instance
cd /app
./update.sh  # Pulls latest images and restarts
```

Or update via docker-compose:
```bash
# Pull latest images
docker-compose pull api worker

# Restart services
docker-compose up -d api worker

# View logs
docker-compose logs -f api
```

## 👤 Platform Admin Access

### Automatic Admin Initialization

The all-in-one deployment automatically creates a default platform administrator:

**Default Credentials**:
- Email: `admin@platform.local`
- Password: `admin`

⚠️ **SECURITY**: Change this password immediately after first login!

### Sign In

```bash
ALB_DNS=$(pulumi stack output albDnsName)

curl -X POST http://$ALB_DNS/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "admin@platform.local"},
      {"id": "password", "value": "admin"}
    ]
  }' \
  -c cookies.txt
```

### Verify Admin Status

```bash
# Check platform_admins table
docker exec rex-postgres psql -U rexadmin -d rex_backend \
  -c "SELECT * FROM platform_admins;"

# Check initialization status
ls -la /app/.admin-initialized
```

**📖 Full documentation**: [ADMIN_INITIALIZATION.md](./ADMIN_INITIALIZATION.md)

## 🔧 Management

### Access the Instance

```bash
# Via SSM (no SSH key needed)
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID
```

### View Logs

```bash
# All services
cd /app
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f supertokens

# Real-time
docker-compose logs --tail=100 -f api
```

### Restart Services

```bash
cd /app

# Restart all
docker-compose restart

# Restart specific service
docker-compose restart api

# Stop and start
docker-compose down
docker-compose up -d
```

### Run Migrations

```bash
cd /app

# Run migration container
docker-compose run --rm api /app/migrate up
```

### Check Status

```bash
cd /app

# Service status
docker-compose ps

# Resource usage
docker stats

# Disk usage
docker system df
```

### Update Images

```bash
cd /app

# Use the update script
./update.sh

# Or manually
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin (your-ecr-repo)
docker-compose pull api worker
docker-compose up -d api worker
docker image prune -f
```

## 📊 Monitoring

### Container Health

```bash
docker-compose ps
# Shows health status of all containers
```

### Resource Usage

```bash
docker stats
# Shows CPU, memory, network usage per container
```

### Application Health

```bash
# API health check
curl http://localhost:8080/health

# SuperTokens health check
curl http://localhost:3567/hello
```

### CloudWatch Logs

Docker logs are automatically rotated and can be sent to CloudWatch Logs using the CloudWatch agent.

## 🐛 Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs service-name

# Check if image exists
docker images | grep rex-backend

# Rebuild specific service
docker-compose up -d --force-recreate service-name
```

### Database Connection Issues

```bash
# Check if postgres is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Connect to database
docker exec -it rex-postgres psql -U rexadmin -d rex_backend
```

### Out of Memory

```bash
# Check memory usage
docker stats

# Restart heavy services
docker-compose restart api worker

# Clean up unused images
docker system prune -a -f
```

### Disk Space Issues

```bash
# Check disk usage
df -h
docker system df

# Clean up
docker system prune -a --volumes -f

# Check Docker logs
du -sh /var/lib/docker/containers/*/*-json.log
```

### Spot Instance Interrupted

```bash
# Check if instance is running
aws ec2 describe-instances --instance-ids $INSTANCE_ID

# If stopped, start it
aws ec2 start-instances --instance-ids $INSTANCE_ID

# Wait for it to come back up (~2-3 minutes)
# Docker Compose will auto-start all services
```

## 🔄 Migration Path

### From Standard Mode to All-in-One

1. **Backup data** (if you have any)
2. Enable low-cost mode:
   ```bash
   pulumi config set rex-backend:lowcost true
   pulumi up
   ```
3. This will destroy Fargate resources and create EC2 instance
4. Build and push images to ECR
5. Wait for services to start (~5 minutes)

### From All-in-One to Standard Mode

1. **Backup data** from EC2 instance
2. Disable low-cost mode:
   ```bash
   pulumi config set rex-backend:lowcost false
   pulumi up
   ```
3. This will destroy EC2 and create Fargate resources
4. Restore data if needed

## ⚠️ Limitations

1. **Single Point of Failure**
   - No high availability
   - If instance goes down, everything goes down
   - Spot interruptions possible (though rare)

2. **Resource Limits**
   - Limited to 2 vCPU, 4 GB RAM
   - All services share resources
   - Not suitable for high traffic

3. **No Auto-Scaling**
   - Fixed capacity
   - Can't scale horizontally
   - Manual vertical scaling (change instance type)

4. **Manual Updates**
   - Must SSH in to update
   - No automatic deployment
   - More operational overhead

## ✅ Best For

- ✅ Development environments
- ✅ Testing/QA environments
- ✅ Staging environments
- ✅ MVPs and small projects
- ✅ Learning and experimentation
- ✅ Cost-conscious deployments

## ❌ Not For

- ❌ Production (unless very small scale)
- ❌ High-availability requirements
- ❌ High-traffic applications
- ❌ Mission-critical systems
- ❌ Compliance-heavy industries

## 📈 Performance Tips

1. **Increase instance size** if needed:
   - t3a.large (2 vCPU, 8 GB RAM)
   - t3a.xlarge (4 vCPU, 16 GB RAM)

2. **Optimize Docker images**:
   - Use multi-stage builds
   - Remove unnecessary dependencies
   - Use Alpine-based images

3. **Tune PostgreSQL**:
   - Adjust max_connections
   - Configure shared_buffers
   - Set up connection pooling

4. **Monitor resources**:
   - Set up CloudWatch alarms
   - Watch docker stats
   - Monitor disk usage

## 🎯 Next Steps

After deployment:
1. ✅ Verify all containers are running
2. ✅ Test API endpoints
3. ✅ Run database migrations
4. ✅ Check application logs
5. ✅ Set up monitoring
6. ✅ Configure backups

---

**Need help?** Check the troubleshooting section or access the instance via SSM!

