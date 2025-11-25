# Custom Secret Vaults

Guide to integrating custom secret management solutions for System User tokens.

## Overview

**Status**: 🚧 Planned Feature - Not yet implemented

When available, this feature will allow integration with custom secret management solutions beyond environment variables.

## Supported Vaults (Planned)

- **HashiCorp Vault**
- **AWS Secrets Manager**
- **Azure Key Vault**
- **Google Secret Manager**
- **Custom implementations**

## Current Workaround: Environment Variables

Until custom vault support is implemented, use environment variables:

```bash
# .env
SYSTEM_USER_TOKEN=sys_your_token_here

# Or export
export SYSTEM_USER_TOKEN="sys_your_token_here"
```

### Best Practices for Current Approach

1. **Never commit tokens to git**
2. **Use .env files locally** (in .gitignore)
3. **Use platform secrets in production**:
   - Kubernetes Secrets
   - Docker Secrets
   - Cloud provider secrets

## Planned Implementation

### HashiCorp Vault Integration

```go
// Future implementation
package vault

import (
	vault "github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	Address string
	Token   string
	Path    string
}

func GetSystemUserToken(config VaultConfig) (string, error) {
	client, err := vault.NewClient(&vault.Config{
		Address: config.Address,
	})
	if err != nil {
		return "", err
	}
	
	client.SetToken(config.Token)
	
	secret, err := client.Logical().Read(config.Path)
	if err != nil {
		return "", err
	}
	
	token := secret.Data["token"].(string)
	return token, nil
}
```

### AWS Secrets Manager

```go
// Future implementation
package awssecrets

import (
	"context"
	
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetSystemUserToken(secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	
	client := secretsmanager.NewFromConfig(cfg)
	
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return "", err
	}
	
	return *result.SecretString, nil
}
```

## Interim Solutions

### Kubernetes Secrets

```yaml
# k8s/system-user-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: system-user-token
type: Opaque
stringData:
  token: sys_your_token_here

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-service
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: SYSTEM_USER_TOKEN
          valueFrom:
            secretKeyRef:
              name: system-user-token
              key: token
```

### Docker Secrets

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    image: your-service
    secrets:
      - system_user_token
    environment:
      - SYSTEM_USER_TOKEN_FILE=/run/secrets/system_user_token

secrets:
  system_user_token:
    file: ./secrets/system_user_token.txt
```

```go
// Read from Docker secret file
func getTokenFromFile() string {
	tokenFile := os.Getenv("SYSTEM_USER_TOKEN_FILE")
	if tokenFile == "" {
		return os.Getenv("SYSTEM_USER_TOKEN")
	}
	
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		log.Fatal(err)
	}
	
	return strings.TrimSpace(string(data))
}
```

## Request for Implementation

If you need custom vault support, please:
1. Open a GitHub issue at [https://github.com/ysaakpr/rex](https://github.com/ysaakpr/rex)
2. Describe your use case
3. Specify which vault system you need
4. Consider contributing to implementation

## Related Documentation

- [System Auth Overview](/system-auth/overview) - System authentication
- [System Auth Usage](/system-auth/usage) - Using system auth
- [System Users API](/x-api/system-users) - API reference
- [Security Guide](/guides/security) - Security best practices
- [Credential Rotation](/examples/credential-rotation) - Token rotation
- [GitHub Repository](https://github.com/ysaakpr/rex) - Project repository
