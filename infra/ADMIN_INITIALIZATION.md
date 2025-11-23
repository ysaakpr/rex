# Platform Admin Initialization

## Overview

The all-in-one deployment mode automatically creates a default platform administrator account during initial setup. This eliminates the need for manual database manipulation to create the first admin user.

## Default Admin Credentials

When deploying using the all-in-one mode (`allinone=true`), a default platform admin is automatically created:

- **Email**: `admin@platform.local`
- **Password**: `admin`
- **User Type**: Platform Administrator

⚠️ **SECURITY WARNING**: Change this default password immediately after first login!

## How It Works

### Initialization Process

The initialization happens automatically after all services are started:

1. **Service Health Checks** (up to 5 minutes timeout)
   - Waits for API to be healthy at `http://localhost:8080/health`
   - Waits for SuperTokens to be healthy at `http://localhost:3567/hello`

2. **SuperTokens User Creation**
   - Creates user via SuperTokens Core API using EmailPassword recipe
   - Email: `admin@platform.local`
   - Password: `admin`
   - If user already exists, retrieves the existing user ID

3. **User Metadata Setup**
   - Sets `is_platform_admin: true` in SuperTokens user metadata
   - Marks as created by system
   - Stores email in metadata for easy identification

4. **Database Entry**
   - Inserts user ID into `platform_admins` table
   - Uses `ON CONFLICT DO NOTHING` to handle re-runs gracefully

5. **Completion Status**
   - Creates `/app/.admin-initialized` file to indicate successful completion
   - Logs admin credentials and user ID
   - Displays security warning

### Initialization Script Location

On the EC2 instance, the initialization script is located at:
```
/app/init-admin.sh
```

### Manual Re-initialization

If you need to re-run the initialization:

```bash
# SSH into the EC2 instance
cd /app
./init-admin.sh
```

The script is idempotent and can be safely re-run.

## First Login

### Using the Frontend

**Note**: The all-in-one mode does not automatically deploy the frontend. You'll need to deploy it manually first (see [ALLINONE_QUICKSTART.md](./ALLINONE_QUICKSTART.md#-frontend-deployment-manual)).

Once your frontend is deployed:

1. Navigate to your deployed application URL
2. Click "Sign In"
3. Enter credentials:
   - Email: `admin@platform.local`
   - Password: `admin`
4. **Immediately change your password** after first login

### Using the API

**1. Sign In**

```bash
# Get your instance public DNS
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

# Sign in (API is on port 8080)
curl -X POST http://$PUBLIC_DNS:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "admin@platform.local"},
      {"id": "password", "value": "admin"}
    ]
  }' \
  -c cookies.txt
```

**2. Verify Platform Admin Access**

```bash
curl -X GET http://$PUBLIC_DNS:8080/api/v1/platform/admin \
  -H "Content-Type: application/json" \
  -b cookies.txt
```

## Changing the Default Password

### Option 1: Via SuperTokens Dashboard (Recommended)

1. Access SuperTokens dashboard (if enabled)
2. Find user `admin@platform.local`
3. Update password
4. Save changes

### Option 2: Via API

**Change Password Endpoint** (requires authentication):

```bash
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

curl -X PUT http://$PUBLIC_DNS:8080/api/v1/auth/user/password \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "oldPassword": "admin",
    "newPassword": "your-secure-password-here"
  }'
```

## Troubleshooting

### Admin User Not Created

**Check initialization logs:**

```bash
# SSH into EC2 instance
cd /app
cat /var/log/cloud-init-output.log | grep "admin"
```

**Manual initialization:**

```bash
cd /app
./init-admin.sh
```

### Services Not Ready Timeout

The initialization waits up to 5 minutes for services. If this timeout is reached:

1. Check if services are running:
   ```bash
   docker-compose ps
   ```

2. Check service logs:
   ```bash
   docker-compose logs api
   docker-compose logs supertokens
   ```

