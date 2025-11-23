# All-in-One Mode - Quick Start

## 🚀 Deploy in 3 Steps

### 1. Deploy Infrastructure
```bash
cd infra
./scripts/allinone-deploy.sh
```

This will:
- ✅ Enable all-in-one mode
- ✅ Deploy EC2 instance with Docker Compose
- ✅ Build and push your app images to ECR
- ✅ Start all services (takes ~5 minutes)

### 2. Get Instance Details

```bash
# Get instance public DNS
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

# Get instance public IP
PUBLIC_IP=$(cd infra && pulumi stack output allInOnePublicIp)

echo "Your instance is accessible at: $PUBLIC_DNS"
```

### 3. Access Admin Account

A default platform admin is automatically created:

**Credentials:**
- Email: `admin@platform.local`
- Password: `admin`

⚠️ **Change this password immediately after first login!**

```bash
# Get the public DNS
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

# Sign in as admin (HTTPS via nginx reverse proxy)
# Note: Use -k flag to accept self-signed certificate
curl -k -X POST https://$PUBLIC_DNS/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "admin@platform.local"},
      {"id": "password", "value": "admin"}
    ]
  }' \
  -c cookies.txt
```

See [ADMIN_INITIALIZATION.md](./ADMIN_INITIALIZATION.md) for details.

### 4. Test Your API
```bash
# Get the public DNS
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

# Test health endpoint (HTTP allowed for monitoring)
curl http://$PUBLIC_DNS/health

# Test via HTTPS (use -k to accept self-signed cert)
curl -k https://$PUBLIC_DNS/health
```

### 5. Enable Let's Encrypt SSL (Optional but Recommended)

By default, a self-signed certificate is used. To enable a trusted Let's Encrypt certificate:

```bash
# SSH into instance
INSTANCE_ID=$(cd infra && pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID

# Once connected, run:
cd /app

# Get your public hostname
PUBLIC_HOSTNAME=$(curl -s http://169.254.169.254/latest/meta-data/public-hostname)
echo "Your hostname: $PUBLIC_HOSTNAME"

# Request Let's Encrypt certificate
docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email \
  -d $PUBLIC_HOSTNAME

# Update nginx.conf to use Let's Encrypt certificates
# Replace the fallback self-signed paths with Let's Encrypt paths
# (The setup-ssl.sh script provides the exact domain)

# Restart nginx
docker-compose restart nginx

# Test HTTPS (no more browser warning!)
curl https://$PUBLIC_HOSTNAME/health
```

**Note**: Let's Encrypt certificates auto-renew via the certbot container every 12 hours.

### 6. View Logs
```bash
# Get instance ID
INSTANCE_ID=$(cd infra && pulumi stack output allInOneInstanceId)

# Connect via SSM
aws ssm start-session --target $INSTANCE_ID

# View logs (once connected)
cd /app
docker-compose logs -f
```

## 🔄 Update Services

After making code changes:

```bash
./infra/scripts/allinone-update.sh
```

Or manually:
```bash
# Build and push
./infra/scripts/build-and-push.sh

# SSH to instance
INSTANCE_ID=$(cd infra && pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID

# Update (once connected)
cd /app
./update.sh
```

## 📊 Common Tasks

### Check Service Status
```bash
aws ssm start-session --target $INSTANCE_ID
cd /app && docker-compose ps
```

### View Specific Service Logs
```bash
# On the instance
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f supertokens
```

### Restart Services
```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart api
```

### Run Database Migration
```bash
# On the instance
cd /app
docker-compose exec api /app/migrate up
```

### Check Resource Usage
```bash
# On the instance
docker stats
```

## 💰 Cost

**Monthly Cost: ~$10**
- EC2 t3a.medium spot: ~$6-8/month
- EBS 30GB: ~$2.40/month
- Data transfer: ~$1/month

**What You DON'T Pay For:**
- ❌ No ALB: Save ~$16/month
- ❌ No NAT Gateway: Save ~$32/month  
- ❌ No Fargate tasks: Save ~$45/month
- ✅ Uses Internet Gateway: **FREE** (included with VPC)

**Total: ~$10/month** (vs ~$100+/month for standard setup = **90% savings!**)

## 🎯 What's Running

