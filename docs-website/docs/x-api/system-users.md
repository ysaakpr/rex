# System Users API

Complete API reference for machine-to-machine (M2M) authentication using System Users.

## Overview

System Users enable secure service-to-service authentication. Each System User has:
- Unique username (identifier)
- Password credential (JWT token)
- Optional expiration date
- Grace period for credential rotation

**Base URL**: `/api/v1/system-users`

**Authentication**: Platform Admin required for management operations

## Create System User

Create a new System User for M2M authentication.

**Request**:
```http
POST /api/v1/system-users
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "username": "analytics-service",
  "display_name": "Analytics Service",
  "description": "Data analytics and reporting service",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Fields**:
- `username` (required): Unique identifier (3-50 chars, alphanumeric + dash/underscore)
- `display_name` (optional): Human-readable name
- `description` (optional): Service description
- `expires_at` (optional): Expiration datetime (ISO 8601 format), null for no expiration

**Response** (201):
```json
{
  "success": true,
  "message": "System User created successfully",
  "data": {
    "id": "sysuser-uuid",
    "username": "analytics-service",
    "display_name": "Analytics Service",
    "description": "Data analytics and reporting service",
    "password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "is_active": true,
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

:::warning Important
The `password` field is **only returned once** during creation. Store it securely - you cannot retrieve it later. If lost, you must regenerate credentials.
:::

## Get System User

Get details about a specific System User (without password).

**Request**:
```http
GET /api/v1/system-users/:id
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "sysuser-uuid",
    "username": "analytics-service",
    "display_name": "Analytics Service",
    "description": "Data analytics and reporting service",
    "is_active": true,
    "expires_at": "2025-12-31T23:59:59Z",
    "old_password_expires_at": null,
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

## List System Users

Get all System Users with optional filtering.

**Request**:
```http
GET /api/v1/system-users?active=true&page=1&page_size=20
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `active` (optional): Filter by active status (true/false)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Response** (200):
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "sysuser-uuid-1",
        "username": "analytics-service",
        "display_name": "Analytics Service",
        "is_active": true,
        "expires_at": "2025-12-31T23:59:59Z",
        "created_at": "2024-11-01T10:00:00Z"
      },
      {
        "id": "sysuser-uuid-2",
        "username": "reporting-service",
        "display_name": "Reporting Service",
        "is_active": true,
        "expires_at": null,
        "created_at": "2024-11-15T14:30:00Z"
      }
    ],
    "total_count": 2,
    "page": 1,
    "page_size": 20
  }
}
```

## Update System User

Update System User metadata (not credentials).

**Request**:
```http
PATCH /api/v1/system-users/:id
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body** (all fields optional):
```json
{
  "display_name": "Analytics Service v2",
  "description": "Updated analytics service",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

**Response** (200):
```json
{
  "success": true,
  "message": "System User updated successfully",
  "data": {
    "id": "sysuser-uuid",
    "username": "analytics-service",
    "display_name": "Analytics Service v2",
    "description": "Updated analytics service",
    "expires_at": "2026-12-31T23:59:59Z",
    "updated_at": "2024-11-20T15:30:00Z"
  }
}
```

## Regenerate Password

Immediately regenerate credentials, invalidating the old password.

**Request**:
```http
POST /api/v1/system-users/:id/regenerate
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Password regenerated successfully",
  "data": {
    "password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

:::warning Immediate Effect
The old password is **immediately invalidated**. Update your service configuration before calling this endpoint.
:::

## Rotate with Grace Period

Rotate credentials with a grace period, allowing both old and new passwords to work temporarily.

**Request**:
```http
POST /api/v1/system-users/:id/rotate
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
{
  "grace_period_hours": 24
}
```

**Fields**:
- `grace_period_hours` (required): Hours both passwords remain valid (1-168, i.e., max 7 days)

**Response** (200):
```json
{
  "success": true,
  "message": "Password rotated with grace period",
  "data": {
    "new_password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "old_password_expires_at": "2024-11-21T10:00:00Z"
  }
}
```

**Grace Period Workflow**:
1. Call rotate endpoint with grace period
2. Update your services to use the new password
3. Both old and new passwords work during grace period
4. After grace period, old password expires automatically
5. Optionally call `/revoke-old` to end grace period early

## Revoke Old Credentials

Manually end the grace period and invalidate the old password.

**Request**:
```http
POST /api/v1/system-users/:id/revoke-old
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "Old credentials revoked successfully"
}
```

## Deactivate System User

Deactivate a System User (soft delete). Can be reactivated by updating `is_active`.

**Request**:
```http
POST /api/v1/system-users/:id/deactivate
Authorization: Bearer <admin-token>
```

**Response** (200):
```json
{
  "success": true,
  "message": "System User deactivated successfully"
}
```

## Get Application Credentials

Verify and decode a System User JWT token payload.

**Request**:
```http
GET /api/v1/system-users/credentials?username=analytics-service
Authorization: Bearer <system-user-token>
```

**Query Parameters**:
- `username` (required): System User username

**Response** (200):
```json
{
  "success": true,
  "data": {
    "system_user_id": "sysuser-uuid",
    "username": "analytics-service",
    "expires_at": "2025-12-31T23:59:59Z",
    "issued_at": "2024-11-20T10:00:00Z"
  }
}
```

:::tip Use Case
Use this endpoint to verify token validity and extract metadata at service startup.
:::

## Authentication Flow

### Initial Setup

```bash
# 1. Create System User (as Platform Admin)
curl -X POST https://api.example.com/api/v1/system-users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{
    "username": "my-service",
    "display_name": "My Service",
    "description": "Backend service",
    "expires_at": "2025-12-31T23:59:59Z"
  }'

# Response includes password (JWT token):
# {
#   "data": {
#     "username": "my-service",
#     "password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
#   }
# }

# 2. Store password securely (environment variable, secrets manager, etc.)
export SYSTEM_USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 3. Use token in API requests
curl https://api.example.com/api/v1/tenants \
  -H "Authorization: Bearer $SYSTEM_USER_TOKEN"
```

### Credential Rotation (Zero Downtime)

```javascript
// Step 1: Initiate rotation with 24-hour grace period
const rotateResp = await fetch('/api/v1/system-users/my-sysuser-id/rotate', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({grace_period_hours: 24})
});

const {data} = await rotateResp.json();
console.log('New password:', data.new_password);
console.log('Old password expires:', data.old_password_expires_at);

// Step 2: Update service configuration
// Deploy new config with new password
// During grace period, both old and new passwords work

// Step 3: After all services updated (or after grace period)
await fetch('/api/v1/system-users/my-sysuser-id/revoke-old', {
  method: 'POST',
  credentials: 'include'
});
```

## Using System Users in Your Service

### Go Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
)

