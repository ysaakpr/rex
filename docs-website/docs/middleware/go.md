# Go Middleware Library

Official Go client library for RBAC authorization middleware.

## Overview

The Go middleware library provides:
- **RBAC Authorization**: Check user permissions before handling requests
- **Tenant Access Control**: Verify user belongs to tenant
- **Platform Admin Checks**: Restrict access to platform admins
- **Session Management**: Extract user information from SuperTokens session
- **Flexible Permission Checking**: Support for single or multiple permissions

## Installation

```bash
# Already included in the backend
# Located in: internal/api/middleware/
```

## Quick Start

### 1. Basic Permission Check

```go
import (
    "github.com/gin-gonic/gin"
    "yourproject/internal/api/middleware"
)

func setupRoutes(router *gin.Engine) {
    // Require specific permission
    router.POST("/api/v1/tenants/:tenant_id/posts",
        middleware.RequirePermission("blog-api", "post", "create"),
        handlers.CreatePost,
    )
}
```

### 2. Multiple Routes with Same Permission

```go
posts := router.Group("/api/v1/tenants/:tenant_id/posts")
posts.Use(middleware.RequirePermission("blog-api", "post", "read"))
{
    posts.GET("", handlers.ListPosts)
    posts.GET("/:id", handlers.GetPost)
}
```

### 3. Check Multiple Permissions (ANY)

```go
// User needs ANY ONE of these permissions
router.PUT("/api/v1/tenants/:tenant_id/posts/:id",
    middleware.RequireAnyPermission([]middleware.Permission{
        {Service: "blog-api", Entity: "post", Action: "update"},
        {Service: "blog-api", Entity: "post", Action: "admin"},
    }),
    handlers.UpdatePost,
)
```

## Available Middleware

### RequirePermission

Check single permission.

**Signature**:
```go
func RequirePermission(service, entity, action string) gin.HandlerFunc
```

**Usage**:
```go
router.POST("/api/v1/tenants/:tenant_id/members",
    middleware.RequirePermission("tenant-api", "member", "invite"),
    handlers.AddMember,
)
```

**Behavior**:
- Extracts `tenant_id` from URL params
- Gets `user_id` from SuperTokens session
- Checks if user has permission in tenant
- Returns 403 if unauthorized

### RequireAnyPermission

Check if user has ANY of the specified permissions.

**Signature**:
```go
func RequireAnyPermission(permissions []Permission) gin.HandlerFunc

type Permission struct {
    Service string
    Entity  string
    Action  string
}
```

**Usage**:
```go
router.DELETE("/api/v1/tenants/:tenant_id/posts/:id",
    middleware.RequireAnyPermission([]middleware.Permission{
        {Service: "blog-api", Entity: "post", Action: "delete"},
        {Service: "blog-api", Entity: "*", Action: "admin"},  // Super admin
    }),
    handlers.DeletePost,
)
```

### RequireTenantAccess

Verify user is a member of the tenant (any role).

**Signature**:
```go
func RequireTenantAccess() gin.HandlerFunc
```

**Usage**:
```go
router.GET("/api/v1/tenants/:tenant_id/dashboard",
    middleware.RequireTenantAccess(),
    handlers.GetDashboard,
)
```

**Behavior**:
- Checks if user is an active member of the tenant
- Returns 403 if not a member
- Does NOT check specific permissions

### RequirePlatformAdmin

Restrict access to platform administrators only.

**Signature**:
```go
func RequirePlatformAdmin() gin.HandlerFunc
```

**Usage**:
```go
platformRoutes := router.Group("/api/v1/platform")
platformRoutes.Use(middleware.RequirePlatformAdmin())
{
    platformRoutes.POST("/roles", handlers.CreateRole)
    platformRoutes.POST("/permissions", handlers.CreatePermission)
}
```

**Behavior**:
- Checks if user is in `platform_admins` table
- Returns 403 if not a platform admin
- Bypasses tenant-level authorization

