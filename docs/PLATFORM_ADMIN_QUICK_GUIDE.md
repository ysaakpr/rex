# 🚀 Platform Admin Quick Guide - Cloud Server

## ⚡ Quick Commands

### On Your Cloud Server (ubuntu@10.20.146.127)

```bash
# 1. SSH to server
ssh ubuntu@10.20.146.127
cd ~/rex

# 2. Get user ID by email
./scripts/get_user_id.sh user@example.com

# 3. Create platform admin
./scripts/create_platform_admin_production.sh USER_ID_HERE

# 4. List all admins
./scripts/list_platform_admins.sh
```

## 📋 Complete Example

```bash
# Step 1: SSH
ssh ubuntu@10.20.146.127
cd ~/rex

# Step 2: Pull latest scripts (if needed)
git pull origin main
chmod +x scripts/*.sh

# Step 3: Find user
./scripts/get_user_id.sh admin@example.com
# Copy the user_id from output

# Step 4: Make them admin
./scripts/create_platform_admin_production.sh 04413f25-fdfa-42a0-a046-c3ad67d135fe

# Done! ✅
```

## 🆘 Troubleshooting

### Container not running?
```bash
docker-compose ps
docker-compose up -d
```

### Can't find user?
```bash
# List recent users
docker-compose exec postgres psql -U utmuser -d utm_backend -c \
  "SELECT user_id, email FROM emailpassword_users ORDER BY time_joined DESC LIMIT 10;"
```

### Script permission denied?
```bash
chmod +x scripts/*.sh
```

## 📚 Full Documentation

See: `docs/PRODUCTION_ADMIN_MANAGEMENT.md`

## ✅ What Changed

**New Scripts:**
- ✅ `create_platform_admin_production.sh` - Production-ready version
- ✅ `get_user_id.sh` - Helper to find user IDs
- ✅ `list_platform_admins.sh` - List all admins

**Old Script Issues:**
- ❌ `create_platform_admin.sh` - Designed for local dev only
- ❌ Required sourcing .env (doesn't work in production)
- ❌ No error handling

**Why It Failed Before:**
1. Script tried to source `.env` file (may not exist in production)
2. No validation of docker-compose availability
3. No check if postgres container was running
4. Weak error handling

**Fixed Now:**
1. ✅ Works without .env file
2. ✅ Validates prerequisites
3. ✅ Checks container status
4. ✅ Better error messages
5. ✅ Idempotent (safe to run multiple times)

---

**Ready to use!** Just SSH to your server and run the commands above. 🎉

