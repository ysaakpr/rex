# Go Implementation

Complete System User Authentication Library for Go applications.

## Installation

```bash
go get github.com/yourorg/utm-backend/pkg/systemauth
```

Or copy the package directly into your project.

## Package Structure

```
pkg/systemauth/
├── vault.go          # Vault interface and implementations
├── client.go         # Main auth client
└── client_test.go    # Tests
```

## Quick Start

```go
package main

import (
    "fmt"
    "io"
    "github.com/yourorg/systemauth"
)

func main() {
    // Create vault (environment variables)
    vault := systemauth.NewEnvVault(
        "SYSTEM_USER_EMAIL",
        "SYSTEM_USER_PASSWORD",
    )
    
    // Create auth client
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: "https://api.yourdomain.com",
    })
    
    // Make authenticated request (login is automatic)
    resp, err := client.MakeAuthenticatedRequest(
        "GET",
        "https://api.yourdomain.com/api/v1/tenants",
        nil,
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

## Vault Interface

### Definition

```go
// SecretVault interface for credential providers
type SecretVault interface {
    GetEmail() (string, error)
    GetPassword() (string, error)
}
```

### EnvVault (Default)

```go
// EnvVault loads credentials from environment variables
type EnvVault struct {
    EmailEnvKey    string
    PasswordEnvKey string
}

func NewEnvVault(emailKey, passwordKey string) *EnvVault {
    return &EnvVault{
        EmailEnvKey:    emailKey,
        PasswordEnvKey: passwordKey,
    }
}

func (v *EnvVault) GetEmail() (string, error) {
    email := os.Getenv(v.EmailEnvKey)
    if email == "" {
        return "", fmt.Errorf("email not found in env: %s", v.EmailEnvKey)
    }
    return email, nil
}

func (v *EnvVault) GetPassword() (string, error) {
    password := os.Getenv(v.PasswordEnvKey)
    if password == "" {
        return "", fmt.Errorf("password not found in env: %s", v.PasswordEnvKey)
    }
    return password, nil
}
```

**Usage**:
```go
vault := systemauth.NewEnvVault("WORKER_EMAIL", "WORKER_PASSWORD")
```

**Environment Setup**:
```bash
export WORKER_EMAIL="background-worker@system.internal"
export WORKER_PASSWORD="sysuser_abc123xyz..."
```

## SystemAuthClient

### Configuration

```go
type SystemAuthConfig struct {
    Vault          SecretVault
    APIURL         string
    RefreshBuffer  time.Duration // Default: 5 minutes before expiry
}
```

### Creating Client

```go
client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
    Vault:         vault,
    APIURL:        "https://api.yourdomain.com",
    RefreshBuffer: 5 * time.Minute, // Refresh 5 min before expiry
})
```

### Making Requests

#### GET Request

```go
resp, err := client.MakeAuthenticatedRequest(
    "GET",
    "https://api.yourdomain.com/api/v1/tenants",
    nil, // No body for GET
)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var result map[string]interface{}
json.NewDecoder(resp.Body).Decode(&result)
fmt.Println(result)
```

#### POST Request

```go
data := map[string]string{
    "name": "New Tenant",
    "slug": "new-tenant",
}
jsonData, _ := json.Marshal(data)

resp, err := client.MakeAuthenticatedRequest(
    "POST",
    "https://api.yourdomain.com/api/v1/tenants",
    bytes.NewBuffer(jsonData),
)
```

#### Custom Headers

```go
// For custom headers, get token and make request manually
token, err := client.GetAccessToken()
if err != nil {
    log.Fatal(err)
}

req, _ := http.NewRequest("POST", url, body)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("st-auth-mode", "header")
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-Custom-Header", "value")

httpClient := &http.Client{}
resp, err := httpClient.Do(req)
```

## Advanced Usage

### Manual Token Management

If you need more control:

```go
// Explicit login
err := client.Login()
if err != nil {
    log.Fatal(err)
}

// Get current access token
token, err := client.GetAccessToken()

// Manual refresh
err = client.RefreshAccessToken()
```

### Custom HTTP Client

```go
client := &SystemAuthClient{
    vault:     vault,
    apiURL:    apiURL,
    httpClient: &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
        },
    },
}
```

### Error Handling

```go
resp, err := client.MakeAuthenticatedRequest("GET", url, nil)
if err != nil {
    if strings.Contains(err.Error(), "401") {
        // Authentication failed (shouldn't happen with auto-retry)
        log.Println("Auth failed after retry")
    } else if strings.Contains(err.Error(), "timeout") {
        // Request timeout
        log.Println("Request timed out")
    } else {
        // Other error
        log.Println("Request failed:", err)
    }
    return
}

// Check response status
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    log.Printf("API error %d: %s", resp.StatusCode, string(body))
}
```

## Complete Example: Background Worker

```go
package main

import (
    "encoding/json"
    "io"
    "log"
    "time"
    
    "github.com/yourorg/systemauth"
)

type Job struct {
    ID       string `json:"id"`
    TenantID string `json:"tenant_id"`
    Type     string `json:"type"`
    Data     map[string]interface{} `json:"data"`
}

