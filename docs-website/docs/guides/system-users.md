# System Users (M2M Authentication)

Complete guide to System Users for machine-to-machine authentication.

## What are System Users?

System Users are service accounts that allow automated processes to authenticate with the API. They're designed for:

- Background workers
- Scheduled jobs (cron)
- API integrations
- Data pipelines
- Automated scripts

**Key Differences from Regular Users**:
- Longer token expiry (24 hours vs 1 hour)
- Special JWT payload flags
- Created by platform admins only
- Email format: `service-name@system.internal`
- Cannot log in via web UI

## System User Model

```go
type SystemUser struct {
    ID              uuid.UUID  `json:"id"`
    Name            string     `json:"name"`
    Email           string     `json:"email"`
    ApplicationName string     `json:"application_name"`
    ServiceType     string     `json:"service_type"`
    IsActive        bool       `json:"is_active"`
    IsPrimary       bool       `json:"is_primary"`
    ExpiresAt       *time.Time `json:"expires_at"`
    CreatedBy       string     `json:"created_by"`
    CreatedAt       time.Time  `json:"created_at"`
}
```

### Properties

**Name**: Display name (e.g., "Background Worker")

**Email**: Unique identifier
- Format: `service-name@system.internal`
- Example: `background-worker@system.internal`

**Application Name**: Groups related system users
- Example: "worker", "cron", "integration"

**Service Type**: Type of service
- `worker` - Background job processor
- `cron` - Scheduled task
- `integration` - External integration
- `pipeline` - Data pipeline

**Is Primary**: Primary credential for an application
- Only one primary per application
- Secondary credentials for rotation

**Expires At**: Optional expiry date
- Credentials automatically disabled after this date

## Creating System Users

### Via API (Platform Admin Only)

**Endpoint**: `POST /api/v1/platform/system-users`

**Request**:
```json
{
  "name": "Background Worker",
  "application_name": "worker",
  "service_type": "worker",
  "is_primary": true,
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response**:
```json
{
  "success": true,
  "message": "System user created",
  "data": {
    "system_user": {
      "id": "uuid",
      "name": "Background Worker",
      "email": "background-worker@system.internal",
      "application_name": "worker",
      "service_type": "worker",
      "is_active": true,
      "is_primary": true
    },
    "credentials": {
      "email": "background-worker@system.internal",
      "password": "sysuser_abc123xyz..."
    }
  }
}
```

**Important**: Save the password immediately! It's only shown once.

### Via cURL

```bash
curl -X POST https://api.yourdomain.com/api/v1/platform/system-users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "st-auth-mode: header" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Background Worker",
    "application_name": "worker",
    "service_type": "worker",
    "is_primary": true
  }'
```

## Authentication

System users authenticate like regular users but receive special JWT tokens.

### Login Flow

```
1. Send email + password to /api/auth/signin
2. SuperTokens creates session
3. Session includes is_system_user flag
4. Access token valid for 24 hours
```

### JWT Token Payload

```json
{
  "sub": "user-id",
  "exp": 1732358400,
  "iat": 1732272000,
  "sessionHandle": "...",
  
  "is_system_user": true,
  "service_name": "background-worker",
  "service_type": "worker"
}
```

### Manual Authentication

```bash
# Login
curl -X POST http://localhost:8080/api/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "formFields": [
      {"id": "email", "value": "background-worker@system.internal"},
      {"id": "password", "value": "sysuser_abc123xyz..."}
    ]
  }'

# Extract access token from response headers
# st-access-token: eyJhbGc...

# Use token in API calls
curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "st-auth-mode: header"
```

## Using System Auth Library

**Recommended**: Use the System Auth Library for automatic authentication.

### Go

```go
import "github.com/yourorg/systemauth"

// Setup
vault := systemauth.NewEnvVault("WORKER_EMAIL", "WORKER_PASSWORD")
client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
    Vault:  vault,
    APIURL: "https://api.yourdomain.com",
})

// Make authenticated request (login automatic)
resp, err := client.MakeAuthenticatedRequest(
    "GET",
    "https://api.yourdomain.com/api/v1/tenants",
    nil,
)
```

**[Full Go Documentation →](/system-auth/go)**

### Java

```java
import com.yourorg.systemauth.SystemAuthClient;

// Setup
SecretVault vault = new EnvVault("WORKER_EMAIL", "WORKER_PASSWORD");
SystemAuthClient.Config config = new SystemAuthClient.Config();
config.vault = vault;
config.apiUrl = "https://api.yourdomain.com";

SystemAuthClient client = new SystemAuthClient(config);

// Make authenticated request (login automatic)
Response response = client.makeAuthenticatedRequest(
    "GET",
    "https://api.yourdomain.com/api/v1/tenants",
    null
);
```

**[Full Java Documentation →](/system-auth/java)**

## Credential Management

### Storing Credentials

**Development**: Environment variables
```bash
export WORKER_EMAIL="background-worker@system.internal"
export WORKER_PASSWORD="sysuser_abc123xyz..."
```

**Production**: Secrets Manager

**AWS Secrets Manager**:
```bash
aws secretsmanager create-secret \
  --name utm-backend/worker-credentials \
  --secret-string '{
    "email": "background-worker@system.internal",
    "password": "sysuser_abc123xyz..."
  }'
