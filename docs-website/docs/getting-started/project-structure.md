# Project Structure

Understanding the codebase organization and conventions.

## Repository Layout

```
utm-backend/
├── cmd/                    # Application entry points
│   ├── api/               # Main API server
│   ├── migrate/           # Migration CLI tool
│   └── worker/            # Background worker
│
├── internal/              # Private application code
│   ├── api/              # HTTP layer
│   │   ├── handlers/     # Request handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── router/       # Route definitions
│   ├── config/           # Configuration management
│   ├── database/         # Database connection
│   ├── jobs/             # Background jobs
│   ├── models/           # Data models (GORM)
│   ├── pkg/              # Internal packages
│   ├── repository/       # Data access layer
│   └── services/         # Business logic
│
├── frontend/              # React application
│   ├── public/           # Static assets
│   └── src/              # React source code
│       ├── components/   # React components
│       ├── App.jsx       # Main app
│       └── main.jsx      # Entry point
│
├── migrations/            # Database migrations
│   ├── *.up.sql         # Migration scripts
│   └── *.down.sql       # Rollback scripts
│
├── scripts/               # Helper scripts
│   ├── seed_rbac.sql    # RBAC seed data
│   └── *.sh             # Bash scripts
│
├── docs/                  # Documentation (old)
├── docs-website/          # New documentation site
├── infra/                 # Infrastructure as code
│   └── *.go             # Pulumi definitions
│
├── .env.example           # Environment template
├── docker-compose.yml     # Development services
├── Dockerfile             # API/Worker image
├── go.mod                 # Go dependencies
├── Makefile              # Common commands
└── README.md             # Project overview
```

## Backend Structure (Go)

### cmd/ - Application Entry Points

**Purpose**: Main packages that can be built into executables

**Files**:
- `cmd/api/main.go` - HTTP API server
- `cmd/worker/main.go` - Background job worker
- `cmd/migrate/main.go` - Database migration tool

**Pattern**:
```go
// cmd/api/main.go
package main

func main() {
    // 1. Load configuration
    cfg := config.Load()
    
    // 2. Initialize dependencies
    db := database.Connect(cfg)
    logger := initLogger()
    
    // 3. Setup application
    router := setupRouter(db, logger)
    
    // 4. Start server
    router.Run(":8080")
}
```

### internal/ - Private Code

**Purpose**: Application code not meant for external import