func main() {
    // Setup auth client
    vault := systemauth.NewEnvVault("WORKER_EMAIL", "WORKER_PASSWORD")
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: "https://api.yourdomain.com",
    })
    
    log.Println("Worker started")
    
    // Worker loop
    for {
        // Fetch pending jobs
        jobs, err := fetchPendingJobs(client)
        if err != nil {
            log.Printf("Error fetching jobs: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }
        
        // Process each job
        for _, job := range jobs {
            if err := processJob(client, job); err != nil {
                log.Printf("Error processing job %s: %v", job.ID, err)
            }
        }
        
        // Wait before next iteration
        time.Sleep(5 * time.Second)
    }
}

func fetchPendingJobs(client *systemauth.SystemAuthClient) ([]Job, error) {
    resp, err := client.MakeAuthenticatedRequest(
        "GET",
        "https://api.yourdomain.com/api/v1/jobs/pending",
        nil,
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Success bool   `json:"success"`
        Data    []Job  `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result.Data, nil
}

func processJob(client *systemauth.SystemAuthClient, job Job) error {
    log.Printf("Processing job %s (type: %s)", job.ID, job.Type)
    
    // Do work...
    time.Sleep(2 * time.Second)
    
    // Mark job as complete
    resp, err := client.MakeAuthenticatedRequest(
        "POST",
        fmt.Sprintf("https://api.yourdomain.com/api/v1/jobs/%s/complete", job.ID),
        nil,
    )
    if err != nil {
        return err
    }
    resp.Body.Close()
    
    log.Printf("Job %s completed", job.ID)
    return nil
}
```

## Example: Scheduled Task

```go
package main

import (
    "bytes"
    "encoding/json"
    "log"
    "time"
    
    "github.com/robfig/cron/v3"
    "github.com/yourorg/systemauth"
)

func main() {
    // Setup auth client
    vault := systemauth.NewEnvVault("CRON_EMAIL", "CRON_PASSWORD")
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: "https://api.yourdomain.com",
    })
    
    // Create cron scheduler
    c := cron.New()
    
    // Run daily at 2am
    c.AddFunc("0 2 * * *", func() {
        log.Println("Starting daily report generation")
        if err := generateDailyReports(client); err != nil {
            log.Printf("Error: %v", err)
        }
    })
    
    // Run hourly
    c.AddFunc("@hourly", func() {
        log.Println("Starting hourly data sync")
        if err := syncData(client); err != nil {
            log.Printf("Error: %v", err)
        }
    })
    
    c.Start()
    
    // Keep running
    select {}
}

func generateDailyReports(client *systemauth.SystemAuthClient) error {
    data := map[string]string{
        "report_type": "daily_summary",
        "date":        time.Now().Format("2006-01-02"),
    }
    jsonData, _ := json.Marshal(data)
    
    resp, err := client.MakeAuthenticatedRequest(
        "POST",
        "https://api.yourdomain.com/api/v1/reports/generate",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    log.Println("Daily report generated")
    return nil
}

func syncData(client *systemauth.SystemAuthClient) error {
    // Implementation...
    return nil
}
```

## Testing

### Mock Vault for Tests

```go
type MockVault struct {
    Email    string
    Password string
}

func (m *MockVault) GetEmail() (string, error) {
    return m.Email, nil
}

func (m *MockVault) GetPassword() (string, error) {
    return m.Password, nil
}

// In tests
func TestWorker(t *testing.T) {
    vault := &MockVault{
        Email:    "test@system.internal",
        Password: "test_password",
    }
    
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: "http://test-api",
    })
    
    // Test your code...
}
```

### Integration Test

```go
func TestIntegration(t *testing.T) {
    // Use real environment variables
    vault := systemauth.NewEnvVault("TEST_EMAIL", "TEST_PASSWORD")
    client := systemauth.NewSystemAuthClient(systemauth.SystemAuthConfig{
        Vault:  vault,
        APIURL: os.Getenv("TEST_API_URL"),
    })
    
    // Test API call
    resp, err := client.MakeAuthenticatedRequest("GET", 
        os.Getenv("TEST_API_URL")+"/api/v1/tenants", nil)
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Deployment

### Docker

**Dockerfile**:
```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o worker ./cmd/worker

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/worker .

CMD ["./worker"]
```

**docker-compose.yml**:
```yaml
services:
  worker:
    build: .
    environment:
      - WORKER_EMAIL=background-worker@system.internal
      - WORKER_PASSWORD=${WORKER_PASSWORD}
      - API_URL=https://api.yourdomain.com
    restart: unless-stopped
```

### Kubernetes

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: background-worker
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: worker
        image: yourorg/worker:latest
        env:
        - name: WORKER_EMAIL
          valueFrom:
            secretKeyRef:
              name: worker-credentials
              key: email
        - name: WORKER_PASSWORD
          valueFrom:
            secretKeyRef:
              name: worker-credentials
              key: password
        - name: API_URL
          value: "https://api.yourdomain.com"
```

## Next Steps

- **[Java Implementation](/system-auth/java)** - Java version of the library
- **[Custom Vaults](/system-auth/custom-vaults)** - Implement AWS, HashiCorp, etc.
- **[Usage Examples](/system-auth/usage)** - More real-world examples

