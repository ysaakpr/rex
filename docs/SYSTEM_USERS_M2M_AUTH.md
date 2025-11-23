# System Users - Machine-to-Machine Authentication

## Overview

System Users (also known as Service Accounts) enable machine-to-machine (M2M) authentication for automated processes, background jobs, integrations, and API access without human interaction.

## Architecture

### Dual Storage Approach

```
SuperTokens UserMetadata (Fast Access):
├── is_system_user: true
├── service_name: "background-worker"
└── service_type: "worker"

UTM Database (system_users table):
├── Full business logic and metadata
├── Audit history
├── Custom configuration
└── Relationships
```

### Key Features

- ✅ **SuperTokens Integration**: Uses EmailPassword recipe with custom metadata
- ✅ **Long-lived Tokens**: 24-hour access tokens (vs 1-hour for regular users)
- ✅ **Token-based Identification**: `is_system_user` flag in JWT payload
- ✅ **Secure Credentials**: Password shown once, cryptographically secure
- ✅ **Platform Admin Only**: Only platform administrators can manage system users
- ✅ **Session Management**: Automatic token refresh and revocation support

## Database Schema

```sql
CREATE TABLE system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,  -- format: service-name@system.internal
    supertokens_user_id VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    service_type VARCHAR(50) NOT NULL,  -- 'worker', 'integration', 'cron', 'api'
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_by VARCHAR(255) NOT NULL,  -- SuperTokens user ID of creator
    last_used_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb,  -- Custom configuration
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
```

## API Endpoints

All endpoints require Platform Admin authentication.

### Create System User

```bash
POST /api/v1/platform/system-users
```

**Request:**
```json
{
  "name": "background-worker",
  "description": "Handles async job processing",
  "service_type": "worker",
  "metadata": {
    "ip_whitelist": ["10.0.0.0/24"],
    "rate_limit": 10000
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "System user created successfully",
  "data": {
    "id": "uuid",
    "name": "background-worker",
    "email": "background-worker@system.internal",
    "password": "sysuser_abc123xyz...",
    "supertokens_user_id": "st-user-123",
    "description": "Handles async job processing",
    "service_type": "worker",
    "is_active": true,
    "created_by": "admin-user-id",
    "metadata": {...},
    "created_at": "2025-11-23T05:19:37Z",
    "message": "Save this password securely. It will not be shown again."
  }
}
```

⚠️ **IMPORTANT**: The password is only shown once. Store it securely!

### List System Users

```bash
GET /api/v1/platform/system-users?active_only=true
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "background-worker",
      "email": "background-worker@system.internal",
      "supertokens_user_id": "st-user-123",
      "description": "Handles async job processing",
      "service_type": "worker",
      "is_active": true,
      "created_by": "admin-user-id",
      "last_used_at": "2025-11-23T10:00:00Z",
      "metadata": {...},
      "created_at": "2025-11-23T05:19:37Z",
      "updated_at": "2025-11-23T05:19:37Z"
    }
  ]
}
```

### Get System User

```bash
GET /api/v1/platform/system-users/{id}
```

### Update System User

```bash
PATCH /api/v1/platform/system-users/{id}
```

**Request:**
```json
{
  "description": "Updated description",
  "is_active": false,
  "metadata": {
    "rate_limit": 20000
  }
}
```

### Regenerate Password

```bash
POST /api/v1/platform/system-users/{id}/regenerate-password
```

**Response:**
```json
{
  "success": true,
  "data": {
    "password": "sysuser_new_password_123...",
    "message": "Save this password securely. It will not be shown again. All existing sessions have been revoked."
  }
}
```

⚠️ All existing sessions for this system user will be revoked immediately.

### Deactivate System User

```bash
DELETE /api/v1/platform/system-users/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "System user deactivated successfully. All sessions have been revoked."
  }
}
```

## Usage Guide

### Step 1: Create System User (Platform Admin)

```bash
curl -X POST http://localhost:8080/api/v1/platform/system-users \
  -H "Content-Type: application/json" \
  -H "st-auth-mode: header" \
  -H "Authorization: Bearer <admin_access_token>" \
  -d '{
    "name": "my-service",
    "description": "My service integration",
    "service_type": "integration"
  }'
```

