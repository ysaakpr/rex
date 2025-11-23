# 🦖 Rex

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

> Happily developed with Cursor and Claude-4.5 🎉

**A comprehensive, production-ready Go backend system for managing multi-tenant applications with authentication, RBAC (Role-Based Access Control), and background job processing.**

## 🚀 Features

### Core Features

- **Multi-Tenant Architecture**: Complete tenant isolation with self-service and managed onboarding
- **Authentication**: SuperTokens integration for secure user authentication
- **RBAC System**: Flexible role-based access control with permissions across multiple services
- **Member Management**: Invite users, manage tenant memberships with different relations
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
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐      ┌──────────────┐
│  API Gateway    │◄────►│ SuperTokens  │
│  (Gin Router)   │      │  Auth Core   │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│   Services      │◄────►│  PostgreSQL  │
│   Layer         │      │   Database   │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Job Queue      │◄────►│    Redis     │
│  (Asynq)        │      │              │
└─────────────────┘      └──────────────┘
```

### Key Design Patterns

- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic encapsulation
- **Dependency Injection**: Loose coupling between components
- **Middleware Pipeline**: Request processing chain

## 📦 Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for using Makefile commands)

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

- **API**: http://localhost:8080
- **MailHog (Email Testing)**: http://localhost:8025
- **SuperTokens Dashboard**: http://localhost:3567

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
| POST | `/api/v1/tenants/:tenant_id/members` | Add member to tenant |
| GET | `/api/v1/tenants/:tenant_id/members` | List tenant members |
| GET | `/api/v1/tenants/:tenant_id/members/:user_id` | Get member details |
| PATCH | `/api/v1/tenants/:tenant_id/members/:user_id` | Update member |
| DELETE | `/api/v1/tenants/:tenant_id/members/:user_id` | Remove member |
| POST | `/api/v1/tenants/:tenant_id/members/:user_id/roles` | Assign roles |
| DELETE | `/api/v1/tenants/:tenant_id/members/:user_id/roles/:role_id` | Remove role |

### Invitations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tenants/:tenant_id/invitations` | Invite user |
| GET | `/api/v1/tenants/:tenant_id/invitations` | List invitations |
| POST | `/api/v1/invitations/:token/accept` | Accept invitation |
| DELETE | `/api/v1/invitations/:id` | Cancel invitation |

### RBAC

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/relations` | List relations |
| POST | `/api/v1/relations` | Create relation |
| GET | `/api/v1/roles` | List roles |
| POST | `/api/v1/roles` | Create role |
| POST | `/api/v1/roles/:id/permissions` | Assign permissions to role |
| GET | `/api/v1/permissions` | List permissions |
| POST | `/api/v1/permissions` | Create permission |
| POST | `/api/v1/authorize` | Check user permission |

## 🗄 Database Schema

### Core Tables

- **tenants**: Tenant information and status
- **relations**: Membership types (Admin, Writer, Viewer, etc.)
- **tenant_members**: User-tenant associations
- **roles**: Permission bundles
- **permissions**: Granular access controls
- **role_permissions**: Role-permission mappings
- **member_roles**: Member-role assignments
- **user_invitations**: Pending user invitations

### Relationships

```
tenants
  ├── tenant_members (1:N)
  │     ├── relation (N:1)
  │     └── roles (N:M)
  └── user_invitations (1:N)
  
roles
  └── permissions (N:M)
```

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

1. User signs up via SuperTokens
2. User creates tenant via `POST /api/v1/tenants`
3. User is automatically added as tenant Admin
4. Background job initializes tenant in all services
5. Tenant status becomes "active"

### Managed Onboarding

1. Super admin creates tenant with admin email
2. System creates invitation for specified user
3. Invitation email is sent
4. User accepts invitation on first login
5. User becomes tenant admin
6. Tenant initialization is triggered

### Invitation Flow

1. Tenant admin invites user via email
2. System creates invitation record
3. Email is sent with invitation link
4. New user signs up and accepts invitation
5. Existing user accepts invitation on login
6. User becomes tenant member with specified relation

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

