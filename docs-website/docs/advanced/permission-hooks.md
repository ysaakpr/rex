# Permission Hooks

Custom logic hooks for permission evaluation.

## Overview

Permission hooks allow you to inject custom logic into the permission evaluation process, enabling dynamic authorization based on runtime conditions.

**Status**: 🚧 Planned - Not yet implemented

**When available**, this system will support:
- Pre-authorization hooks (before permission check)
- Post-authorization hooks (after permission check)
- Custom permission logic
- Context-aware authorization
- Resource-level permissions
- Dynamic permission evaluation

## Planned Features

### Hook Types

```typescript
// Future hook types
type HookType =
  | 'pre_authorization'   // Before permission check
  | 'post_authorization'  // After permission check  
  | 'permission_denied'   // When permission is denied
  | 'permission_granted'; // When permission is granted
```

### Pre-Authorization Hooks

Evaluate before standard permission check:

```javascript
// Future API
await registerHook({
  type: 'pre_authorization',
  name: 'business_hours_check',
  description: 'Only allow operations during business hours',
  
  handler: async (context) => {
    const hour = new Date().getHours();
    
    // Only allow 9 AM to 5 PM
    if (hour < 9 || hour >= 17) {
      return {
        allowed: false,
        reason: 'Operation only allowed during business hours (9 AM - 5 PM)'
      };
    }
    
    return {allowed: true};
  },
  
  // Apply to specific permissions
  filters: {
    services: ['blog-api'],
    actions: ['delete']
  }
});
```

### Resource-Level Permissions

Check ownership or other resource properties:

```javascript
// Future API
await registerHook({
  type: 'pre_authorization',
  name: 'post_ownership_check',
  description: 'Users can only edit their own posts',
  
  handler: async (context) => {
    const {userId, resource} = context;
    
    // Fetch resource
    const post = await getPost(resource.id);
    
    // Check ownership
    if (post.author_id !== userId) {
      return {
        allowed: false,
        reason: 'You can only edit your own posts'
      };
    }
    
    return {allowed: true};
  },
  
  filters: {
    services: ['blog-api'],
    entities: ['post'],
    actions: ['update', 'delete']
  }
});
```

### Context-Aware Authorization

Use request context for decisions:

```javascript
// Future API
await registerHook({
  type: 'pre_authorization',
  name: 'ip_whitelist_check',
  description: 'Restrict sensitive operations to office IPs',
  
  handler: async (context) => {
    const {request} = context;
    const clientIP = request.ip;
    
    const allowedIPs = ['10.0.0.0/8', '172.16.0.0/12'];
    
    if (!isIPInRange(clientIP, allowedIPs)) {
      return {
        allowed: false,
        reason: 'Sensitive operations only allowed from office network'
      };
    }
    
    return {allowed: true};
  },
  
  filters: {
    entities: ['system_user'],
    actions: ['create', 'rotate']
  }
});
```

### Post-Authorization Hooks

Add logging or side effects after authorization:

```javascript
// Future API
await registerHook({
  type: 'post_authorization',
  name: 'audit_log',
  description: 'Log all authorization decisions',
  
  handler: async (context) => {
    const {userId, tenantId, permission, authorized} = context;
    
    await logAuditEvent({
      event: 'authorization_check',
      user_id: userId,
      tenant_id: tenantId,
      permission: `${permission.service}:${permission.entity}:${permission.action}`,
      result: authorized ? 'granted' : 'denied',
      timestamp: new Date()
    });
    
    // Don't modify the authorization result
    return {allowed: authorized};
  }
});
```

### Permission Denied Hooks

Custom handling for denied permissions:

```javascript
// Future API
await registerHook({
  type: 'permission_denied',
  name: 'notify_admin',
  description: 'Notify admin of critical permission denials',
  
  handler: async (context) => {
    const {userId, permission, reason} = context;
    
    // Critical permissions
    const critical = ['system_user:create', 'role:delete'];
    const permKey = `${permission.entity}:${permission.action}`;
    
    if (critical.includes(permKey)) {
      await notifyAdmin({
        subject: 'Critical Permission Denied',
        message: `User ${userId} attempted ${permKey} and was denied: ${reason}`,
        priority: 'high'
      });
    }
  }
});
```

## Hook Context

Hooks receive a context object with useful information:

```typescript
interface HookContext {
  // User information
  userId: string;
  tenantId: string;
  
  // Permission being checked
  permission: {
    service: string;
    entity: string;
    action: string;
  };
  
  // Resource information (if provided)
  resource?: {
    id: string;
    type: string;
    attributes?: Record<string, any>;
  };
  
  // Request information
  request: {
    ip: string;
    userAgent: string;
    method: string;
    path: string;
    headers: Record<string, string>;
  };
  
  // Current authorization result (for post-hooks)
  authorized?: boolean;
  reason?: string;
  
  // Additional metadata
  metadata: Record<string, any>;
}
```

## Current Workaround: Middleware

Until hooks are implemented, use custom middleware:

```go
// internal/api/middleware/custom_authorization.go
package middleware

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

func BusinessHoursOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        hour := time.Now().Hour()
        
        if hour < 9 || hour >= 17 {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Operation only allowed during business hours (9 AM - 5 PM)",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Apply to specific routes
protected.DELETE("/posts/:id",
    middleware.BusinessHoursOnly(),
    handlers.DeletePost,
)
```

### Resource Ownership Check

```go
func RequirePostOwnership() gin.HandlerFunc {
    return func(c *gin.Context) {
        postID := c.Param("id")
        userID := c.GetString("user_id")
        
        // Get post
        post, err := postService.GetByID(c, postID)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
            c.Abort()
            return
        }
        
        // Check ownership
        if post.AuthorID != userID {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You can only edit your own posts",
            })
            c.Abort()
            return
        }
        
        // Store in context for handler
        c.Set("post", post)
        c.Next()
    }
}
```

### Audit Logging Middleware

```go
func AuditLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Process request
        c.Next()
        
        // Log after request
        duration := time.Since(start)
        
        auditLog := map[string]interface{}{
            "user_id":    c.GetString("user_id"),
            "tenant_id":  c.GetString("tenant_id"),
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "duration":   duration.Milliseconds(),
            "ip":         c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        }
        
        logger.Info("API request", zap.Any("audit", auditLog))
    }
}
```

## Planned Management API

### Register Hook

```javascript
const hook = await fetch('/api/v1/hooks', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    type: 'pre_authorization',
    name: 'business_hours_check',
    description: 'Business hours restriction',
    code: `
      const hour = new Date().getHours();
      if (hour < 9 || hour >= 17) {
        return {allowed: false, reason: 'Outside business hours'};
      }
      return {allowed: true};
    `,
    filters: {
      services: ['blog-api'],
      actions: ['delete']
    },
    active: true
  })
});
```

### List Hooks

```javascript
const hooks = await fetch('/api/v1/hooks', {
  credentials: 'include'
});

// Returns all registered hooks
```

### Test Hook

```javascript
const result = await fetch(`/api/v1/hooks/${hookId}/test`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    context: {
      userId: 'test-user',
      tenantId: 'test-tenant',
      permission: {
        service: 'blog-api',
        entity: 'post',
        action: 'delete'
      }
    }
  })
});

// Returns hook execution result
```

## Use Cases

### 1. Time-Based Restrictions
```
- Business hours only
- Weekend restrictions
- Maintenance windows
```

### 2. Resource Ownership
```
- Edit own posts only
- Delete own comments
- View own data
```

### 3. Geographic Restrictions
```
- Region-based access
- IP whitelisting
- Country restrictions
```

### 4. Rate Limiting
```
- Per-user rate limits
- Per-tenant quotas
- Action throttling
```

### 5. Compliance & Audit
```
- GDPR data access logs
- SOC2 audit trails
- Sensitive operation logs
```

### 6. Dynamic Permissions
```
- Subscription-based features
- Trial limitations
- Feature flags
```

## Best Practices

### 1. Keep Hooks Fast
Hooks run on every permission check - keep them performant

### 2. Handle Errors Gracefully
Don't crash on hook failures - log and continue

### 3. Cache Results
Cache hook results when appropriate

### 4. Test Thoroughly
Test hooks with various contexts and edge cases

### 5. Document Behavior
Clearly document what each hook does

## Security Considerations

- Validate all hook inputs
- Sandbox hook execution
- Limit hook execution time
- Prevent infinite loops
- Audit hook changes

## Contributing

Interested in implementing permission hooks? See:
- [Backend Integration Guide](/guides/backend-integration)
- [Custom Middleware](/advanced/custom-middleware)
- [RBAC API](/x-api/rbac)

## Request for Implementation

If you need this feature, please:
1. Open a GitHub issue
2. Describe your use case
3. Share expected hook types

**Implementation Priority**: Based on community demand

## Related Documentation

- [Webhooks](/advanced/webhooks) - Event notifications
- [RBAC Overview](/guides/rbac-overview) - Authorization system
- [Custom Middleware](/advanced/custom-middleware) - Current workaround
