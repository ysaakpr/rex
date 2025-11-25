# Backend Integration

Complete guide for integrating and extending the Go backend.

## Overview

This guide covers:
- Project structure
- Adding new API endpoints
- Database operations with GORM
- Background jobs with Asynq
- Service layer patterns
- Error handling
- Testing strategies

## Project Structure

```
internal/
├── api/              # HTTP layer
│   ├── handlers/     # Request handlers
│   ├── middleware/   # Middleware functions
│   └── router/       # Route definitions
├── models/           # Database models (GORM)
├── services/         # Business logic
├── repository/       # Data access layer
├── jobs/             # Background jobs
├── config/           # Configuration
└── database/         # DB connection & migrations
```

### Layered Architecture

```
HTTP Request
  ↓
Handler (validation, response formatting)
  ↓
Service (business logic)
  ↓
Repository (database access)
  ↓
Database
```

## Adding New API Endpoints

### Step 1: Define Model

```go
// internal/models/article.go
package models

import (
    "time"
    "github.com/google/uuid"
)

type Article struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
    Title       string    `gorm:"type:varchar(200);not null" json:"title"`
    Content     string    `gorm:"type:text" json:"content"`
    Status      string    `gorm:"type:varchar(20);default:'draft'" json:"status"`
    AuthorID    string    `gorm:"type:varchar(255);not null" json:"author_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateArticleInput struct {
    Title   string `json:"title" binding:"required,min=3,max=200"`
    Content string `json:"content" binding:"required"`
}

type UpdateArticleInput struct {
    Title   *string `json:"title,omitempty" binding:"omitempty,min=3,max=200"`
    Content *string `json:"content,omitempty"`
    Status  *string `json:"status,omitempty" binding:"omitempty,oneof=draft published archived"`
}
```

### Step 2: Create Repository

```go
// internal/repository/article_repository.go
package repository

import (
    "gorm.io/gorm"
    "yourproject/internal/models"
    "github.com/google/uuid"
)

type ArticleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
    return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *models.Article) error {
    return r.db.Create(article).Error
}

func (r *ArticleRepository) FindByID(id uuid.UUID) (*models.Article, error) {
    var article models.Article
    err := r.db.First(&article, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &article, nil
}

func (r *ArticleRepository) FindByTenantID(tenantID uuid.UUID, page, pageSize int) ([]models.Article, int64, error) {
    var articles []models.Article
    var total int64
    
    offset := (page - 1) * pageSize
    
    err := r.db.Model(&models.Article{}).
        Where("tenant_id = ?", tenantID).
        Count(&total).
        Offset(offset).
        Limit(pageSize).
        Order("created_at DESC").
        Find(&articles).
        Error
    
    return articles, total, err
}

func (r *ArticleRepository) Update(article *models.Article) error {
    return r.db.Save(article).Error
}

func (r *ArticleRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.Article{}, "id = ?", id).Error
}
```

### Step 3: Create Service

```go
// internal/services/article_service.go
package services

import (
    "fmt"
    "yourproject/internal/models"
    "yourproject/internal/repository"
    "github.com/google/uuid"
)

type ArticleService struct {
    repo *repository.ArticleRepository
}

func NewArticleService(repo *repository.ArticleRepository) *ArticleService {
    return &ArticleService{repo: repo}
}

func (s *ArticleService) CreateArticle(tenantID uuid.UUID, authorID string, input models.CreateArticleInput) (*models.Article, error) {
    article := &models.Article{
        TenantID: tenantID,
        AuthorID: authorID,
        Title:    input.Title,
        Content:  input.Content,
        Status:   "draft",
    }
    
    if err := s.repo.Create(article); err != nil {
        return nil, fmt.Errorf("failed to create article: %w", err)
    }
    
    return article, nil
}

func (s *ArticleService) GetArticle(id uuid.UUID) (*models.Article, error) {
    article, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("article not found: %w", err)
    }
    return article, nil
}

