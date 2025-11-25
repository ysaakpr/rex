# Monitoring and Alerting

Complete guide to monitoring your Rex in production.

## Overview

This guide covers:
- CloudWatch Logs and Metrics
- Application Performance Monitoring (APM)
- Health checks and uptime monitoring
- Alert configuration
- Log aggregation and analysis
- Performance optimization

## CloudWatch Logs

### Log Groups

**Default Log Groups**:
```
/aws/ecs/utm-api          # API application logs
/aws/ecs/utm-worker       # Background worker logs
/aws/ecs/utm-frontend     # Frontend logs
/aws/rds/utm-db          # Database logs
/aws/elasticache/utm-redis # Redis logs
```

### Viewing Logs

**CLI**:
```bash
# Stream API logs
aws logs tail /aws/ecs/utm-api --follow

# Last 1 hour
aws logs tail /aws/ecs/utm-api --since 1h

# Filter errors
aws logs tail /aws/ecs/utm-api --filter-pattern "ERROR"

# Specific time range
aws logs tail /aws/ecs/utm-api \
  --start-time "2024-11-25T10:00:00Z" \
  --end-time "2024-11-25T11:00:00Z"
```

**Console**: AWS Console → CloudWatch → Log Groups

### Log Insights Queries

**Find Errors**:
```
fields @timestamp, @message
| filter @message like /ERROR/
| sort @timestamp desc
| limit 100
```

**Slow API Requests**:
```
fields @timestamp, method, path, duration
| filter duration > 1000
| sort duration desc
| limit 50
```

**Failed Login Attempts**:
```
fields @timestamp, user_id, ip_address
| filter action = "login" and success = false
| stats count() by user_id, ip_address
```

**Request Volume by Endpoint**:
```
fields @timestamp, path
| stats count() by path
| sort count desc
```

## CloudWatch Metrics

### Key Metrics to Monitor

**Application Metrics**:
- `CPUUtilization` - CPU usage percentage
- `MemoryUtilization` - Memory usage percentage
- `RequestCount` - Total API requests
- `TargetResponseTime` - API response time
- `HealthyHostCount` - Number of healthy targets
- `UnhealthyHostCount` - Number of unhealthy targets

**Database Metrics**:
- `DatabaseConnections` - Active connections
- `ReadLatency` - Read query latency
- `WriteLatency` - Write query latency
- `FreeableMemory` - Available memory
- `CPUUtilization` - Database CPU usage

**Cache Metrics**:
- `CacheHits` - Redis cache hits
- `CacheMisses` - Redis cache misses
- `Evictions` - Items evicted from cache
- `CPUUtilization` - Redis CPU usage

### Custom Metrics

**Application-Level Metrics**:

```go
// internal/pkg/metrics/metrics.go
package metrics

import (
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type MetricsClient struct {
    cw *cloudwatch.Client
}

func (m *MetricsClient) PublishMetric(name string, value float64, unit types.StandardUnit) error {
    _, err := m.cw.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
        Namespace: aws.String("UTM/Application"),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String(name),
                Value:      aws.Float64(value),
                Unit:       unit,
                Timestamp:  aws.Time(time.Now()),
            },
        },
    })
    return err
}

// Track business metrics
func TrackTenantCreated(metricsClient *MetricsClient) {
    metricsClient.PublishMetric("TenantCreated", 1, types.StandardUnitCount)
}

func TrackInvitationSent(metricsClient *MetricsClient) {
    metricsClient.PublishMetric("InvitationSent", 1, types.StandardUnitCount)
}

func TrackAPILatency(metricsClient *MetricsClient, duration time.Duration) {
    metricsClient.PublishMetric("APILatency", float64(duration.Milliseconds()), types.StandardUnitMilliseconds)
}
```

**Usage in Handlers**:
```go
func CreateTenantHandler(c *gin.Context) {
    start := time.Now()
    
    // Create tenant logic
    tenant, err := tenantService.CreateTenant(input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Track metrics
    metricsClient.TrackTenantCreated()
    metricsClient.TrackAPILatency(time.Since(start))
    
    c.JSON(201, gin.H{"success": true, "data": tenant})
}
```

## Alerts

### Critical Alerts

**High Error Rate**:
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name utm-high-error-rate \
  --alarm-description "API error rate above 5%" \
  --metric-name 5XXError \
  --namespace AWS/ApplicationELB \
  --statistic Sum \
  --period 300 \
  --threshold 50 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions <sns-topic-arn>
