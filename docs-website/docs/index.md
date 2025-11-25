---
layout: home

hero:
  name: "Rex"
  text: "Open-Source User Management Platform"
  tagline: Complete multi-tenant user management with authentication, RBAC, and tenant isolation built-in
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started/quick-start

features:
  - icon: 🔐
    title: Authentication
    details: SuperTokens integration with cookie-based sessions, JWT tokens, Google OAuth, and M2M authentication for services

  - icon: 🏢
    title: Multi-Tenancy
    details: Complete tenant isolation with self-service and managed onboarding, member invitations, and role management

  - icon: 🛡️
    title: Flexible RBAC
    details: 3-tier authorization system (Roles → Policies → Permissions) with fine-grained access control

  - icon: ⚡
    title: Background Jobs
    details: Redis-backed async job processing with Asynq for tenant initialization, emails, and scheduled tasks

  - icon: 👨‍💼
    title: Platform Admin
    details: Super admin capabilities with system-wide access, tenant management, and system user creation

  - icon: ⚛️
    title: React Frontend
    details: Modern React 18 frontend with TailwindCSS, SuperTokens SDK, and beautiful pre-built components

  - icon: 🐳
    title: Docker Ready
    details: Complete Docker Compose setup with PostgreSQL, Redis, SuperTokens, and all services configured

  - icon: 📡
    title: RESTful API
    details: Comprehensive REST API with consistent response formats, pagination, and full CRUD operations

  - icon: 🔧
    title: Production Ready
    details: Database migrations, environment configuration, logging, monitoring, and deployment guides
---

## Quick Start

Get up and running in 5 minutes:

<div style="margin: 20px 0;">
  <DemoLink />
  <span style="margin: 0 10px;">or</span>
  <a href="https://github.com/ysaakpr/rex" class="VPButton medium alt" target="_blank" rel="noopener">
    <span class="text">View on GitHub</span>
  </a>
</div>

```bash
# Clone the repository
git clone https://github.com/ysaakpr/rex
cd rex

# Configure environment
cp .env.example .env

# Start all services
docker-compose up -d

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# MailHog: http://localhost:8025
```

## What You Get

### 🔐 Complete Authentication System
- **Email/Password + Google OAuth**: Multiple sign-in methods
- **Session Management**: Secure cookie-based and header-based auth
- **System Users**: M2M authentication for background services
- **Automatic Token Refresh**: Built-in token lifecycle management

### 🏢 Multi-Tenant Architecture
- **Tenant Isolation**: Complete data separation per organization
- **Self-Service Onboarding**: Users create and manage their own tenants
- **Managed Tenants**: Platform admins can create tenants for customers
- **Member Invitations**: Email-based team member invites with role assignment

### 🛡️ Advanced Authorization
- **Roles**: User positions in tenants (Admin, Writer, Viewer, Basic)
- **Policies**: Groups of permissions for easy management
- **Permissions**: Fine-grained access control (service:entity:action)
- **Dynamic RBAC**: Create custom roles and permissions via API

### ⚙️ Developer Friendly
- **Multiple Language Support**: Go, Node.js, Python, Java, C# middleware examples
- **System Auth Library**: Ready-to-use authentication clients for services
- **Comprehensive Docs**: Every endpoint documented with examples
- **Code Examples**: Complete user journeys and integration patterns

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   Browser                        │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│          React Frontend (:3000)                  │
│  - SuperTokens Auth React SDK                   │
│  - Session management                            │
│  - Protected routes                              │
└────────────────┬────────────────────────────────┘
                 │ HTTP/REST + Cookies
                 ↓
┌─────────────────────────────────────────────────┐
│         Go Backend API (:8080)                   │
│  - Gin HTTP framework                            │
│  - SuperTokens middleware                        │
│  - Auth, Tenant, RBAC middleware                 │
│  - RESTful API endpoints                         │
└──┬──────┬──────┬──────────────────────┬─────────┘
   │      │      │                      │
   ↓      ↓      ↓                      ↓
┌────┐ ┌────┐ ┌─────────┐      ┌──────────────┐
│ PG │ │Redis│ │SuperTokens│    │   Worker     │
│:5432│ │:6379│ │Core :3567│    │ (Background) │
└────┘ └────┘ └─────────┘      └──────────────┘
```

## Key Features Breakdown

### Authentication Modes

| Mode | Use Case | Example |
|------|----------|---------|
| **Cookie-based** | Web applications | React frontend with user login |
| **Header-based** | API clients, mobile apps | Mobile app with token storage |
| **System Users** | M2M, background jobs | Worker service processing data |

### RBAC Hierarchy

```
User → Tenant Member → Role → Policy → Permission
                         ↓       ↓         ↓
                      Admin  → Full    → tenant-api:*:*
                      Writer → Content → tenant-api:content:*
                      Viewer → Read    → tenant-api:*:read
```

### API Endpoints Summary

- **Authentication**: Sign up, sign in, sign out, session management
- **Tenants**: CRUD operations, status checking, member listing
- **Members**: Add, update, remove, role assignment
- **Invitations**: Create, accept, list, cancel
- **RBAC**: Manage roles, policies, permissions
- **System Users**: Create, rotate credentials, manage M2M accounts
- **Platform Admin**: Super admin management and operations

## Technology Stack

**Backend**:
- Go 1.23+
- Gin (HTTP framework)
- GORM (ORM)
- SuperTokens (Authentication)
- Asynq (Background jobs)
- PostgreSQL (Database)
- Redis (Cache & Queue)

**Frontend**:
- React 18
- Vite (Build tool)
- TailwindCSS (Styling)
- SuperTokens React SDK
- React Router v6
- Lucide React (Icons)

**DevOps**:
- Docker & Docker Compose
- Nginx (Reverse proxy)
- golang-migrate (Migrations)
- MailHog (Email testing)

## Use Cases

### SaaS Applications
Build multi-tenant SaaS products with complete user isolation, team collaboration, and role-based access control.

### B2B Platforms
Create platforms where businesses can manage their organizations, invite team members, and control access.

### API Platforms
Develop API services with M2M authentication for integrations and programmatic access.

### Enterprise Applications
Build internal tools with platform admin capabilities and fine-grained permission management.

## Next Steps

1. **[Quick Start Guide](/getting-started/quick-start)** - Get the system running locally
2. **[Authentication Guide](/guides/authentication)** - Understand the auth system
3. **[API Reference](/x-api/overview)** - Explore all available endpoints
4. **[Tenant Management](/guides/multi-tenancy)** - Learn about multi-tenancy
5. **[RBAC System](/guides/rbac-overview)** - Master authorization

## Community & Support

- 📖 [Documentation](/) - Comprehensive guides and API reference
- 💬 [GitHub Discussions](https://github.com/ysaakpr/rex/discussions) - Ask questions
- 🐛 [Issue Tracker](https://github.com/ysaakpr/rex/issues) - Report bugs
- 📧 Email Support - support@example.com

## License

Rex is released under the MIT License. See [LICENSE](https://github.com/ysaakpr/rex/blob/main/LICENSE) for details.

