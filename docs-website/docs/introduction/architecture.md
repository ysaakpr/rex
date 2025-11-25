# Architecture

This page explains the overall architecture, design decisions, and how different components work together.

## System Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────┐
│                   Browser                        │
└────────────────┬────────────────────────────────┘
                 │ HTTPS
                 ↓
┌─────────────────────────────────────────────────┐
│              Nginx Reverse Proxy                 │
│  - SSL/TLS Termination                          │
│  - Route to frontend/backend                    │
│  - Static file serving                          │
└──────────┬─────────────────┬────────────────────┘
           │                 │
           ↓                 ↓
┌──────────────────┐  ┌──────────────────────────┐
│  React Frontend  │  │    Go Backend API        │
│  Port: 3000      │  │    Port: 8080            │
│                  │  │                          │
│  - SuperTokens   │  │  - Gin Framework         │
│    React SDK     │  │  - SuperTokens SDK       │
│  - UI Components │  │  - Middleware Chain      │
│  - State Mgmt    │  │  - Business Logic        │
└──────────────────┘  └────┬──────┬──────┬───────┘
                           │      │      │
                ┌──────────┘      │      └────────┐
                ↓                 ↓               ↓
        ┌──────────────┐  ┌─────────────┐  ┌─────────┐
        │  PostgreSQL  │  │    Redis    │  │SuperTokens│
        │  Port: 5432  │  │  Port: 6379 │  │Core :3567│
        │              │  │             │  │          │
        │  - Main DB   │  │  - Sessions │  │  - Auth  │
        │  - GORM      │  │  - Job Queue│  │  - Users │
        └──────────────┘  └─────────────┘  └─────────┘
                                 │
                                 ↓
                         ┌──────────────┐
                         │    Worker    │
                         │  (Background)│
                         │              │
                         │  - Asynq     │
                         │  - Jobs      │
                         └──────────────┘
```

## Component Breakdown

### 1. Frontend (React)

**Technology**: React 18 + Vite + TailwindCSS

**Responsibilities**:
- User interface and experience
- SuperTokens authentication UI
- Protected route management
- API communication
- State management

**Key Features**:
- Server-side rendering ready (Vite)
- Component-based architecture
- TailwindCSS for styling
- SuperTokens pre-built auth UI
- Cookie-based session management

**Directory Structure**:
```
frontend/src/
├── components/
│   ├── layout/         # Layout components
│   ├── pages/          # Page components
│   └── ui/             # Reusable UI components
├── App.jsx             # Main app with routing
├── main.jsx            # Entry point
└── index.css           # Global styles
```

### 2. Backend API (Go)

**Technology**: Go 1.23+ + Gin + GORM

**Responsibilities**:
- Business logic
- Authentication verification
- Authorization enforcement
- Data persistence
- Job enqueueing

**Architecture Pattern**: Clean Architecture

```
cmd/api/main.go         # Entry point
    ↓
internal/api/router     # Route definitions
    ↓
internal/api/middleware # Auth, RBAC, logging
    ↓
internal/api/handlers   # HTTP handlers
    ↓
internal/services       # Business logic
    ↓
internal/repository     # Data access
    ↓
internal/models         # Data models
```

**Middleware Chain**:
```
Request
  ↓
1. Logger              # Log all requests
  ↓
2. CORS                # Handle cross-origin
  ↓
3. SuperTokens         # SuperTokens routes
  ↓
4. AuthMiddleware      # Verify session
  ↓
5. TenantAccess        # Check tenant membership
  ↓
6. RBAC (optional)     # Check permissions
  ↓
Handler
```

### 3. PostgreSQL Database

**Purpose**: Primary data store

**Schema Organization**:
```sql
-- Core tables
tenants
tenant_members
user_invitations

