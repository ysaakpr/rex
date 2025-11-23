# Debug: Docker Services Not Starting

## Quick Diagnosis

### Step 1: Check if deployment completed

```bash
cd infra
pulumi stack output allInOneInstanceId

# If this returns "No instance ID found", the deployment hasn't completed
# Run: pulumi up --yes
```

### Step 2: SSH to instance and check logs

```bash
# Get instance ID
INSTANCE_ID=$(cd infra && pulumi stack output allInOneInstanceId)

# Connect via SSM
aws ssm start-session --target $INSTANCE_ID

# Once connected, check cloud-init logs
sudo tail -100 /var/log/cloud-init-output.log

# Check if user data script finished
sudo grep -i "All-in-one setup complete" /var/log/cloud-init-output.log
```

### Step 3: Check Docker and Docker Compose

```bash
# On the instance:

# Is Docker running?
sudo systemctl status docker

# Are Docker Compose files created?
ls -la /app/docker-compose.yml

# Check Docker Compose status
cd /app
docker-compose ps

# Check container logs
docker-compose logs --tail=50

# Check specific service logs
docker-compose logs postgres
docker-compose logs nginx
docker-compose logs api
```

## Common Issues and Fixes

### Issue 1: ECR Login Failed

**Symptoms:**
```
Error response from daemon: pull access denied
```

**Cause:** Can't pull API/Worker images from ECR

**Fix:**
```bash
# Check if images exist in ECR
aws ecr describe-images --repository-name rex-backend-dev-api --region ap-south-1

# If no images, you need to build and push first
cd /path/to/utm-backend
cd infra
./scripts/build-and-push.sh
```

### Issue 2: Docker Compose File Not Created

**Symptoms:**
```
Can't find a suitable configuration file
```

**Cause:** User data script didn't complete

**Check:**
```bash
# View full cloud-init log
sudo cat /var/log/cloud-init-output.log

# Check for errors
sudo grep -i error /var/log/cloud-init-output.log
```

**Fix:**
```bash
# Manually create the docker-compose file by running the user data script
# Or redeploy: pulumi destroy && pulumi up
```

### Issue 3: Port Conflicts

**Symptoms:**
```
bind: address already in use
```

**Cause:** Port already used by another service

**Fix:**
```bash
# Check what's using the ports
sudo netstat -tlnp | grep -E ':(80|443|8080|3567|5432|6379)'

# Kill conflicting processes or change ports
```

### Issue 4: Health Checks Failing

**Symptoms:**
Services keep restarting

**Check:**
```bash
docker-compose logs api
docker-compose logs postgres
docker-compose logs supertokens

# Check health status
docker inspect rex-api | grep -A 10 Health
docker inspect rex-postgres | grep -A 10 Health
```

**Common causes:**
- Database not ready yet
- Environment variables incorrect
- Network connectivity issues

### Issue 5: Nginx Configuration Error

**Symptoms:**
```
nginx: [emerg] ...
```

**Check:**
```bash
# Test nginx config
docker-compose exec nginx nginx -t

# View nginx logs
docker-compose logs nginx

# Check nginx.conf exists
ls -la /app/nginx.conf
```

**Fix:**
```bash
# If config missing, it may not have been created
# Check cloud-init logs for creation errors
sudo grep "Create nginx configuration" /var/log/cloud-init-output.log
```

## Manual Service Start

If services didn't start automatically, you can start them manually:

```bash
cd /app

# 1. Login to ECR (if needed)
aws ecr get-login-password --region ap-south-1 | \
  docker login --username AWS --password-stdin $(docker-compose config | grep image | head -1 | cut -d':' -f2 | cut -d'/' -f1)

# 2. Pull images
docker-compose pull postgres redis supertokens

# Get API and Worker image URLs from Pulumi
API_IMAGE=$(cd /infra && pulumi stack output apiRepositoryUrl)
WORKER_IMAGE=$(cd /infra && pulumi stack output workerRepositoryUrl)

docker pull $API_IMAGE:latest
docker pull $WORKER_IMAGE:latest

# 3. Start PostgreSQL and Redis first
docker-compose up -d postgres redis

# Wait for them to be healthy
sleep 15

# 4. Initialize database
./init-db.sh

# 5. Start remaining services
docker-compose up -d

# 6. Setup SSL
./setup-ssl.sh

# 7. Wait for services to stabilize
sleep 30

# 8. Initialize admin user
./init-admin.sh
```

## Check Service Health

```bash
# Check all containers
docker-compose ps

# Expected output:
# NAME              STATUS                    PORTS
# rex-postgres      Up (healthy)              
# rex-redis         Up (healthy)              
# rex-supertokens   Up (healthy)              3567/tcp
# rex-api           Up (healthy)              8080/tcp
# rex-worker        Up                        
# rex-nginx         Up (healthy)              80/tcp, 443/tcp
# rex-certbot       Up

# Check if services are responding
curl http://localhost:8080/health  # API
curl http://localhost:3567/hello   # SuperTokens
curl http://localhost/health       # Nginx
```

## Restart Everything

If services are in a bad state:

```bash
cd /app

# Stop everything
docker-compose down

# Remove orphaned containers
docker-compose down --remove-orphans

# Start fresh
docker-compose up -d

# Watch logs
docker-compose logs -f
```

## Complete Reset

If nothing works, clean start:

```bash
cd /app

# Stop and remove everything
docker-compose down -v  # WARNING: This removes volumes/data!

# Remove all containers
docker container prune -f

# Restart from scratch
docker-compose up -d postgres redis
sleep 15
./init-db.sh
docker-compose up -d
./setup-ssl.sh
sleep 30
./init-admin.sh
```

## Check Specific Error Scenarios

### Postgres Won't Start

```bash
docker-compose logs postgres

# Common issues:
# - Data directory permissions
# - Port 5432 already in use
# - Memory insufficient

# Fix permissions
docker-compose down
sudo rm -rf /var/lib/docker/volumes/app_postgres-data
docker-compose up -d postgres
```

### API Won't Start

```bash
docker-compose logs api

# Common issues:
# - Can't connect to database
# - Can't connect to SuperTokens
# - Environment variables missing

# Check environment
docker-compose exec api env | grep -E '(DB_|SUPERTOKENS_|REDIS_)'
```

### Nginx Won't Start

```bash
docker-compose logs nginx

# Common issues:
# - nginx.conf syntax error
# - SSL certificate files missing
# - Port 80/443 already in use

# Test config
docker-compose run --rm nginx nginx -t -c /etc/nginx/nginx.conf

# Check SSL files
ls -la /app/ssl/
```

## Get Help Information

```bash
# System info
uname -a
docker --version
docker-compose --version

# Instance info
curl -s http://169.254.169.254/latest/meta-data/instance-id
curl -s http://169.254.169.254/latest/meta-data/public-ipv4

# Disk space
df -h

# Memory
free -h

# Running processes
ps aux | grep docker
```

## Contact Point

If you're still stuck, gather this information:

```bash
# Create diagnostic bundle
cd /app
tar czf ~/docker-debug.tar.gz \
  /var/log/cloud-init-output.log \
  /app/docker-compose.yml \
  /app/nginx.conf \
  $(docker-compose logs 2>&1)
  
echo "Diagnostic bundle created: ~/docker-debug.tar.gz"
```

Then share:
1. Output of `docker-compose ps`
2. Output of `docker-compose logs`
3. Output of `cat /var/log/cloud-init-output.log | tail -200`

