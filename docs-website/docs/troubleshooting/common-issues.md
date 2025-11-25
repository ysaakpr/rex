# Common Issues

Solutions to frequently encountered problems.

## Authentication Issues

### Cookies Not Being Set

**Symptoms**:
- API returns 401 Unauthorized
- SuperTokens session cookies missing in browser

**Causes and Solutions**:

1. **Missing `credentials: 'include'` in fetch**:
```javascript
// ❌ Wrong
fetch('/api/v1/tenants')

// ✅ Correct
fetch('/api/v1/tenants', {credentials: 'include'})
```

2. **CORS configuration issue**:
```go
// Backend: internal/api/router/router.go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"content-type"},
    AllowCredentials: true,  // ← Must be true for cookies
}))
```

3. **Domain mismatch** (localhost vs 127.0.0.1):
```
Use localhost:3000 for frontend
Use localhost:8080 for backend
NOT 127.0.0.1
```

4. **SuperTokens cookie configuration**:
```go
session.Init(&sessmodels.TypeInput{
    CookieSecure:   ptrBool(false),  // false for localhost HTTP
    CookieSameSite: ptrString("lax"),
    CookieDomain:   ptrString("localhost"),  // Match your domain
})
```

### Session Expired

**Symptoms**:
- Redirect to login unexpectedly
- 401 errors after period of inactivity

**Solutions**:

1. **Configure session refresh**:
```javascript
// Frontend SuperTokens init
Session.init({
  sessionTokenFrontendDomain: "localhost",
  sessionExpiredStatusCode: 401
})
```

2. **Automatic token refresh** (SuperTokens handles this):
```javascript
// SuperTokens SDK automatically refreshes tokens
// Ensure you're using Session.init() correctly
```

3. **Check session expiry settings** (backend):
```go
session.Init(&sessmodels.TypeInput{
    SessionExpiredStatusCode: ptrInt(401),
    // Session duration defaults to 24 hours
})
```

### Email Verification Required

**Symptoms**:
- "Email verification required" error
- Cannot access tenant features

**Solutions**:

1. **Check verification status**:
```javascript
const user = await fetch('/api/v1/users/me', {credentials: 'include'})
  .then(res => res.json());
console.log('Verified:', user.data.email_verified);
```

2. **Resend verification email**:
```javascript
await fetch('/auth/user/email/verify/token', {
  method: 'POST',
  credentials: 'include'
});
```

3. **Verify email** (development):
```
Check Mailhog: http://localhost:8025
Click verification link in email
```

4. **Skip verification** (development only):
```sql
-- Manually mark as verified
UPDATE supertokens.emailpassword_users
SET email_verified = true
WHERE email = 'user@example.com';
```

## Permission Issues

### Access Denied / 403 Forbidden

**Symptoms**:
- API returns 403 Forbidden
- "Permission denied" error

**Diagnosis Steps**:

1. **Check user's tenant membership**:
```javascript
const response = await fetch(
  `/api/v1/tenants/${tenantId}/members/${userId}`,
  {credentials: 'include'}
);
const member = await response.json();
console.log('Role:', member.data.role_name);
```

2. **Check user's permissions**:
```javascript
const response = await fetch(
  `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
  {credentials: 'include'}
);
const permissions = await response.json();
console.log('Permissions:', permissions.data);
```

3. **Verify permission exists**:
```javascript
const response = await fetch('/api/v1/platform/permissions?service=blog-api', {
  credentials: 'include'
});
```

4. **Check role → policy → permission chain**:
```javascript
// Get role details
const role = await fetch(`/api/v1/platform/roles/${roleId}`, {
  credentials: 'include'
}).then(r => r.json());

console.log('Role policies:', role.data.policies);

// Get policy details
const policy = await fetch(`/api/v1/platform/policies/${policyId}`, {
  credentials: 'include'
}).then(r => r.json());

