# Example: Machine-to-Machine Integration

Complete example of integrating external services using System User authentication.

## Scenario

You're building an analytics service that needs to:
- Fetch tenant data from the backend
- Access member information
- Read tenant activity
- Do this automatically without human interaction

## Solution: System User (M2M Authentication)

System Users provide bearer tokens for service-to-service authentication.

## Complete Implementation

### Step 1: Create System User

As a platform administrator:

```bash
# Using curl
curl -X POST https://api.yourdomain.com/api/v1/system-users \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "analytics-service",
    "description": "Analytics data collection service",
    "expires_in_days": 90
  }'

# Response
{
  "success": true,
  "data": {
    "id": "sys-user-uuid",
    "name": "analytics-service",
    "token": "sys_abc123...",  # Save this securely!
    "expires_at": "2025-02-23T10:00:00Z",
    "created_at": "2024-11-25T10:00:00Z"
  }
}
```

**Save the token securely** - it won't be shown again!

### Step 2: Store Token Securely

```bash
# Option 1: Environment variable
export SYSTEM_USER_TOKEN="sys_abc123..."

# Option 2: AWS Secrets Manager
aws secretsmanager create-secret \
  --name /analytics/system-user-token \
  --secret-string "sys_abc123..."

# Option 3: HashiCorp Vault
vault kv put secret/analytics token="sys_abc123..."

# Option 4: Kubernetes Secret
kubectl create secret generic analytics-token \
  --from-literal=token="sys_abc123..."
```

### Step 3: Implement Client Library

#### Go Implementation

```go
// pkg/client/utm_client.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UTMClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewUTMClient(baseURL, token string) *UTMClient {
	return &UTMClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *UTMClient) request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}
	
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	
	return c.httpClient.Do(req)
}

// Get all tenants
func (c *UTMClient) GetTenants() ([]Tenant, error) {
	resp, err := c.request("GET", "/api/v1/tenants", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
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

// Get tenant members
func (c *UTMClient) GetTenantMembers(tenantID string) ([]Member, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/members", tenantID)
	resp, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}
	
	var result struct {
		Data struct {
			Data []Member `json:"data"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Data.Data, nil
}

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	RoleName string `json:"role_name"`
	Status   string `json:"status"`
}
```

#### Usage in Analytics Service

```go
// cmd/analytics/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"your-org/analytics/pkg/client"
)

func main() {
	// Get token from environment
	token := os.Getenv("SYSTEM_USER_TOKEN")
	if token == "" {
		log.Fatal("SYSTEM_USER_TOKEN not set")
	}
	
	// Create client
	client := client.NewUTMClient("https://api.yourdomain.com", token)
	
	// Collect analytics data
	if err := collectAnalytics(client); err != nil {
		log.Fatalf("Analytics collection failed: %v", err)
	}
	
	log.Println("Analytics collection completed successfully")
}

func collectAnalytics(client *client.UTMClient) error {
	// Get all active tenants
	tenants, err := client.GetTenants()
	if err != nil {
		return fmt.Errorf("failed to get tenants: %w", err)
	}
	
	log.Printf("Found %d tenants", len(tenants))
	
	// Collect data for each tenant
	for _, tenant := range tenants {
		if tenant.Status != "active" {
			continue
		}
		
		log.Printf("Processing tenant: %s", tenant.Name)
		
		// Get members
		members, err := client.GetTenantMembers(tenant.ID)
		if err != nil {
			log.Printf("Warning: Failed to get members for %s: %v", tenant.ID, err)
			continue
		}
		
		// Store analytics
		analytics := TenantAnalytics{
			TenantID:    tenant.ID,
			TenantName:  tenant.Name,
			MemberCount: len(members),
			ActiveMembers: countActiveMembers(members),
			CollectedAt: time.Now(),
		}
		
		if err := storeAnalytics(analytics); err != nil {
			log.Printf("Warning: Failed to store analytics for %s: %v", tenant.ID, err)
		}
	}
	
	return nil
}

func countActiveMembers(members []client.Member) int {
	count := 0
	for _, m := range members {
		if m.Status == "active" {
			count++
		}
	}
	return count
}

type TenantAnalytics struct {
	TenantID      string
	TenantName    string
	MemberCount   int
	ActiveMembers int
	CollectedAt   time.Time
}

func storeAnalytics(analytics TenantAnalytics) error {
	// Store in your analytics database
	fmt.Printf("Storing analytics: %+v\n", analytics)
	return nil
}
```

### Step 4: Node.js Implementation

```typescript
// src/client/utm-client.ts
import axios, { AxiosInstance } from 'axios';

export class UTMClient {
  private client: AxiosInstance;
  
  constructor(baseURL: string, token: string) {
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
    const { data } = await this.client.get('/api/v1/tenants');
    return data.data.data;
  }
  
  async getTenantMembers(tenantId: string) {
    const { data } = await this.client.get(`/api/v1/tenants/${tenantId}/members`);
    return data.data.data;
  }
  
  async getTenantDetails(tenantId: string) {
    const { data } = await this.client.get(`/api/v1/tenants/${tenantId}`);
    return data.data;
  }
}

// Usage
import { UTMClient } from './client/utm-client';

const client = new UTMClient(
  process.env.API_URL!,
  process.env.SYSTEM_USER_TOKEN!
);

