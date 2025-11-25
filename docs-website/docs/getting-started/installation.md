# Installation

Detailed installation guide for Rex.

## Prerequisites

Before installing, ensure you have:

- **Docker Desktop** 20.10+ with Docker Compose V2
- **Git** for version control
- **8GB+ RAM** for running all services
- **10GB+ Disk Space** for Docker images and volumes

**Optional** (for development without Docker):
- **Go** 1.23+
- **Node.js** 18+
- **PostgreSQL** 15+
- **Redis** 7+

## Installation Methods

Choose the method that suits your needs:

### Method 1: Docker Compose (Recommended)

**Best for**: Quick setup, development, testing

```bash
# Clone repository
git clone https://github.com/yourusername/utm-backend.git
cd utm-backend

# Configure environment
cp .env.example .env

# Start all services
docker-compose up -d

# Run migrations
docker-compose exec api /app/migrate up
```

✅ **Advantages**:
- All services configured and connected
- No local dependencies needed
- Easy to reset and restart
- Matches production environment

### Method 2: Local Development

**Best for**: Active development, debugging

**Backend**:
```bash
# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with local database credentials

# Run migrations
make migrate-up

# Start API
go run cmd/api/main.go

# Start Worker (separate terminal)
go run cmd/worker/main.go
```

**Frontend**:
```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

**Required Services**:
- PostgreSQL running on localhost:5432
- Redis running on localhost:6379
- SuperTokens Core running on localhost:3567

### Method 3: Kubernetes

**Best for**: Production deployments, cloud environments

See [Deployment Guide](/deployment/docker) for Kubernetes manifests.

## Step-by-Step Installation

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/utm-backend.git
cd utm-backend
```

### 2. Environment Configuration

```bash
cp .env.example .env
```

Edit `.env`:

**Required Variables**:
```bash
# App
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=utmuser
DB_PASSWORD=utmpassword  # Change in production!
DB_NAME=utm_backend

# SuperTokens
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=            # Optional for development
API_DOMAIN=http://localhost:8080
WEBSITE_DOMAIN=http://localhost:3000

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=                 # Set in production!

# Email (MailHog for development)
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
EMAIL_FROM=noreply@localhost

# Optional: Google OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

### 3. Start Services

```bash
docker-compose up -d
```

Wait for services to be ready (~30 seconds):

```bash
docker-compose ps
```

All services should show "Up" status.

### 4. Database Migrations

```bash
docker-compose exec api /app/migrate up
```

Expected output:
```
Applying migrations...
✓ 20241101000001_create_tenants.up.sql
✓ 20241102000001_create_roles.up.sql
...
✓ All migrations applied
```

### 5. Verify Installation

**Check API Health**:
```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

**Check Frontend**:
Open http://localhost:3000

**Check MailHog**:
Open http://localhost:8025

### 6. Create First User

Navigate to http://localhost:3000 and sign up with:
- Email: your@email.com
- Password: At least 8 characters

### 7. Create Platform Admin

```bash
# Get your user ID from SuperTokens
docker-compose exec supertokens-db psql -U supertokens -d supertokens -c \
  "SELECT user_id, email FROM emailpassword_users;"

# Create platform admin (replace USER_ID)
docker-compose exec db psql -U utmuser -d utm_backend -c \
  "INSERT INTO platform_admins (user_id, created_by) VALUES ('USER_ID', 'system');"
```

Or use the script:
```bash
./scripts/create-platform-admin.sh your@email.com
```

## Verifying Installation

### Check Services

```bash
docker-compose ps
```

Expected output:
```
NAME                    STATUS              PORTS
utm-backend-api         Up                  0.0.0.0:8080->8080/tcp
utm-backend-frontend    Up                  0.0.0.0:3000->3000/tcp
utm-backend-postgres    Up                  5432/tcp
utm-backend-redis       Up                  6379/tcp
utm-backend-supertokens Up                  3567/tcp
utm-backend-worker      Up
utm-backend-mailhog     Up                  0.0.0.0:8025->8025/tcp
```

### Check Logs

```bash
# API logs
docker-compose logs -f api

# Worker logs
docker-compose logs -f worker

# All logs
docker-compose logs -f
```

### Test API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Sign up (should return tokens)
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "test@example.com"},
      {"id": "password", "value": "testpass123"}
    ]
  }'
```

## Troubleshooting Installation

### Services Won't Start

**Check Docker Resources**:
```bash
docker system info
```

Ensure sufficient resources:
- Memory: 8GB+
- CPU: 4+ cores
- Disk: 10GB+ free

**Check Port Conflicts**:
```bash
# Check if ports are in use
lsof -i :3000  # Frontend
lsof -i :8080  # API
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
```

**Solution**: Stop conflicting services or change ports in docker-compose.yml

### Migration Errors

**"no such file or directory"**:
```bash
# Ensure migrations directory is mounted
docker-compose exec api ls -la /app/migrations
```

**"migration already applied"**:
```bash
# Check migration status
docker-compose exec api /app/migrate version

# Force to specific version if needed
docker-compose exec api /app/migrate force VERSION
```

### Database Connection Errors

**"connection refused"**:
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Wait for ready message
docker-compose logs postgres | grep "ready to accept"
```

### SuperTokens Errors

**"Core not reachable"**:
```bash
# Check SuperTokens Core
docker-compose ps supertokens

# Check health
curl http://localhost:3567/hello
```

### Frontend Not Loading

**"Cannot GET /"**:
```bash
# Check frontend logs
docker-compose logs frontend

# Rebuild if needed
docker-compose up -d --build frontend
```

## Uninstallation

### Stop Services

```bash
docker-compose down
```

### Remove All Data

```bash
# Remove containers, networks, and volumes
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

### Complete Cleanup

```bash
# Remove everything including unused Docker resources
docker system prune -a --volumes

# Warning: This removes all Docker data!
```

## Next Steps

After successful installation:

1. **[Quick Start Guide](/getting-started/quick-start)** - Get familiar with the system
2. **[Configuration](/getting-started/configuration)** - Customize settings
3. **[Project Structure](/getting-started/project-structure)** - Understand the codebase
4. **[Authentication Guide](/guides/authentication)** - Learn about auth

## Getting Help

- **Check logs**: `docker-compose logs -f [service]`
- **GitHub Issues**: Report installation problems
- **Discussions**: Ask for help from community

