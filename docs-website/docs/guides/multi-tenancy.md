# Multi-Tenancy

Understanding the multi-tenant architecture and how to work with tenants.

## What is Multi-Tenancy?

Multi-tenancy is an architecture where a single application serves multiple isolated customer organizations (tenants).

### Key Principles

**Data Isolation**:
- Each tenant's data is completely separate
- No cross-tenant data access
- Enforced at database and application level

**Resource Sharing**:
- Same codebase and infrastructure
- Shared database with tenant_id filtering
- Cost-effective scaling

**Flexibility**:
- Tenants can customize their experience
- Different permissions per tenant
- Independent tenant lifecycles

## Tenant Model

```go
type Tenant struct {
    ID        uuid.UUID    `json:"id"`
    Name      string       `json:"name"`
    Slug      string       `json:"slug"`
    Status    TenantStatus `json:"status"`
    Metadata  JSONMap      `json:"metadata"`
    CreatedBy string       `json:"created_by"`
    CreatedAt time.Time    `json:"created_at"`
    UpdatedAt time.Time    `json:"updated_at"`
}
```

### Tenant Properties

**ID**: Unique identifier (UUID)
- Used in all database queries
- Never changes
- Example: `123e4567-e89b-12d3-a456-426614174000`

**Name**: Display name
- Shown in UI
- Can be changed
- Example: "Acme Corporation"

**Slug**: URL-friendly identifier
- Used in URLs
- Unique across system
- Example: "acme-corp"

**Status**: Lifecycle state
- `pending` - Initial creation, setup in progress
- `active` - Fully operational
- `suspended` - Temporarily disabled
- `deleted` - Soft deleted

**Metadata**: Custom JSON data
- Flexible storage for tenant-specific config
- Example: `{"industry": "Technology", "size": "50-100"}`

## Tenant Creation Methods

### 1. Self-Service Onboarding

**Who**: Any authenticated user

**When**: User wants to create their own workspace

**Flow**:
```
1. User signs up/logs in
2. User creates tenant
3. User automatically becomes Admin
4. Tenant initialized in background
5. User can start using tenant
```

**API Call**:
```javascript
POST /api/v1/tenants
{
  "name": "My Startup",
  "slug": "my-startup",
  "metadata": {
    "industry": "SaaS",
    "plan": "free"
  }
}
```

**Benefits**:
- User in control
- Immediate access
- No manual approval needed

### 2. Managed Onboarding

**Who**: Platform admins only

**When**: Creating tenant for a customer

**Flow**:
```
1. Platform admin creates tenant
2. Admin specifies customer email
3. Invitation sent to customer
4. Customer signs up/logs in
5. Customer accepts invitation
6. Customer becomes Admin of tenant
```

**API Call**:
```javascript
POST /api/v1/tenants/managed
{
  "name": "Enterprise Customer Inc",
  "slug": "enterprise-customer",
  "admin_email": "admin@enterprise-customer.com",
  "metadata": {
    "plan": "enterprise",
    "contract_value": 50000
  }
}
```

**Benefits**:
- Controlled onboarding
- Pre-configuration possible
- Better for enterprise customers

## Tenant Isolation

### Database Level

All queries include `tenant_id` filter:

```sql
-- Example query
SELECT * FROM content 
WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000'
AND status = 'published';

-- Enforced by application
```

### Application Level

**TenantAccessMiddleware** enforces access:

```go
func TenantAccessMiddleware(memberRepo, db) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := GetUserID(c)
        tenantID := c.Param("id")
        
        // Check platform admin (bypass membership check)
        if isPlatformAdmin(userID, db) {
            c.Next()
            return
        }
        
        // Check tenant membership
        member := memberRepo.GetByTenantAndUser(tenantID, userID)
        if member == nil || member.Status != "active" {
            c.JSON(403, "Access denied")
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### API Level

Routes require tenant access:

```go
// Tenant-scoped routes
tenantScoped := tenants.Group("/:id")
tenantScoped.Use(middleware.TenantAccessMiddleware(...))
{
    tenantScoped.GET("", handler.GetTenant)
    tenantScoped.GET("/members", handler.ListMembers)
}
```

## Tenant Lifecycle

### Creation

```
POST /api/v1/tenants
    ↓
1. Validate input (name, slug)
2. Check slug uniqueness
3. Create tenant record (status: pending)
4. Add creator as Admin member
5. Enqueue initialization job
6. Return tenant info
```

### Initialization

Background job runs:

```
1. Create default resources
2. Set up integrations
3. Send welcome email
4. Update status to 'active'
```

Configuration:
```bash
# .env
TENANT_INIT_SERVICES=email,storage,analytics
```

### Updates

```
PATCH /api/v1/tenants/:id
    ↓
- Update name
- Update metadata
- Change status
```

### Suspension

```
PATCH /api/v1/tenants/:id
{
  "status": "suspended"
}
    ↓
- Tenant becomes read-only
- Users can view but not modify
- Can be reactivated
```

### Deletion

```
DELETE /api/v1/tenants/:id
    ↓
