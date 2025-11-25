# Bootstrapping First Platform Admin

Guide to creating the first platform administrator after initial deployment.

## Overview

After deploying the system for the first time, you need to bootstrap the first platform administrator who can then manage other admins, roles, and permissions.

## Methods

There are 3 ways to create the first platform admin:

1. **Database Direct** - Direct SQL insert (Recommended for first setup)
2. **API Script** - Using existing admin token
3. **Migration** - Include in migrations (for controlled deployments)

---

## Method 1: Database Direct (Recommended)

###Step 1: Register User via Frontend

```bash
# 1. Open the application
open http://your-domain.com

# 2. Click "Sign Up"
# 3. Register with email and password
# 4. Verify email (if email verification enabled)
# 5. Complete registration
```

### Step 2: Get User ID

```bash
# Connect to database
docker exec -it utm-postgres psql -U postgres -d utm_backend

# Or in production
psql "postgresql://user:password@host:5432/dbname"
```

```sql
-- Find your user ID from SuperTokens
SELECT user_id, email, time_joined
FROM all_auth_recipe_users
WHERE email = 'your-email@example.com';

-- Note the user_id (e.g., '123e4567-e89b-12d3-a456-426614174000')
```

### Step 3: Insert Platform Admin Record

```sql
-- Insert platform admin
INSERT INTO platform_admins (user_id, granted_by, is_active)
VALUES (
    'your-user-id-from-step-2',
    'system',  -- Or your user_id if self-granting
    true
);

-- Verify
SELECT * FROM platform_admins WHERE user_id = 'your-user-id';
```

### Step 4: Verify Access

```bash
# Log out from frontend
# Log in again with the admin email

# Navigate to platform admin section
open http://your-domain.com/platform-admin

# You should now see platform admin features
```

---

## Method 2: API Script

### Prerequisites

- Existing platform admin token OR temporary system setup

### Script

```bash
#!/bin/bash
# scripts/create-first-admin.sh

set -e

API_URL="${API_URL:-http://localhost:8080}"
ADMIN_EMAIL="${1}"

if [ -z "$ADMIN_EMAIL" ]; then
  echo "Usage: $0 <admin-email>"
  exit 1
fi

echo "Creating first platform admin for: $ADMIN_EMAIL"

# Step 1: Get user ID from email
echo "Looking up user..."
USER_ID=$(psql "$DATABASE_URL" -t -c "
  SELECT user_id FROM all_auth_recipe_users
  WHERE email = '$ADMIN_EMAIL'
  LIMIT 1;
" | tr -d '[:space:]')

if [ -z "$USER_ID" ]; then
  echo "Error: User not found. Please register first."
  exit 1
fi

echo "Found user ID: $USER_ID"

# Step 2: Insert platform admin
echo "Granting platform admin access..."
psql "$DATABASE_URL" -c "
  INSERT INTO platform_admins (user_id, granted_by, is_active)
  VALUES ('$USER_ID', 'system', true)
  ON CONFLICT (user_id) DO UPDATE
  SET is_active = true, updated_at = NOW();
"

echo "✓ Platform admin access granted successfully!"
echo ""
echo "Next steps:"
echo "1. Log out from the application"
echo "2. Log in with: $ADMIN_EMAIL"
echo "3. Navigate to /platform-admin"
```

### Usage

```bash
# Set database connection
export DATABASE_URL="postgresql://user:pass@host:5432/dbname"

# Run script
./scripts/create-first-admin.sh admin@example.com
```

---

## Method 3: Migration (Production)

For controlled production deployments, include admin creation in migrations.

### Create Migration

```sql
-- migrations/YYYYMMDDHHMMSS_bootstrap_admin.up.sql

-- This assumes you know the user_id in advance
-- (e.g., from a controlled registration process)

INSERT INTO platform_admins (user_id, granted_by, is_active)
VALUES (
    'known-user-id-from-registration',
    'system',
    true
)
ON CONFLICT (user_id) DO NOTHING;
```

