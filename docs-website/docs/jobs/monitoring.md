# Job Monitoring and Debugging

Guide to monitoring and troubleshooting background jobs.

## Monitoring Tools

### 1. Asynqmon (Web UI)

**Best for**: Real-time monitoring, manual intervention

```bash
# Install
go install github.com/hibiken/asynqmon@latest

# Run
asynqmon --redis-addr=localhost:6379 --port=8081

# Open in browser
open http://localhost:8081
```

**Features**:
- View all queues and task counts
- See active, pending, failed tasks
- Retry or delete failed tasks
- View task details and payloads
- Monitor processing rates
- Search tasks by type or ID

### 2. Docker Logs

**Best for**: Debugging, error investigation

```bash
# View worker logs
docker logs utm-worker

# Follow logs in real-time
docker logs -f utm-worker

# Last 100 lines
docker logs --tail 100 utm-worker

# With timestamps
docker logs -t utm-worker

# Filter by keyword
docker logs utm-worker | grep "ERROR"
docker logs utm-worker | grep "tenant:initialize"
```

### 3. Redis CLI

**Best for**: Direct queue inspection

```bash
# Connect to Redis
docker exec -it utm-redis redis-cli

# List all keys
KEYS *

# View queue lengths
LLEN asynq:{default}:pending
LLEN asynq:{critical}:pending
LLEN asynq:{low}:pending

# View scheduled tasks
ZCARD asynq:{default}:scheduled

# View task details
GET asynq:task:task-id

# Clear queue (CAUTION)
DEL asynq:{default}:pending
```

## Metrics and KPIs

### Queue Metrics

Monitor these key metrics:

**Queue Depth**:
```bash
# Check pending tasks per queue
watch -n 5 'docker exec utm-redis redis-cli LLEN asynq:{critical}:pending'
```

**Processing Rate**:
- Tasks processed per minute
- Average processing time
- Success rate vs failure rate

**Failure Rate**:
- Failed tasks per hour
- Retry count distribution
- Permanent failures

### Example Monitoring Script

```bash
#!/bin/bash
# scripts/monitor-jobs.sh

echo "=== Job Queue Status ==="
echo ""

# Critical queue
CRITICAL=$(docker exec utm-redis redis-cli LLEN asynq:{critical}:pending)
echo "Critical Queue: $CRITICAL pending"

# Default queue
DEFAULT=$(docker exec utm-redis redis-cli LLEN asynq:{default}:pending)
echo "Default Queue:  $DEFAULT pending"

# Low queue
LOW=$(docker exec utm-redis redis-cli LLEN asynq:{low}:pending)
echo "Low Queue:      $LOW pending"

echo ""
echo "=== Active Tasks ==="
ACTIVE=$(docker exec utm-redis redis-cli LLEN asynq:servers:active)
echo "Active: $ACTIVE"

echo ""
echo "=== Failed Tasks ==="
FAILED=$(docker exec utm-redis redis-cli ZCARD asynq:dead)
echo "Dead: $FAILED"
```

## Common Issues

### Issue 1: Tasks Stuck in Pending

**Symptoms**:
- Tasks not processing
- Queue depth increasing
- No worker activity logs

**Diagnosis**:
```bash
# Check if worker is running
docker ps | grep worker

# Check worker logs for errors
docker logs utm-worker --tail 50

# Check Redis connection
docker exec utm-worker ping -c 1 redis
```

**Solutions**:
```bash
# Restart worker
docker-compose restart worker

# Check Redis health
docker exec utm-redis redis-cli PING

# Verify Redis config
docker exec utm-redis redis-cli CONFIG GET maxmemory
```

### Issue 2: Tasks Failing Repeatedly

**Symptoms**:
- High retry count
- Same task failing repeatedly
- Error logs repeating

**Diagnosis**:
```bash
# View failed tasks in Asynqmon
open http://localhost:8081

# Check error logs
docker logs utm-worker | grep "ERROR"

# Get task details
docker exec utm-redis redis-cli GET asynq:task:<task-id>
```

**Solutions**:
```bash
# Fix the underlying issue (DB, service, etc.)
# Then retry from Asynqmon UI

# Or delete permanently failed tasks
# (Do this from Asynqmon after investigating)
```

### Issue 3: Slow Processing

**Symptoms**:
- Tasks taking too long
- Queue backlog growing
- Timeouts

**Diagnosis**:
```bash
# Check worker concurrency
docker exec utm-worker ps aux

# Monitor system resources
docker stats utm-worker

# Check task duration in logs
docker logs utm-worker | grep "Task completed"
```

**Solutions**:
```bash
# Increase worker concurrency
# In .env:
ASYNQ_CONCURRENCY=20  # Increase from 10

# Or scale workers
docker-compose up --scale worker=3

# Optimize task handlers (database queries, external calls)
```

### Issue 4: Memory Issues

**Symptoms**:
- Worker OOM (Out of Memory)
- Redis memory warnings
- Container restarts

**Diagnosis**:
```bash
# Check Redis memory usage
docker exec utm-redis redis-cli INFO memory

# Check worker memory
docker stats utm-worker --no-stream

# Check Redis max memory
docker exec utm-redis redis-cli CONFIG GET maxmemory
```