func main() {
    token := os.Getenv("SYSTEM_USER_TOKEN")
    if token == "" {
        panic("SYSTEM_USER_TOKEN not set")
    }

    // Create HTTP client with authentication
    client := &http.Client{}
    
    req, _ := http.NewRequest("GET", "https://api.example.com/api/v1/tenants", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        panic(fmt.Sprintf("API error: %d", resp.StatusCode))
    }
    
    // Process response...
}
```

### Node.js Example

```javascript
const SYSTEM_USER_TOKEN = process.env.SYSTEM_USER_TOKEN;

async function callAPI(endpoint) {
  const response = await fetch(`https://api.example.com${endpoint}`, {
    headers: {
      'Authorization': `Bearer ${SYSTEM_USER_TOKEN}`
    }
  });
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }
  
  return response.json();
}

// Usage
const tenants = await callAPI('/api/v1/tenants');
console.log('Tenants:', tenants.data);
```

### Python Example

```python
import os
import requests

SYSTEM_USER_TOKEN = os.getenv('SYSTEM_USER_TOKEN')

def call_api(endpoint):
    headers = {
        'Authorization': f'Bearer {SYSTEM_USER_TOKEN}'
    }
    response = requests.get(
        f'https://api.example.com{endpoint}',
        headers=headers
    )
    response.raise_for_status()
    return response.json()

# Usage
tenants = call_api('/api/v1/tenants')
print('Tenants:', tenants['data'])
```

## Security Best Practices

### Store Tokens Securely

✅ **DO**:
- Use environment variables
- Use secrets managers (AWS Secrets Manager, HashiCorp Vault)
- Encrypt at rest
- Limit access to credentials

❌ **DON'T**:
- Hardcode in source code
- Commit to version control
- Log token values
- Share tokens between services

### Rotate Regularly

- Set expiration dates on System Users
- Rotate credentials every 90 days
- Use grace period for zero-downtime rotation
- Monitor `old_password_expires_at`

### Monitor and Audit

- Log all System User API calls
- Alert on failed authentication attempts
- Review active System Users regularly
- Deactivate unused System Users

### Principle of Least Privilege

- Create separate System Users per service
- Assign minimal required permissions
- Use descriptive usernames
- Document purpose in description field

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid username format | Username must be 3-50 alphanumeric chars |
| 400 | Invalid grace period | Must be 1-168 hours |
| 401 | Invalid credentials | Token expired or invalid |
| 403 | System User inactive | Deactivated or expired |
| 404 | System User not found | Invalid System User ID |
| 409 | Username already exists | Choose a different username |

## Next Steps

- [System Auth Overview](/system-auth/overview) - Understanding System Auth
- [Go Library](/system-auth/go) - Go authentication library
- [Java Library](/system-auth/java) - Java authentication library
- [RBAC Guide](/guides/rbac-overview) - Managing permissions