All services in Docker Compose:
- **Nginx** - Reverse proxy with HTTPS (ports 80, 443)
- **PostgreSQL 14** - Main database
- **Redis 7** - Cache and queue
- **SuperTokens** - Authentication (internal port 3567)
- **API** - Your Go backend (internal port 8080)
- **Worker** - Background jobs
- **Certbot** - SSL certificate renewal (for Let's Encrypt)

**Nginx Configuration:**
- Listens on ports 80 (HTTP) and 443 (HTTPS)
- HTTP → HTTPS redirect (except `/health` and Let's Encrypt challenges)
- Path-based routing: `/api/*` → API, `/auth/*` → SuperTokens
- Self-signed certificate by default
- Optional Let's Encrypt integration

**Note**: Frontend is **NOT** deployed automatically in all-in-one mode. Deploy manually via Amplify Console or your preferred hosting service.

## 📍 Endpoints

**Nginx reverse proxy with HTTPS:**
- **Base URL**: `https://YOUR-PUBLIC-DNS`
- **API**: `https://YOUR-PUBLIC-DNS/api`
- **SuperTokens**: `https://YOUR-PUBLIC-DNS/auth`
- **Health Check**: `http://YOUR-PUBLIC-DNS/health` (HTTP allowed)

**SSL Certificate:**
- Self-signed certificate active by default (browser will show security warning)
- To enable Let's Encrypt: see `/app/setup-ssl.sh` on instance

Get your public DNS:
```bash
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)
echo "API: https://$PUBLIC_DNS/api"
echo "Note: Browser will warn about self-signed certificate"
```

**Internal routing (within Docker network):**
- nginx → api:8080 (HTTP)
- nginx → supertokens:3567 (HTTP)

## 🌐 Frontend Deployment (Manual)

The all-in-one mode focuses on backend services only. Deploy your frontend separately:

### Option 1: AWS Amplify Console (Recommended)

1. **Open Amplify Console**: https://console.aws.amazon.com/amplify/
2. **New App** → **Host web app** → Connect GitHub
3. **Configure Build Settings**:
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
4. **Add Environment Variables**:
   - `VITE_API_URL`: `https://YOUR-PUBLIC-DNS` (from `pulumi stack output allInOnePublicDns`)
   - `VITE_SUPERTOKENS_URL`: `https://YOUR-PUBLIC-DNS` (same base URL, nginx routes by path)
5. **Deploy** and access via Amplify URL

### Option 2: Vercel

```bash
cd frontend
vercel --prod
```

Add environment variables in Vercel dashboard:
- `VITE_API_URL`: `https://YOUR-PUBLIC-DNS`
- `VITE_SUPERTOKENS_URL`: `https://YOUR-PUBLIC-DNS`

### Option 3: Netlify

```bash
cd frontend
netlify deploy --prod
```

Add environment variables in Netlify dashboard (same as above).

### Option 4: Add to Docker Compose (Co-locate Frontend)

If you want everything on one instance, add the frontend to `docker-compose.yml`:

```yaml
  frontend:
    image: nginx:alpine
    container_name: rex-frontend
    volumes:
      - ./frontend/dist:/usr/share/nginx/html:ro
      - ./frontend/nginx.conf:/etc/nginx/conf.d/default.conf:ro
    networks:
      - app-network
    ports:
      - "80:80"
    restart: unless-stopped
```

Then access at `https://YOUR-PUBLIC-DNS` (or update nginx config to serve frontend on port 80).

## 🔧 Troubleshooting

### Services Won't Start
```bash
# Check logs
docker-compose logs

# Check if images were pulled
docker images

# Manually pull images
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $(docker-compose config | grep image | head -1 | cut -d'/' -f1)
docker-compose pull
docker-compose up -d
```

### Can't Connect to API
```bash
# Check if containers are running
INSTANCE_ID=$(cd infra && pulumi stack output allInOneInstanceId)
aws ssm start-session --target $INSTANCE_ID
# Once connected:
docker-compose ps

# Check security group allows traffic on port 8080
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)
curl -v http://$PUBLIC_DNS:8080/health

# Check if instance has public IP
cd infra && pulumi stack output allInOnePublicIp
```

### Out of Memory
```bash
# Check usage
docker stats

# Increase instance size in ec2_allinone.go
# Change: t3a.medium → t3a.large
# Then: pulumi up
```

## ✅ Next Steps

1. ✅ Deploy: `./scripts/allinone-deploy.sh`
2. ✅ Sign in with default admin credentials
3. ✅ **Change the default admin password immediately**
4. ✅ Test API
5. ✅ Run migrations
6. ✅ Deploy frontend (Amplify or separately)
7. ✅ Set up monitoring
8. ✅ Configure backups

## 📚 Full Documentation

- **Complete Guide**: [LOWCOST_ALLINONE.md](./LOWCOST_ALLINONE.md)
- **Implementation**: [ALLINONE_IMPLEMENTATION_STATUS.md](./ALLINONE_IMPLEMENTATION_STATUS.md)

---

**Questions?** Check the full documentation or connect to the instance via SSM!

