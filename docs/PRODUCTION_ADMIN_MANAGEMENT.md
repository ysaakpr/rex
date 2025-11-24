# Production Platform Admin Management

## Overview

This guide covers how to manage platform administrators on your production/cloud server.

## Scripts Available

| Script | Purpose |
|--------|---------|
| `create_platform_admin_production.sh` | Add a new platform admin |
| `get_user_id.sh` | Get a user's ID by email |
| `list_platform_admins.sh` | List all current platform admins |

## Quick Start

### 1. SSH to Production Server

```bash
ssh ubuntu@10.20.146.127  # or your server IP
cd ~/rex  # or your deployment directory
```

### 2. Get User ID (if you don't have it)

**Method A: By Email**
```bash
./scripts/get_user_id.sh user@example.com
```

**Method B: From Browser Console**
1. Log in as the user
2. Open browser console (F12)
3. Run:
   ```javascript
   Session.getUserId().then(id => console.log(id))
   ```
4. Copy the ID that appears

**Method C: From Database**
```bash
docker-compose exec postgres psql -U utmuser -d utm_backend -c \
  "SELECT user_id, email FROM emailpassword_users WHERE email = 'user@example.com';"
```

### 3. Create Platform Admin

```bash
./scripts/create_platform_admin_production.sh <USER_ID>
```

**Example:**
```bash
./scripts/create_platform_admin_production.sh 04413f25-fdfa-42a0-a046-c3ad67d135fe
```

**Expected Output:**
```
==========================================
Platform Admin Creation - Production
==========================================

User ID: 04413f25-fdfa-42a0-a046-c3ad67d135fe

✓ Docker Compose found
✓ PostgreSQL container is running

Creating platform admin...

NOTICE:  ✅ Platform admin created successfully!
 user_id                              | created_by | created_at                  | status
--------------------------------------+------------+-----------------------------+---------------
 04413f25-fdfa-42a0-a046-c3ad67d135fe | system     | 2025-11-24 13:00:00.123456 | 🆕 Just created

==========================================
✅ Success!
==========================================

User 04413f25-fdfa-42a0-a046-c3ad67d135fe is now a platform admin.

They can now access:
  • Platform Admin Management
  • Roles & Policies Management
  • Permissions Management

Access URLs:
  • https://rex.stage.fauda.dream11.in/platform/admins
  • https://rex.stage.fauda.dream11.in/roles
  • https://rex.stage.fauda.dream11.in/permissions
```

### 4. Verify Platform Admin Was Created

**List all admins:**
```bash
./scripts/list_platform_admins.sh
```

**Or check directly:**
```bash
docker-compose exec postgres psql -U utmuser -d utm_backend -c \
  "SELECT * FROM platform_admins;"
```

## Complete Workflow Example

### Scenario: Add your first platform admin

```bash
# 1. SSH to server
ssh ubuntu@10.20.146.127
cd ~/rex

# 2. Pull latest scripts
git pull origin main

# 3. Get user ID by email
./scripts/get_user_id.sh admin@example.com

# Output shows:
# user_id: 04413f25-fdfa-42a0-a046-c3ad67d135fe
# email: admin@example.com

# 4. Create platform admin
./scripts/create_platform_admin_production.sh 04413f25-fdfa-42a0-a046-c3ad67d135fe

# 5. Verify
./scripts/list_platform_admins.sh

# 6. Test - Log in as that user and access:
# https://rex.stage.fauda.dream11.in/platform/admins
```

## Troubleshooting

### Issue: "docker-compose not found"

**Solution:**
```bash
# Check if docker-compose is installed
docker-compose --version

# If not, install it
sudo apt-get update
sudo apt-get install docker-compose
```

### Issue: "PostgreSQL container is not running"

**Solution:**
```bash
# Check container status
docker-compose ps

# Start services
docker-compose up -d

# Check logs if it fails
docker-compose logs postgres
```

### Issue: "User is already a platform admin"

This is **normal** if you run the script twice. The script is idempotent - it won't create duplicates.

To verify:
```bash
./scripts/list_platform_admins.sh
```

### Issue: "psql: could not connect to server"

**Check database container:**
```bash
docker-compose ps postgres
docker-compose logs postgres --tail=50
```