### Safer Approach: Environment Variable

```sql
-- migrations/YYYYMMDDHHMMSS_bootstrap_admin.up.sql

-- Insert admin based on email (requires enabling extensions)
INSERT INTO platform_admins (user_id, granted_by, is_active)
SELECT user_id, 'system', true
FROM all_auth_recipe_users
WHERE email = CURRENT_SETTING('app.bootstrap_admin_email', true)
LIMIT 1
ON CONFLICT (user_id) DO NOTHING;
```

```bash
# Run migration with config
psql "$DATABASE_URL" -c "SET app.bootstrap_admin_email = 'admin@example.com';" \
  -f migrations/YYYYMMDDHHMMSS_bootstrap_admin.up.sql
```

---

## Automated Setup Script

Complete automation for first-time setup:

```bash
#!/bin/bash
# scripts/first-time-setup.sh

set -e

echo "=== First-Time Platform Setup ==="
echo ""

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
DATABASE_URL="${DATABASE_URL}"
ADMIN_EMAIL="${ADMIN_EMAIL}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"

if [ -z "$DATABASE_URL" ] || [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
  echo "Error: Required environment variables not set"
  echo ""
  echo "Required:"
  echo "  DATABASE_URL - PostgreSQL connection string"
  echo "  ADMIN_EMAIL - Email for first admin"
  echo "  ADMIN_PASSWORD - Password for first admin"
  exit 1
fi

echo "Configuration:"
echo "  API URL: $API_URL"
echo "  Admin Email: $ADMIN_EMAIL"
echo ""

# Step 1: Check if services are running
echo "1. Checking services..."
if ! curl -f "$API_URL/health" > /dev/null 2>&1; then
  echo "Error: API is not responding at $API_URL"
  exit 1
fi
echo "✓ API is running"
echo ""

# Step 2: Register admin user
echo "2. Registering admin user..."
REGISTER_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$API_URL/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{
    \"formFields\": [
      {\"id\": \"email\", \"value\": \"$ADMIN_EMAIL\"},
      {\"id\": \"password\", \"value\": \"$ADMIN_PASSWORD\"}
    ]
  }")

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)

if [ "$HTTP_CODE" != "200" ] && [ "$HTTP_CODE" != "201" ]; then
  echo "Warning: User registration returned code $HTTP_CODE (may already exist)"
else
  echo "✓ User registered"
fi
echo ""

# Step 3: Get user ID
echo "3. Looking up user ID..."
USER_ID=$(psql "$DATABASE_URL" -t -c "
  SELECT user_id FROM all_auth_recipe_users
  WHERE email = '$ADMIN_EMAIL'
  LIMIT 1;
" | tr -d '[:space:]')

if [ -z "$USER_ID" ]; then
  echo "Error: Failed to find user"
  exit 1
fi

echo "✓ Found user ID: $USER_ID"
echo ""

# Step 4: Grant platform admin
echo "4. Granting platform admin access..."
psql "$DATABASE_URL" > /dev/null 2>&1 <<EOF
  INSERT INTO platform_admins (user_id, granted_by, is_active)
  VALUES ('$USER_ID', 'system', true)
  ON CONFLICT (user_id) DO UPDATE
  SET is_active = true, updated_at = NOW();
EOF

echo "✓ Platform admin access granted"
echo ""

# Step 5: Verify
echo "5. Verifying setup..."
ADMIN_COUNT=$(psql "$DATABASE_URL" -t -c "
  SELECT COUNT(*) FROM platform_admins WHERE is_active = true;
" | tr -d '[:space:]')

echo "✓ Active platform admins: $ADMIN_COUNT"
echo ""

echo "=== Setup Complete! ==="
echo ""
echo "Admin Credentials:"
echo "  Email: $ADMIN_EMAIL"
echo "  Password: [as configured]"
echo ""
echo "Next Steps:"
echo "  1. Log in at: $API_URL"
echo "  2. Navigate to: /platform-admin"
echo "  3. Set up RBAC: roles, policies, permissions"
echo "  4. Create additional admins if needed"
echo ""
```