3. Restart services if needed:
   ```bash
   docker-compose restart
   ```

4. Re-run initialization:
   ```bash
   ./init-admin.sh
   ```

### User Already Exists

If you see "EMAIL_ALREADY_EXISTS_ERROR", the user was already created. The script will:
- Attempt to sign in with the credentials
- Retrieve the existing user ID
- Update metadata and database entry
- Continue successfully

### Cannot Log In

**Verify user exists in SuperTokens:**

```bash
docker exec rex-api curl -X GET "http://supertokens:3567/users?email=admin@platform.local" \
  -H "api-key: YOUR_SUPERTOKENS_API_KEY"
```

**Verify user in platform_admins:**

```bash
docker exec rex-postgres psql -U rexadmin -d rex_backend \
  -c "SELECT * FROM platform_admins;"
```

### Reset Admin User

If you need to completely reset the admin user:

```bash
# 1. Remove from database
docker exec rex-postgres psql -U rexadmin -d rex_backend \
  -c "DELETE FROM platform_admins WHERE user_id IN (
    SELECT user_id FROM system_users WHERE email = 'admin@platform.local'
  );"

# 2. Delete from SuperTokens (via dashboard or API)

# 3. Re-run initialization
cd /app
rm .admin-initialized
./init-admin.sh
```

## Security Best Practices

### After Initial Deployment

1. ✅ **Change the default password immediately**
2. ✅ **Use a strong password** (min 12 characters, mixed case, numbers, symbols)
3. ✅ **Enable MFA** (if available in your SuperTokens configuration)
4. ✅ **Limit access** to the admin account
5. ✅ **Create additional admin users** with personal accounts
6. ✅ **Audit admin activities** regularly

### For Production Deployments

1. **Disable default user creation** by modifying the Pulumi script
2. **Use SSO/OAuth** for admin authentication
3. **Implement IP whitelisting** for admin access
4. **Enable audit logging** for all admin actions
5. **Set up alerts** for admin user activities
6. **Regular password rotation** policies

## Customization

### Change Default Credentials

Edit the `init-admin.sh` script in the Pulumi infrastructure code:

```go
// In ec2_allinone.go, modify the init-admin.sh script:

cat > init-admin.sh <<'ADMIN_EOF'
#!/bin/bash
# ... (health checks) ...

# Change these values:
ADMIN_EMAIL="your-admin@example.com"
ADMIN_PASSWORD="your-secure-password"

# Create admin user via SuperTokens Core API
SIGNUP_RESPONSE=$(curl -s -X POST http://localhost:3567/recipe/signup \
  -H "Content-Type: application/json" \
  -H "api-key: %s" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")
# ... (rest of script) ...
```

### Disable Auto-Initialization

To disable automatic admin creation, comment out the initialization call in `ec2_allinone.go`:

```go
# Start remaining services
docker-compose up -d

# Wait a bit for services to stabilize, then initialize admin user
# sleep 30
# echo "Running admin initialization..."
# ./init-admin.sh || echo "Admin initialization failed, check logs at /app/admin-init.log"
```

## Related Documentation

- [Platform Admin API](../docs/API_EXAMPLES.md#platform-admin-endpoints)
- [SuperTokens Authentication](../docs/AUTHENTICATION_IMPLEMENTATION.md)
- [All-in-One Deployment](./ALLINONE_QUICKSTART.md)
- [Security Best Practices](../README.md#security-best-practices)

## Notes

- The initialization script uses `jq` for JSON parsing (automatically installed during EC2 setup)
- The script is idempotent and safe to run multiple times
- Database insertion uses `ON CONFLICT DO NOTHING` to prevent duplicates
- The `.admin-initialized` file can be used to check if initialization has run
- All initialization output is logged to system logs for auditing

---

**Last Updated**: November 23, 2025
**Pulumi Version**: Latest
**Deployment Mode**: All-in-One only

