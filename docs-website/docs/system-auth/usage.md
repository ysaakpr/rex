# System Auth Usage

Practical examples for implementing System User authentication in various scenarios.

## Overview

System Users enable machine-to-machine (M2M) authentication. This guide provides practical examples for:
- Backend service authentication
- Scheduled jobs
- API integrations
- Microservice communication
- CI/CD pipelines

## Basic Usage Pattern

### 1. Create System User

```bash
# As Platform Admin
curl -X POST http://localhost:8080/api/v1/system-users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{
    "username": "analytics-service",
    "display_name": "Analytics Service",
    "description": "Service for analytics and reporting",
    "expires_at": "2025-12-31T23:59:59Z"
  }'

# Response includes password (JWT token) - save this!
{
  "data": {
    "id": "sysuser-uuid",
    "username": "analytics-service",
    "password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 2. Store Token Securely

```bash
# Environment variable
export SYSTEM_USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# AWS Secrets Manager
aws secretsmanager create-secret \
  --name analytics-service/token \
  --secret-string "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# .env file (for development only)
SYSTEM_USER_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3. Use Token in API Calls

```bash
# cURL
curl -H "Authorization: Bearer $SYSTEM_USER_TOKEN" \
  http://localhost:8080/api/v1/tenants

# With specific headers
curl -H "Authorization: Bearer $SYSTEM_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -X POST http://localhost:8080/api/v1/tenants \
  -d '{"name":"New Tenant","slug":"new-tenant"}'
```

## Language-Specific Examples

### Go Service

```go
// main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

type SystemUserClient struct {
    token      string
    baseURL    string
    httpClient *http.Client
}

func NewSystemUserClient(token, baseURL string) *SystemUserClient {
    return &SystemUserClient{
        token:   token,
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *SystemUserClient) doRequest(method, endpoint string, body []byte) (*http.Response, error) {
    req, err := http.NewRequest(method, c.baseURL+endpoint, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")
    
    return c.httpClient.Do(req)
}

func (c *SystemUserClient) GetTenants() ([]Tenant, error) {
    resp, err := c.doRequest("GET", "/api/v1/tenants", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %d", resp.StatusCode)
    }
    
    var result struct {
        Data struct {
            Data []Tenant `json:"data"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result.Data.Data, nil
}

func main() {
    token := os.Getenv("SYSTEM_USER_TOKEN")
    if token == "" {
        log.Fatal("SYSTEM_USER_TOKEN not set")
    }
    
    client := NewSystemUserClient(token, "http://localhost:8080")
    
    tenants, err := client.GetTenants()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d tenants\n", len(tenants))
}
```

### Node.js Service

```javascript
// systemUserClient.js
const axios = require('axios');

class SystemUserClient {
  constructor(token, baseURL = 'http://localhost:8080') {
    this.token = token;
    this.baseURL = baseURL;
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      timeout: 30000
    });
  }
  
  async getTenants() {
    const response = await this.client.get('/api/v1/tenants');
    return response.data.data.data;
  }
  
  async createTenant(data) {
    const response = await this.client.post('/api/v1/tenants', data);
    return response.data.data;
  }
  
  async getMembers(tenantId) {
    const response = await this.client.get(`/api/v1/tenants/${tenantId}/members`);
    return response.data.data.data;
  }
}

// Usage
const token = process.env.SYSTEM_USER_TOKEN;
const client = new SystemUserClient(token);

async function run() {
  try {
    const tenants = await client.getTenants();
    console.log(`Found ${tenants.length} tenants`);
    
    for (const tenant of tenants) {
      const members = await client.getMembers(tenant.id);
      console.log(`${tenant.name}: ${members.length} members`);
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}

run();
```

### Python Service

```python
# system_user_client.py
import os
import requests
from typing import List, Dict, Optional

class SystemUserClient:
    def __init__(self, token: str, base_url: str = "http://localhost:8080"):
        self.token = token
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        })
    
    def _request(self, method: str, endpoint: str, **kwargs) -> Dict:
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response.json()
    
    def get_tenants(self) -> List[Dict]:
        result = self._request("GET", "/api/v1/tenants")
        return result["data"]["data"]
    
    def create_tenant(self, data: Dict) -> Dict:
        result = self._request("POST", "/api/v1/tenants", json=data)
        return result["data"]
    
    def get_members(self, tenant_id: str) -> List[Dict]:
        result = self._request("GET", f"/api/v1/tenants/{tenant_id}/members")
        return result["data"]["data"]

