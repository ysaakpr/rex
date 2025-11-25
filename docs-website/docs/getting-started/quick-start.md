# Quick Start

Get Rex running on your machine in 5 minutes.

## Prerequisites

Before you begin, ensure you have:

- ✅ **Docker Desktop** installed and running
- ✅ **Docker Compose** (included with Docker Desktop)
- ✅ **Git** for cloning the repository
- ✅ **8GB+ RAM** for running all services

**Optional (for development)**:
- Go 1.23+ (if you want to run API without Docker)
- Node.js 18+ (if you want to run frontend without Docker)

## Step 1: Clone the Repository

```bash
git clone https://github.com/yourusername/utm-backend.git
cd utm-backend
```

## Step 2: Configure Environment

Copy the example environment file:

```bash
cp .env.example .env
```

The defaults work for local development. **For production**, you'll need to update:
- Database credentials
- SuperTokens API key
- Redis password
- SMTP settings
- OAuth credentials (if using Google login)

**Example `.env` for local development**:
```bash
# App
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=utmuser
DB_PASSWORD=utmpassword
DB_NAME=utm_backend

# SuperTokens
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=
API_DOMAIN=http://localhost:8080
WEBSITE_DOMAIN=http://localhost:3000

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Email (MailHog for development)
SMTP_HOST=mailhog
SMTP_PORT=1025
EMAIL_FROM=noreply@localhost
```

## Step 3: Start All Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (database)
- Redis (cache & queue)
- SuperTokens Core (authentication)
- Backend API (Go)
- Worker (background jobs)
- Frontend (React)
- MailHog (email testing)

**Wait for services to be ready** (~30 seconds):

```bash
# Check status
docker-compose ps

# Should show all services as "Up"
```

## Step 4: Run Database Migrations

```bash
# Apply migrations
docker-compose exec api /app/migrate up

# Or using Make
make migrate-up
```

Expected output:
```
Applying migrations...
✓ 000001_create_tenants.up.sql
✓ 000002_create_relations.up.sql
✓ 000003_create_tenant_members.up.sql
... (more migrations)
✓ All migrations applied successfully
```

## Step 5: Access the Application

Open your browser:

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | React application |
| **Backend API** | http://localhost:8080 | REST API |
| **API Health** | http://localhost:8080/health | Health check |
| **MailHog** | http://localhost:8025 | Email viewer |

## Step 6: Create Your First User

1. **Navigate to Frontend**: http://localhost:3000

2. **Click "Sign Up"** (or go to http://localhost:3000/auth)

3. **Create Account**:
   - Email: `your.email@example.com`
   - Password: At least 8 characters

4. **Verify Email**: Check MailHog at http://localhost:8025 for verification email (if enabled)

5. **You're logged in!** 🎉

## Step 7: Create Your First Platform Admin

Currently, your user is just a regular user. To access platform admin features, you need to be added as a platform admin.

**Option 1: SQL (Recommended for first admin)**

```bash
# Get your user ID
docker-compose exec supertokens-db psql -U supertokens -d supertokens -c \
  "SELECT user_id, email FROM emailpassword_users;"

# Copy the user_id, then:
docker-compose exec db psql -U utmuser -d utm_backend -c \
  "INSERT INTO platform_admins (user_id, created_by) VALUES ('YOUR_USER_ID', 'system');"
```

**Option 2: Using Script**

```bash
# scripts/create-platform-admin.sh
./scripts/create-platform-admin.sh your.email@example.com
```

## Step 8: Create Your First Tenant

Now that you're a platform admin:

### Via Frontend (Recommended)

1. Navigate to http://localhost:3000/tenants
2. Click "**Create Tenant**"
3. Fill in:
   - **Name**: "My First Company"
   - **Slug**: "my-first-company" (URL-friendly)
4. Click "**Create**"

### Via API (using curl)

First, get your access token from browser cookies, then:

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -b "sAccessToken=YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "My First Company",
    "slug": "my-first-company"
  }'
```

**Response**:
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "tenant-uuid",
    "name": "My First Company",
    "slug": "my-first-company",
    "status": "pending",
    "member_count": 1
  }
}
```

