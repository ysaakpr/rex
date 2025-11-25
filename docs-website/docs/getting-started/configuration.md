# Configuration

Complete guide to configuring Rex via environment variables.

## Configuration File

Rex uses environment variables for configuration, loaded from `.env` file.

```bash
# Create from example
cp .env.example .env

# Edit configuration
nano .env
```

## Environment Variables

### App Configuration

```bash
# Application environment
APP_ENV=development          # development | production
APP_PORT=8080               # API server port
LOG_LEVEL=info              # debug | info | warn | error
```

**APP_ENV**:
- `development`: Debug logging, verbose errors, MailHog
- `production`: Production logging, secure cookies, real SMTP

### Database Configuration

```bash
DB_HOST=postgres            # Database host
DB_PORT=5432                # Database port
DB_USER=utmuser            # Database user
DB_PASSWORD=utmpassword     # Database password (change in production!)
DB_NAME=utm_backend         # Database name
DB_SSL_MODE=disable         # disable | require
```

**Production Settings**:
```bash
DB_HOST=your-rds-endpoint.region.rds.amazonaws.com
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=<strong-password>
DB_NAME=utm_production
DB_SSL_MODE=require
```

### SuperTokens Configuration

```bash
# SuperTokens Core connection
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=                    # Optional, recommended for production

# Domains (must match your setup)
API_DOMAIN=http://localhost:8080        # Backend URL
WEBSITE_DOMAIN=http://localhost:3000    # Frontend URL

# Google OAuth (optional)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

**Production Settings**:
```bash
SUPERTOKENS_CONNECTION_URI=https://supertokens.yourdomain.com
SUPERTOKENS_API_KEY=your-api-key-here
API_DOMAIN=https://api.yourdomain.com
WEBSITE_DOMAIN=https://yourdomain.com
```

### Redis Configuration

```bash
REDIS_HOST=redis            # Redis host
REDIS_PORT=6379             # Redis port
REDIS_PASSWORD=             # Redis password (set in production!)
REDIS_DB=0                  # Redis database number
```

**Production Settings**:
```bash
REDIS_HOST=your-elasticache-endpoint.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=<strong-password>
REDIS_DB=0
```

### SMTP/Email Configuration

```bash
# Email settings
SMTP_HOST=mailhog           # SMTP host (mailhog for dev, real SMTP for prod)
SMTP_PORT=1025              # SMTP port (1025 for mailhog, 587 for production)
SMTP_USER=                  # SMTP username
SMTP_PASSWORD=              # SMTP password
SMTP_FROM=noreply@localhost # From address
EMAIL_FROM_NAME=Rex # From name
```

**Production Settings (AWS SES)**:
```bash
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=<your-ses-username>
SMTP_PASSWORD=<your-ses-password>
SMTP_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your Company
```

**Production Settings (SendGrid)**:
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=<your-sendgrid-api-key>
SMTP_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your Company
```

### Asynq (Background Jobs)

```bash
# Job queue configuration
ASYNQ_REDIS_ADDR=redis:6379             # Redis address for Asynq
ASYNQ_REDIS_PASSWORD=                    # Redis password
ASYNQ_REDIS_DB=0                        # Redis DB
ASYNQ_CONCURRENCY=10                    # Worker concurrency
```

### Invitation Configuration

```bash
# Invitation settings
INVITATION_EXPIRY_HOURS=72              # Default: 72 hours (3 days)
INVITATION_BASE_URL=http://localhost:3000/invitations
```

**Production**:
```bash
INVITATION_EXPIRY_HOURS=168             # 7 days
INVITATION_BASE_URL=https://yourdomain.com/invitations
```

## Complete Example Configurations

### Development (.env)

```bash
# App
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=debug

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=utmuser
DB_PASSWORD=utmpassword
DB_NAME=utm_backend
DB_SSL_MODE=disable

# SuperTokens
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=
API_DOMAIN=http://localhost:8080
WEBSITE_DOMAIN=http://localhost:3000

# Google OAuth (optional)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Email (MailHog for development)
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@localhost
EMAIL_FROM_NAME=Rex

# Asynq
ASYNQ_REDIS_ADDR=redis:6379
ASYNQ_REDIS_PASSWORD=
ASYNQ_REDIS_DB=0
ASYNQ_CONCURRENCY=5

# Invitations
INVITATION_EXPIRY_HOURS=72
INVITATION_BASE_URL=http://localhost:3000/invitations
```