Save the returned password securely!

### Step 2: Authenticate (Service)

```bash
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {
        "id": "email",
        "value": "my-service@system.internal"
      },
      {
        "id": "password",
        "value": "sysuser_abc123xyz..."
      }
    ]
  }'
```

**Response Headers:**
```
st-access-token: eyJhbGc...
st-refresh-token: eyJhbGc...
```

### Step 3: Make API Calls

```bash
curl -X GET http://localhost:8080/api/v1/tenants \
  -H "st-auth-mode: header" \
  -H "Authorization: Bearer <access_token>"
```

### Step 4: Refresh Token (Before Expiry)

```bash
curl -X POST http://localhost:8080/auth/session/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

## Token Payload

When a system user authenticates, the access token contains:

```json
{
  "userId": "st-user-123",
  "is_system_user": true,
  "service_name": "background-worker",
  "service_type": "worker",
  "sessionHandle": "...",
  "iat": 1732272000,
  "exp": 1732358400
}
```

Regular user token (for comparison):

```json
{
  "userId": "st-user-456",
  "sessionHandle": "...",
  "iat": 1732272000,
  "exp": 1732275600
}
```

## Middleware Usage

Check if a request is from a system user:

```go
func MyHandler(c *gin.Context) {
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    payload := sessionContainer.GetAccessTokenPayload()
    
    if isSystemUser, ok := payload["is_system_user"].(bool); ok && isSystemUser {
        // This is a system user
        serviceName := payload["service_name"].(string)
        serviceType := payload["service_type"].(string)
        
        // Apply system user specific logic
        // - Higher rate limits
        // - Different logging
        // - Special permissions
    } else {
        // This is a regular user
    }
}
```

## Security Best Practices

### 1. Credential Storage

Store credentials in secure locations:

**Environment Variables (Docker/Kubernetes):**
```bash
SYSTEM_USER_EMAIL=my-service@system.internal
SYSTEM_USER_PASSWORD=sysuser_abc123xyz...
```

**AWS Secrets Manager:**
```json
{
  "system_user_my_service": {
    "email": "my-service@system.internal",
    "password": "sysuser_abc123xyz..."
  }
}
```

**HashiCorp Vault:**
```bash
vault kv put secret/system-users/my-service \
  email="my-service@system.internal" \
  password="sysuser_abc123xyz..."
```

### 2. Password Rotation

Rotate passwords quarterly:

```bash
# Generate new password
curl -X POST http://localhost:8080/api/v1/platform/system-users/{id}/regenerate-password \
  -H "Authorization: Bearer <admin_token>"

# Update your secrets manager
# All old sessions are automatically revoked
```

### 3. IP Whitelisting (Optional)

Store allowed IPs in metadata:

```json
{
  "metadata": {
    "ip_whitelist": ["10.0.0.0/24", "192.168.1.100"]
  }
}
```

Implement in middleware:

```go
if isSystemUser {
    systemUser, err := systemUserService.GetSystemUserBySuperTokensID(userID)
    if err == nil && systemUser.Metadata["ip_whitelist"] != nil {
        // Check if request IP is in whitelist
    }
}
```

### 4. Rate Limiting

Apply different rate limits for system users:

```go
if isSystemUser {
    rateLimit = 10000 // per hour
} else {
    rateLimit = 100 // per hour
}
```

### 5. Audit Logging

Track system user activity:

```go
if isSystemUser {
    logger.Info("System user API call",
        zap.String("service_name", serviceName),
        zap.String("endpoint", c.Request.URL.Path),
        zap.String("method", c.Request.Method),
        zap.String("ip", c.ClientIP()),
    )
}
```

## Common Use Cases

### 1. Background Worker

```go
// worker/main.go
func main() {
    email := os.Getenv("SYSTEM_USER_EMAIL")
    password := os.Getenv("SYSTEM_USER_PASSWORD")
    
    // Authenticate
    accessToken, refreshToken := authenticate(email, password)
    
    // Process jobs
    for job := range jobs {
        processJob(job, accessToken)
    }
}
```

### 2. Cron Job

```bash
#!/bin/bash
# cron-job.sh