## Advanced Usage

### Combine Multiple Middleware

```go
// Require tenant access AND specific permission
router.POST("/api/v1/tenants/:tenant_id/posts",
    middleware.RequireTenantAccess(),
    middleware.RequirePermission("blog-api", "post", "create"),
    handlers.CreatePost,
)
```

### Custom Authorization Logic

```go
func CustomAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.Param("tenant_id")
        sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
        userID := sessionContainer.GetUserID()
        
        // Custom logic
        isOwner, err := checkIfOwner(userID, tenantID)
        if err != nil || !isOwner {
            c.JSON(403, gin.H{"error": "Not the owner"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Extract User Info in Handler

```go
func CreatePostHandler(c *gin.Context) {
    // Get user ID from session
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    userID := sessionContainer.GetUserID()
    
    // Get tenant ID from params
    tenantID := c.Param("tenant_id")
    
    // Your handler logic
    post := models.Post{
        TenantID: tenantID,
        AuthorID: userID,
        // ...
    }
    
    db.Create(&post)
    c.JSON(201, gin.H{"success": true, "data": post})
}
```

## Permission Patterns

### Standard CRUD

```go
posts := router.Group("/api/v1/tenants/:tenant_id/posts")
{
    posts.POST("",
        middleware.RequirePermission("blog-api", "post", "create"),
        handlers.CreatePost,
    )
    posts.GET("",
        middleware.RequirePermission("blog-api", "post", "read"),
        handlers.ListPosts,
    )
    posts.GET("/:id",
        middleware.RequirePermission("blog-api", "post", "read"),
        handlers.GetPost,
    )
    posts.PUT("/:id",
        middleware.RequirePermission("blog-api", "post", "update"),
        handlers.UpdatePost,
    )
    posts.DELETE("/:id",
        middleware.RequirePermission("blog-api", "post", "delete"),
        handlers.DeletePost,
    )
}
```

### Admin-Only Routes

```go
admin := router.Group("/api/v1/tenants/:tenant_id/admin")
admin.Use(middleware.RequirePermission("tenant-api", "admin", "access"))
{
    admin.GET("/analytics", handlers.GetAnalytics)
    admin.POST("/bulk-import", handlers.BulkImport)
}
```

### Public Routes (No Auth)

```go
// No middleware = public access
router.GET("/api/v1/public/posts", handlers.ListPublicPosts)
router.GET("/api/v1/health", handlers.HealthCheck)
```

## Error Handling

### Permission Denied Response

```json
{
  "success": false,
  "error": "Permission denied: blog-api:post:create"
}
```

**Status Code**: `403 Forbidden`

### Unauthorized (No Session)

```json
{
  "message": "Unauthorized"
}
```

**Status Code**: `401 Unauthorized`

### Tenant Access Denied

```json
{
  "success": false,
  "error": "Access denied: Not a member of this tenant"
}
```

**Status Code**: `403 Forbidden`

## Testing

### Unit Test Example

```go
package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "yourproject/internal/api/middleware"
)

