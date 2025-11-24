# Database Query Fix for Platform Admin Scripts

## Problem

The scripts were failing with:
```
ERROR: relation "emailpassword_users" does not exist
```

## Root Cause

SuperTokens stores its data in a **separate database**, not in the main application database.

### Database Structure

```
Docker Compose Services:
├── postgres (utm_backend database)
│   ├── tenants
│   ├── platform_admins ✅
│   ├── roles
│   └── ... (application tables)
│
└── supertokens-db (supertokens database)
    ├── emailpassword_users ✅
    ├── all_auth_recipe_users
    └── ... (SuperTokens tables)
```

## Solution Applied

Updated scripts to query the correct database:

### Before (Incorrect)
```bash
docker-compose exec postgres psql -U utmuser -d utm_backend <<EOF
SELECT * FROM emailpassword_users;  -- ❌ Wrong database!
EOF
```

### After (Correct)
```bash
# Query SuperTokens database for user info
docker-compose exec supertokens-db psql -U supertokens -d supertokens <<EOF
SELECT * FROM emailpassword_users;  -- ✅ Correct!
EOF

# Query main database for platform admin status
docker-compose exec postgres psql -U utmuser -d utm_backend <<EOF
SELECT * FROM platform_admins;  -- ✅ Correct!
EOF
```

## Updated Scripts

### 1. `get_user_id.sh`
- ✅ Now queries `supertokens-db` container for user info
- ✅ Then checks `postgres` container for admin status
- ✅ Shows both email and admin status

### 2. `list_platform_admins.sh`
- ✅ Queries both databases and joins results
- ✅ Shows user_id, email, created info, and status
- ✅ Formatted table output

### 3. `list_platform_admins_simple.sh` (NEW)
- ✅ Only queries main database (no SuperTokens dependency)
- ✅ Shows user_id and metadata (no emails)
- ✅ Use as fallback if SuperTokens db is inaccessible

## How to Use (Updated)

### Get User ID by Email
```bash
./scripts/get_user_id.sh vyshakh.p@dream11.com
```

**Expected Output:**
```
==========================================
Get User ID by Email
==========================================

Searching for user: vyshakh.p@dream11.com

 user_id                              | email                  | created_at
--------------------------------------+------------------------+-------------------------
 abc123-def-456-ghi-789              | vyshakh.p@dream11.com  | 2025-11-24 10:30:00

Checking if user is a platform admin...

     status      
-----------------
 👑 Platform Admin
```

### List All Platform Admins
```bash
./scripts/list_platform_admins.sh
```

**Expected Output:**
```
==========================================
Platform Admins List
==========================================

Fetching platform admins...

USER_ID                                EMAIL                          CREATED_BY      CREATED_AT                STATUS
--------------------------------------------------------------------------------------------------------
abc123-def-456-ghi-789                vyshakh.p@dream11.com         system          2025-11-24 10:30:00      🆕
def456-ghi-789-jkl-012                admin@platform.local          system          2025-11-23 15:20:00      ✓

 total_platform_admins 
-----------------------
                     2
```

### Simple List (No Email)
```bash
./scripts/list_platform_admins_simple.sh
```

## Database Connection Details

### Main Application Database
- **Container**: `postgres`
- **Database**: `utm_backend`
- **User**: `utmuser`
- **Tables**: `platform_admins`, `tenants`, `roles`, etc.

### SuperTokens Database
- **Container**: `supertokens-db`
- **Database**: `supertokens`
- **User**: `supertokens`
- **Tables**: `emailpassword_users`, `all_auth_recipe_users`, etc.

## Manual Database Queries

### Get User by Email
```bash
docker-compose exec supertokens-db psql -U supertokens -d supertokens -c \
  "SELECT user_id, email FROM emailpassword_users WHERE email = 'user@example.com';"
```

### Check Platform Admin Status
```bash
docker-compose exec postgres psql -U utmuser -d utm_backend -c \
  "SELECT * FROM platform_admins WHERE user_id = 'USER_ID_HERE';"
```

### List All Users
```bash
docker-compose exec supertokens-db psql -U supertokens -d supertokens -c \
  "SELECT user_id, email, to_timestamp(time_joined/1000) as joined FROM emailpassword_users ORDER BY time_joined DESC LIMIT 10;"
```

## Troubleshooting

### Issue: "supertokens-db: command not found"

Check container name:
```bash
docker-compose ps
```

If container has different name, update scripts or use:
```bash
docker-compose exec <ACTUAL_CONTAINER_NAME> psql -U supertokens -d supertokens
```

### Issue: "FATAL: password authentication failed"

SuperTokens database credentials are in docker-compose.yml:
```yaml
supertokens-db:
  environment:
    POSTGRES_USER: supertokens
    POSTGRES_PASSWORD: supertokenspass
    POSTGRES_DB: supertokens
```

### Issue: Container not running

Start all services:
```bash
docker-compose up -d
```

Check status:
```bash
docker-compose ps
```

### Issue: Permission denied

Scripts need execute permission:
```bash
chmod +x scripts/*.sh
```

## Testing the Fix

```bash
# 1. Test get_user_id
./scripts/get_user_id.sh vyshakh.p@dream11.com
# Should show user_id and admin status without errors

# 2. Test list_platform_admins
./scripts/list_platform_admins.sh
# Should show table with emails

# 3. Test simple list (fallback)
./scripts/list_platform_admins_simple.sh
# Should show table without emails

# All three should work without "relation does not exist" errors
```

## Related Documentation

- [Production Admin Management](PRODUCTION_ADMIN_MANAGEMENT.md)
- [Platform Admin Quick Guide](PLATFORM_ADMIN_QUICK_GUIDE.md)

---

**Last Updated**: November 24, 2025  
**Issue**: Fixed database query errors  
**Status**: ✅ Resolved

