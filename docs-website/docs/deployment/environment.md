# Environment Configuration

Complete guide to configuring environment variables for all environments.

## Overview

The application uses environment variables for configuration across different environments:
- **Development**: `.env` file
- **Production**: AWS Secrets Manager / Environment variables
- **Testing**: `.env.test` file

## Configuration File Location

```
.env                    # Development (not committed)
.env.example            # Template (committed)
.env.test               # Testing (optional)
```

## Required Variables

### Application

```bash
# Environment mode
APP_ENV=development  # development, staging, production
APP_PORT=8080

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

### Database (PostgreSQL)

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_NAME=utm_backend
DB_SSL_MODE=disable  # disable (dev), require (prod)
```

### SuperTokens

```bash
# SuperTokens connection
SUPERTOKENS_CONNECTION_URI=http://localhost:3567
SUPERTOKENS_API_KEY=your_supertokens_api_key_here

# API configuration
SUPERTOKENS_APP_NAME=Rex
SUPERTOKENS_API_DOMAIN=http://localhost:8080
SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:3000
SUPERTOKENS_API_BASE_PATH=/auth
SUPERTOKENS_WEBSITE_BASE_PATH=/auth
```

### Redis

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=  # Optional, leave empty for no password
REDIS_DB=0
```

### JWT

```bash
JWT_SECRET_KEY=your_jwt_secret_key_minimum_32_characters
```

### CORS

```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Logging

```bash
LOG_LEVEL=info  # debug, info, warn, error
LOG_FORMAT=json  # json, console
```

## Optional Variables

### Email (SMTP)

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@yourdomain.com
```

### AWS (Production)

```bash
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key

# S3 for file uploads
S3_BUCKET=your-bucket-name
S3_REGION=us-east-1

# CloudWatch Logs
CLOUDWATCH_LOG_GROUP=/aws/utm-backend
```

### Google OAuth

```bash
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

### Feature Flags

```bash
ENABLE_REGISTRATION=true
ENABLE_INVITATIONS=true
ENABLE_GOOGLE_AUTH=true
MAX_TENANTS_PER_USER=10
```

## Environment-Specific Configurations

### Development (.env)

```bash
APP_ENV=development
APP_PORT=8080
FRONTEND_URL=http://localhost:3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=utm_backend
DB_SSL_MODE=disable

SUPERTOKENS_CONNECTION_URI=http://localhost:3567
SUPERTOKENS_API_KEY=test_api_key
SUPERTOKENS_API_DOMAIN=http://localhost:8080
SUPERTOKENS_WEBSITE_DOMAIN=http://localhost:3000

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET_KEY=development_jwt_secret_key_change_in_production

CORS_ALLOWED_ORIGINS=http://localhost:3000

LOG_LEVEL=debug
LOG_FORMAT=console
```

### Staging

```bash
APP_ENV=staging
APP_PORT=8080
FRONTEND_URL=https://staging.yourdomain.com

DB_HOST=staging-db.xxxxx.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=${STAGING_DB_PASSWORD}  # From Secrets Manager
DB_NAME=utm_backend_staging
DB_SSL_MODE=require

SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=${STAGING_SUPERTOKENS_API_KEY}
SUPERTOKENS_API_DOMAIN=https://api-staging.yourdomain.com
SUPERTOKENS_WEBSITE_DOMAIN=https://staging.yourdomain.com

REDIS_HOST=staging-redis.xxxxx.cache.amazonaws.com
REDIS_PORT=6379

JWT_SECRET_KEY=${STAGING_JWT_SECRET}

CORS_ALLOWED_ORIGINS=https://staging.yourdomain.com

LOG_LEVEL=info
LOG_FORMAT=json

# AWS
AWS_REGION=us-east-1
CLOUDWATCH_LOG_GROUP=/aws/utm-backend/staging
```

### Production

