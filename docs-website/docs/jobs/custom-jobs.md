# Creating Custom Background Jobs

Guide to adding new background jobs to the system.

## Overview

Adding a new job involves 4 steps:
1. Define the task type constant
2. Create the task handler
3. Register the handler with the worker
4. Add client method to enqueue the task

## Step-by-Step Guide

### Step 1: Define Task Type

Add your task type constant to `internal/jobs/client.go`:

```go
// internal/jobs/client.go
package jobs

const (
    // Existing tasks
    TypeTenantInitialization = "tenant:initialize"
    TypeUserInvitation       = "user:invitation"
    TypeSystemUserExpiry     = "system_user:expiry"
    
    // Your new task
    TypeDataExport = "data:export"  // Add this
)
```

**Naming Convention**: `category:action` (e.g., `tenant:cleanup`, `report:generate`)

### Step 2: Create Task Handler

Create a new file in `internal/jobs/tasks/`:

```go
// internal/jobs/tasks/data_export.go
package tasks

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/hibiken/asynq"
    "go.uber.org/zap"
    "gorm.io/gorm"
    
    "github.com/ysaakpr/rex/internal/config"
    "github.com/ysaakpr/rex/internal/models"
)

// Handler struct
type DataExportHandler struct {
    db     *gorm.DB
    cfg    *config.Config
    logger *zap.Logger
}

// Constructor
func NewDataExportHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *DataExportHandler {
    return &DataExportHandler{
        db:     db,
        cfg:    cfg,
        logger: logger,
    }
}

// Payload struct
type DataExportPayload struct {
    TenantID  string `json:"tenant_id"`
    UserID    string `json:"user_id"`
    ExportType string `json:"export_type"`  // e.g., "full", "partial"
}

// Handler function
func (h *DataExportHandler) HandleDataExport(ctx context.Context, task *asynq.Task) error {
    // 1. Parse payload
    var payload DataExportPayload
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    h.logger.Info("Starting data export",
        zap.String("tenant_id", payload.TenantID),
        zap.String("user_id", payload.UserID),
        zap.String("export_type", payload.ExportType),
    )
    
    // 2. Validate input
    tenantID, err := uuid.Parse(payload.TenantID)
    if err != nil {
        return fmt.Errorf("invalid tenant ID: %w", err)
    }
    
    userID, err := uuid.Parse(payload.UserID)
    if err != nil {
        return fmt.Errorf("invalid user ID: %w", err)
    }
    
    // 3. Fetch tenant
    var tenant models.Tenant
    if err := h.db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
        return fmt.Errorf("tenant not found: %w", err)
    }
    
    // 4. Perform export
    exportPath, err := h.performExport(ctx, &tenant, userID, payload.ExportType)
    if err != nil {
        return fmt.Errorf("export failed: %w", err)
    }
    
    // 5. Notify user (optional)
    if err := h.notifyUser(userID, exportPath); err != nil {
        h.logger.Error("Failed to notify user", zap.Error(err))
        // Don't fail the task if notification fails
    }
    
    h.logger.Info("Data export completed",
        zap.String("export_path", exportPath),
    )
    
    return nil
}

func (h *DataExportHandler) performExport(
    ctx context.Context,
    tenant *models.Tenant,
    userID uuid.UUID,
    exportType string,
) (string, error) {
    // Your export logic here
    // ...
    
    exportPath := fmt.Sprintf("/exports/%s_%s.zip", tenant.Slug, time.Now().Format("20060102"))
    return exportPath, nil
}

func (h *DataExportHandler) notifyUser(userID uuid.UUID, exportPath string) error {
    // Send notification to user
    // Could enqueue another job here
    return nil
}
```

### Step 3: Register with Worker

Update `internal/jobs/worker.go`:

```go
// internal/jobs/worker.go
package jobs

func NewWorker(cfg *config.Config, db *gorm.DB, logger *zap.Logger) (*Worker, error) {
    // ... existing setup ...
    
    mux := asynq.NewServeMux()
    
    // Existing handlers
    tenantInitHandler := tasks.NewTenantInitHandler(db, cfg)
    mux.HandleFunc(TypeTenantInitialization, tenantInitHandler.HandleTenantInitialization)
    
    invitationHandler := tasks.NewInvitationHandler(db, cfg)
    mux.HandleFunc(TypeUserInvitation, invitationHandler.HandleUserInvitation)
    
    systemUserExpiryTask := tasks.NewSystemUserExpiryTask(db, logger)
    mux.HandleFunc(TypeSystemUserExpiry, systemUserExpiryTask.HandleSystemUserExpiry)
    
    // Add your new handler
    dataExportHandler := tasks.NewDataExportHandler(db, cfg, logger)
    mux.HandleFunc(TypeDataExport, dataExportHandler.HandleDataExport)
    
    return &Worker{
        server:    server,
        mux:       mux,
        scheduler: scheduler,
    }, nil
}
```