# Authenticate
RESPONSE=$(curl -s -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "'$SYSTEM_USER_EMAIL'"},
      {"id": "password", "value": "'$SYSTEM_USER_PASSWORD'"}
    ]
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.st-access-token')

# Call API
curl -X POST http://localhost:8080/api/v1/jobs/cleanup \
  -H "st-auth-mode: header" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 3. External Integration

```python
# integration.py
import requests

class SystemUserClient:
    def __init__(self, email, password):
        self.email = email
        self.password = password
        self.access_token = None
        self.refresh_token = None
        
    def authenticate(self):
        response = requests.post('http://localhost:8080/auth/signin', json={
            'formFields': [
                {'id': 'email', 'value': self.email},
                {'id': 'password', 'value': self.password}
            ]
        })
        self.access_token = response.headers['st-access-token']
        self.refresh_token = response.headers['st-refresh-token']
        
    def call_api(self, endpoint):
        headers = {
            'st-auth-mode': 'header',
            'Authorization': f'Bearer {self.access_token}'
        }
        return requests.get(f'http://localhost:8080{endpoint}', headers=headers)

# Usage
client = SystemUserClient(
    email=os.environ['SYSTEM_USER_EMAIL'],
    password=os.environ['SYSTEM_USER_PASSWORD']
)
client.authenticate()
response = client.call_api('/api/v1/tenants')
```

## Troubleshooting

### Issue: "Invalid credentials"

**Solution**: Ensure you're using the correct email and password. Remember, the password is shown only once during creation.

### Issue: "Session expired"

**Solution**: Refresh the token before it expires (24h for system users):

```bash
curl -X POST http://localhost:8080/auth/session/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

### Issue: "Unauthorized"

**Possible causes:**
1. System user is deactivated (check `is_active` status)
2. Token has expired (refresh the token)
3. Endpoint requires specific permissions (check RBAC)

### Issue: "Too many requests"

**Solution**: You may have hit the rate limit. Check the system user's metadata for `rate_limit` configuration.

## Service Types

| Type | Description | Example Use Case |
|------|-------------|------------------|
| `worker` | Background job processors | Async task workers, queue consumers |
| `integration` | External service integrations | Third-party API connectors, webhooks |
| `cron` | Scheduled jobs | Cleanup scripts, data sync, reports |
| `api` | API-to-API communication | Microservices, internal APIs |

## Best Practices Summary

✅ Store credentials in secure vaults (AWS Secrets Manager, Vault)
✅ Rotate passwords quarterly
✅ Use IP whitelisting for sensitive services
✅ Implement proper rate limiting
✅ Log all system user activity
✅ Use descriptive service names
✅ Document the purpose in the description field
✅ Deactivate unused system users
✅ Monitor last_used_at timestamp
✅ Use RBAC to limit permissions

❌ Never hardcode credentials in source code
❌ Never commit credentials to git
❌ Never share system user credentials
❌ Never use system users for human actions
❌ Never skip monitoring and logging

## Migration from Other Systems

### From API Keys

If you currently use API keys, migrate to system users:

```bash
# For each API key:
# 1. Create a system user
# 2. Update services to use email/password authentication
# 3. Deactivate old API keys
# 4. Monitor and verify
```

### From Service Accounts

If you have existing service accounts:

```bash
# 1. Create equivalent system users
# 2. Update authentication flow
# 3. Test thoroughly
# 4. Switch over
# 5. Retire old accounts
```

## Support

For questions or issues:
- Check the troubleshooting section
- Review API logs: `docker-compose logs api`
- Verify SuperTokens configuration
- Check database for system_users table
- Ensure platform admin access

## Related Documentation

- [Authentication Implementation](./AUTHENTICATION_IMPLEMENTATION.md)
- [RBAC System](../README.md#rbac)
- [Platform Admin](../README.md#platform-admin)
- [API Examples](./API_EXAMPLES.md)