```bash
APP_ENV=production
APP_PORT=8080
FRONTEND_URL=https://app.yourdomain.com

DB_HOST=prod-db.xxxxx.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=${PROD_DB_PASSWORD}
DB_NAME=utm_backend
DB_SSL_MODE=require

SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=${PROD_SUPERTOKENS_API_KEY}
SUPERTOKENS_API_DOMAIN=https://api.yourdomain.com
SUPERTOKENS_WEBSITE_DOMAIN=https://app.yourdomain.com

REDIS_HOST=prod-redis.xxxxx.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=${PROD_REDIS_PASSWORD}

JWT_SECRET_KEY=${PROD_JWT_SECRET}

CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com

LOG_LEVEL=warn
LOG_FORMAT=json

# AWS
AWS_REGION=us-east-1
CLOUDWATCH_LOG_GROUP=/aws/utm-backend/production

# Feature flags
ENABLE_REGISTRATION=true
ENABLE_INVITATIONS=true
MAX_TENANTS_PER_USER=50
```

## Loading Configuration

### Go Application

```go
// internal/config/config.go
package config

import (
    "os"
    "strconv"
    
    "github.com/joho/godotenv"
)

type Config struct {
    AppEnv      string
    AppPort     string
    FrontendURL string
    
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    DBSSLMode  string
    
    SuperTokensConnectionURI string
    SuperTokensAPIKey        string
    SuperTokensAPIDomain     string
    SuperTokensWebsiteDomain string
    
    RedisHost     string
    RedisPort     string
    RedisPassword string
    
    JWTSecretKey string
    
    CORSAllowedOrigins []string
    
    LogLevel  string
    LogFormat string
}

func Load() (*Config, error) {
    // Load .env file (development)
    _ = godotenv.Load()
    
    return &Config{
        AppEnv:      getEnv("APP_ENV", "development"),
        AppPort:     getEnv("APP_PORT", "8080"),
        FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
        
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "utm_backend"),
        DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
        
        SuperTokensConnectionURI: getEnv("SUPERTOKENS_CONNECTION_URI", ""),
        SuperTokensAPIKey:        getEnv("SUPERTOKENS_API_KEY", ""),
        SuperTokensAPIDomain:     getEnv("SUPERTOKENS_API_DOMAIN", ""),
        SuperTokensWebsiteDomain: getEnv("SUPERTOKENS_WEBSITE_DOMAIN", ""),
        
        RedisHost:     getEnv("REDIS_HOST", "localhost"),
        RedisPort:     getEnv("REDIS_PORT", "6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        
        JWTSecretKey: getEnv("JWT_SECRET_KEY", ""),
        
        LogLevel:  getEnv("LOG_LEVEL", "info"),
        LogFormat: getEnv("LOG_FORMAT", "console"),
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}
```

### Frontend (React/Vite)

```javascript
// frontend/.env.development
VITE_API_URL=http://localhost:8080
VITE_APP_NAME=Rex
VITE_ENABLE_DEBUG=true
```

```javascript
// frontend/.env.production
VITE_API_URL=https://api.yourdomain.com
VITE_APP_NAME=Rex
VITE_ENABLE_DEBUG=false
```

**Access in code**:
```javascript
const API_URL = import.meta.env.VITE_API_URL;
const APP_NAME = import.meta.env.VITE_APP_NAME;
const DEBUG = import.meta.env.VITE_ENABLE_DEBUG === 'true';
```

## AWS Secrets Manager

### Store Secrets

```bash
# Store database password
aws secretsmanager create-secret \
  --name utm/production/db-password \
  --secret-string "your-secure-password"

# Store SuperTokens API key
aws secretsmanager create-secret \
  --name utm/production/supertokens-api-key \
  --secret-string "your-api-key"

# Store JWT secret
aws secretsmanager create-secret \
  --name utm/production/jwt-secret \
  --secret-string "your-jwt-secret"
```

### Retrieve in Application