**Restart if needed:**
```bash
docker-compose restart postgres
sleep 5  # Wait for it to start
./scripts/create_platform_admin_production.sh <USER_ID>
```

### Issue: Can't find user ID

**All methods to get user ID:**

1. **From email (recommended):**
   ```bash
   ./scripts/get_user_id.sh user@example.com
   ```

2. **From SuperTokens database:**
   ```bash
   docker-compose exec postgres psql -U utmuser -d utm_backend -c \
     "SELECT user_id, email, time_joined FROM emailpassword_users ORDER BY time_joined DESC LIMIT 10;"
   ```

3. **From browser (user must be logged in):**
   - F12 → Console
   - `await Session.getUserId()`

4. **From API (user must be logged in):**
   ```bash
   curl -X GET https://rex.stage.fauda.dream11.in/api/v1/users/me \
     -H "Cookie: YOUR_SESSION_COOKIE" | jq .
   ```

### Issue: Script has no execute permission

```bash
chmod +x ./scripts/create_platform_admin_production.sh
chmod +x ./scripts/get_user_id.sh
chmod +x ./scripts/list_platform_admins.sh
```

## Removing a Platform Admin

To remove platform admin access:

```bash
# Get user ID first
USER_ID="04413f25-fdfa-42a0-a046-c3ad67d135fe"

# Remove from platform_admins table
docker-compose exec postgres psql -U utmuser -d utm_backend <<EOF
DELETE FROM platform_admins WHERE user_id = '$USER_ID';
SELECT 'Removed' as status;
EOF
```

Or use the API (as a platform admin):
```bash
curl -X DELETE https://rex.stage.fauda.dream11.in/api/v1/platform/admins/$USER_ID \
  -H "Cookie: YOUR_SESSION_COOKIE"
```

## Security Best Practices

1. **Limit platform admins**: Only grant to trusted users
2. **Audit regularly**: Run `list_platform_admins.sh` periodically
3. **Use email verification**: Verify the user's identity before granting access
4. **Log actions**: Platform admin actions are logged in the database
5. **Rotate regularly**: Review and remove inactive platform admins

## Database Schema

The `platform_admins` table:

```sql
CREATE TABLE platform_admins (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR(255) NOT NULL UNIQUE,
  created_by VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
```

## API Endpoints

Platform admins can access these endpoints:

- `GET /api/v1/platform/admins/check` - Check if current user is admin
- `GET /api/v1/platform/admins` - List all platform admins
- `POST /api/v1/platform/admins` - Create new platform admin
- `DELETE /api/v1/platform/admins/:userId` - Remove platform admin

## Automation

### Create Admin on User Signup (Optional)

To automatically make the first user a platform admin:

```bash
# Add to your deployment script
FIRST_USER_EMAIL="admin@example.com"

# Wait for user to sign up, then:
USER_ID=$(docker-compose exec -T postgres psql -U utmuser -d utm_backend -t -c \
  "SELECT user_id FROM emailpassword_users WHERE email = '$FIRST_USER_EMAIL' LIMIT 1;" | xargs)

if [ ! -z "$USER_ID" ]; then
  ./scripts/create_platform_admin_production.sh "$USER_ID"
fi
```

### Environment-Based Admin

Set in `.env`:
```bash
DEFAULT_ADMIN_EMAIL=admin@example.com
```

Then in your startup script:
```bash
if [ ! -z "$DEFAULT_ADMIN_EMAIL" ]; then
  ./scripts/get_user_id.sh "$DEFAULT_ADMIN_EMAIL" | grep "user_id" | \
    awk '{print $2}' | xargs ./scripts/create_platform_admin_production.sh
fi
```

## Related Documentation

- [Platform Admin Design](changedoc/06-PLATFORM_ADMIN_DESIGN.md)
- [Platform Admin Complete](changedoc/07-PLATFORM_ADMIN_COMPLETE.md)
- [RBAC Authorization Guide](RBAC_AUTHORIZATION_GUIDE.md)

---

**Last Updated**: November 24, 2025  
**Tested On**: Ubuntu 22.04 LTS, Docker Compose v2  
**Status**: ✅ Production Ready