# Usage
if __name__ == "__main__":
    token = os.getenv("SYSTEM_USER_TOKEN")
    if not token:
        raise ValueError("SYSTEM_USER_TOKEN not set")
    
    client = SystemUserClient(token)
    
    # List tenants
    tenants = client.get_tenants()
    print(f"Found {len(tenants)} tenants")
    
    # Get members for each tenant
    for tenant in tenants:
        members = client.get_members(tenant["id"])
        print(f"{tenant['name']}: {len(members)} members")
```

## Use Case Examples

### Scheduled Analytics Report

```python
# analytics_report.py
import os
from datetime import datetime
from system_user_client import SystemUserClient

def generate_analytics_report():
    client = SystemUserClient(os.getenv("SYSTEM_USER_TOKEN"))
    
    # Collect data
    tenants = client.get_tenants()
    
    report = {
        "generated_at": datetime.utcnow().isoformat(),
        "total_tenants": len(tenants),
        "tenants": []
    }
    
    for tenant in tenants:
        members = client.get_members(tenant["id"])
        report["tenants"].append({
            "name": tenant["name"],
            "member_count": len(members),
            "status": tenant["status"]
        })
    
    # Save report
    with open(f"report-{datetime.now().date()}.json", "w") as f:
        json.dump(report, f, indent=2)
    
    print(f"Report generated: {report['total_tenants']} tenants")

if __name__ == "__main__":
    generate_analytics_report()
```

### Data Sync Service

```javascript
// dataSyncService.js
const {SystemUserClient} = require('./systemUserClient');
const cron = require('node-cron');

class DataSyncService {
  constructor() {
    this.client = new SystemUserClient(process.env.SYSTEM_USER_TOKEN);
  }
  
  async syncTenants() {
    console.log('Starting tenant sync...');
    
    const tenants = await this.client.getTenants();
    
    for (const tenant of tenants) {
      try {
        await this.syncTenant(tenant);
        console.log(`✓ Synced ${tenant.name}`);
      } catch (error) {
        console.error(`✗ Failed to sync ${tenant.name}:`, error.message);
      }
    }
    
    console.log('Sync complete');
  }
  
  async syncTenant(tenant) {
    const members = await this.client.getMembers(tenant.id);
    
    // Sync to external system
    await externalSystem.updateTenant({
      id: tenant.id,
      name: tenant.name,
      member_count: members.length
    });
  }
}

// Run every hour
const service = new DataSyncService();
cron.schedule('0 * * * *', () => {
  service.syncTenants();
});

console.log('Data sync service started');
```

### Tenant Provisioning Service

```go
// provisioning_service.go
package main

import (
    "log"
    "time"
)

type ProvisioningService struct {
    client *SystemUserClient
}

func NewProvisioningService(token string) *ProvisioningService {
    return &ProvisioningService{
        client: NewSystemUserClient(token, "http://localhost:8080"),
    }
}

