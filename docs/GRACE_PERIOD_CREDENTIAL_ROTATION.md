# Grace Period Credential Rotation

## Overview

Traditional credential rotation immediately revokes old credentials, causing service interruptions if applications haven't updated yet. This implementation uses a **grace period** approach where both old and new credentials work simultaneously during a transition window.

## The Problem

**Immediate Rotation (Old Approach)**:
```
Before: app-worker@system.internal (working) ✅
Rotate: app-worker@system.internal (password changed)
Result: All existing deployments BREAK immediately ❌
```

**Why This Fails**:
- Multi-instance deployments (K8s, Docker Swarm)
- CI/CD pipelines still using old credentials
- Distributed services need time to update
- No rollback capability

## The Solution: Grace Period Rotation

**Grace Period Rotation (New Approach)**:
```
Before: app-worker-v1@system.internal (primary) ✅

Rotate: 
  - app-worker-v1@system.internal (expires in 7 days) ✅
  - app-worker-v2@system.internal (new primary) ✅
  
Result: BOTH credentials work for 7 days ✅

After Grace Period:
  - app-worker-v1@system.internal (deactivated) ❌
  - app-worker-v2@system.internal (still working) ✅
```

## How It Works

### 1. Application Structure

Each **logical application** can have multiple **system users** (credentials):

```
Application: "my-background-worker"
├── Credential 1 (v1) - expires 2025-11-30
│   ├── Email: my-background-worker-v1@system.internal
│   ├── User ID: st-abc123
│   ├── Status: Active
│   ├── Primary: false
│   └── Expires: 2025-11-30 12:00:00
└── Credential 2 (v2) - current
    ├── Email: my-background-worker-v2@system.internal
    ├── User ID: st-xyz789
    ├── Status: Active
    ├── Primary: true
    └── Expires: null (never)
```

### 2. Rotation Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 1: User Initiates Rotation                                    │
├─────────────────────────────────────────────────────────────────────┤
│ User clicks "Rotate Credentials" on application                     │
│ Sets grace period: 7 days (default)                                │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 2: Backend Creates New Credential                             │
├─────────────────────────────────────────────────────────────────────┤
│ 1. Mark old credential(s) as non-primary                           │
│ 2. Set expiry date = NOW + grace_period_days                       │
│ 3. Create new SuperTokens user (new email)                         │
│ 4. Create new system_user record (is_primary=true)                 │
│ 5. Return new credentials + old credential info                    │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 3: Transition Period (Grace Period)                           │
├─────────────────────────────────────────────────────────────────────┤
│ ✅ Old credential: Still works (expires_at in future)              │
│ ✅ New credential: Works (primary)                                 │
│                                                                     │
│ User updates credentials in:                                       │
│   - Production servers                                             │
│   - Staging environments                                           │
│   - CI/CD pipelines                                                │
│   - Development machines                                           │
│   - Documentation                                                  │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 4: Grace Period Expires                                       │
├─────────────────────────────────────────────────────────────────────┤
│ Background job runs (daily):                                       │
│   1. Find all system_users where expires_at < NOW                  │
│   2. Set is_active = false                                         │
│   3. Revoke all sessions (SuperTokens)                             │
│                                                                     │
│ Result:                                                             │
│   ❌ Old credential: No longer works                               │
│   ✅ New credential: Still works                                   │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 5: Manual Revocation (Optional)                               │
├─────────────────────────────────────────────────────────────────────┤
│ If user updates all services before grace period ends:             │
│   - User clicks "Revoke Old Credentials" button                    │
│   - Old credentials immediately deactivated                        │
│   - No need to wait for grace period                               │
└─────────────────────────────────────────────────────────────────────┘
```

### 3. Database Schema

```sql
CREATE TABLE system_users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,           -- Unique technical name (e.g., my-worker-v2)
    application_name VARCHAR(100) NOT NULL,      -- Logical app name (e.g., my-worker)
    email VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL UNIQUE,        -- SuperTokens user ID
    is_primary BOOLEAN DEFAULT true NOT NULL,    -- Current/recommended credential
    expires_at TIMESTAMP,                        -- Grace period expiry (NULL = never)
    is_active BOOLEAN DEFAULT true NOT NULL,
    ...
);