```

**HashiCorp Vault**:
```bash
vault kv put secret/utm-backend/worker \
  email=background-worker@system.internal \
  password=sysuser_abc123xyz...
```

### Rotating Credentials

**With Grace Period** (Recommended):

```bash
POST /api/v1/platform/system-users/:id/rotate
{
  "grace_period_days": 7
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "old_credentials": {
      "email": "background-worker-old@system.internal",
      "valid_until": "2024-12-07T00:00:00Z"
    },
    "new_credentials": {
      "email": "background-worker@system.internal",
      "password": "sysuser_new123xyz..."
    }
  }
}
```

**Process**:
1. Old credentials remain valid for grace period
2. Deploy new credentials to services
3. After grace period, old credentials revoked
4. Services using old credentials fail
5. Manual intervention required if needed

### Regenerating Password

**No Grace Period**:

```bash
POST /api/v1/platform/system-users/:id/regenerate-password
```

**Response**:
```json
{
  "success": true,
  "data": {
    "email": "background-worker@system.internal",
    "password": "sysuser_regenerated456..."
  }
}
```

⚠️ **Warning**: Old password immediately invalid!

## Use Cases

### Background Worker

```go
// worker/main.go
package main

import (
    "log"
    "time"
    "github.com/yourorg/systemauth"
)

func main() {
    // Setup auth client
    vault := systemauth.NewEnvVault("WORKER_EMAIL", "WORKER_PASSWORD")
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: os.Getenv("API_URL"),
    })
    
    log.Println("Worker started")
    
    // Worker loop
    for {
        // Fetch work (authentication automatic)
        resp, err := client.MakeAuthenticatedRequest(
            "GET",
            os.Getenv("API_URL")+"/api/v1/jobs/pending",
            nil,
        )
        if err != nil {
            log.Printf("Error: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }
        
        // Process work...
        
        time.Sleep(5 * time.Second)
    }
}
```

### Scheduled Job (Cron)

```bash
#!/bin/bash
# daily-report.sh

# Load credentials
export WORKER_EMAIL="cron-job@system.internal"
export WORKER_PASSWORD="$(aws secretsmanager get-secret-value --secret-id cron-credentials --query SecretString --output text | jq -r .password)"

# Run script
./generate-reports
```

### API Integration

```python
# integration.py
import os
import requests
from systemauth import SystemAuthClient

# Setup
client = SystemAuthClient(
    email=os.getenv('INTEGRATION_EMAIL'),
    password=os.getenv('INTEGRATION_PASSWORD'),
    api_url=os.getenv('API_URL')
)

# Fetch data (authentication automatic)
response = client.get('/api/v1/data')
data = response.json()

# Send to external system
external_api.send(data)
```

## Security Best Practices

### 1. Principle of Least Privilege

Only grant necessary permissions:

```sql
-- Example: Read-only system user
-- (Implement via RBAC in future)
```

### 2. Regular Rotation

Rotate credentials every 90 days:

```bash
# Automate rotation
0 0 1 */3 * /scripts/rotate-system-users.sh
```

### 3. Monitor Usage

Log system user activity:

```go
if isSystemUser {
    log.Info("System user request",
        zap.String("service", serviceName),
        zap.String("endpoint", endpoint),
    )
}
```

### 4. Set Expiry Dates

Use `expires_at` for temporary credentials:

```json
{
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### 5. Network Restrictions

Restrict by IP or VPC (implement at infrastructure level):

```nginx
# Nginx example
location /api/v1/system {
    allow 10.0.0.0/8;  # Internal network only
    deny all;
}
```

## Monitoring

### List System Users

```bash
GET /api/v1/platform/system-users

Response:
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Background Worker",
      "email": "background-worker@system.internal",
      "is_active": true,
      "last_used": "2024-11-20T10:30:00Z"
    }
  ]
}
```

### Check Activity

```bash
# View logs
docker-compose logs api | grep "is_system_user"

# Look for system user requests
```

### Deactivate System User

```bash
DELETE /api/v1/platform/system-users/:id

# System user marked inactive
# Can be reactivated later
```

## Troubleshooting

### "Login failed with status 401"

**Cause**: Invalid credentials

**Solution**:
- Verify email and password are correct
- Check system user is active
- Ensure not expired

### "Token expired"

**Cause**: Access token expired (after 24 hours)

**Solution**: System Auth Library automatically re-authenticates

### "Permission denied"

**Cause**: System user doesn't have required permissions

**Solution**: Grant appropriate RBAC permissions (future feature)

## Next Steps

- **[System Auth Library Overview](/system-auth/overview)** - Automated authentication
- **[Go Implementation](/system-auth/go)** - Complete Go library
- **[Java Implementation](/system-auth/java)** - Complete Java library
- **[Custom Vaults](/system-auth/custom-vaults)** - Implement secret storage

