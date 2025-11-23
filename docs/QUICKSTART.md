# Quick Start Guide

Get your UTM Backend up and running in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- Terminal/Command line access

## Step-by-Step Setup

### 1. Start the Services

```bash
# Start all services
make run

# Or if you don't have Make installed:
docker-compose up -d
```

This will start:
- PostgreSQL database
- SuperTokens authentication service
- Redis for job queue
- API server
- Background worker
- MailHog for email testing

### 2. Run Database Migrations

```bash
make migrate-up

# Or manually:
docker-compose exec postgres psql -U utmuser -d utm_backend -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

### 3. Verify Services

Check that all services are running:

```bash
docker-compose ps
```

You should see:
- ✓ utm-postgres (healthy)
- ✓ utm-supertokens-db (healthy)
- ✓ utm-supertokens (healthy)
- ✓ utm-redis (healthy)
- ✓ utm-api (running)
- ✓ utm-worker (running)
- ✓ utm-mailhog (running)

### 4. Test the API

Health check:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok"
}
```

## 🎯 Your First Tenant

### 1. Sign Up a User (via SuperTokens)

You'll need to integrate SuperTokens in your frontend, or use the SuperTokens test dashboard at `http://localhost:3567`.

For testing, you can manually create a user in SuperTokens and get the user ID.

### 2. Create a Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN" \
  -d '{
    "name": "My First Tenant",
    "slug": "my-first-tenant",
    "metadata": {
      "industry": "technology"
    }
  }'
```

### 3. Check Tenant Status

```bash
curl http://localhost:8080/api/v1/tenants/TENANT_ID/status \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN"
```

### 4. List Your Tenants

```bash
curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN"
```

## 📨 Viewing Emails

All emails sent by the system are captured by MailHog. View them at:

**http://localhost:8025**

## 🔍 Exploring the System

### List Default Relations

```bash
curl http://localhost:8080/api/v1/relations \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN"
```

Relations included:
- Admin (full access)
- Writer (can create/edit)
- Viewer (read-only)
- Basic (basic access)

### List Default Roles

```bash
curl http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN"
```

Roles included:
- Tenant Admin
- Content Manager
- Analytics Viewer
- User Manager

### View Permissions

```bash
curl http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN"
```

## 👥 Inviting Users

### Invite a User to Your Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants/TENANT_ID/invitations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_SUPERTOKENS_ACCESS_TOKEN" \
  -d '{
    "email": "user@example.com",
    "relation_id": "RELATION_UUID"
  }'
```

The user will receive an invitation email (visible in MailHog).

## 🛠 Development Workflow

### View Logs

```bash
# All services
make logs

# Just API
make logs-api

# Just worker
make logs-worker
```

### Access Database

```bash
make shell-db
# Or
docker-compose exec postgres psql -U utmuser -d utm_backend
```

### Stop Services

```bash
make stop
# Or
docker-compose stop
```

### Clean Everything

```bash
make clean
# This removes all containers and volumes
```

## 🐛 Troubleshooting

### Services Won't Start

1. Check if ports are already in use:
   ```bash
   lsof -i :8080  # API port
   lsof -i :5432  # PostgreSQL
   lsof -i :3567  # SuperTokens
   lsof -i :6379  # Redis
   ```

2. Check Docker logs:
   ```bash
   docker-compose logs
   ```

### Database Connection Issues

```bash
# Check if PostgreSQL is healthy
docker-compose ps postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Migration Errors

```bash
# Check current migration version
docker-compose exec postgres psql -U utmuser -d utm_backend -c "SELECT version FROM schema_migrations;"

# Force migration down and up
make migrate-down
make migrate-up
```

## 📚 Next Steps

1. **Frontend Integration**: Set up SuperTokens in your frontend application
2. **Custom Relations**: Create tenant-specific relations
3. **Custom Roles**: Define roles with specific permission sets
4. **Service Integration**: Configure `TENANT_INIT_SERVICES` to call your other backend services
5. **Email Service**: Configure SMTP or SendGrid for production emails

## 🔗 Useful URLs

- **API Server**: http://localhost:8080
- **MailHog UI**: http://localhost:8025
- **SuperTokens**: http://localhost:3567
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 📖 Documentation

- [Full README](./README.md)
- [API Endpoints](./README.md#-api-endpoints)
- [Database Schema](./README.md#-database-schema)
- [Configuration](./README.md#-configuration)

## 💡 Tips

1. **Postman Collection**: Import the API endpoints into Postman for easier testing
2. **Environment Variables**: Copy `.env.example` to `.env` and customize for your setup
3. **Dev Container**: Use VS Code's dev container for a consistent development environment
4. **Background Jobs**: Watch the worker logs to see tenant initialization in action

---

Happy coding! 🚀