func (s *ArticleService) ListArticles(tenantID uuid.UUID, page, pageSize int) ([]models.Article, int64, error) {
    return s.repo.FindByTenantID(tenantID, page, pageSize)
}

func (s *ArticleService) UpdateArticle(id uuid.UUID, input models.UpdateArticleInput) (*models.Article, error) {
    article, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("article not found: %w", err)
    }
    
    if input.Title != nil {
        article.Title = *input.Title
    }
    if input.Content != nil {
        article.Content = *input.Content
    }
    if input.Status != nil {
        article.Status = *input.Status
    }
    
    if err := s.repo.Update(article); err != nil {
        return nil, fmt.Errorf("failed to update article: %w", err)
    }
    
    return article, nil
}

func (s *ArticleService) DeleteArticle(id uuid.UUID) error {
    return s.repo.Delete(id)
}

func (s *ArticleService) PublishArticle(id uuid.UUID) (*models.Article, error) {
    article, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("article not found: %w", err)
    }
    
    article.Status = "published"
    
    if err := s.repo.Update(article); err != nil {
        return nil, fmt.Errorf("failed to publish article: %w", err)
    }
    
    return article, nil
}
```

### Step 4: Create Handler

```go
// internal/api/handlers/article_handler.go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/supertokens/supertokens-golang/recipe/session"
    
    "yourproject/internal/models"
    "yourproject/internal/pkg/response"
    "yourproject/internal/services"
)

type ArticleHandler struct {
    service *services.ArticleService
}

func NewArticleHandler(service *services.ArticleService) *ArticleHandler {
    return &ArticleHandler{service: service}
}

func (h *ArticleHandler) CreateArticle(c *gin.Context) {
    // Get tenant ID from URL
    tenantID, err := uuid.Parse(c.Param("tenant_id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid tenant ID")
        return
    }
    
    // Get user ID from session
    sessionContainer := session.GetSessionFromRequestContext(c.Request.Context())
    authorID := sessionContainer.GetUserID()
    
    // Parse request body
    var input models.CreateArticleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        response.ValidationError(c, err)
        return
    }
    
    // Create article
    article, err := h.service.CreateArticle(tenantID, authorID, input)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.Success(c, http.StatusCreated, "Article created successfully", article)
}

func (h *ArticleHandler) GetArticle(c *gin.Context) {
    articleID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid article ID")
        return
    }
    
    article, err := h.service.GetArticle(articleID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "Article not found")
        return
    }
    
    response.Success(c, http.StatusOK, "", article)
}

func (h *ArticleHandler) ListArticles(c *gin.Context) {
    tenantID, err := uuid.Parse(c.Param("tenant_id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid tenant ID")
        return
    }
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    articles, total, err := h.service.ListArticles(tenantID, page, pageSize)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.PaginatedSuccess(c, articles, total, page, pageSize)
}

func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
    articleID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid article ID")
        return
    }
    
    var input models.UpdateArticleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        response.ValidationError(c, err)
        return
    }
    
    article, err := h.service.UpdateArticle(articleID, input)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.Success(c, http.StatusOK, "Article updated successfully", article)
}

func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
    articleID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid article ID")
        return
    }
    
    if err := h.service.DeleteArticle(articleID); err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.Success(c, http.StatusOK, "Article deleted successfully", nil)
}