### Step 4: Add Client Method

Update `internal/jobs/client.go`:

```go
// internal/jobs/client.go
package jobs

// Update interface
type Client interface {
    EnqueueTenantInitialization(tenantID uuid.UUID) error
    EnqueueUserInvitation(invitationID uuid.UUID) error
    EnqueueDataExport(tenantID, userID uuid.UUID, exportType string) error  // Add this
    Close() error
}

// Implement method
func (c *client) EnqueueDataExport(
    tenantID, userID uuid.UUID,
    exportType string,
) error {
    payload, err := json.Marshal(map[string]interface{}{
        "tenant_id":   tenantID.String(),
        "user_id":     userID.String(),
        "export_type": exportType,
    })
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }
    
    task := asynq.NewTask(TypeDataExport, payload)
    
    info, err := c.asynqClient.Enqueue(
        task,
        asynq.Queue(QueueDefault),                    // Choose appropriate queue
        asynq.MaxRetry(3),                           // Set retry count
        asynq.Timeout(10*time.Minute),               // Set timeout
        asynq.Retention(7*24*time.Hour),             // Keep for 7 days
    )
    if err != nil {
        return fmt.Errorf("failed to enqueue task: %w", err)
    }
    
    fmt.Printf("Enqueued data export task: id=%s, queue=%s\n", info.ID, info.Queue)
    return nil
}
```

### Step 5: Use in API Handler

```go
// internal/api/handlers/export_handler.go
package handlers

func (h *ExportHandler) RequestExport(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    userID := c.GetString("user_id")
    
    var input struct {
        ExportType string `json:"export_type"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "Invalid input"})
        return
    }
    
    // Enqueue export job
    err := h.jobClient.EnqueueDataExport(
        uuid.MustParse(tenantID),
        uuid.MustParse(userID),
        input.ExportType,
    )
    
    if err != nil {
        h.logger.Error("Failed to enqueue export", zap.Error(err))
        c.JSON(500, gin.H{"error": "Failed to start export"})
        return
    }
    
    c.JSON(202, gin.H{
        "success": true,
        "message": "Export job started",
    })
}
```

## Job Options

### Queue Selection

```go
// High priority - user-facing operations
asynq.Queue(QueueCritical)

// Normal priority - background operations
asynq.Queue(QueueDefault)

// Low priority - cleanup, analytics
asynq.Queue(QueueLow)
```

### Retry Configuration

```go
// No retries
asynq.MaxRetry(0)

// Limited retries
asynq.MaxRetry(3)

// Many retries (for critical operations)
asynq.MaxRetry(10)

// Custom retry delay
asynq.MaxRetry(5)
asynq.RetryDelay(time.Minute)  // Wait 1 minute between retries
```

### Timeouts

```go
// Short timeout (quick operations)
asynq.Timeout(30 * time.Second)

// Medium timeout (API calls)
asynq.Timeout(5 * time.Minute)

// Long timeout (large exports, processing)
asynq.Timeout(30 * time.Minute)
```

### Scheduling

```go
// Process immediately
c.asynqClient.Enqueue(task, ...)

// Process after delay
asynq.ProcessIn(5 * time.Minute)

// Process at specific time
processAt := time.Now().Add(24 * time.Hour)
asynq.ProcessAt(processAt)
```

### Task Retention

```go
// Delete immediately after completion
asynq.Retention(0)

// Keep for 1 day (debugging)
asynq.Retention(24 * time.Hour)