CREATE INDEX idx_system_users_application_name ON system_users(application_name);
CREATE INDEX idx_system_users_expires_at ON system_users(expires_at);
CREATE INDEX idx_system_users_is_primary ON system_users(is_primary);
```

### 4. API Changes

#### Rotate Credentials (NEW)

```http
POST /api/v1/platform/system-users/:id/rotate
Content-Type: application/json

{
  "grace_period_days": 7  // Optional, default 7
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "new_credential": {
      "id": "uuid-new",
      "name": "my-worker-v2",
      "application_name": "my-worker",
      "email": "my-worker-v2@system.internal",
      "password": "sysuser_new123...",
      "user_id": "st-xyz789",
      "is_primary": true,
      "expires_at": null
    },
    "old_credentials": [
      {
        "email": "my-worker-v1@system.internal",
        "expires_at": "2025-11-30T12:00:00Z",
        "message": "This credential will stop working on 2025-11-30"
      }
    ],
    "message": "New credential created. Old credentials will expire on 2025-11-30. Both work during grace period."
  }
}
```

#### List Application Credentials

```http
GET /api/v1/platform/applications/:application_name/credentials
```

**Response**:
```json
{
  "success": true,
  "data": {
    "application_name": "my-worker",
    "credentials": [
      {
        "id": "uuid-v2",
        "email": "my-worker-v2@system.internal",
        "is_primary": true,
        "is_active": true,
        "expires_at": null,
        "last_used_at": "2025-11-23T10:30:00Z",
        "created_at": "2025-11-23T09:00:00Z"
      },
      {
        "id": "uuid-v1",
        "email": "my-worker-v1@system.internal",
        "is_primary": false,
        "is_active": true,
        "expires_at": "2025-11-30T09:00:00Z",
        "last_used_at": "2025-11-23T10:29:00Z",
        "created_at": "2025-11-20T14:00:00Z",
        "warning": "Expires in 7 days"
      }
    ]
  }
}
```

#### Manually Revoke Old Credentials

```http
POST /api/v1/platform/applications/:application_name/revoke-old
```

Immediately deactivates all non-primary credentials for the application.

## Benefits

### 1. Zero-Downtime Rotation ✅
- No service interruption
- Gradual rollout possible
- Rollback capability during grace period

### 2. Multi-Instance Support ✅
- K8s rolling updates work seamlessly
- Canary deployments supported
- Blue/green deployments compatible

### 3. CI/CD Friendly ✅
- Update credentials in secret manager
- Let pipelines naturally pick up new creds
- Old pipelines still work during transition

### 4. Audit Trail ✅
- Track when credentials are rotated
- Monitor last_used_at for both credentials
- Know when old credentials can be safely removed

### 5. Security Best Practices ✅
- Regular rotation without risk
- Automatic expiry enforcement
- Manual revocation when ready

## Configuration

### Default Grace Period
```env
# .env
CREDENTIAL_GRACE_PERIOD_DAYS=7  # Default: 7 days
```

### Background Job
```bash
# Runs daily to deactivate expired credentials
# Configured in: internal/jobs/tasks/expire_credentials.go
```

## Usage Examples

### Example 1: Rotate with Default Grace Period

```bash
# 1. Rotate credentials (7-day grace period)
curl -X POST http://localhost:8080/api/v1/platform/system-users/{id}/rotate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Response includes new password - save it!

# 2. Update your application config
# Option A: Environment variables
export API_USER="my-worker-v2@system.internal"
export API_PASSWORD="sysuser_new123..."

# Option B: Secret manager (AWS Secrets Manager)
aws secretsmanager update-secret \
  --secret-id my-worker-creds \
  --secret-string '{"username":"my-worker-v2@system.internal","password":"sysuser_new123..."}'

# 3. Deploy updated configuration
kubectl set env deployment/my-worker \
  API_USER="my-worker-v2@system.internal" \
  API_PASSWORD="sysuser_new123..."

# 4. Wait for rollout
kubectl rollout status deployment/my-worker