### Production (.env)

```bash
# App
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Database (AWS RDS)
DB_HOST=utm-prod.abc123.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=utm_admin
DB_PASSWORD=<strong-random-password>
DB_NAME=utm_production
DB_SSL_MODE=require

# SuperTokens
SUPERTOKENS_CONNECTION_URI=https://supertokens.yourdomain.com
SUPERTOKENS_API_KEY=<your-supertokens-api-key>
API_DOMAIN=https://api.yourdomain.com
WEBSITE_DOMAIN=https://yourdomain.com

# Google OAuth
GOOGLE_CLIENT_ID=123456789-abc123xyz.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=<your-google-client-secret>

# Redis (AWS ElastiCache)
REDIS_HOST=utm-prod.abc123.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=<strong-random-password>
REDIS_DB=0

# Email (AWS SES)
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=<your-ses-smtp-username>
SMTP_PASSWORD=<your-ses-smtp-password>
SMTP_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your Company

# Asynq
ASYNQ_REDIS_ADDR=utm-prod.abc123.cache.amazonaws.com:6379
ASYNQ_REDIS_PASSWORD=<strong-random-password>
ASYNQ_REDIS_DB=0
ASYNQ_CONCURRENCY=20

# Invitations
INVITATION_EXPIRY_HOURS=168
INVITATION_BASE_URL=https://yourdomain.com/invitations
```

## Security Best Practices

### 1. Strong Passwords

Use strong, random passwords:

```bash
# Generate strong password
openssl rand -base64 32

# Use in .env
DB_PASSWORD=<generated-password>
```

### 2. Never Commit .env

Ensure `.gitignore` includes:

```
.env
.env.local
.env.production
```

### 3. Use Secrets Manager

For production, use a secrets manager:

**AWS Secrets Manager**:
```bash
aws secretsmanager get-secret-value \
  --secret-id utm-backend/prod \
  --query SecretString \
  --output text > .env
```

**HashiCorp Vault**:
```bash
vault kv get -format=json secret/utm-backend | \
  jq -r '.data.data | to_entries | .[] | "\(.key)=\(.value)"' > .env
```

### 4. Rotate Credentials

Regularly rotate:
- Database passwords (every 90 days)
- API keys (every 90 days)
- SMTP passwords (every 90 days)

## Loading Configuration

Configuration is loaded in `internal/config/config.go`:

```go
func Load() *Config {
    // Load .env file
    viper.SetConfigFile(".env")
    
    // Override with environment variables
    viper.AutomaticEnv()
    
    // Read config
    if err := viper.ReadInConfig(); err != nil {
        log.Printf("Warning: .env file not found: %v", err)
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        log.Fatalf("Failed to unmarshal config: %v", err)
    }
    
    return &cfg
}
```

## Environment-Specific Config

### Override for Testing

```bash
# .env.test
APP_ENV=test
DB_NAME=utm_test
SMTP_HOST=localhost
SMTP_PORT=1025
```

### Override with Docker

```yaml
# docker-compose.override.yml
services:
  api:
    environment:
      - LOG_LEVEL=debug
      - DB_PASSWORD=custom_password
```

## Validation

The configuration is validated on startup:

```go
func (c *Config) Validate() error {
    if c.Database.Host == "" {
        return errors.New("DB_HOST is required")
    }
    if c.SuperTokens.ConnectionURI == "" {
        return errors.New("SUPERTOKENS_CONNECTION_URI is required")
    }
    // ... more validations
    return nil
}
```

## Next Steps

- **[Installation](/getting-started/installation)** - Install the system
- **[Quick Start](/getting-started/quick-start)** - Get started quickly
- **[Deployment](/deployment/docker)** - Deploy to production