async function collectAnalytics() {
  const tenants = await client.getTenants();
  
  for (const tenant of tenants) {
    if (tenant.status !== 'active') continue;
    
    const members = await client.getTenantMembers(tenant.id);
    
    console.log(`${tenant.name}: ${members.length} members`);
    
    // Store analytics
    await storeAnalytics({
      tenantId: tenant.id,
      tenantName: tenant.name,
      memberCount: members.length,
      collectedAt: new Date()
    });
  }
}

collectAnalytics().catch(console.error);
```

### Step 5: Python Implementation

```python
# utm_client.py
import os
import requests
from typing import List, Dict, Any
from datetime import datetime

class UTMClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })
        self.session.timeout = 30
    
    def get_tenants(self) -> List[Dict[str, Any]]:
        """Get all tenants"""
        response = self.session.get(f'{self.base_url}/api/v1/tenants')
        response.raise_for_status()
        return response.json()['data']['data']
    
    def get_tenant_members(self, tenant_id: str) -> List[Dict[str, Any]]:
        """Get members of a tenant"""
        response = self.session.get(
            f'{self.base_url}/api/v1/tenants/{tenant_id}/members'
        )
        response.raise_for_status()
        return response.json()['data']['data']

# Usage
from utm_client import UTMClient

client = UTMClient(
    os.getenv('API_URL'),
    os.getenv('SYSTEM_USER_TOKEN')
)

def collect_analytics():
    tenants = client.get_tenants()
    
    for tenant in tenants:
        if tenant['status'] != 'active':
            continue
        
        members = client.get_tenant_members(tenant['id'])
        
        print(f"{tenant['name']}: {len(members)} members")
        
        # Store analytics
        store_analytics({
            'tenant_id': tenant['id'],
            'tenant_name': tenant['name'],
            'member_count': len(members),
            'collected_at': datetime.now().isoformat()
        })

if __name__ == '__main__':
    collect_analytics()
```

## Production Deployment

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  analytics:
    build: .
    environment:
      - API_URL=https://api.yourdomain.com
      - SYSTEM_USER_TOKEN_FILE=/run/secrets/system_user_token
    secrets:
      - system_user_token
    restart: unless-stopped

secrets:
  system_user_token:
    file: ./secrets/system_user_token.txt
```

### Kubernetes Deployment

```yaml
# k8s/analytics-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: analytics-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: analytics
  template:
    metadata:
      labels:
        app: analytics
    spec:
      containers:
      - name: analytics
        image: your-org/analytics:latest
        env:
        - name: API_URL
          value: "https://api.yourdomain.com"
        - name: SYSTEM_USER_TOKEN
          valueFrom:
            secretKeyRef:
              name: analytics-secrets
              key: system-user-token
---
apiVersion: v1
kind: Secret
metadata:
  name: analytics-secrets
type: Opaque
stringData:
  system-user-token: "sys_your_token_here"
```

## Error Handling

### Retry Logic

```go
func (c *UTMClient) requestWithRetry(method, path string, body interface{}, maxRetries int) (*http.Response, error) {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		resp, err := c.request(method, path, body)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		
		lastErr = err
		if resp != nil {
			resp.Body.Close()
		}
		
		// Exponential backoff
		time.Sleep(time.Duration(1<<uint(i)) * time.Second)
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### Token Expiry Handling

```go
func (c *UTMClient) checkTokenExpiry() error {
	resp, err := c.request("GET", "/api/v1/system-users/token-info", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result struct {
		Data struct {
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"data"`
	}
	
	json.NewDecoder(resp.Body).Decode(&result)
	
	daysLeft := time.Until(result.Data.ExpiresAt).Hours() / 24
	
	if daysLeft < 7 {
		// Alert: Token expiring soon!
		return fmt.Errorf("token expires in %.0f days - rotate immediately", daysLeft)
	}
	
	return nil
}
```

## Monitoring

### Health Check Endpoint

```go
func (c *UTMClient) HealthCheck() error {
	resp, err := c.request("GET", "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	
	return nil
}
```

### Metrics Collection

```go
import "github.com/prometheus/client_golang/prometheus"

var (
	apiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "utm_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	apiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "utm_api_request_duration_seconds",
			Help: "API request duration",
		},
		[]string{"method", "endpoint"},
	)
)

func (c *UTMClient) instrumentedRequest(method, path string, body interface{}) (*http.Response, error) {
	start := time.Now()
	
	resp, err := c.request(method, path, body)
	
	duration := time.Since(start).Seconds()
	apiRequestDuration.WithLabelValues(method, path).Observe(duration)
	
	if err == nil {
		apiRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", resp.StatusCode)).Inc()
	}
	
	return resp, err
}
```

## Best Practices

1. **Secure Token Storage** - Use secret managers
2. **Implement Retries** - Handle transient failures
3. **Monitor Token Expiry** - Rotate before expiration
4. **Rate Limiting** - Respect API limits
5. **Error Logging** - Log all failures for debugging
6. **Health Checks** - Verify connectivity regularly
7. **Graceful Shutdown** - Handle termination signals

## Related Documentation

- [System Users API](/x-api/system-users) - API reference
- [System Auth Usage](/system-auth/usage) - Authentication details
- [Credential Rotation](/examples/credential-rotation) - Token rotation
