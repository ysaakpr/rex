# Overview

Rex is an **open-source user management platform** built with Go, featuring complete authentication, authorization (RBAC), and multi-tenant capabilities built-in.

## What is Rex?

Rex provides a **complete user management foundation** with tenant management and RBAC built-in. It handles all the complex authentication, authorization, and tenant isolation logic, so you can focus on your core business features.

## Key Capabilities

### 🔐 Authentication
- **SuperTokens Integration**: Production-grade authentication with session management
- **Multiple Auth Methods**: Email/password + Google OAuth (extensible)
- **Dual Auth Modes**: Cookie-based (web) and header-based (API/mobile)
- **System Users**: M2M authentication for background services and integrations
- **Automatic Session Management**: Token refresh and expiry handling

### 🏢 Multi-Tenancy
- **Complete Isolation**: Data separation at the tenant level
- **Flexible Onboarding**: Self-service or managed tenant creation
- **Member Management**: Invite users, assign roles, manage access
- **Team Collaboration**: Multiple users per tenant with different permissions

### 🛡️ Authorization (RBAC)
- **3-Tier System**: Roles → Policies → Permissions
- **Fine-Grained Control**: Permission format: `service:entity:action`
- **Dynamic Management**: Create custom roles and permissions via API
- **Platform vs Tenant**: System-wide and tenant-specific roles

### ⚡ Background Jobs
- **Async Processing**: Redis-backed queue with Asynq
- **Built-in Jobs**: Tenant initialization, email sending
- **Scheduled Tasks**: Cron-like periodic job execution
- **Extensible**: Easy to add custom background jobs

## When to Use Rex

### ✅ Perfect For

**SaaS Applications**
- Building B2B SaaS products
- Need multi-tenant architecture
- Require team collaboration features
- Want fine-grained permissions

**API Platforms**
- Creating API services with authentication
- Need M2M (machine-to-machine) auth
- Building integrations platform
- Require system user management

**Enterprise Tools**
- Internal business applications
- Department/team isolation
- Role-based access control
- Audit and compliance needs

**Marketplaces**
- Multi-vendor platforms
- Separate seller/buyer workspaces
- Complex permission requirements
- Invitation-based onboarding

### ❌ Not Ideal For

- **Simple Single-User Apps**: Overkill for single-user applications
- **Public Websites**: No authentication needed
- **Microservices**: If you need per-service auth (though system users can help)
- **Real-time Only**: Primary focus is REST API, not WebSocket/real-time

## How It Works

### Request Flow

```
1. User makes request to API
   ↓
2. SuperTokens middleware verifies session
   ↓
3. AuthMiddleware extracts user ID
   ↓
4. TenantAccessMiddleware checks tenant membership
   ↓
5. RBACMiddleware verifies permissions (if needed)
   ↓
6. Handler processes request
   ↓
7. Response returned to user
```

### Data Model

```
Platform Level:
├── Users (SuperTokens)
├── Platform Admins
└── System Users

Tenant Level:
├── Tenants
├── Tenant Members (User + Role)
├── Invitations
└── Tenant-specific data

RBAC:
├── Roles (Admin, Writer, Viewer, Basic)
├── Policies (groups of permissions)
└── Permissions (service:entity:action)
```

## Core Concepts

### Tenant
An isolated workspace/organization. Each tenant has:
- Unique slug (identifier)
- Members with roles
- Isolated data
- Status (pending, active, suspended, deleted)

### Member
A user belonging to a tenant with a specific role:
- One user can be member of multiple tenants
- Each membership has one role
- Status: active, inactive, pending

### Role
User's position in a tenant (previously called "Relation"):
- Examples: Admin, Writer, Viewer, Basic
- Contains one or more policies
- Can be system-wide or tenant-specific

### Policy
A group of related permissions (previously called "Role"):
- Examples: "Tenant Admin Policy", "Content Writer Policy"
- Makes permission management easier
- Reusable across multiple roles

### Permission
A specific action that can be performed:
- Format: `service:entity:action`
- Examples: `tenant-api:member:invite`, `tenant-api:content:delete`
- Atomic and composable

### System User
Service account for M2M authentication:
- Used by background workers, integrations, cron jobs
- Authenticates like regular users (email/password)
- Special flags in JWT token
- Longer token expiry (24 hours)

### Platform Admin
Super users with system-wide access:
- Can access any tenant without membership
- Manage system users
- Create managed tenants
- Manage RBAC configuration

## Technology Stack

### Backend
- **Go 1.23+**: Modern, fast, and efficient
- **Gin**: High-performance HTTP framework
- **GORM**: Powerful ORM with migrations
- **SuperTokens**: Authentication and session management
- **PostgreSQL**: Reliable relational database
- **Redis**: Fast cache and job queue
- **Asynq**: Background job processing

### Frontend
- **React 18**: Modern UI library
- **Vite**: Fast build tool
- **TailwindCSS**: Utility-first styling
- **SuperTokens React SDK**: Pre-built auth UI
- **React Router v6**: Client-side routing

### DevOps
- **Docker Compose**: Easy local development
- **Nginx**: Reverse proxy and SSL termination
- **golang-migrate**: Database migrations
- **MailHog**: Email testing (development)

## Architecture Highlights

### Clean Architecture
```
presentation/ (handlers, middleware)
    ↓
business/     (services)
    ↓
data/         (repositories, models)
    ↓
infrastructure/ (database, cache, queue)
```

### Security
- **No Password Storage**: SuperTokens handles all authentication
- **HTTP-only Cookies**: XSS protection
- **CSRF Protection**: Built into SuperTokens
- **SQL Injection Protection**: GORM prepared statements
- **Input Validation**: Gin binding with validation tags

### Scalability
- **Stateless API**: Horizontal scaling ready
- **Session Store**: Redis for distributed sessions
- **Job Queue**: Redis for background processing
- **Database**: PostgreSQL with proper indexing

## Getting Started

Ready to dive in? Here's your path:

1. **[Quick Start](/getting-started/quick-start)** - Get it running in 5 minutes
2. **[Architecture](/introduction/architecture)** - Understand the system design
3. **[Core Concepts](/introduction/core-concepts)** - Learn the terminology
4. **[Authentication Guide](/guides/authentication)** - Master the auth system

## Project Status

- ✅ **Production Ready**: Battle-tested authentication and authorization
- ✅ **Actively Maintained**: Regular updates and bug fixes
- ✅ **Well Documented**: Comprehensive guides and API reference
- ✅ **Docker Support**: Easy deployment and development
- ✅ **Migration Support**: Database versioning included

## Community

- **GitHub**: [github.com/yourusername/utm-backend](https://github.com/yourusername/utm-backend)
- **Discussions**: Ask questions and share ideas
- **Issues**: Report bugs and request features
- **Contributing**: Pull requests welcome!

## License

Rex is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

