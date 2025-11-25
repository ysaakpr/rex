# Security Best Practices

Comprehensive security guide for the Rex.

## Overview

This guide covers security best practices across all layers:
- Authentication and session management
- Authorization and RBAC
- Data protection
- API security
- Infrastructure security
- Monitoring and auditing

## Authentication Security

### Session Management

**Use SuperTokens Best Practices**:

```go
// Production session configuration
session.Init(&sessmodels.TypeInput{
    CookieSecure:   ptrBool(true),  // HTTPS only
    CookieSameSite: ptrString("lax"),  // CSRF protection
    CookieDomain:   ptrString(".yourdomain.com"),
    
    SessionExpiredStatusCode: ptrInt(401),
    AntiCsrf: ptrString("VIA_TOKEN"),  // Enable CSRF protection
    
    GetTokenTransferMethod: func(req *http.Request, forCreateNewSession bool, userContext supertokens.UserContext) sessmodels.TokenTransferMethod {
        return sessmodels.CookieTransferMethod  // Prefer cookies over headers
    },
})
```

### Password Security

**Enforce Strong Passwords** (frontend):
```javascript
EmailPassword.init({
  signUpFeature: {
    formFields: [
      {
        id: "password",
        validate: async (value) => {
          // Minimum 12 characters
          if (value.length < 12) {
            return "Password must be at least 12 characters";
          }
          
          // Require uppercase
          if (!/[A-Z]/.test(value)) {
            return "Password must contain uppercase letter";
          }
          
          // Require lowercase
          if (!/[a-z]/.test(value)) {
            return "Password must contain lowercase letter";
          }
          
          // Require number
          if (!/[0-9]/.test(value)) {
            return "Password must contain number";
          }
          
          // Require special character
          if (!/[^A-Za-z0-9]/.test(value)) {
            return "Password must contain special character";
          }
          
          // Check against common passwords
          if (isCommonPassword(value)) {
            return "Password is too common";
          }
          
          return undefined;
        }
      }
    ]
  }
});
```

### Email Verification

**Enforce Verification**:
```go
func RequireEmailVerification() gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        // Get user from SuperTokens
        user, err := emailpassword.GetUserByID(userID)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to get user"})
            c.Abort()
            return
        }
        
        if !user.Email.IsVerified {
            c.JSON(403, gin.H{
                "success": false,
                "error": "Email verification required",
                "code": "EMAIL_NOT_VERIFIED"
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Apply to protected routes
protected.Use(middleware.RequireEmailVerification())
```

### Multi-Factor Authentication (MFA)

**Configure SuperTokens MFA** (future enhancement):
```javascript
// Enable MFA in dashboard or via SDK
Session.init({
  override: {
    functions: (originalImplementation) => {
      return {
        ...originalImplementation,
        createNewSession: async function (input) {
          // Check if MFA required
          if (requiresMFA(input.userId)) {
            // Redirect to MFA setup/verification
          }
          return originalImplementation.createNewSession(input);
        }
      };
    }
  }
});
```

## Authorization Security

### Principle of Least Privilege

**Grant minimum required permissions**:

```javascript
// ✅ Good: Specific permissions
const writerPolicy = {
  name: "Content Writer",
  permissions: [
    "blog-api:post:create",
    "blog-api:post:update",
    "blog-api:post:read"
  ]
};

// ❌ Bad: Overly broad
const badPolicy = {
  name: "Writer",
  permissions: [
    "blog-api:*:*",  // Too broad!
    "tenant-api:*:*"
  ]
};
```

### Permission Checks

**Always verify permissions before sensitive operations**:

```go
func DeletePostHandler(c *gin.Context) {
    postID := c.Param("id")
    tenantID := c.Param("tenant_id")
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    userID := sessionContainer.GetUserID()
    
    // 1. Check if user has permission
    authorized, err := rbacService.CheckUserPermission(
        userID, tenantID, "blog-api", "post", "delete",
    )
    if !authorized || err != nil {
        c.JSON(403, gin.H{"error": "Permission denied"})
        return
    }
    
    // 2. Check ownership (additional check)
    post, _ := postService.GetPost(postID)
    if post.AuthorID != userID && !isAdmin(userID, tenantID) {
        c.JSON(403, gin.H{"error": "You can only delete your own posts"})
        return
    }
    
    // 3. Proceed with deletion
    postService.DeletePost(postID)
}
```

### Prevent Privilege Escalation

**Restrict role assignment**:

```go
func AssignRoleHandler(c *gin.Context) {
    var input struct {
        UserID string `json:"user_id"`
        RoleID string `json:"role_id"`
    }
    c.BindJSON(&input)
    
    // Get role details
    role, _ := rbacService.GetRole(input.RoleID)
    
    // Prevent non-admins from assigning admin role
    if role.Name == "Admin" {
        currentUserIsAdmin := isAdmin(getUserID(c), getTenantID(c))
        if !currentUserIsAdmin {
            c.JSON(403, gin.H{"error": "Only admins can assign admin role"})
            return
        }
    }
    
    // Proceed with assignment
}
```

