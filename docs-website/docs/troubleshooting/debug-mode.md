# Debug Mode

Complete guide to enabling and using debug mode for troubleshooting.

## Overview

Debug mode provides detailed logging and diagnostic information to help troubleshoot issues during development and production. This guide covers:
- Enabling debug logging
- Backend debug tools
- Frontend debugging
- SuperTokens debug mode
- Performance profiling

## Backend Debug Mode

### Enable Debug Logging

**Environment Variable**:
```bash
# .env
LOG_LEVEL=debug
LOG_FORMAT=console  # More readable for debugging

# Or export
export LOG_LEVEL=debug
```

**Restart Application**:
```bash
# Docker Compose
docker-compose restart api

# Or
make restart-api
```

### Log Levels

```go
// internal/pkg/logger/logger.go
const (
    DebugLevel = "debug"  // Detailed information
    InfoLevel  = "info"   // General information
    WarnLevel  = "warn"   // Warning messages
    ErrorLevel = "error"  // Error messages
    FatalLevel = "fatal"  // Fatal errors
)
```

### Debug Logging Examples

**In Handlers**:
```go
func CreateTenantHandler(c *gin.Context) {
    logger.Debug("create tenant handler called",
        zap.String("user_id", getUserID(c)),
        zap.Any("request_body", c.Request.Body),
    )
    
    var input models.CreateTenantInput
    if err := c.ShouldBindJSON(&input); err != nil {
        logger.Debug("validation failed",
            zap.Error(err),
            zap.Any("input", input),
        )
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    logger.Debug("creating tenant",
        zap.String("name", input.Name),
        zap.String("slug", input.Slug),
    )
    
    tenant, err := tenantService.CreateTenant(input)
    if err != nil {
        logger.Error("failed to create tenant",
            zap.Error(err),
            zap.Any("input", input),
        )
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    logger.Debug("tenant created successfully",
        zap.String("tenant_id", tenant.ID.String()),
    )
    
    c.JSON(201, gin.H{"success": true, "data": tenant})
}
```

### SQL Query Logging

**Enable GORM Debug Mode**:
```go
// internal/database/database.go
func Connect() (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // Log all SQL queries
    })
    
    // In production, use Silent or Warn
    if os.Getenv("APP_ENV") == "production" {
        db.Logger = logger.Default.LogMode(logger.Silent)
    }
    
    return db, err
}
```

**Log Specific Queries**:
```go
// Debug a specific query
db.Debug().Where("email = ?", email).First(&user)

// See the SQL that would be executed
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Where("status = ?", "active").Find(&users)
})
logger.Debug("SQL query", zap.String("sql", sql))
```

### Request/Response Logging

**Middleware**:
```go
// internal/api/middleware/logger.go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Log request
        logger.Debug("incoming request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.String("query", c.Request.URL.RawQuery),
            zap.String("ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        )
        
        // Process request
        c.Next()
        
        // Log response
        logger.Debug("request completed",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}

// Apply middleware
router.Use(middleware.RequestLogger())
```

## Frontend Debug Mode

### Enable SuperTokens Debug

```javascript
// src/main.jsx
SuperTokens.init({
  appInfo: {
    // ... config
  },
  enableDebugLogs: true,  // Enable debug mode
  recipeList: [
    EmailPassword.init(),
    Session.init()
  ]
});
```

**Output**:
```
[SuperTokens] Session verification successful
[SuperTokens] Refresh token endpoint called
[SuperTokens] Cookie set: sAccessToken
```

### Browser Console Debugging

**Add Debug Logs**:
```javascript
// src/lib/api.js
export async function apiCall(endpoint, options = {}) {
  console.debug('[API] Request:', {
    endpoint,
    method: options.method || 'GET',
    body: options.body
  });
  
  const response = await fetch(endpoint, {
    credentials: 'include',
    ...options
  });
  
  console.debug('[API] Response:', {
    endpoint,
    status: response.status,
    ok: response.ok
  });
  
  const data = await response.json();
  console.debug('[API] Data:', data);
  
  return data;
}
```

### React DevTools

**Install**:
- Chrome: React Developer Tools extension
- Firefox: React Developer Tools add-on

**Usage**:
1. Open DevTools (F12)
2. Select "Components" tab
3. Inspect component props and state
4. Select "Profiler" tab to analyze performance

### Redux DevTools (if using Redux)

```javascript
// src/store.js
import {configureStore} from '@reduxjs/toolkit';

const store = configureStore({
  reducer: rootReducer,
  devTools: process.env.NODE_ENV !== 'production'
});
```

## Database Debugging

### PostgreSQL Query Logging

**Enable in postgresql.conf**:
```
log_statement = 'all'
log_duration = on
log_min_duration_statement = 100  # Log queries > 100ms
```

**Or via SQL**:
```sql
-- Enable logging for current session
SET log_statement = 'all';
SET log_min_duration_statement = 0;

-- Run your queries
SELECT * FROM tenants WHERE id = 'uuid';

-- Check logs
SELECT * FROM pg_stat_activity;
```

### Slow Query Analysis

```sql
-- Find slow queries
SELECT 
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### Connection Pool Debugging

```go
// Monitor connection pool
func LogConnectionPoolStats(db *gorm.DB) {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()
    
    logger.Debug("connection pool stats",
        zap.Int("open_connections", stats.OpenConnections),
        zap.Int("in_use", stats.InUse),
        zap.Int("idle", stats.Idle),
        zap.Int64("wait_count", stats.WaitCount),
        zap.Duration("wait_duration", stats.WaitDuration),
    )
}