```

**Database Connection Pool Exhausted**:
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name utm-db-connections-high \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --dimensions Name=DBInstanceIdentifier,Value=utm-db \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions <sns-topic-arn>
```

**High CPU Usage**:
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name utm-high-cpu \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --dimensions Name=ServiceName,Value=utm-api \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 3 \
  --alarm-actions <sns-topic-arn>
```

**Unhealthy Targets**:
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name utm-unhealthy-targets \
  --metric-name UnhealthyHostCount \
  --namespace AWS/ApplicationELB \
  --dimensions Name=TargetGroup,Value=<target-group-arn> \
  --statistic Average \
  --period 60 \
  --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --evaluation-periods 2 \
  --alarm-actions <sns-topic-arn>
```

### SNS Topics for Alerts

**Create SNS Topic**:
```bash
# Create topic
aws sns create-topic --name utm-alerts

# Subscribe email
aws sns subscribe \
  --topic-arn arn:aws:sns:us-east-1:ACCOUNT:utm-alerts \
  --protocol email \
  --notification-endpoint ops@yourdomain.com

# Subscribe Slack (via AWS Chatbot)
# Configure in AWS Console: AWS Chatbot → Slack
```

## Health Checks

### Application Health Endpoint

```go
// internal/api/handlers/health_handler.go
func HealthCheck(c *gin.Context) {
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "version": os.Getenv("APP_VERSION"),
    }
    
    // Check database
    if err := db.Ping(); err != nil {
        health["status"] = "unhealthy"
        health["database"] = "down"
        c.JSON(503, health)
        return
    }
    health["database"] = "up"
    
    // Check Redis
    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        health["status"] = "unhealthy"
        health["redis"] = "down"
        c.JSON(503, health)
        return
    }
    health["redis"] = "up"
    
    c.JSON(200, health)
}

// Register route
router.GET("/health", HealthCheck)
```

### ALB Health Check Configuration

```hcl
resource "aws_lb_target_group" "api" {
  health_check {
    enabled             = true
    path                = "/health"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }
}
```

### Uptime Monitoring

**Third-Party Services**:
- UptimeRobot (free tier available)
- Pingdom
- StatusCake
- AWS Route 53 Health Checks

**Example: UptimeRobot**:
```bash
# Monitor API health
Monitor: https://api.yourdomain.com/health
Interval: 5 minutes
Alert: Email, SMS, Slack

# Monitor frontend
Monitor: https://app.yourdomain.com
Interval: 5 minutes
Alert: Email
```

## Application Performance Monitoring (APM)

### Structured Logging

**Use Zap Logger**:
```go
// internal/pkg/logger/logger.go
package logger

import "go.uber.org/zap"

func NewLogger() (*zap.Logger, error) {
    if os.Getenv("APP_ENV") == "production" {
        return zap.NewProduction()
    }
    return zap.NewDevelopment()
}

// Usage in handlers
logger.Info("tenant created",
    zap.String("tenant_id", tenantID.String()),
    zap.String("user_id", userID),
    zap.Duration("duration", time.Since(start)),
)

logger.Error("failed to create tenant",
    zap.Error(err),
    zap.String("user_id", userID),
)
```

### Request Tracing

**Add Request ID Middleware**:
```go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// Log with request ID
func LogRequest(c *gin.Context, message string) {
    requestID, _ := c.Get("request_id")
    logger.Info(message,
        zap.String("request_id", requestID.(string)),
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
    )
}
```

### Performance Profiling

**pprof Integration**:
```go
import _ "net/http/pprof"

// Enable pprof in development
if os.Getenv("APP_ENV") == "development" {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}

// Access profiles:
// CPU: http://localhost:6060/debug/pprof/profile?seconds=30
// Heap: http://localhost:6060/debug/pprof/heap
// Goroutines: http://localhost:6060/debug/pprof/goroutine
```

## Dashboards

### CloudWatch Dashboard

**Create Dashboard**:
```bash
aws cloudwatch put-dashboard \
  --dashboard-name UTM-Production \
  --dashboard-body file://dashboard.json
```

**dashboard.json**:
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ApplicationELB", "RequestCount", {"stat": "Sum"}],
          [".", "TargetResponseTime", {"stat": "Average"}]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "API Metrics"
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/RDS", "CPUUtilization", {"stat": "Average"}],
          [".", "DatabaseConnections", {"stat": "Average"}]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "Database Metrics"
      }
    }
  ]
}
```

### Grafana (Optional)

**Setup**:
1. Deploy Grafana on EC2 or use Grafana Cloud
2. Add CloudWatch data source
3. Import pre-built dashboards
4. Create custom dashboards

## Log Retention

**Configure Retention**:
```bash
# Set log retention to 30 days
aws logs put-retention-policy \
  --log-group-name /aws/ecs/utm-api \
  --retention-in-days 30