## Data Protection

### Encryption at Rest

**Database Encryption**:
```bash
# Enable RDS encryption
aws rds modify-db-instance \
  --db-instance-identifier utm-db \
  --storage-encrypted \
  --apply-immediately
```

**Application-Level Encryption** (for sensitive fields):
```go
import "golang.org/x/crypto/nacl/secretbox"

func EncryptSensitiveData(plaintext, key []byte) ([]byte, error) {
    var secretKey [32]byte
    copy(secretKey[:], key)
    
    var nonce [24]byte
    if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
        return nil, err
    }
    
    encrypted := secretbox.Seal(nonce[:], plaintext, &nonce, &secretKey)
    return encrypted, nil
}
```

### Encryption in Transit

**Enforce HTTPS**:
```go
// Redirect HTTP to HTTPS
router.Use(func(c *gin.Context) {
    if c.Request.Header.Get("X-Forwarded-Proto") != "https" && 
       os.Getenv("APP_ENV") == "production" {
        c.Redirect(301, "https://"+c.Request.Host+c.Request.RequestURI)
        c.Abort()
        return
    }
    c.Next()
})
```

**Database SSL**:
```go
dsn := fmt.Sprintf(
    "host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
    config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort,
)
```

### Sensitive Data Handling

**Mask Sensitive Data in Logs**:
```go
type User struct {
    Email    string `json:"email" log:"mask"`
    Password string `json:"-" log:"omit"`  // Never log
    APIKey   string `json:"api_key" log:"mask"`
}

func (u *User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("email", maskEmail(u.Email)),
        // Don't log password or API key
    )
}
```

**Sanitize User Input**:
```go
import "html"

func SanitizeInput(input string) string {
    // HTML escape
    sanitized := html.EscapeString(input)
    
    // Remove null bytes
    sanitized = strings.ReplaceAll(sanitized, "\x00", "")
    
    return strings.TrimSpace(sanitized)
}
```

## API Security

### Rate Limiting

**Implement rate limiting**:

```go
import "github.com/ulule/limiter/v3"
import "github.com/ulule/limiter/v3/drivers/middleware/gin"
import "github.com/ulule/limiter/v3/drivers/store/memory"

func SetupRateLimiting() gin.HandlerFunc {
    rate := limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  60,  // 60 requests per minute
    }
    
    store := memory.NewStore()
    instance := limiter.New(store, rate)
    
    return middleware.NewMiddleware(instance)
}

// Apply to routes
router.Use(SetupRateLimiting())
```

**Different limits for different endpoints**:
```go
// Authentication endpoints: stricter limits
authRoutes.Use(RateLimitMiddleware(5, time.Minute))  // 5/minute

// API endpoints: standard limits
apiRoutes.Use(RateLimitMiddleware(60, time.Minute))  // 60/minute

// Public endpoints: more lenient
publicRoutes.Use(RateLimitMiddleware(100, time.Minute))  // 100/minute
```

### Input Validation

**Always validate input**:
```go
type CreatePostInput struct {
    Title   string `json:"title" binding:"required,min=3,max=200"`
    Content string `json:"content" binding:"required,min=10"`
    Tags    []string `json:"tags" binding:"max=10,dive,max=50"`
}

func (h *PostHandler) CreatePost(c *gin.Context) {
    var input CreatePostInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "Validation failed", "details": err.Error()})
        return
    }
    
    // Additional validation
    if containsHTMLTags(input.Content) {
        c.JSON(400, gin.H{"error": "HTML tags not allowed"})
        return
    }
    
    // Process...
}
```

### SQL Injection Prevention

**Use parameterized queries** (GORM does this automatically):
```go
// ✅ Safe: Parameterized
db.Where("email = ?", email).First(&user)

// ❌ Dangerous: String concatenation
db.Where(fmt.Sprintf("email = '%s'", email)).First(&user)  // DON'T DO THIS!
```

### CORS Configuration

**Restrict origins**:
```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{
        "https://app.yourdomain.com",
        "https://www.yourdomain.com",
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))

// Development only
if os.Getenv("APP_ENV") == "development" {
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"*"},
        AllowHeaders:     []string{"*"},
        AllowCredentials: true,
    }))
}
```

## Infrastructure Security

### Network Security

**Security Groups** (AWS):
```hcl
# Application security group
resource "aws_security_group" "app" {
  name = "utm-app-sg"
  
  # Allow HTTPS from internet
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  # Allow HTTP (redirect to HTTPS)
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Database security group
resource "aws_security_group" "db" {
  name = "utm-db-sg"
  
  # Only from application
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }
  
  # No outbound needed
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

### Secrets Management

**Use AWS Secrets Manager**:
```bash
# Store secrets
aws secretsmanager create-secret \
  --name utm/production/db-password \
  --secret-string "$(openssl rand -base64 32)"

# Rotate automatically
aws secretsmanager rotate-secret \
  --secret-id utm/production/db-password \
  --rotation-lambda-arn <lambda-arn> \
  --rotation-rules AutomaticallyAfterDays=90