// Keep for 1 week (compliance)
asynq.Retention(7 * 24 * time.Hour)
```

## Periodic Tasks

### Scheduling Periodic Jobs

```go
// internal/jobs/worker.go
func setupPeriodicTasks(scheduler *asynq.Scheduler) error {
    // Daily cleanup at 2 AM
    _, err := scheduler.Register(
        "0 2 * * *",  // Cron expression
        asynq.NewTask(TypeDailyCleanup, nil),
        asynq.Queue(QueueLow),
    )
    if err != nil {
        return err
    }
    
    // Weekly report on Sundays at 8 AM
    _, err = scheduler.Register(
        "0 8 * * 0",
        asynq.NewTask(TypeWeeklyReport, nil),
        asynq.Queue(QueueDefault),
    )
    if err != nil {
        return err
    }
    
    // Every 15 minutes
    _, err = scheduler.Register(
        "*/15 * * * *",
        asynq.NewTask(TypeHealthCheck, nil),
        asynq.Queue(QueueDefault),
    )
    
    return err
}

// Call from NewWorker
func NewWorker(...) (*Worker, error) {
    // ... existing setup ...
    
    if err := setupPeriodicTasks(scheduler); err != nil {
        return nil, fmt.Errorf("failed to setup periodic tasks: %w", err)
    }
    
    return &Worker{...}, nil
}
```

## Best Practices

### 1. Make Tasks Idempotent

```go
func (h *Handler) HandleTask(ctx context.Context, task *asynq.Task) error {
    // Check if already processed
    var record ProcessRecord
    err := h.db.Where("task_id = ?", task.ResultWriter().TaskID()).
        First(&record).Error
    
    if err == nil {
        // Already processed
        return nil
    }
    
    // Process
    if err := h.process(); err != nil {
        return err
    }
    
    // Record completion
    h.db.Create(&ProcessRecord{TaskID: task.ResultWriter().TaskID()})
    
    return nil
}
```

### 2. Handle Context Cancellation

```go
func (h *Handler) HandleTask(ctx context.Context, task *asynq.Task) error {
    // Check context before long operations
    select {
    case <-ctx.Done():
        return ctx.Err()  // Task cancelled
    default:
        // Continue
    }
    
    // Use context in DB calls
    result := h.db.WithContext(ctx).Where(...).Find(&records)
    
    return nil
}
```

### 3. Log Progress

```go
func (h *Handler) HandleTask(ctx context.Context, task *asynq.Task) error {
    h.logger.Info("Task started")
    
    // Log steps
    h.logger.Info("Fetching data")
    data := h.fetchData()
    
    h.logger.Info("Processing data", zap.Int("count", len(data)))
    h.processData(data)
    
    h.logger.Info("Task completed")
    return nil
}
```

### 4. Handle Partial Failures

```go
func (h *Handler) HandleBatchTask(ctx context.Context, task *asynq.Task) error {
    items := h.getItems()
    
    var errors []error
    for _, item := range items {
        if err := h.processItem(item); err != nil {
            // Log but continue
            h.logger.Error("Item failed", zap.Error(err))
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        // Return error to trigger retry for failed items only
        return fmt.Errorf("%d items failed", len(errors))
    }
    
    return nil
}
```

### 5. Use Structured Payloads

```go
// Good: Typed struct
type WellDefinedPayload struct {
    ResourceID string `json:"resource_id"`
    Action     string `json:"action"`
    Options    map[string]interface{} `json:"options"`
}

// Bad: Generic map
payload := map[string]interface{}{
    "id": "123",
    "do_thing": true,
}
```

## Testing

### Unit Testing Handlers

```go
// internal/jobs/tasks/data_export_test.go
package tasks

import (
    "context"
    "testing"
    
    "github.com/hibiken/asynq"
    "github.com/stretchr/testify/assert"
)

func TestDataExportHandler(t *testing.T) {
    // Setup test DB
    db := setupTestDB(t)
    
    // Create handler
    handler := NewDataExportHandler(db, testConfig, testLogger)
    
    // Create test payload
    payload := DataExportPayload{
        TenantID:   "tenant-id",
        UserID:     "user-id",
        ExportType: "full",
    }
    payloadBytes, _ := json.Marshal(payload)
    
    // Create task
    task := asynq.NewTask(TypeDataExport, payloadBytes)
    
    // Execute
    err := handler.HandleDataExport(context.Background(), task)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify results
    // ...
}
```

### Integration Testing

```bash
# Start dependencies
docker-compose up -d postgres redis

# Run worker
go run cmd/worker/main.go &

# Enqueue test task
go run scripts/test_job.go

# Check logs
docker logs -f utm-worker
```

## Related Documentation

- [Jobs Architecture](/jobs/architecture) - System design
- [Available Jobs](/jobs/available-jobs) - Existing jobs
- [Job Monitoring](/jobs/monitoring) - Monitoring and debugging