func (h *ArticleHandler) PublishArticle(c *gin.Context) {
    articleID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid article ID")
        return
    }
    
    article, err := h.service.PublishArticle(articleID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.Success(c, http.StatusOK, "Article published successfully", article)
}
```

### Step 5: Register Routes

```go
// internal/api/router/router.go
func SetupRouter(
    // ... existing dependencies
    articleHandler *handlers.ArticleHandler,
) *gin.Engine {
    router := gin.Default()
    
    // ... existing routes
    
    // Article routes
    articles := v1.Group("/tenants/:tenant_id/articles")
    articles.Use(middleware.RequireTenantAccess())
    {
        articles.POST("",
            middleware.RequirePermission("blog-api", "article", "create"),
            articleHandler.CreateArticle,
        )
        articles.GET("",
            middleware.RequirePermission("blog-api", "article", "read"),
            articleHandler.ListArticles,
        )
        articles.GET("/:id",
            middleware.RequirePermission("blog-api", "article", "read"),
            articleHandler.GetArticle,
        )
        articles.PUT("/:id",
            middleware.RequirePermission("blog-api", "article", "update"),
            articleHandler.UpdateArticle,
        )
        articles.DELETE("/:id",
            middleware.RequirePermission("blog-api", "article", "delete"),
            articleHandler.DeleteArticle,
        )
        articles.POST("/:id/publish",
            middleware.RequirePermission("blog-api", "article", "publish"),
            articleHandler.PublishArticle,
        )
    }
    
    return router
}
```

### Step 6: Wire Dependencies

```go
// cmd/api/main.go
func main() {
    // ... existing setup
    
    // Initialize article components
    articleRepo := repository.NewArticleRepository(db)
    articleService := services.NewArticleService(articleRepo)
    articleHandler := handlers.NewArticleHandler(articleService)
    
    // Setup router
    router := router.SetupRouter(
        // ... existing dependencies
        articleHandler,
    )
    
    router.Run(":8080")
}
```

## Database Operations

### GORM Basics

```go
// Create
article := &models.Article{Title: "Test", Content: "Content"}
db.Create(article)

// Read
var article models.Article
db.First(&article, "id = ?", id)  // Find by ID
db.Where("status = ?", "published").Find(&articles)  // Find all matching

// Update
db.Model(&article).Update("status", "published")
db.Model(&article).Updates(models.Article{Status: "published", Title: "New"})

// Delete
db.Delete(&article, "id = ?", id)
```

### Transactions

```go
func (s *ArticleService) PublishWithNotification(articleID uuid.UUID) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Update article
        var article models.Article
        if err := tx.First(&article, "id = ?", articleID).Error; err != nil {
            return err
        }
        
        article.Status = "published"
        if err := tx.Save(&article).Error; err != nil {
            return err
        }
        
        // Create notification
        notification := &models.Notification{
            UserID:  article.AuthorID,
            Message: "Your article has been published",
        }
        if err := tx.Create(notification).Error; err != nil {
            return err
        }
        
        return nil  // Commit
    })
}
```

### Associations

```go
// Preload related data
db.Preload("Author").Find(&articles)

// Join tables
db.Joins("JOIN users ON users.id = articles.author_id").
    Where("users.status = ?", "active").
    Find(&articles)
```

## Background Jobs

### Create Job Task

```go
// internal/jobs/tasks/article_tasks.go
package tasks

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/hibiken/asynq"
    "github.com/google/uuid"
)

const (
    TypeArticlePublish = "article:publish"
    TypeArticleArchive = "article:archive"
)

type ArticlePublishPayload struct {
    ArticleID uuid.UUID `json:"article_id"`
    TenantID  uuid.UUID `json:"tenant_id"`
}

func NewArticlePublishTask(articleID, tenantID uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal(ArticlePublishPayload{
        ArticleID: articleID,
        TenantID:  tenantID,
    })
    if err != nil {
        return nil, err
    }
    return asynq.NewTask(TypeArticlePublish, payload), nil
}

