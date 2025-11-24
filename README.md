# 🦖 Rex

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

> Happily developed with Cursor and Claude-4.5 🎉

**A comprehensive, production-ready Go backend system for managing multi-tenant applications with authentication, RBAC (Role-Based Access Control), and background job processing.**

## 🚀 Features

### Core Features

- **Multi-Tenant Architecture**: Complete tenant isolation with self-service and managed onboarding
- **Authentication**: SuperTokens integration with cookie-based sessions and optional Google OAuth
- **RBAC System**: Flexible role-based access control with policies and permissions across multiple services
- **Member Management**: Invite users, manage tenant memberships with different roles (Admin, Writer, Viewer, Basic)
- **Background Jobs**: Reliable asynchronous job processing with Redis/Asynq
- **Tenant Initialization**: Automated tenant setup across multiple backend services

### Technical Highlights

- **Clean Architecture**: Layered design with clear separation of concerns
- **Database**: PostgreSQL with GORM ORM
- **Migrations**: Version-controlled database schema management
- **Docker**: Full containerization with Docker Compose
- **Dev Container**: VS Code dev container support for consistent development
- **API Design**: RESTful API with comprehensive endpoints

## 📋 Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Configuration](#configuration)
- [Development](#development)
- [Deployment](#deployment)

## 🏗 Architecture

### System Components

```
┌──────────────┐
│   Browser    │
└──────┬───────┘
       │ http://localhost
       ↓
┌──────────────┐
│    Nginx     │  ← Reverse Proxy
│   Port 80    │
└──────┬───────┘
       │
   ┌───┴────┬──────────────┐
   │        │              │
   ↓        ↓              ↓
┌────────┐ ┌─────────┐ ┌──────────────┐
│Frontend│ │Backend  │ │ SuperTokens  │
│ :3000  │ │API :8080│ │ Core :3567   │
│(React) │ │  (Gin)  │ │  (Internal)  │
└────────┘ └────┬────┘ └──────────────┘
                │
        ┌───────┼────────┐
        ↓       ↓        ↓
   ┌─────────┐ ┌─────┐ ┌────────┐
   │PostgreSQL│ │Redis│ │MailHog │
   │  :5432  │ │:6379│ │  :8025 │
   └─────────┘ └─────┘ └────────┘
```

### Key Design Patterns

- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic encapsulation
- **Dependency Injection**: Loose coupling between components
- **Middleware Pipeline**: Request processing chain (Auth → Tenant Access → RBAC)

### Security Architecture

- **Authentication**: SuperTokens with cookie-based sessions
- **Authorization**: RBAC with roles, policies, and permissions
- **Network Isolation**: Services communicate through internal Docker network
- **API Gateway**: Nginx reverse proxy as single entry point
- **Session Management**: HTTP-only cookies with anti-CSRF protection

## 📦 Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)
- Make (optional, for using Makefile commands)

**Supported Architectures**:
- ✅ AMD64 (x86_64) - Intel/AMD processors
- ✅ ARM64 (aarch64) - Apple Silicon (M1/M2/M3), AWS Graviton

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd utm-backend
```

### 2. Configure Environment

Create a `.env` file:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Application
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=utmuser
DB_PASSWORD=utmpassword
DB_NAME=utm_backend

# SuperTokens
SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
SUPERTOKENS_API_KEY=your-generated-api-key
API_DOMAIN=http://localhost:8080
WEBSITE_DOMAIN=http://localhost:3000

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Other services
TENANT_INIT_SERVICES=http://service1:8080,http://service2:8080
```

### 3. Start Services

```bash
# Build and start all services
make run

# Or with logs
make dev
```

### 4. Run Migrations

```bash
make migrate-up
```

### 5. Access Services

- **Main Application: http://localhost** ⭐ (via Nginx reverse proxy)
  - Frontend: `http://localhost/`
  - API: `http://localhost/api`
  - Auth: `http://localhost/auth`
- **MailHog (Email Testing)**: http://localhost:8025

**Note**: All services are now accessible through Nginx on port 80. See [Nginx Proxy Guide](docs/NGINX_PROXY_GUIDE.md) for routing details.

## 📁 Project Structure

```
utm-backend/
├── cmd/
│   ├── api/                    # API server entrypoint
│   └── worker/                 # Background worker entrypoint
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP request handlers
│   │   ├── middleware/        # HTTP middleware
│   │   └── router/            # Route definitions
│   ├── config/                # Configuration management
│   ├── database/              # Database connection
│   ├── jobs/                  # Background job client & worker
│   │   └── tasks/            # Job task implementations
│   ├── models/                # Data models & DTOs
│   ├── pkg/                   # Utility packages
│   │   └── response/         # HTTP response helpers
│   ├── repository/            # Data access layer
│   └── services/              # Business logic layer
├── migrations/                # Database migrations
├── scripts/                   # Utility scripts
├── .devcontainer/            # VS Code dev container config
├── docker-compose.yml        # Service orchestration
├── Dockerfile                # API/Worker image
├── Makefile                  # Development commands
└── README.md                 # This file
```

## 🔌 API Endpoints

### Tenant Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants` | Create tenant (self-onboarding) |
| POST | `/api/v1/tenants/managed` | Create managed tenant |
| GET | `/api/v1/tenants` | List user's tenants |
| GET | `/api/v1/tenants/:id` | Get tenant details |
| PATCH | `/api/v1/tenants/:id` | Update tenant |
| DELETE | `/api/v1/tenants/:id` | Delete tenant |
| GET | `/api/v1/tenants/:id/status` | Get tenant status |

### Member Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants/:id/members` | Add member to tenant with role |
| GET | `/api/v1/tenants/:id/members` | List tenant members |
| GET | `/api/v1/tenants/:id/members/:user_id` | Get member details |
| PATCH | `/api/v1/tenants/:id/members/:user_id` | Update member (change role) |
| DELETE | `/api/v1/tenants/:id/members/:user_id` | Remove member from tenant |

### Invitations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants/:id/invitations` | Invite user with role |
| GET | `/api/v1/tenants/:id/invitations` | List tenant invitations |
| GET | `/api/v1/invitations/:token` | Get invitation details (public) |
| POST | `/api/v1/invitations/:token/accept` | Accept invitation |
| POST | `/api/v1/invitations/check-pending` | Auto-accept pending invitations |
| DELETE | `/api/v1/invitations/:id` | Cancel invitation |

### RBAC (Role-Based Access Control)

| Method | Endpoint | Description |
|--------|----------|-------------|
| **Roles** (User's role in tenant) | | |
| GET | `/api/v1/roles` | List roles (Admin, Writer, Viewer, Basic) |
| GET | `/api/v1/roles/:id` | Get role details |
| POST | `/api/v1/platform/roles` | Create role (platform admin only) |
| **Policies** (Groups of permissions) | | |
| GET | `/api/v1/policies` | List policies |
| GET | `/api/v1/policies/:id` | Get policy details |
| POST | `/api/v1/platform/policies` | Create policy (platform admin only) |
| **Permissions** (Individual permissions) | | |
| GET | `/api/v1/permissions` | List permissions |
| GET | `/api/v1/permissions/:id` | Get permission details |
| POST | `/api/v1/platform/permissions` | Create permission (platform admin only) |
| **Authorization** | | |
| POST | `/api/v1/authorize` | Check user permission ⭐ |
| GET | `/api/v1/permissions/user` | Get user's permissions |

**📖 For detailed RBAC implementation guide (backend & frontend examples), see [RBAC Authorization Guide](docs/RBAC_AUTHORIZATION_GUIDE.md)**

**RBAC Hierarchy**: `User → Member → Role → Policies → Permissions`

## 🗄 Database Schema

### Core Tables

- **tenants**: Tenant information and status
- **roles**: User roles in tenant (Admin, Writer, Viewer, Basic)
- **tenant_members**: User-tenant associations with role
- **policies**: Groups of permissions (FullAccess, ReadOnly, etc.)
- **permissions**: Individual permissions (service:entity:action format)
- **role_policies**: Role-to-policy mappings
- **policy_permissions**: Policy-to-permission mappings
- **invitations**: Pending user invitations
- **platform_admins**: Platform-level administrators

### RBAC Hierarchy

```
User
  ↓
tenant_members (has role_id)
  ↓
roles (Admin, Writer, Viewer, Basic)
  ↓ (N:M via role_policies)
policies (FullAccess, ReadOnly, etc.)
  ↓ (N:M via policy_permissions)
permissions (tenant-api:member:create, etc.)
```

### Tenant Relationships

```
tenants
  ├── tenant_members (1:N)
  │     └── role (N:1) → Admin, Writer, Viewer, Basic
  └── invitations (1:N)

roles
  └── policies (N:M via role_policies)
        └── permissions (N:M via policy_permissions)
```

**Example**: An Admin role has a FullAccess policy, which contains permissions like `tenant-api:member:create`, `tenant-api:member:delete`, etc.

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `APP_PORT` | API server port | 8080 |
| `DB_HOST` | PostgreSQL host | postgres |
| `DB_PORT` | PostgreSQL port | 5432 |
| `SUPERTOKENS_CONNECTION_URI` | SuperTokens URL | http://supertokens:3567 |
| `REDIS_HOST` | Redis host | redis |
| `INVITATION_EXPIRY_HOURS` | Invitation validity | 72 |
| `TENANT_INIT_SERVICES` | Comma-separated service URLs | - |

## 🛠 Development

### Local Development Setup

1. **Install Dependencies**:
   ```bash
   go mod download
   ```

2. **Run Services**:
   ```bash
   make dev
   ```

3. **Run Migrations**:
   ```bash
   make migrate-up
   ```

### Using Dev Container

Open the project in VS Code and select "Reopen in Container" when prompted.

### Running Tests

```bash
make test
```

### Creating Migrations

```bash
make migrate-create name=add_new_feature
```

### Viewing Logs

```bash
# All services
make logs

# API only
make logs-api

# Worker only
make logs-worker
```

### Database Access

```bash
make shell-db
```

## 🔐 Authentication Flow

### Self-Onboarding

1. User signs up via SuperTokens (email/password or Google OAuth)
2. User creates tenant via `POST /api/v1/tenants`
3. User is automatically added as tenant member with **Admin** role
4. Background job initializes tenant in all services
5. Tenant status becomes "active"

### Managed Onboarding

1. Platform admin creates tenant with owner email via `POST /api/v1/tenants/managed`
2. System creates invitation for specified user with **Admin** role
3. Invitation email is sent
4. User signs up (if new) or logs in (if existing) and accepts invitation
5. User becomes tenant member with **Admin** role
6. Tenant initialization is triggered

### Invitation Flow

1. Tenant admin invites user via `POST /api/v1/tenants/:id/invitations`
2. System creates invitation record with specified role (Admin, Writer, Viewer, or Basic)
3. Email is sent with invitation link
4. New user signs up and accepts invitation
5. Existing user accepts invitation on login
6. User becomes tenant member with the specified role

## 📊 Background Jobs

### Tenant Initialization Job

- **Queue**: critical
- **Retry**: 5 times
- **Purpose**: Initialize tenant configuration across all backend services
- **Trigger**: After tenant creation or admin acceptance

### User Invitation Job

- **Queue**: default
- **Retry**: 3 times
- **Purpose**: Send invitation emails to users
- **Trigger**: When invitation is created

## 🚢 Deployment

### Production Build

```bash
docker build -t utm-backend:latest .
```

### Environment Setup

1. Set `APP_ENV=production`
2. Configure production database
3. Set secure API keys
4. Configure email service (SMTP/SendGrid)
5. Set up Redis for jobs

### Database Migration

```bash
migrate -path=./migrations \
  -database "postgres://user:pass@host:port/dbname" \
  up
```

## 📈 Monitoring & Logging

- Structured logging with Zap
- Request/response logging middleware
- Job execution tracking
- Database query logging (development)

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Run tests: `make test`
4. Run linter: `make lint`
5. Submit pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**MIT License Summary**:
- ✅ Commercial use allowed
- ✅ Modification allowed
- ✅ Distribution allowed
- ✅ Private use allowed
- ⚠️ License and copyright notice must be included
- ⚠️ No liability or warranty provided

## 📧 Support

For issues and questions, please open a GitHub issue or contact the development team.

---

**Built with ❤️ using Go, Gin, GORM, SuperTokens, and PostgreSQL**