func (s *ProvisioningService) ProvisionTenant(name, slug string) error {
    log.Printf("Provisioning tenant: %s", name)
    
    // Create tenant
    tenant, err := s.client.CreateTenant(map[string]interface{}{
        "name": name,
        "slug": slug,
    })
    if err != nil {
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    log.Printf("Created tenant: %s (ID: %s)", tenant.Name, tenant.ID)
    
    // Additional setup (databases, resources, etc.)
    if err := s.setupTenantResources(tenant.ID); err != nil {
        return fmt.Errorf("failed to setup resources: %w", err)
    }
    
    log.Printf("Provisioning complete for %s", name)
    return nil
}

func (s *ProvisioningService) setupTenantResources(tenantID string) error {
    // Create database
    // Setup S3 bucket
    // Configure resources
    return nil
}

func main() {
    token := os.Getenv("SYSTEM_USER_TOKEN")
    service := NewProvisioningService(token)
    
    if err := service.ProvisionTenant("New Company", "new-company"); err != nil {
        log.Fatal(err)
    }
}
```

## Credential Rotation

### Zero-Downtime Rotation

```bash
# Step 1: Rotate with grace period
curl -X POST http://localhost:8080/api/v1/system-users/{id}/rotate \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"grace_period_hours": 24}'

# Response includes new password
{
  "data": {
    "new_password": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "old_password_expires_at": "2024-11-26T10:00:00Z"
  }
}

# Step 2: Update service configuration
# Deploy new version with new token

# Step 3: After verification, revoke old token
curl -X POST http://localhost:8080/api/v1/system-users/{id}/revoke-old \
  -H "Authorization: Bearer <admin-token>"
```

### Automated Rotation Script

```bash
#!/bin/bash
# rotate-system-user-token.sh

SYSTEM_USER_ID="$1"
ADMIN_TOKEN="$2"
SECRET_NAME="$3"

# Rotate token
echo "Rotating token for system user: $SYSTEM_USER_ID"
RESPONSE=$(curl -s -X POST \
  "http://localhost:8080/api/v1/system-users/$SYSTEM_USER_ID/rotate" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"grace_period_hours": 24}')

NEW_TOKEN=$(echo "$RESPONSE" | jq -r '.data.new_password')

# Update secret in AWS Secrets Manager
aws secretsmanager update-secret \
  --secret-id "$SECRET_NAME" \
  --secret-string "$NEW_TOKEN"

echo "Token rotated successfully"
echo "Old token expires: $(echo "$RESPONSE" | jq -r '.data.old_password_expires_at')"
```

## Error Handling

### Retry Logic

```javascript
async function apiCallWithRetry(client, method, ...args) {
  const maxRetries = 3;
  let lastError;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await client[method](...args);
    } catch (error) {
      lastError = error;
      
      // Check if token expired
      if (error.response?.status === 401) {
        console.error('Token expired - please rotate credentials');
        throw error;
      }
      
      // Retry on server errors
      if (error.response?.status >= 500) {
        const delay = Math.pow(2, i) * 1000;
        console.log(`Retry ${i + 1}/${maxRetries} after ${delay}ms`);
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      
      throw error;
    }
  }
  
  throw lastError;
}

// Usage
const tenants = await apiCallWithRetry(client, 'getTenants');
```

## Best Practices

1. **Store Tokens Securely**
   - Use environment variables or secrets manager
   - Never commit to version control
   - Encrypt at rest

2. **Rotate Regularly**
   - Set expiration dates
   - Rotate every 90 days
   - Use grace period for zero downtime

3. **Monitor Usage**
   - Log all API calls
   - Alert on failures
   - Track token expiration

4. **Handle Errors**
   - Implement retry logic
   - Check for 401 (expired token)
   - Log errors for debugging

5. **One Token Per Service**
   - Don't share tokens between services
   - Create separate System Users
   - Isolate permissions

## Testing

### Test System User Authentication

```bash
# Test token validity
curl -H "Authorization: Bearer $SYSTEM_USER_TOKEN" \
  http://localhost:8080/api/v1/system-users/credentials?username=analytics-service

# Expected response
{
  "success": true,
  "data": {
    "system_user_id": "uuid",
    "username": "analytics-service",
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

## Next Steps

- [System Users API](/x-api/system-users) - API reference
- [System Auth Overview](/system-auth/overview) - Architecture
- [Go Library](/system-auth/go) - Go authentication library
- [Security Guide](/guides/security) - Security best practices