**Why internal/**: Go toolchain prevents external packages from importing `internal/` code

### internal/api/ - HTTP Layer

**handlers/**:
```
handlers/
├── tenant_handler.go       # Tenant endpoints
├── member_handler.go       # Member endpoints
├── invitation_handler.go   # Invitation endpoints
├── rbac_handler.go         # RBAC endpoints
├── system_user_handler.go  # System user endpoints
├── platform_admin_handler.go
└── user_handler.go
```

**Pattern**:
```go
type TenantHandler struct {
    service services.TenantService
    logger  *zap.Logger
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // 1. Parse request
    var input models.CreateTenantInput
    if err := c.ShouldBindJSON(&input); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    
    // 2. Get authenticated user
    userID, _ := middleware.GetUserID(c)
    
    // 3. Call service
    tenant, err := h.service.CreateTenant(userID, &input)
    if err != nil {
        response.InternalServerError(c, err)
        return
    }
    
    // 4. Return response
    response.Success(c, "Tenant created", tenant)
}
```

**middleware/**:
```
middleware/
├── auth.go              # Authentication
├── tenant.go            # Tenant access control
├── rbac.go              # Permission checking
├── platform_admin.go    # Platform admin check
├── cors.go              # CORS handling
└── logger.go            # Request logging
```

**router/**:
- `router.go` - All route definitions

**Pattern**:
```go
func SetupRouter(deps *RouterDeps) *gin.Engine {
    router := gin.New()
    
    // Global middleware
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    
    // Route groups
    v1 := router.Group("/api/v1")
    {
        // Public routes
        v1.GET("/invitations/:token", handler.GetInvitation)
        
        // Authenticated routes
        auth := v1.Group("")
        auth.Use(middleware.AuthMiddleware())
        {
            auth.GET("/tenants", handler.ListTenants)
        }
    }
    
    return router
}
```

### internal/models/ - Data Models

**Purpose**: GORM models and API structures

**Files**:
```
models/
├── common.go              # Shared fields
├── tenant.go              # Tenant model
├── tenant_member.go       # Member model
├── role.go                # Role model
├── policy.go              # Policy model
├── permission.go          # Permission model
├── system_user.go         # System user model
├── invitation.go          # Invitation model
└── platform_admin.go      # Platform admin model
```

**Pattern**:
```go
// Model (database)
type Tenant struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name      string    `gorm:"size:255;not null"`
    Slug      string    `gorm:"size:255;not null;uniqueIndex"`
    Status    string    `gorm:"size:50;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Input (API request)
type CreateTenantInput struct {
    Name     string  `json:"name" binding:"required,min=2,max=255"`
    Slug     string  `json:"slug" binding:"required,min=3,max=255,alphanum"`
    Metadata JSONMap `json:"metadata"`
}

// Response (API response)
type TenantResponse struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Slug        string    `json:"slug"`
    Status      string    `json:"status"`
    MemberCount int       `json:"member_count"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### internal/services/ - Business Logic

**Purpose**: Core business operations

**Files**:
```
services/
├── tenant_service.go      # Tenant operations
├── member_service.go      # Member operations
├── invitation_service.go  # Invitation operations
├── rbac_service.go        # RBAC operations
├── system_user_service.go # System user operations
└── user_service.go        # User operations
```

**Pattern**:
```go
type TenantService interface {
    CreateTenant(userID string, input *CreateTenantInput) (*Tenant, error)
    GetTenant(id uuid.UUID) (*Tenant, error)
    UpdateTenant(id uuid.UUID, input *UpdateTenantInput) error
    DeleteTenant(id uuid.UUID) error
}

type tenantService struct {
    repo   repository.TenantRepository
    queue  *asynq.Client
    logger *zap.Logger
}

func (s *tenantService) CreateTenant(...) (*Tenant, error) {
    // 1. Validate
    // 2. Create tenant
    // 3. Add creator as admin
    // 4. Enqueue initialization job
    // 5. Return tenant
}
```

### internal/repository/ - Data Access

**Purpose**: Database queries (GORM)

**Files**:
```
repository/
├── tenant_repository.go
├── member_repository.go
├── role_repository.go
├── policy_repository.go
├── permission_repository.go
└── system_user_repository.go
```

**Pattern**:
```go
type TenantRepository interface {
    Create(tenant *Tenant) error
    FindByID(id uuid.UUID) (*Tenant, error)
    FindBySlug(slug string) (*Tenant, error)
    Update(tenant *Tenant) error
    Delete(id uuid.UUID) error
}

type tenantRepository struct {
    db *gorm.DB
}

func (r *tenantRepository) FindByID(id uuid.UUID) (*Tenant, error) {
    var tenant Tenant
    err := r.db.Where("id = ?", id).First(&tenant).Error
    return &tenant, err
}
```

### internal/jobs/ - Background Jobs

**Purpose**: Async task processing

**Files**:
```
jobs/
├── worker.go              # Worker setup
├── tenant_init.go         # Tenant initialization
├── invitation_email.go    # Send invitation
└── system_user_expiry.go  # Cleanup expired users
```

**Pattern**:
```go
// Define task type
const TypeTenantInit = "tenant:init"

// Task payload
type TenantInitPayload struct {
    TenantID uuid.UUID `json:"tenant_id"`
}

// Handler
func HandleTenantInit(ctx context.Context, t *asynq.Task) error {
    var payload TenantInitPayload
    json.Unmarshal(t.Payload(), &payload)
    
    // Do work...
    
    return nil
}

// Enqueue (from service)
client.Enqueue(
    asynq.NewTask(TypeTenantInit, payloadBytes),
    asynq.Queue("default"),
)
```

### internal/config/ - Configuration

**Purpose**: Environment variable management

**File**: `config/config.go`

**Pattern**:
```go
type Config struct {
    App struct {
        Env  string
        Port string
    }
    Database struct {
        Host     string
        Port     string
        User     string
        Password string
        Name     string
    }
    // ... more config
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    viper.ReadInConfig()
    
    var cfg Config
    viper.Unmarshal(&cfg)
    return &cfg
}
```

## Frontend Structure (React)

```
frontend/src/
├── components/
│   ├── layout/
│   │   ├── DashboardLayout.jsx
│   │   ├── Header.jsx
│   │   └── Sidebar.jsx
│   ├── pages/
│   │   ├── TenantsPage.jsx
│   │   ├── TenantDetailPage.jsx
│   │   ├── MembersPage.jsx
│   │   └── ...
│   └── ui/
│       ├── Button.jsx
│       ├── Card.jsx
│       └── ...
├── App.jsx           # Main app with routing
├── main.jsx          # Entry point
└── index.css         # Global styles
```

**Pattern**:
```jsx
// Component structure
export default function TenantsPage() {
    const [tenants, setTenants] = useState([]);
    const [loading, setLoading] = useState(true);
    
    useEffect(() => {
        fetchTenants();
    }, []);
    
    const fetchTenants = async () => {
        const response = await fetch('/api/v1/tenants', {
            credentials: 'include'
        });
        const data = await response.json();
        setTenants(data.data);
        setLoading(false);
    };
    
    return (
        <div>
            {/* JSX */}
        </div>
    );
}
```

## File Naming Conventions

### Go Files
- **lowercase_with_underscores.go** for all files
- **_test.go** suffix for tests
- **Package name** = directory name

Examples:
- `tenant_handler.go`
- `tenant_service.go`
- `tenant_repository_test.go`

### React Files
- **PascalCase.jsx** for components
- **camelCase.js** for utilities
- **kebab-case.css** for styles

Examples:
- `TenantsPage.jsx`
- `DashboardLayout.jsx`
- `apiClient.js`

## Import Organization

### Go
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"
    
    // 2. External packages
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    // 3. Internal packages
    "github.com/yourorg/utm-backend/internal/models"
    "github.com/yourorg/utm-backend/internal/services"
)
```

### JavaScript/React
```javascript
// 1. External packages
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

// 2. Internal components
import DashboardLayout from '../layout/DashboardLayout';
import Button from '../ui/Button';

// 3. Styles
import './TenantsPage.css';
```

## Configuration Files

### Root Level
- `.env` - Environment variables (not committed)
- `.env.example` - Environment template (committed)
- `.gitignore` - Git ignore rules
- `go.mod`, `go.sum` - Go dependencies
- `Makefile` - Common commands
- `docker-compose.yml` - Development services
- `Dockerfile` - Production image

### Frontend
- `package.json` - npm dependencies
- `vite.config.js` - Vite configuration
- `tailwind.config.js` - TailwindCSS config

## Next Steps

- **[Configuration](/getting-started/configuration)** - Environment variables
- **[Authentication](/guides/authentication)** - Auth system
- **[API Reference](/x-api/overview)** - API endpoints