-- RBAC tables
roles            (user's role in tenant)
policies         (group of permissions)
permissions      (atomic actions)
role_policies    (M2M mapping)
policy_permissions (M2M mapping)

-- Platform tables
platform_admins
system_users
```

**Connection**:
- GORM ORM for type-safe queries
- Connection pooling enabled
- Automatic retry on connection failure
- Migration support via golang-migrate

### 4. Redis

**Purpose**: Cache and job queue

**Usage**:
- **Session Store**: SuperTokens session data
- **Job Queue**: Asynq background jobs
- **Cache** (optional): Rate limiting, temporary data

**Configuration**:
```
Host: redis:6379
Password: (configured)
DB: 0 (default)
```

### 5. SuperTokens Core

**Purpose**: Authentication service

**What it Does**:
- User registration and login
- Session management
- Token generation and verification
- OAuth provider integration
- Password reset flows

**Why SuperTokens?**:
- ✅ Production-ready authentication
- ✅ Multiple auth methods (email/password, OAuth)
- ✅ Built-in session management
- ✅ Security best practices
- ✅ Pre-built UI components
- ✅ Multi-framework support

**Communication**:
```
Backend ←→ SuperTokens Core
  - Verify sessions
  - Get user info
  - Manage users

Frontend ←→ SuperTokens Core (via Backend)
  - Sign up/in/out
  - Session refresh
  - OAuth flows
```

### 6. Background Worker

**Purpose**: Async job processing

**Technology**: Asynq (Redis-backed)

**Jobs**:
- Tenant initialization
- Email sending (invitations, notifications)
- System user credential expiry
- Cleanup tasks

**Architecture**:
```
API Handler
    ↓
Enqueue Job → Redis Queue
                   ↓
              Worker picks up
                   ↓
              Executes handler
                   ↓
              Updates DB/sends email
```

## Data Flow

### 1. User Authentication Flow

```
1. User submits email/password
   ↓
2. Frontend sends to SuperTokens endpoint (/api/auth/signin)
   ↓
3. SuperTokens verifies credentials
   ↓
4. SuperTokens creates session
   ↓
5. Session cookies set in response
   ↓
6. Frontend stores cookies (automatic)
   ↓
7. Subsequent requests include cookies
   ↓
8. Backend verifies session with SuperTokens
```

### 2. Tenant Creation Flow

```
1. Authenticated user POSTs to /api/v1/tenants
   ↓
2. AuthMiddleware verifies session
   ↓
3. Handler extracts user ID
   ↓
4. Service creates tenant in DB
   ↓
5. Service adds user as Admin member
   ↓
6. Service enqueues initialization job
   ↓
7. Response returned to user
   ↓
8. Worker processes initialization job (async)
```

### 3. Permission Check Flow

```
1. Request to protected endpoint
   ↓
2. AuthMiddleware verifies user
   ↓
3. TenantAccessMiddleware checks membership
   ↓
4. RBACMiddleware checks permission
   ↓
5. Query: Does user have permission in tenant?
   │
   ├─ Get user's role in tenant
   ├─ Get role's policies
   ├─ Get policies' permissions
   └─ Check if required permission exists
   ↓
6. If yes → continue, if no → 403 Forbidden
```

### 4. Invitation Flow

```
1. Admin creates invitation
   ↓
2. Invitation stored in DB with token
   ↓
3. Background job sends email
   ↓
4. Invited user clicks link
   ↓
5. Frontend shows invitation details (public endpoint)
   ↓
6. User signs up/logs in
   ↓
7. User accepts invitation
   ↓
8. Backend creates tenant membership
   ↓
9. User can now access tenant
```

## Security Architecture

### Authentication Security

**Session Management**:
- HTTP-only cookies (no JavaScript access)
- Secure flag for HTTPS
- SameSite=Lax (CSRF protection)
- Automatic expiry and refresh

**Token Security**:
- JWT for access tokens
- Refresh tokens stored separately
- Token rotation on refresh
- Revocation support

### Authorization Security

**Principle of Least Privilege**:
- Default roles have minimal permissions
- Explicit permission grants
- No implicit access

**Defense in Depth**:
```
Layer 1: Authentication (Is user logged in?)
Layer 2: Tenant Access (Is user member of tenant?)
Layer 3: RBAC (Does user have permission?)
Layer 4: Resource Ownership (Does resource belong to tenant?)
```

### Network Security

**Nginx as Gateway**:
- Single entry point
- SSL/TLS termination
- Rate limiting capability
- DDoS protection

**Internal Network**:
- Services communicate via Docker network
- No external exposure except Nginx
- Database not exposed to internet

## Scalability Considerations

### Horizontal Scaling

**Stateless API**:
- No server-side state (except sessions in Redis)
- Can run multiple API instances
- Load balancer distributes requests

**Worker Scaling**:
- Multiple worker instances possible
- Jobs distributed via Redis
- Automatic failover

### Vertical Scaling

**Database**:
- PostgreSQL with proper indexing
- Connection pooling
- Read replicas for read-heavy workloads

**Redis**:
- In-memory performance
- Persistence for durability
- Clustering for large datasets

### Performance Optimizations

**Database**:
- Indexes on foreign keys
- Composite indexes for common queries
- Pagination on list endpoints
- Soft deletes (no actual row deletion)

**Caching**:
- Redis for session data
- Optional caching for frequent queries
- Cache invalidation strategies

**API**:
- Gin's high performance
- Connection pooling
- Async jobs for slow operations

## Deployment Architecture

### Development

```
docker-compose up
    ↓
All services start locally
    ↓
Hot reload enabled
    ↓
Development URLs:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- MailHog: http://localhost:8025
```

### Production

```
Nginx (443, 80)
    ↓
Frontend (container)
Backend API (container)
    ↓
PostgreSQL (managed service or container)
Redis (managed service or container)
SuperTokens (container or managed)
    ↓
Worker (container)
```

**Recommended Setup**:
- **API**: 2+ instances behind load balancer
- **Database**: Managed PostgreSQL (AWS RDS, etc.)
- **Redis**: Managed Redis (ElastiCache, etc.)
- **Worker**: 1-2 instances
- **Frontend**: CDN + static hosting

## Design Decisions

### Why Go?
- **Performance**: Fast execution and low memory
- **Concurrency**: Excellent goroutine support
- **Type Safety**: Compile-time error checking
- **Standard Library**: Comprehensive and well-designed
- **Deployment**: Single binary, easy to deploy

### Why SuperTokens?
- **Complete Solution**: Authentication + session management
- **Security**: Best practices built-in
- **Flexibility**: Multiple auth methods
- **Open Source**: Self-hosted option
- **SDKs**: Multi-language support

### Why GORM?
- **Productivity**: Less boilerplate code
- **Type Safety**: Go structs as models
- **Migrations**: Built-in migration support
- **Associations**: Easy relationship management
- **Performance**: Efficient query generation

### Why Asynq?
- **Reliability**: Redis-backed durability
- **Features**: Retries, scheduling, priority queues
- **Monitoring**: Built-in inspection tools
- **Simple**: Easy to use and understand
- **Go Native**: Written in Go

### Why Multi-Tenant?
- **SaaS Standard**: Most B2B apps need it
- **Data Isolation**: Security and compliance
- **Resource Efficiency**: Shared infrastructure
- **Scalability**: Add tenants without code changes

## Extension Points

### Adding Custom Middleware
Place in `internal/api/middleware/`

### Adding Custom Jobs
1. Define task type in `internal/jobs/`
2. Create handler in `internal/jobs/tasks/`
3. Register in worker
4. Enqueue from service

### Adding Custom Permissions
Via API: `POST /api/v1/platform/permissions`

### Adding External Services
Integrate in service layer, use repository pattern

## Next Steps

- **[Core Concepts](/introduction/core-concepts)** - Understand the terminology
- **[Quick Start](/getting-started/quick-start)** - Get it running
- **[Project Structure](/getting-started/project-structure)** - Explore the code