```go
// internal/config/secrets.go
package config

import (
    "context"
    "encoding/json"
    
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func LoadFromSecretsManager() (*Config, error) {
    ctx := context.Background()
    
    // Load AWS config
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }
    
    client := secretsmanager.NewFromConfig(cfg)
    
    // Retrieve secrets
    dbPassword, _ := getSecret(ctx, client, "utm/production/db-password")
    stApiKey, _ := getSecret(ctx, client, "utm/production/supertokens-api-key")
    jwtSecret, _ := getSecret(ctx, client, "utm/production/jwt-secret")
    
    return &Config{
        DBPassword:        dbPassword,
        SuperTokensAPIKey: stApiKey,
        JWTSecretKey:      jwtSecret,
        // ... other config from env vars
    }, nil
}

func getSecret(ctx context.Context, client *secretsmanager.Client, secretName string) (string, error) {
    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    })
    if err != nil {
        return "", err
    }
    
    return *result.SecretString, nil
}
```

## Docker Compose

### Environment File

```yaml
# docker-compose.yml
services:
  api:
    build: .
    env_file:
      - .env
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
```

### Pass-through Variables

```yaml
services:
  api:
    environment:
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET_KEY=${JWT_SECRET_KEY}
```

## Validation

### Validate on Startup

```go
func (c *Config) Validate() error {
    if c.DBPassword == "" {
        return errors.New("DB_PASSWORD is required")
    }
    
    if c.SuperTokensAPIKey == "" {
        return errors.New("SUPERTOKENS_API_KEY is required")
    }
    
    if c.JWTSecretKey == "" || len(c.JWTSecretKey) < 32 {
        return errors.New("JWT_SECRET_KEY must be at least 32 characters")
    }
    
    if c.AppEnv == "production" && c.DBSSLMode != "require" {
        return errors.New("DB_SSL_MODE must be 'require' in production")
    }
    
    return nil
}

// In main.go
func main() {
    config, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    if err := config.Validate(); err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    // Continue with app initialization
}
```

## Security Best Practices

### 1. Never Commit Secrets

```bash
# .gitignore
.env
.env.local
.env.*.local
*.key
*.pem
```

### 2. Use Strong Secrets

```bash
# Generate strong secrets
openssl rand -base64 32  # JWT secret
openssl rand -base64 24  # Database password
openssl rand -base64 32  # SuperTokens API key
```

### 3. Rotate Secrets Regularly

```bash
# Rotate database password
aws rds modify-db-instance \
  --db-instance-identifier utm-db \
  --master-user-password "new-password"

# Update secret
aws secretsmanager update-secret \
  --secret-id utm/production/db-password \
  --secret-string "new-password"

# Restart application
aws ecs update-service --force-new-deployment \
  --cluster utm-cluster \
  --service utm-api
```

### 4. Restrict Access

```json
// IAM policy for secrets
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:*:*:secret:utm/production/*"
      ]
    }
  ]
}
```

## Troubleshooting

### Missing Environment Variable

**Symptom**: Application fails to start with "required variable not set"

**Solution**:
```bash
# Check if variable is set
echo $DB_PASSWORD

# Check .env file
cat .env | grep DB_PASSWORD

# Verify loading
docker-compose config  # Shows resolved config
```

### Incorrect Value Type

**Symptom**: Panic or parsing errors

**Solution**:
```go
// Add type validation
if _, err := strconv.Atoi(config.DBPort); err != nil {
    log.Fatal("DB_PORT must be a number")
}
```

### Environment Variable Not Updating

**Symptom**: Changes to .env not reflected

**Solution**:
```bash
# Restart services
docker-compose restart

# Or rebuild
docker-compose down
docker-compose up --build
```

## Next Steps

- [Production Setup](/deployment/production-setup) - Deploy to production
- [Docker Deployment](/deployment/docker) - Docker configuration
- [AWS Deployment](/deployment/aws) - AWS infrastructure
- [Security Guide](/guides/security) - Security best practices