// Call periodically
ticker := time.NewTicker(10 * time.Second)
go func() {
    for range ticker.C {
        LogConnectionPoolStats(db)
    }
}()
```

## Performance Profiling

### Go pprof

**Enable**:
```go
// cmd/api/main.go
import _ "net/http/pprof"

func main() {
    // Enable pprof endpoint
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Start main application
    router.Run(":8080")
}
```

**Collect Profiles**:
```bash
# CPU profile (30 seconds)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze with pprof
go tool pprof cpu.prof

# Commands in pprof:
(pprof) top10      # Show top 10 functions
(pprof) list main  # Show source for main
(pprof) web        # Open visualization in browser
```

### Memory Leak Detection

```go
// Add memory tracking
import "runtime"

func LogMemoryStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    logger.Debug("memory stats",
        zap.Uint64("alloc_mb", m.Alloc/1024/1024),
        zap.Uint64("total_alloc_mb", m.TotalAlloc/1024/1024),
        zap.Uint64("sys_mb", m.Sys/1024/1024),
        zap.Uint32("num_gc", m.NumGC),
    )
}

// Call periodically
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        LogMemoryStats()
    }
}()
```

## Redis Debugging

### Monitor Redis Commands

```bash
# Connect to Redis
docker exec -it utm-redis redis-cli

# Monitor all commands
MONITOR

# Or via CLI
redis-cli MONITOR
```

### Check Redis Stats

```bash
# Get info
redis-cli INFO

# Check memory usage
redis-cli INFO memory

# Check connected clients
redis-cli CLIENT LIST

# Check slow log
redis-cli SLOWLOG GET 10
```

## SuperTokens Debugging

### Core Logs

**Enable in docker-compose.yml**:
```yaml
services:
  supertokens:
    environment:
      - LOG_LEVEL=debug
```

**Check Logs**:
```bash
docker-compose logs supertokens

# Or follow logs
docker-compose logs -f supertokens
```

### Test SuperTokens Endpoints

```bash
# Health check
curl http://localhost:3567/hello

# List users
curl http://localhost:3567/users/count \
  -H "api-key: your-api-key"
```

## Network Debugging

### Inspect HTTP Requests

**Browser DevTools**:
1. Open DevTools (F12)
2. Go to "Network" tab
3. Filter by "Fetch/XHR"
4. Click request to see details

**Check Cookies**:
1. DevTools → Application → Cookies
2. Look for SuperTokens cookies:
   - `sAccessToken`
   - `sRefreshToken`
   - `sIdRefreshToken`

### cURL Debugging

```bash
# Verbose output
curl -v http://localhost:8080/api/v1/tenants

# Include cookies
curl -b cookies.txt -c cookies.txt http://localhost:8080/api/v1/tenants

# Show headers
curl -I http://localhost:8080/health

# Time request
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/tenants

# curl-format.txt:
time_namelookup:  %{time_namelookup}\n
time_connect:  %{time_connect}\n
time_appconnect:  %{time_appconnect}\n
time_pretransfer:  %{time_pretransfer}\n
time_redirect:  %{time_redirect}\n
time_starttransfer:  %{time_starttransfer}\n
----------\n
time_total:  %{time_total}\n
```

## Docker Debugging

### Container Logs

```bash
# View logs
docker-compose logs api
docker-compose logs worker
docker-compose logs postgres

# Follow logs
docker-compose logs -f api

# Last 100 lines
docker-compose logs --tail=100 api

# Since timestamp
docker-compose logs --since="2024-11-25T10:00:00" api
```

### Execute Commands in Container

```bash
# Interactive shell
docker-compose exec api /bin/sh

# Run command
docker-compose exec api ls -la

# Check environment variables
docker-compose exec api env

# Test database connection
docker-compose exec api psql -h postgres -U postgres -d utm_backend
```

### Inspect Container

```bash
# Container details
docker inspect utm-api

# Resource usage
docker stats utm-api

# Network info
docker network inspect utm-backend_default
```

## Debugging Checklist

### Authentication Issues

- [ ] Check SuperTokens logs
- [ ] Verify cookies are set (DevTools)
- [ ] Check `credentials: 'include'` in fetch
- [ ] Verify CORS configuration
- [ ] Test SuperTokens `/hello` endpoint
- [ ] Check API key configuration

### Permission Issues

- [ ] Enable debug logging
- [ ] Check user's role assignment
- [ ] Verify role → policy → permission chain
- [ ] Test authorization endpoint directly
- [ ] Check RBAC middleware logs

### Database Issues

- [ ] Enable SQL query logging
- [ ] Check connection pool stats
- [ ] Verify migrations applied
- [ ] Test database connection
- [ ] Check slow query log

### Performance Issues

- [ ] Enable CPU profiling
- [ ] Check memory usage
- [ ] Monitor database query time
- [ ] Check Redis hit rate
- [ ] Profile slow endpoints

## Disabling Debug Mode

### Production

```bash
# Set appropriate log level
LOG_LEVEL=warn  # or error
LOG_FORMAT=json

# Disable SQL logging
# Remove .Debug() from GORM queries

# Disable pprof
# Comment out pprof import and server

# Disable SuperTokens debug
enableDebugLogs: false
```

## Best Practices

1. **Don't Leave Debug On**: Always disable in production
2. **Sanitize Logs**: Never log sensitive data (passwords, tokens)
3. **Use Structured Logging**: JSON format for easy parsing
4. **Add Context**: Include request IDs, user IDs in logs
5. **Monitor Performance**: Debug mode can slow things down
6. **Use Appropriate Levels**: Debug for dev, Info for staging, Warn for prod

## Next Steps

- [Common Issues](/troubleshooting/common-issues) - Troubleshooting guide
- [Monitoring](/deployment/monitoring) - Production monitoring
- [Security](/guides/security) - Security best practices