func TestRequirePermission(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    router.GET("/test/:tenant_id",
        middleware.RequirePermission("test-api", "resource", "read"),
        func(c *gin.Context) {
            c.JSON(200, gin.H{"success": true})
        },
    )
    
    // Test with authorized user
    req := httptest.NewRequest("GET", "/test/tenant-123", nil)
    // Add session cookies/headers
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

### Integration Test

```go
func TestCreatePost_Integration(t *testing.T) {
    // Setup test database and services
    db := setupTestDB()
    defer db.Close()
    
    // Create test user with permission
    user := createTestUser(db)
    role := createRoleWithPermission(db, "blog-api:post:create")
    assignRoleToUser(db, user.ID, role.ID)
    
    // Make authenticated request
    router := setupRouter(db)
    req := httptest.NewRequest("POST", "/api/v1/tenants/test-tenant/posts", body)
    addSessionCookie(req, user.SessionToken)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

## Best Practices

### 1. Apply Middleware at Route Group Level

```go
// ✅ Good: Apply once to entire group
api := router.Group("/api/v1/tenants/:tenant_id")
api.Use(middleware.RequireTenantAccess())
{
    api.GET("/dashboard", handlers.Dashboard)
    api.GET("/settings", handlers.Settings)
}

// ❌ Bad: Repeat for each route
router.GET("/api/v1/tenants/:tenant_id/dashboard",
    middleware.RequireTenantAccess(),
    handlers.Dashboard,
)
router.GET("/api/v1/tenants/:tenant_id/settings",
    middleware.RequireTenantAccess(),
    handlers.Settings,
)
```

### 2. Order Middleware Correctly

```go
// Correct order: Session → Tenant Access → Permissions
router.POST("/api/v1/tenants/:tenant_id/posts",
    middleware.AuthMiddleware(),           // 1. Verify session
    middleware.RequireTenantAccess(),      // 2. Check tenant membership
    middleware.RequirePermission(...),     // 3. Check specific permission
    handlers.CreatePost,                   // 4. Handler
)
```

### 3. Use Specific Permissions

```go
// ✅ Good: Specific permission
middleware.RequirePermission("blog-api", "post", "publish")

// ❌ Bad: Too broad
middleware.RequirePermission("blog-api", "*", "*")
```

### 4. Document Required Permissions

```go
// CreatePost creates a new blog post
// Required permission: blog-api:post:create
// Required role: Writer or Admin
func CreatePost(c *gin.Context) {
    // ...
}
```

## Performance Optimization

### Cache Permission Checks

The middleware automatically uses the RBAC service which can implement caching:

```go
// In your RBAC service
type RBACService struct {
    cache *cache.Cache
    ttl   time.Duration
}

func (s *RBACService) CheckUserPermission(userID, tenantID, service, entity, action string) (bool, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("perm:%s:%s:%s:%s:%s", userID, tenantID, service, entity, action)
    if cached, found := s.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }
    
    // Check database
    authorized := s.checkDatabase(...)
    
    // Cache result
    s.cache.Set(cacheKey, authorized, s.ttl)
    
    return authorized, nil
}
```

### Batch Permission Checks

For handlers that need multiple permissions:

```go
func ComplexHandler(c *gin.Context) {
    permissions := []Permission{
        {Service: "blog-api", Entity: "post", Action: "create"},
        {Service: "media-api", Entity: "image", Action: "upload"},
    }
    
    // Batch check (implement in service)
    authorized, err := rbacService.CheckMultiplePermissions(userID, tenantID, permissions)
    if err != nil || !authorized {
        c.JSON(403, gin.H{"error": "Insufficient permissions"})
        return
    }
    
    // Handle request
}
```

## Troubleshooting

### Middleware Not Applied

**Issue**: Routes accessible without authentication

**Solution**: Ensure middleware is applied:
```go
// Check middleware order
router.Use(middleware.AuthMiddleware())  // Must be before routes
```

### Permission Always Denied

**Debug**:
```go
func DebugPermissionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.Param("tenant_id")
        userID := getUserID(c)
        
        // Log permission check
        log.Printf("Checking permission for user=%s, tenant=%s", userID, tenantID)
        
        // Get user's actual permissions
        perms, _ := rbacService.GetUserPermissions(userID, tenantID)
        log.Printf("User permissions: %v", perms)
        
        c.Next()
    }
}
```

### 401 vs 403 Errors

- **401 Unauthorized**: No valid session (SuperTokens auth failed)
  - Solution: User needs to sign in
  
- **403 Forbidden**: Valid session but insufficient permissions
  - Solution: User needs appropriate role/permission

## Related Documentation

- [RBAC Overview](/guides/rbac-overview) - Authorization system
- [Permissions](/guides/permissions) - Permission management
- [Backend Integration](/guides/backend-integration) - Full backend guide
- [API Reference](/x-api/rbac) - RBAC API endpoints