::: tip
The tenant status is initially "pending" because a background job is initializing it. After a few seconds, it will change to "active".
:::

## Step 9: Test API Access

**Health Check**:
```bash
curl http://localhost:8080/health
```

**List Your Tenants** (requires authentication):
```bash
curl http://localhost:8080/api/v1/tenants \
  -H "Cookie: sAccessToken=YOUR_TOKEN"
```

## Verify Everything Works

### ✅ Checklist

- [ ] All Docker containers running (`docker-compose ps`)
- [ ] Frontend accessible at http://localhost:3000
- [ ] Backend API responding at http://localhost:8080/health
- [ ] User account created
- [ ] Platform admin status granted
- [ ] First tenant created
- [ ] MailHog showing emails at http://localhost:8025

## Common Commands

### Docker Compose

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f api        # API logs
docker-compose logs -f worker     # Worker logs
docker-compose logs -f frontend   # Frontend logs

# Restart a service
docker-compose restart api

# Remove everything (including volumes)
docker-compose down -v
```

### Using Makefile

```bash
# Start all services
make run

# Stop all services
make stop

# View API logs
make logs-api

# View worker logs
make logs-worker

# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Access database shell
make shell-db

# Clean everything
make clean
```

## Next Steps

Now that you have Rex running:

1. **[Authentication Guide](/guides/authentication)** - Understand the auth system
2. **[Tenant Management](/guides/multi-tenancy)** - Learn about multi-tenancy
3. **[RBAC System](/guides/rbac-overview)** - Master authorization
4. **[API Reference](/x-api/overview)** - Explore all endpoints
5. **[Frontend Integration](/frontend/react-setup)** - Build your UI

## Troubleshooting

### Services Won't Start

**Check Docker resources**:
```bash
# Ensure Docker has enough memory (8GB+)
docker system df
docker system prune  # Clean up if needed
```

**Check port conflicts**:
```bash
# Ensure ports are not in use: 3000, 8080, 5432, 6379, 3567
lsof -i :3000  # macOS/Linux
netstat -ano | findstr :3000  # Windows
```

### Database Connection Errors

**Wait for PostgreSQL to be ready**:
```bash
docker-compose logs postgres

# Look for: "database system is ready to accept connections"
```

**Reset database** (⚠️ loses all data):
```bash
docker-compose down -v
docker-compose up -d
make migrate-up
```

### Migration Errors

**Check migration status**:
```bash
docker-compose exec api /app/migrate version
```

**Force migration** (if stuck):
```bash
docker-compose exec api /app/migrate force VERSION_NUMBER
```

### Frontend Won't Load

**Check if Vite is running**:
```bash
docker-compose logs frontend

# Should show: "Local: http://localhost:3000/"
```

**Rebuild frontend**:
```bash
docker-compose up -d --build frontend
```

### Authentication Not Working

**Check SuperTokens**:
```bash
# SuperTokens should be running
docker-compose logs supertokens

# Check SuperTokens database
docker-compose logs supertokens-db
```

**Verify configuration**:
```bash
# API_DOMAIN and WEBSITE_DOMAIN should match your setup
cat .env | grep DOMAIN
```

## Development vs Production

**Development** (current setup):
- Uses MailHog instead of real SMTP
- HTTP instead of HTTPS
- Debug logging enabled
- Hot reload for code changes
- All services in Docker Compose

**Production**: See [Deployment Guide](/deployment/docker) for:
- Real SMTP configuration
- HTTPS with SSL certificates
- Production database (managed)
- Load balancing
- Environment-specific configs

## Getting Help

- **Documentation**: You're reading it! 📖
- **Issues**: [GitHub Issues](https://github.com/yourusername/utm-backend/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/utm-backend/discussions)

## What's Next?

You now have a fully functional multi-tenant backend! Here are some ideas:

- **Invite a team member**: Create an invitation and test the flow
- **Create custom roles**: Set up your own RBAC structure
- **Explore the API**: Try different endpoints with Postman
- **Build a feature**: Add custom endpoints for your use case
- **Deploy to production**: Follow the deployment guide

Happy coding! 🚀