```

**Never hardcode secrets**:
```go
// ✅ Good: From environment
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ Bad: Hardcoded
dbPassword := "my-password"  // DON'T DO THIS!
```

### Container Security

**Run as non-root user**:
```dockerfile
# Dockerfile
FROM golang:1.21 AS builder
# ... build steps ...

FROM alpine:latest
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /app/main /app/main
CMD ["/app/main"]
```

**Scan images for vulnerabilities**:
```bash
# Using Trivy
docker run aquasec/trivy image your-image:latest

# Using AWS ECR scanning
aws ecr start-image-scan --repository-name utm-api --image-id imageTag=latest
```

## Monitoring and Auditing

### Audit Logging

**Log all security-relevant events**:

```go
type AuditLog struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID      string    `gorm:"not null"`
    TenantID    uuid.UUID `gorm:"type:uuid"`
    Action      string    `gorm:"not null"`
    Resource    string    `gorm:"not null"`
    ResourceID  string
    IPAddress   string
    UserAgent   string
    Success     bool
    ErrorReason string
    Metadata    datatypes.JSON
    CreatedAt   time.Time
}

func LogAudit(c *gin.Context, action, resource, resourceID string, success bool, err error) {
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    
    log := AuditLog{
        ID:         uuid.New(),
        UserID:     sessionContainer.GetUserID(),
        TenantID:   getTenantID(c),
        Action:     action,
        Resource:   resource,
        ResourceID: resourceID,
        IPAddress:  c.ClientIP(),
        UserAgent:  c.Request.UserAgent(),
        Success:    success,
        CreatedAt:  time.Now(),
    }
    
    if err != nil {
        log.ErrorReason = err.Error()
    }
    
    db.Create(&log)
}

// Usage
func DeleteTenantHandler(c *gin.Context) {
    tenantID := c.Param("id")
    
    err := tenantService.DeleteTenant(tenantID)
    
    LogAudit(c, "DELETE", "tenant", tenantID, err == nil, err)
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"success": true})
}
```

### Security Monitoring

**Alert on suspicious activity**:
```go
func DetectSuspiciousActivity(userID, action string) {
    // Check for unusual patterns
    recentActions := getRecentActions(userID, 5*time.Minute)
    
    // Too many failed attempts
    if countFailed(recentActions) > 5 {
        alertSecurityTeam("Multiple failed attempts", userID)
        lockAccount(userID, 15*time.Minute)
    }
    
    // Unusual access patterns
    if isUnusualTime(userID) || isUnusualLocation(userID) {
        alertSecurityTeam("Unusual access pattern", userID)
        requireMFA(userID)
    }
}
```

## Compliance

### GDPR Compliance

**Right to Access**:
```go
func ExportUserData(userID string) ([]byte, error) {
    // Collect all user data
    userData := make(map[string]interface{})
    
    // Profile
    user, _ := userService.GetUser(userID)
    userData["profile"] = user
    
    // Tenants
    tenants, _ := tenantService.GetUserTenants(userID)
    userData["tenants"] = tenants
    
    // Activity logs
    logs, _ := auditService.GetUserLogs(userID)
    userData["activity"] = logs
    
    // Convert to JSON
    return json.MarshalIndent(userData, "", "  ")
}
```

**Right to Deletion**:
```go
func DeleteUserData(userID string) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // Delete from all tables
        tx.Where("user_id = ?", userID).Delete(&models.TenantMember{})
        tx.Where("user_id = ?", userID).Delete(&models.Invitation{})
        tx.Where("author_id = ?", userID).Delete(&models.Post{})
        
        // Anonymize audit logs (don't delete for compliance)
        tx.Model(&models.AuditLog{}).
            Where("user_id = ?", userID).
            Update("user_id", "deleted_user")
        
        // Delete from SuperTokens
        return emailpassword.DeleteUser(userID)
    })
}
```

## Security Checklist

### Development

- [ ] Never commit secrets to version control
- [ ] Use `.env` files for local development
- [ ] Keep dependencies up to date
- [ ] Run security linters (gosec, npm audit)

### Deployment

- [ ] HTTPS enforced (production)
- [ ] Database SSL enabled
- [ ] Secrets in AWS Secrets Manager
- [ ] Security groups configured
- [ ] Container scanning enabled

### Application

- [ ] Strong password requirements
- [ ] Email verification enforced
- [ ] Rate limiting implemented
- [ ] Input validation on all endpoints
- [ ] RBAC permissions properly configured
- [ ] Audit logging enabled

### Monitoring

- [ ] CloudWatch alarms configured
- [ ] Failed login attempts monitored
- [ ] Suspicious activity alerts
- [ ] Regular security audits scheduled

## Next Steps

- [Production Setup](/deployment/production-setup) - Deploy securely
- [Monitoring](/guides/monitoring) - Set up monitoring
- [RBAC Overview](/guides/rbac-overview) - Authorization system
- [Environment Configuration](/deployment/environment) - Manage secrets

