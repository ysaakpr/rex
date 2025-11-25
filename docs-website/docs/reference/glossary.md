# Glossary

Key terms and concepts used throughout the documentation.

## A

**API Key**  
A secret token used for authentication in API requests. In this system, System Users use API keys (tokens) for machine-to-machine authentication.

**Asynq**  
Redis-backed distributed task queue used for background job processing. Powers the worker system for asynchronous operations like tenant initialization and email sending.

**Authorization**  
The process of determining what actions an authenticated user is allowed to perform. Implemented through the RBAC system.

## C

**Cookie-based Authentication**  
Primary authentication method using HTTP-only cookies managed by SuperTokens. Used for web browser sessions.

**Credential Rotation**  
The practice of periodically replacing authentication credentials (tokens) to maintain security. System Users support automatic expiry and rotation.

**CORS (Cross-Origin Resource Sharing)**  
HTTP mechanism that allows web applications from one origin to access resources from another origin. Configured to allow frontend-backend communication.

## E

**Entity**  
In RBAC, the resource type being accessed (e.g., `post`, `comment`, `tenant`). Part of the permission structure: `service:entity:action`.

## G

**Gin**  
Go HTTP web framework used for building the REST API. Provides routing, middleware, and request handling.

**GORM**  
Go ORM (Object-Relational Mapping) library used for database operations. Provides model definitions and query building.

**Grace Period**  
Time window before System User credential expiry where warnings are sent. Allows time for rotation before actual expiry.

## H

**Handler**  
Function that processes HTTP requests in the API layer. Maps to specific endpoints and contains business logic orchestration.

## I

**Idempotent**  
Property where performing an operation multiple times has the same effect as performing it once. Background jobs are designed to be idempotent for safe retries.

**Invitation**  
A request sent to a user to join a tenant. Creates a `UserInvitation` record and sends an email with an acceptance link.

## J

**JWT (JSON Web Token)**  
Token format used by SuperTokens for session management. Contains claims about the authenticated user.

## M

**M2M (Machine-to-Machine)**  
Authentication and communication between services without human interaction. Implemented using System Users.

**Managed Tenant**  
Tenant created by platform admin on behalf of a user/organization. Contrasts with self-service tenant creation.

**Member**  
A user who belongs to a tenant. Represented by the `TenantMember` model linking users to tenants with specific roles.

**Middleware**  
Function that processes requests before they reach handlers or after handlers return responses. Used for authentication, authorization, logging, etc.

## P

**Permission**  
Atomic authorization unit granting ability to perform a specific action. Format: `service:entity:action` (e.g., `blog-api:post:create`).

**Platform Admin**  
Special user role with access to platform-wide administrative functions. Can manage tenants, users, and RBAC configuration.

**Policy**  
Collection of permissions. Intermediate layer in RBAC between roles and permissions. A policy groups related permissions.

**Preload**  
GORM feature to eagerly load related records. Example: `db.Preload("Role").Find(&members)` loads roles with members.

## Q

**Queue**  
In Asynq, a named channel for background tasks. Three priority levels: `critical`, `default`, and `low`.

## R

**RBAC (Role-Based Access Control)**  
Authorization system using roles, policies, and permissions. Three-tier model: Users → Roles → Policies → Permissions.

**Redis**  
In-memory data store used for Asynq task queue. Stores pending, active, and scheduled tasks.

**Repository**  
Data access layer that encapsulates database operations. Separates business logic from database queries.

**Role**  
Named collection of policies assigned to users within a tenant. Examples: Admin, Editor, Writer, Viewer.

**Rotation**  
See Credential Rotation.

## S

**Self-Service Tenant**  
Tenant created by a user through the registration flow. User becomes the first admin member automatically.

**Service**  
In RBAC, represents a backend microservice or functional area. Examples: `tenant-api`, `blog-api`, `media-api`.

**Session**  
Authenticated user session managed by SuperTokens. Includes access token, refresh token, and session data.

**Slug**  
URL-friendly identifier for a tenant. Must be unique, lowercase, and contain only letters, numbers, and hyphens.

**SuperTokens**  
Open-source authentication solution providing session management, user authentication, and OAuth integration.

**System User**  
Service account for machine-to-machine authentication. Has a token (bearer) and can access APIs programmatically.

## T

**Tenant**  
Isolated workspace/organization in the multi-tenant system. Has its own members, data, and configuration.

**Tenant ID**  
UUID uniquely identifying a tenant. Used in all tenant-scoped API calls.

**Token**  
Authentication credential. Can be:
- **Access Token**: Short-lived JWT for API requests (SuperTokens)
- **System User Token**: Long-lived bearer token for M2M auth
- **Invitation Token**: One-time token for accepting invitations

## U

**UUID (Universally Unique Identifier)**  
128-bit identifier used as primary keys for most database records. Format: `123e4567-e89b-12d3-a456-426614174000`.

## V

**Viper**  
Go configuration library used for loading environment variables and config files.

## W

**Worker**  
Background process that executes asynchronous jobs from the Asynq queue. Runs tenant initialization, email sending, and cleanup tasks.

## Common Abbreviations

| Abbreviation | Full Term |
|--------------|-----------|
| API | Application Programming Interface |
| CORS | Cross-Origin Resource Sharing |
| CRUD | Create, Read, Update, Delete |
| DB | Database |
| FK | Foreign Key |
| HTTP | HyperText Transfer Protocol |
| ID | Identifier |
| JSON | JavaScript Object Notation |
| JWT | JSON Web Token |
| M2M | Machine-to-Machine |
| ORM | Object-Relational Mapping |
| PK | Primary Key |
| RBAC | Role-Based Access Control |
| REST | Representational State Transfer |
| SQL | Structured Query Language |
| SMTP | Simple Mail Transfer Protocol |
| TTL | Time To Live |
| UI | User Interface |
| URL | Uniform Resource Locator |
| UUID | Universally Unique Identifier |

## Permission Format

Permissions follow the format:  
```
service:entity:action
```

**Examples**:
- `tenant-api:tenant:create` - Create a tenant
- `blog-api:post:update` - Update a blog post
- `tenant-api:member:invite` - Invite a member to tenant

## Status Values

### Tenant Status
- `pending` - Awaiting initialization
- `active` - Fully operational
- `suspended` - Temporarily disabled
- `deleted` - Soft deleted

### Invitation Status
- `pending` - Awaiting acceptance
- `accepted` - User accepted and joined
- `expired` - Past expiration date
- `cancelled` - Cancelled by admin

### System User Status
- `active` - Currently valid
- `inactive` - Deactivated (expired or manually)

## Related Documentation

- [Core Concepts](/introduction/core-concepts) - System concepts
- [Architecture](/introduction/architecture) - System architecture
- [RBAC Overview](/guides/rbac-overview) - Authorization system
- [API Overview](/x-api/overview) - API structure