- Soft delete (status: deleted)
- Data retained
- Can be restored by platform admin
```

## Tenant Membership

### Member Model

```go
type TenantMember struct {
    ID        uuid.UUID    `json:"id"`
    TenantID  uuid.UUID    `json:"tenant_id"`
    UserID    string       `json:"user_id"`
    RoleID    uuid.UUID    `json:"role_id"`
    Status    MemberStatus `json:"status"`
    InvitedBy *string      `json:"invited_by"`
    JoinedAt  time.Time    `json:"joined_at"`
}
```

### Member Statuses

- `active` - Full access
- `inactive` - Suspended, no access
- `pending` - Invitation not yet accepted

### Adding Members

**Direct Add** (if user exists):
```javascript
POST /api/v1/tenants/:id/members
{
  "user_id": "existing-user-id",
  "role_id": "writer-role-id"
}
```

**Via Invitation** (if user doesn't exist):
```javascript
POST /api/v1/tenants/:id/invitations
{
  "email": "newuser@example.com",
  "role_id": "writer-role-id"
}
```

### Multi-Tenant Users

One user can belong to multiple tenants:

```
User: john@example.com
  ├─ Acme Corp (tenant-1) → Admin
  ├─ Startup Inc (tenant-2) → Writer
  └─ Beta Co (tenant-3) → Viewer
```

## Best Practices

### 1. Slug Guidelines

**Good Slugs**:
- `acme-corp`
- `my-startup`
- `tech-company-123`

**Bad Slugs**:
- `Acme Corp` (not URL-safe)
- `acme_corp` (use hyphens)
- `a` (too short)

**Validation**:
```go
// Alphanumeric + hyphens, 3-255 chars
slug: "required,min=3,max=255,alphanum"
```

### 2. Metadata Usage

Store tenant-specific configuration:

```json
{
  "industry": "Technology",
  "company_size": "50-100",
  "plan": "enterprise",
  "features": {
    "advanced_analytics": true,
    "custom_branding": true,
    "api_access": true
  },
  "billing": {
    "subscription_id": "sub_123",
    "seats": 50
  }
}
```

### 3. Naming Conventions

- Use business names for `name`
- Use technical identifiers for `slug`
- Keep names professional
- Avoid special characters

### 4. Tenant Limits

Consider implementing limits:

```go
const (
    MaxMembersPerTenant = 1000
    MaxTenantsPerUser   = 50
    MaxSlugLength       = 255
)
```

### 5. Audit Logging

Log tenant operations:

```go
log.Info("Tenant created",
    zap.String("tenant_id", tenant.ID.String()),
    zap.String("name", tenant.Name),
    zap.String("created_by", userID),
)
```

## Common Patterns

### Get User's Tenants

```javascript
GET /api/v1/tenants
// Returns all tenants where user is a member
```

### Switch Tenant (Frontend)

```javascript
function switchTenant(tenantId) {
  localStorage.setItem('currentTenant', tenantId);
  navigate(`/tenants/${tenantId}/dashboard`);
}
```

### Check Tenant Access

```go
func CheckTenantAccess(tenantID, userID string) bool {
    // Platform admin?
    if IsPlatformAdmin(userID) {
        return true
    }
    
    // Tenant member?
    member := GetMembership(tenantID, userID)
    return member != nil && member.Status == "active"
}
```

### Tenant Context in Requests

```javascript
// Include tenant ID in API calls
fetch(`/api/v1/tenants/${tenantId}/content`, {
  credentials: 'include'
})
```

## Platform Admin Access

Platform admins can access any tenant:

```go
// TenantAccessMiddleware
if IsPlatformAdmin(userID) {
    c.Set("isPlatformAdmin", true)
    c.Next() // Skip membership check
    return
}
```

**Use Cases**:
- Customer support
- System administration
- Compliance audits
- Troubleshooting

## Scaling Considerations

### Database

**Indexing**:
```sql
CREATE INDEX idx_tenant_members_tenant_id ON tenant_members(tenant_id);
CREATE INDEX idx_tenant_members_user_id ON tenant_members(user_id);
CREATE INDEX idx_tenants_slug ON tenants(slug);
```

**Partitioning** (for very large scale):
- Partition tables by tenant_id
- Improves query performance
- Complex to implement

### Caching

Cache tenant metadata:

```go
// Redis cache
key := fmt.Sprintf("tenant:%s", tenantID)
cachedTenant := redis.Get(key)
```

### Rate Limiting

Per-tenant rate limits:

```go
rateLimiter.AllowN(tenantID, requestCost)
```

## Security Considerations

### Tenant Isolation

**Critical**: Never expose tenant_id as a trust signal

```go
// ❌ WRONG
tenantID := c.Query("tenant_id") // User controlled!

// ✅ CORRECT
tenantID := c.Param("id")
CheckTenantMembership(tenantID, userID) // Verify access
```

### Data Leakage Prevention

```sql
-- Always include tenant_id in WHERE clause
SELECT * FROM content 
WHERE tenant_id = $1 
AND id = $2;

-- Never trust client-provided tenant_id
```

### Access Control

```
Request → Auth → TenantAccess → RBAC → Handler
```

## Next Steps

- **[Creating Tenants](/guides/creating-tenants)** - Step-by-step tenant creation
- **[Member Management](/guides/member-management)** - Managing tenant members
- **[Invitations](/guides/invitations)** - Invitation system details
- **[RBAC](/guides/rbac-overview)** - Authorization within tenants