# 5. Verify new credential works
kubectl logs deployment/my-worker | grep "authenticated"

# 6. (Optional) Manually revoke old credentials early
curl -X POST http://localhost:8080/api/v1/platform/applications/my-worker/revoke-old \
  -H "Authorization: Bearer $TOKEN"
```

### Example 2: Custom Grace Period

```bash
# For critical services, use longer grace period
curl -X POST http://localhost:8080/api/v1/platform/system-users/{id}/rotate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"grace_period_days": 30}'
```

### Example 3: Check Active Credentials

```bash
# List all credentials for an application
curl http://localhost:8080/api/v1/platform/applications/my-worker/credentials \
  -H "Authorization: Bearer $TOKEN"

# Response shows:
# - Primary (current) credential
# - Old credentials with expiry dates
# - Last usage timestamps
```

## Monitoring

### Track Credential Usage

```sql
-- See which credentials are still being used
SELECT 
  application_name,
  email,
  is_primary,
  expires_at,
  last_used_at,
  CASE 
    WHEN expires_at IS NULL THEN 'Never expires'
    WHEN expires_at < NOW() THEN 'EXPIRED'
    ELSE 'Expires in ' || EXTRACT(DAY FROM expires_at - NOW()) || ' days'
  END as expiry_status
FROM system_users
WHERE is_active = true
  AND application_name = 'my-worker'
ORDER BY is_primary DESC, created_at DESC;
```

### Dashboard Metrics

- **Active Credentials**: Count per application
- **Expiring Soon**: Credentials expiring in < 24 hours
- **Unused Old Credentials**: Old creds not used in last 24h (safe to revoke)
- **Rotation Frequency**: How often credentials are rotated

## Best Practices

### 1. Choose Appropriate Grace Period
- **Development**: 1-3 days
- **Staging**: 3-7 days
- **Production**: 7-30 days (depending on deployment frequency)

### 2. Monitor Last Used Timestamps
- If old credential hasn't been used in 24h, likely safe to revoke early
- Monitor for zero usage before expiry

### 3. Automate Updates
```yaml
# Example: Automated rotation in CI/CD
- name: Rotate credentials
  run: |
    NEW_CREDS=$(curl -X POST .../rotate)
    kubectl create secret generic my-worker-creds \
      --from-literal=username=$(echo $NEW_CREDS | jq -r '.data.new_credential.email') \
      --from-literal=password=$(echo $NEW_CREDS | jq -r '.data.new_credential.password') \
      --dry-run=client -o yaml | kubectl apply -f -
```

### 4. Document Rotation Schedule
- Rotate credentials every 90 days (compliance)
- Grace period: 7 days
- Notification: Email teams 1 day before rotation

## Troubleshooting

### Q: Old credential stopped working before grace period
**A**: Check if someone manually revoked it:
```sql
SELECT * FROM system_users WHERE email = 'old-email@system.internal';
-- If is_active = false but expires_at is in future, it was manually revoked
```

### Q: How do I rollback a rotation?
**A**: During grace period:
1. Mark old credential as primary again
2. Use old credential in applications
3. Manually revoke new credential

### Q: Can I rotate multiple times during grace period?
**A**: Yes! Each rotation creates a new credential. All non-expired credentials work.

### Q: What happens if grace period expires but services still use old credential?
**A**: Services will fail authentication. Monitor `last_used_at` to prevent this.

## Security Considerations

### Pros
✅ Reduces rotation risk (no immediate breakage)
✅ Encourages regular rotation
✅ Provides audit trail
✅ Supports compliance requirements

### Cons
⚠️ Multiple valid credentials exist temporarily
⚠️ Requires monitoring to ensure old creds aren't forgotten

### Mitigations
- Background job auto-expires old credentials
- Dashboard warnings for expiring credentials
- Email notifications before expiry
- Monitor last_used_at to detect stale credentials

## Future Enhancements

- [ ] Email notifications before expiry
- [ ] Slack/webhook notifications
- [ ] Automatic rotation schedule (every N days)
- [ ] Rotation approval workflow (for prod)
- [ ] Credential usage analytics dashboard

---

**Last Updated**: 2025-11-23
**Version**: 1.0