### Usage

```bash
# Export configuration
export DATABASE_URL="postgresql://user:pass@host:5432/dbname"
export API_URL="https://api.yourdomain.com"
export ADMIN_EMAIL="admin@yourdomain.com"
export ADMIN_PASSWORD="secure-password-here"

# Run setup
chmod +x scripts/first-time-setup.sh
./scripts/first-time-setup.sh
```

---

## Kubernetes/Cloud Deployments

### Using Init Container

```yaml
# k8s/init-admin-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: bootstrap-admin
spec:
  template:
    spec:
      initContainers:
      - name: wait-for-api
        image: busybox
        command: ['sh', '-c', 'until wget -q --spider http://api:8080/health; do sleep 2; done']
      
      containers:
      - name: bootstrap
        image: postgres:15
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: ADMIN_EMAIL
          valueFrom:
            secretKeyRef:
              name: admin-secret
              key: email
        command:
        - /bin/sh
        - -c
        - |
          psql "$DATABASE_URL" <<EOF
            INSERT INTO platform_admins (user_id, granted_by, is_active)
            SELECT user_id, 'system', true
            FROM all_auth_recipe_users
            WHERE email = '$ADMIN_EMAIL'
            LIMIT 1
            ON CONFLICT (user_id) DO NOTHING;
          EOF
      restartPolicy: OnFailure
```

### Apply

```bash
kubectl apply -f k8s/init-admin-job.yaml
kubectl wait --for=condition=complete job/bootstrap-admin --timeout=60s
```

---

## Security Considerations

### 1. Secure Initial Password

```bash
# Generate secure password
ADMIN_PASSWORD=$(openssl rand -base64 32)

# Store securely
echo "$ADMIN_PASSWORD" | gpg --encrypt --recipient admin@example.com > admin-password.gpg
```

### 2. Force Password Change

After first login, require password change:

```sql
-- Add flag to track first login
ALTER TABLE platform_admins ADD COLUMN must_change_password BOOLEAN DEFAULT true;

-- Update on first login
UPDATE platform_admins
SET must_change_password = false
WHERE user_id = 'user-id';
```

### 3. Audit Logging

```sql
-- Log admin creation
INSERT INTO audit_logs (event_type, user_id, details, timestamp)
VALUES (
  'platform_admin_granted',
  'user-id',
  '{"granted_by": "system", "method": "bootstrap"}',
  NOW()
);
```

### 4. Multi-Factor Authentication

Enable MFA for platform admins immediately after setup.

---

## Verification Checklist

After bootstrapping, verify:

- [ ] User can log in with admin credentials
- [ ] `/platform-admin` route is accessible
- [ ] Platform admin features are visible
- [ ] Can create additional platform admins
- [ ] Can manage tenants
- [ ] Can configure RBAC
- [ ] Audit logs are working

---

## Troubleshooting

### "User not found" after registration

**Check SuperTokens table**:
```sql
SELECT * FROM all_auth_recipe_users WHERE email = 'admin@example.com';
```

If missing, check SuperTokens configuration and connection.

### "Access denied" after granting admin

**Verify record**:
```sql
SELECT * FROM platform_admins WHERE user_id = 'user-id';
```

**Check middleware**:
- Clear browser cache/cookies
- Log out and log back in
- Check middleware logs

### Admin features not showing

**Check frontend routing**:
- Verify `is_platform_admin` claim in session
- Check role-based rendering logic
- Inspect browser console for errors

---

## Related Documentation

- [Platform Admin API](/x-api/platform-admin) - API reference
- [Production Setup](/deployment/production-setup) - Production deployment
- [Security](/guides/security) - Security best practices