console.log('Policy permissions:', policy.data.permissions);
```

**Solutions**:

1. **Assign correct role**:
```javascript
await fetch(`/api/v1/tenants/${tenantId}/members/${userId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({role_id: correctRoleId})
});
```

2. **Create missing permission**:
```javascript
await fetch('/api/v1/platform/permissions', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    service: 'blog-api',
    entity: 'post',
    action: 'create',
    description: 'Create blog posts'
  })
});
```

3. **Assign permission to policy**:
```javascript
await fetch(`/api/v1/platform/policies/${policyId}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({permission_ids: [permissionId]})
});
```

### Platform Admin Access Required

**Symptoms**:
- "Platform admin access required" error
- Cannot access `/api/v1/platform/*` endpoints

**Solutions**:

1. **Check if user is platform admin**:
```javascript
const response = await fetch('/api/v1/platform-admins/check', {
  credentials: 'include'
});
const {data} = await response.json();
console.log('Is platform admin:', data.is_platform_admin);
```

2. **Promote user to platform admin** (via database):
```sql
-- Get user ID from SuperTokens
SELECT user_id FROM supertokens.emailpassword_users WHERE email = 'admin@example.com';

-- Add to platform_admins table
INSERT INTO platform_admins (id, user_id, notes, created_at, updated_at)
VALUES (gen_random_uuid(), 'USER_ID_HERE', 'Manual promotion', NOW(), NOW());
```

3. **Or via API** (if you already have an admin):
```javascript
await fetch('/api/v1/platform-admins', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    user_id: 'USER_ID_HERE',
    notes: 'Promoted via API'
  })
});
```

## Database Issues

### Connection Refused

**Symptoms**:
- "connection refused" error
- API won't start

**Solutions**:

1. **Check PostgreSQL is running**:
```bash
# Docker Compose
docker-compose ps

# Check logs
docker-compose logs postgres

# Restart if needed
docker-compose restart postgres
```

2. **Verify connection string**:
```bash
# Check .env file
cat .env | grep DB_

# Test connection
psql -h localhost -U postgres -d utm_backend
```

3. **Check port availability**:
```bash
# Check if port 5432 is in use
lsof -i :5432

# Or
netstat -an | grep 5432
```

### Migration Failed

**Symptoms**:
- "migration failed" error
- Database schema out of sync

**Solutions**:

1. **Check migration status**:
```bash
make migrate-status
```

2. **Rollback and retry**:
```bash
# Rollback last migration
make migrate-down

# Reapply
make migrate-up
```

3. **Reset database** (development only):
```bash
# Drop and recreate
docker-compose down -v
docker-compose up -d postgres
make migrate-up
```

4. **Fix dirty migration**:
```sql
-- Check migration state
SELECT * FROM schema_migrations;

-- If dirty=true, reset it
UPDATE schema_migrations SET dirty = false WHERE version = XXX;

-- Then retry migration
make migrate-up
```

### Foreign Key Constraint Error

**Symptoms**:
- "violates foreign key constraint" error
- Cannot delete record

**Solutions**:

1. **Find related records**:
```sql
-- Example: Can't delete tenant
SELECT * FROM tenant_members WHERE tenant_id = 'UUID';
SELECT * FROM invitations WHERE tenant_id = 'UUID';
```

2. **Delete related records first**:
```javascript
// Delete members first
await fetch(`/api/v1/tenants/${tenantId}/members/${userId}`, {
  method: 'DELETE',
  credentials: 'include'
});

// Then delete tenant
await fetch(`/api/v1/tenants/${tenantID}`, {
  method: 'DELETE',
  credentials: 'include'
});
```

## Docker Issues

### Container Won't Start

**Symptoms**:
- Container exits immediately
- `docker-compose up` fails

**Solutions**:

1. **Check logs**:
```bash
docker-compose logs api
docker-compose logs worker
docker-compose logs frontend
```

2. **Check environment variables**:
```bash
# Ensure .env file exists
ls -la .env

# Check if variables are loaded
docker-compose config
```

3. **Rebuild containers**:
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up
```

4. **Check port conflicts**:
```bash
# Check if ports are already in use
lsof -i :8080  # API
lsof -i :3000  # Frontend
lsof -i :5432  # PostgreSQL
```

### Volume Permission Issues

**Symptoms**:
- "permission denied" errors
- Database initialization fails

**Solutions**:

1. **Reset volumes**:
```bash
docker-compose down -v
docker-compose up
```

2. **Fix permissions** (Linux/Mac):
```bash
sudo chown -R $(whoami):$(whoami) ./data
```

## Frontend Issues

### Build Failures

**Symptoms**:
- `npm run build` fails
- Vite errors

**Solutions**:

1. **Clear cache and reinstall**:
```bash
cd frontend/
rm -rf node_modules package-lock.json
npm install
```

2. **Check Node version**:
```bash
node --version  # Should be 18+
nvm use 18  # If using nvm
```

3. **Fix dependency conflicts**:
```bash
npm install --legacy-peer-deps
```

### Hot Reload Not Working

**Symptoms**:
- Changes not reflected
- Must manually refresh

**Solutions**:

1. **Check Vite config**:
```javascript
// vite.config.js
export default {
  server: {
    watch: {
      usePolling: true  // For Docker
    },
    host: true,  // Expose to host
    port: 3000
  }
}
```

2. **Restart dev server**:
```bash
docker-compose restart frontend
```

### API Calls Fail (CORS)

**Symptoms**:
- CORS policy error in browser console
- Network requests blocked

**Solutions**:

1. **Check CORS config** (backend):
```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"content-type"},
    AllowCredentials: true,
}))
```

2. **Use proxy** (frontend):
```javascript
// vite.config.js
export default {
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
}
```

## Production Issues

### SSL Certificate Errors

**Symptoms**:
- HTTPS not working
- Certificate validation failed

**Solutions**:

1. **Verify certificate** (AWS ACM):
```bash
aws acm describe-certificate --certificate-arn <arn>
```

2. **Check DNS validation**:
```bash
dig CNAME _validation.yourdomain.com
```

3. **Check ALB listener**:
```bash
aws elbv2 describe-listeners --load-balancer-arn <arn>
```

### High Memory Usage

**Symptoms**:
- Out of memory errors
- Container restarts

**Solutions**:

1. **Check memory usage**:
```bash
docker stats
```

2. **Increase memory limits** (docker-compose.yml):
```yaml
services:
  api:
    deploy:
      resources:
        limits:
          memory: 1G
```

3. **Optimize database queries**:
```go
// Use pagination
db.Limit(pageSize).Offset(offset).Find(&records)

// Use Select to load only needed fields
db.Select("id", "name").Find(&users)

// Use Preload carefully
db.Preload("Author").Find(&articles)  // Only if needed
```

### Database Connection Pool Exhausted

**Symptoms**:
- "too many connections" error
- Slow API responses

**Solutions**:

1. **Configure connection pool**:
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(time.Hour)
```

2. **Check for connection leaks**:
```sql
-- PostgreSQL: Check active connections
SELECT count(*) FROM pg_stat_activity WHERE datname = 'utm_backend';

-- Check idle connections
SELECT * FROM pg_stat_activity WHERE state = 'idle';
```

3. **Close connections properly**:
```go
defer rows.Close()  // Always close result sets
```

## Getting Help

### Enable Debug Logging

**Backend**:
```bash
# Set in .env
LOG_LEVEL=debug

# Or export
export LOG_LEVEL=debug
```

**Frontend**:
```javascript
// Enable SuperTokens debug mode
SuperTokens.init({
  // ... config
  enableDebugLogs: true
});
```

### Collect Diagnostic Information

```bash
# System info
uname -a
docker --version
docker-compose --version

# Container status
docker-compose ps
docker-compose logs --tail=100

# Database status
docker-compose exec postgres psql -U postgres -d utm_backend -c "SELECT version();"

# Environment check
docker-compose exec api env | grep -E "DB_|SUPERTOKENS_|JWT_"
```

### Report Issues

When reporting issues, include:
1. Error message (full stack trace)
2. Steps to reproduce
3. Expected vs actual behavior
4. Environment (Docker, OS, versions)
5. Logs from relevant services
6. Configuration (sanitized, no secrets)

## Next Steps

- [Debug Mode](/troubleshooting/debug-mode) - Enable detailed logging
- [Security Guide](/guides/security) - Security best practices
- [Monitoring](/guides/monitoring) - Set up monitoring and alerts