func HandleArticlePublish(ctx context.Context, t *asynq.Task) error {
    var payload ArticlePublishPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    // Your publish logic here
    fmt.Printf("Publishing article %s in tenant %s\n", 
        payload.ArticleID, payload.TenantID)
    
    // Update article status, send notifications, etc.
    
    return nil
}
```

### Enqueue Job from Service

```go
func (s *ArticleService) SchedulePublish(articleID, tenantID uuid.UUID, publishAt time.Time) error {
    task, err := tasks.NewArticlePublishTask(articleID, tenantID)
    if err != nil {
        return err
    }
    
    // Schedule for specific time
    info, err := s.asynqClient.Enqueue(task, asynq.ProcessAt(publishAt))
    if err != nil {
        return fmt.Errorf("failed to enqueue task: %w", err)
    }
    
    log.Printf("Scheduled article publish: task_id=%s", info.ID)
    return nil
}
```

### Register Worker Handler

```go
// internal/jobs/worker.go
func StartWorker(db *gorm.DB, redisAddr string) {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr},
        asynq.Config{Concurrency: 10},
    )
    
    mux := asynq.NewServeMux()
    
    // Register handlers
    mux.HandleFunc(tasks.TypeArticlePublish, tasks.HandleArticlePublish)
    mux.HandleFunc(tasks.TypeArticleArchive, tasks.HandleArticleArchive)
    
    if err := srv.Run(mux); err != nil {
        log.Fatal(err)
    }
}
```

## Error Handling

### Custom Errors

```go
// internal/pkg/errors/errors.go
package errors

import "fmt"

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

var (
    ErrNotFound      = &AppError{Code: "NOT_FOUND", Message: "Resource not found"}
    ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized"}
    ErrForbidden     = &AppError{Code: "FORBIDDEN", Message: "Forbidden"}
    ErrValidation    = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed"}
)

func NotFound(message string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: message}
}
```

### Error Handling in Handlers

```go
func (h *ArticleHandler) GetArticle(c *gin.Context) {
    article, err := h.service.GetArticle(articleID)
    if err != nil {
        // Check error type
        var appErr *errors.AppError
        if errors.As(err, &appErr) {
            switch appErr.Code {
            case "NOT_FOUND":
                response.Error(c, http.StatusNotFound, appErr.Message)
            case "FORBIDDEN":
                response.Error(c, http.StatusForbidden, appErr.Message)
            default:
                response.Error(c, http.StatusInternalServerError, "Internal error")
            }
            return
        }
        
        // Unknown error
        response.Error(c, http.StatusInternalServerError, "An error occurred")
        return
    }
    
    response.Success(c, http.StatusOK, "", article)
}
```

## Testing

### Unit Test (Service)

```go
// internal/services/article_service_test.go
package services_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "yourproject/internal/models"
    "yourproject/internal/services"
)

type MockArticleRepository struct {
    mock.Mock
}

func (m *MockArticleRepository) Create(article *models.Article) error {
    args := m.Called(article)
    return args.Error(0)
}

func TestCreateArticle(t *testing.T) {
    // Setup
    mockRepo := new(MockArticleRepository)
    service := services.NewArticleService(mockRepo)
    
    input := models.CreateArticleInput{
        Title:   "Test Article",
        Content: "Test content",
    }
    
    mockRepo.On("Create", mock.AnythingOfType("*models.Article")).Return(nil)
    
    // Execute
    article, err := service.CreateArticle(tenantID, authorID, input)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, article)
    assert.Equal(t, "Test Article", article.Title)
    mockRepo.AssertExpectations(t)
}
```

### Integration Test (Handler)

```go
// internal/api/handlers/article_handler_test.go
func TestCreateArticle_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // Create dependencies
    repo := repository.NewArticleRepository(db)
    service := services.NewArticleService(repo)
    handler := handlers.NewArticleHandler(service)
    
    // Setup router
    router := gin.New()
    router.POST("/tenants/:tenant_id/articles", handler.CreateArticle)
    
    // Create test request
    body := `{"title":"Test","content":"Content"}`
    req := httptest.NewRequest("POST", "/tenants/"+tenantID.String()+"/articles", 
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, 201, w.Code)
}
```

## Best Practices

1. **Layered Architecture**: Keep handlers thin, business logic in services
2. **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error context
3. **Transactions**: Use transactions for multi-step operations
4. **Logging**: Log important events, errors, and slow queries
5. **Validation**: Validate input at handler layer
6. **Testing**: Write unit tests for services, integration tests for handlers
7. **Documentation**: Add godoc comments for public functions

## Next Steps

- [GORM Documentation](https://gorm.io/docs/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [Asynq Documentation](https://github.com/hibiken/asynq)
- [Testing Guide](/guides/testing)