# Delete old log streams
aws logs delete-log-stream \
  --log-group-name /aws/ecs/utm-api \
  --log-stream-name old-stream
```

## Cost Optimization

**Reduce Logging Costs**:
1. Set appropriate retention periods
2. Filter logs at source (don't log everything)
3. Use log sampling for high-volume endpoints
4. Archive old logs to S3

**Example: Log Sampling**:
```go
func ShouldLog(endpoint string) bool {
    // Log all errors
    if statusCode >= 400 {
        return true
    }
    
    // Sample high-volume endpoints
    if endpoint == "/api/v1/health" {
        return rand.Float64() < 0.01  // 1% sampling
    }
    
    return true  // Log everything else
}
```

## Security Monitoring

### Monitor Failed Login Attempts

```go
func TrackFailedLogin(userID, ipAddress string) {
    logger.Warn("failed login attempt",
        zap.String("user_id", userID),
        zap.String("ip_address", ipAddress),
    )
    
    // Check for brute force
    recentAttempts := countRecentAttempts(ipAddress, 5*time.Minute)
    if recentAttempts > 5 {
        alertSecurityTeam("Potential brute force attack", ipAddress)
        blockIP(ipAddress, 1*time.Hour)
    }
}
```

### Audit Log Monitoring

```go
// Monitor suspicious activities
func MonitorAuditLogs() {
    // Check for privilege escalation
    escalations := db.Where(
        "action = ? AND resource = ? AND created_at > ?",
        "UPDATE", "roles", time.Now().Add(-1*time.Hour),
    ).Find(&auditLogs)
    
    if len(escalations) > 0 {
        alertSecurityTeam("Privilege escalation detected", escalations)
    }
    
    // Check for unusual deletion patterns
    deletions := db.Where(
        "action = ? AND created_at > ?",
        "DELETE", time.Now().Add(-1*time.Hour),
    ).Find(&auditLogs)
    
    if len(deletions) > 10 {
        alertSecurityTeam("High number of deletions", deletions)
    }
}
```

## Troubleshooting with Monitoring

### High Response Time

**Investigate**:
```bash
# Find slow queries in CloudWatch Logs
aws logs filter-pattern --log-group-name /aws/ecs/utm-api \
  --filter-pattern '{ $.duration > 1000 }'

# Check database performance
aws rds describe-db-log-files \
  --db-instance-identifier utm-db

# Check slow query log
aws rds download-db-log-file-portion \
  --db-instance-identifier utm-db \
  --log-file-name slowquery/postgresql.log
```

### Memory Leaks

**Monitor Memory**:
```go
// Add memory metrics
var m runtime.MemStats
runtime.ReadMemStats(&m)

metricsClient.PublishMetric("MemoryAlloc", float64(m.Alloc), types.StandardUnitBytes)
metricsClient.PublishMetric("NumGoroutines", float64(runtime.NumGoroutine()), types.StandardUnitCount)
```

### Database Connection Pool

**Monitor Connections**:
```go
stats := db.DB.Stats()
metricsClient.PublishMetric("DBOpenConnections", float64(stats.OpenConnections), types.StandardUnitCount)
metricsClient.PublishMetric("DBInUse", float64(stats.InUse), types.StandardUnitCount)
metricsClient.PublishMetric("DBIdle", float64(stats.Idle), types.StandardUnitCount)
```

## Best Practices

1. **Set Up Alerts Early** - Don't wait for incidents
2. **Monitor Business Metrics** - Track tenant creation, invitations, etc.
3. **Use Structured Logging** - JSON format for easy parsing
4. **Add Request IDs** - Trace requests across services
5. **Create Runbooks** - Document response procedures
6. **Regular Reviews** - Weekly metric reviews
7. **Cost Monitoring** - Track CloudWatch costs
8. **Test Alerts** - Ensure alerts work before incidents

## Next Steps

- [Production Setup](/deployment/production-setup) - Deploy to production
- [AWS Deployment](/deployment/aws) - AWS infrastructure
- [Security Guide](/guides/security) - Security monitoring
- [Troubleshooting](/troubleshooting/common-issues) - Debug issues