**Solutions**:
```bash
# Increase Redis memory limit
# In docker-compose.yml:
redis:
  command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru

# Increase worker memory
# In docker-compose.yml:
worker:
  deploy:
    resources:
      limits:
        memory: 1G

# Reduce task retention
asynq.Retention(24 * time.Hour)  # Instead of 7 days
```

### Issue 5: Database Connection Issues

**Symptoms**:
- "Too many connections" errors
- Task failures with DB errors
- Connection timeouts

**Diagnosis**:
```bash
# Check DB connections
docker exec utm-postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Check worker DB connections
docker logs utm-worker | grep "database"
```

**Solutions**:
```go
// Configure DB connection pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, _ := db.DB()

sqlDB.SetMaxOpenConns(25)    // Limit connections
sqlDB.SetMaxIdleConns(5)     // Idle connections
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

## Debugging Tasks

### Enable Debug Logging

```go
// internal/jobs/worker.go
import "go.uber.org/zap"

// Create debug logger
logger, _ := zap.NewDevelopment()  // Instead of Production

// Use in handlers
func (h *Handler) HandleTask(ctx context.Context, task *asynq.Task) error {
    h.logger.Debug("Task started",
        zap.String("type", task.Type()),
        zap.ByteString("payload", task.Payload()),
    )
    
    // ... processing ...
    
    h.logger.Debug("Task completed")
    return nil
}
```

### View Task Payload

```bash
# Get task from Redis
docker exec utm-redis redis-cli GET asynq:task:<task-id>

# Or from Asynqmon UI
open http://localhost:8081
# Click on task → View details
```

### Manual Task Execution

Create a test script:

```go
// scripts/test_task.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/hibiken/asynq"
    "github.com/ysaakpr/rex/internal/jobs"
    "github.com/ysaakpr/rex/internal/jobs/tasks"
)

func main() {
    // Setup
    db := setupDB()
    cfg := loadConfig()
    logger := setupLogger()
    
    // Create handler
    handler := tasks.NewTenantInitHandler(db, cfg)
    
    // Create test payload
    payload := tasks.TenantInitPayload{
        TenantID: "test-tenant-id",
    }
    payloadBytes, _ := json.Marshal(payload)
    
    // Create task
    task := asynq.NewTask(jobs.TypeTenantInitialization, payloadBytes)
    
    // Execute
    err := handler.HandleTenantInitialization(context.Background(), task)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Println("Success!")
    }
}
```

## Performance Optimization

### 1. Batch Processing

```go
func (h *Handler) HandleBatchTask(ctx context.Context, task *asynq.Task) error {
    // Process in batches instead of one-by-one
    items := h.fetchItems()
    
    batchSize := 100
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        h.processBatch(batch)
    }
    
    return nil
}
```

### 2. Database Query Optimization

```go
// Bad: N+1 queries
for _, item := range items {
    user := h.db.First(&User{}, item.UserID)
    process(user)
}

// Good: Single query with preload
var items []Item
h.db.Preload("User").Find(&items)
for _, item := range items {
    process(item.User)
}
```

### 3. Connection Pooling

```go
// Configure HTTP client with connection pool
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 4. Caching

```go
var cache = make(map[string]interface{})

func (h *Handler) getCachedData(key string) interface{} {
    if val, ok := cache[key]; ok {
        return val
    }
    
    val := h.fetchFromDB(key)
    cache[key] = val
    return val
}
```

## Alerts and Notifications

### CloudWatch Alarms (AWS)

```go
// Monitor queue depth
{
  "AlarmName": "HighJobQueueDepth",
  "MetricName": "QueueDepth",
  "Threshold": 1000,
  "ComparisonOperator": "GreaterThanThreshold",
  "EvaluationPeriods": 2,
  "AlarmActions": ["arn:aws:sns:..."]
}
```

### Custom Monitoring Script

```bash
#!/bin/bash
# scripts/alert-on-failures.sh

THRESHOLD=10
FAILED=$(docker exec utm-redis redis-cli ZCARD asynq:dead)

if [ "$FAILED" -gt "$THRESHOLD" ]; then
    echo "ALERT: $FAILED failed tasks detected!"
    # Send notification (email, Slack, etc.)
    curl -X POST https://hooks.slack.com/... \
        -d "{\"text\": \"⚠️ $FAILED failed background jobs!\"}"
fi
```

## Best Practices

### 1. Monitor Key Metrics
- Queue depths
- Processing rates
- Failure rates
- Average task duration

### 2. Set Up Alerts
- High queue depth
- High failure rate
- Worker down
- Memory issues

### 3. Regular Maintenance
- Review failed tasks weekly
- Clean up old tasks
- Optimize slow tasks
- Update retry strategies

### 4. Log Effectively
- Log task start/end
- Log errors with context
- Use structured logging
- Include task IDs

### 5. Test in Staging
- Test new jobs thoroughly
- Verify retry behavior
- Check error handling
- Monitor performance

## Related Documentation

- [Jobs Architecture](/jobs/architecture) - System design
- [Available Jobs](/jobs/available-jobs) - Existing jobs
- [Custom Jobs](/jobs/custom-jobs) - Creating new jobs
