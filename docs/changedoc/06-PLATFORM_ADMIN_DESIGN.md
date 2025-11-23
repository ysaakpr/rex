# Platform Admin Architecture Design

**Date:** November 21, 2025  
**Purpose:** Design platform-level entity management separate from tenant-level operations

## Problem Statement

Currently, platform-level entities (permissions, roles, relations) are mixed with tenant-level operations. We need:

1. **Platform-level entities** that don't belong to any tenant:
   - Permissions
   - Roles
   - Relations
   - Role-to-Permission mappings
   - Relation-to-Role mappings

2. **Platform Admins** - special users who:
   - Are NOT part of any tenant
   - Can manage platform-level entities
   - Cannot be regular tenant members

## Proposed Solution

### 1. Entity Separation

**Platform-Level Entities** (managed by platform admins):
- `permissions` - service:entity:action permissions
- `roles` - groups of permissions
- `relations` - membership types (Admin, Writer, Viewer, etc.)
- `role_permissions` - many-to-many mapping
- `relation_roles` - NEW: which roles come with each relation

**Tenant-Level Entities** (managed by tenant admins):
- `tenants`
- `tenant_members` - uses relations to define membership type
- Member-specific role assignments (additions to relation defaults)

### 2. Platform Admin Model

```go
// Platform admins are tracked separately
type PlatformAdmin struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID    string    `gorm:"type:varchar(255);unique;not null"` // SuperTokens user ID
    CreatedBy string    `gorm:"type:varchar(255)"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 3. Relation-to-Role Mapping

```sql
CREATE TABLE relation_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relation_id UUID NOT NULL REFERENCES relations(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(relation_id, role_id)
);
```

**Purpose:** When a user is assigned a relation in a tenant, they automatically get all roles associated with that relation.

**Example:**
- "Admin" relation → includes "Tenant Manager" + "User Manager" roles
- "Writer" relation → includes "Content Creator" role
- "Viewer" relation → includes "Read Only" role

### 4. API Structure

```
Platform-Level APIs (platform-api scope):
  POST   /api/v1/platform/permissions
  GET    /api/v1/platform/permissions
  PATCH  /api/v1/platform/permissions/:id
  DELETE /api/v1/platform/permissions/:id
  
  POST   /api/v1/platform/roles
  GET    /api/v1/platform/roles
  PATCH  /api/v1/platform/roles/:id
  DELETE /api/v1/platform/roles/:id
  POST   /api/v1/platform/roles/:id/permissions
  
  POST   /api/v1/platform/relations
  GET    /api/v1/platform/relations
  PATCH  /api/v1/platform/relations/:id
  DELETE /api/v1/platform/relations/:id
  POST   /api/v1/platform/relations/:id/roles  ← NEW
  
  POST   /api/v1/platform/admins
  GET    /api/v1/platform/admins
  DELETE /api/v1/platform/admins/:user_id

Tenant-Level APIs:
  POST   /api/v1/tenants
  GET    /api/v1/tenants
  GET    /api/v1/tenants/:id
  GET    /api/v1/tenants/:id/members
  POST   /api/v1/tenants/:id/members
  POST   /api/v1/tenants/:id/members/:user_id/roles  ← Additional roles beyond relation
```

### 5. Authorization Flow

```
Platform Admin Check Middleware:
┌─────────────────────────────────────┐
│ 1. Verify SuperTokens session       │
│ 2. Get user ID from session         │
│ 3. Check if user is platform admin  │
│    - Query platform_admins table    │
│ 4. If yes → proceed                 │
│    If no → 403 Forbidden            │
└─────────────────────────────────────┘

Tenant Access Check Middleware:
┌─────────────────────────────────────┐
│ 1. Verify SuperTokens session       │
│ 2. Get user ID from session         │
│ 3. Check if user is member of tenant│
│ 4. Check user's relation            │
│ 5. Load roles from relation         │
│ 6. Check required permission        │
│ 7. If authorized → proceed          │
│    If not → 403 Forbidden           │
└─────────────────────────────────────┘
```

### 6. Permission Scopes

```
Platform-API Permissions:
- platform-api:permission:create
- platform-api:permission:read
- platform-api:permission:update
- platform-api:permission:delete
- platform-api:role:create
- platform-api:role:read
- platform-api:role:update
- platform-api:role:delete
- platform-api:relation:create
- platform-api:relation:read
- platform-api:relation:update
- platform-api:relation:delete
- platform-api:admin:create
- platform-api:admin:read
- platform-api:admin:delete

Tenant-API Permissions:
- tenant-api:member:create
- tenant-api:member:read
- tenant-api:member:update
- tenant-api:member:delete
- tenant-api:tenant:create
- tenant-api:tenant:read
- tenant-api:tenant:update
- tenant-api:tenant:delete
```

### 7. SuperTokens Metadata

Store platform admin flag in SuperTokens user metadata:

```json
{
  "user_id": "uuid",
  "metadata": {
    "is_platform_admin": true
  }
}
```

## Implementation Plan

### Phase 1: Database Schema
1. Create `platform_admins` table
2. Create `relation_roles` join table
3. Migration to add platform-admin permissions

### Phase 2: Backend Changes
1. Create `PlatformAdmin` model
2. Create `PlatformAdminMiddleware`
3. Create platform admin handlers
4. Update RBAC service to handle relation-to-role mappings
5. Move existing /roles, /permissions, /relations to /platform/*

### Phase 3: Authorization
1. Implement platform admin check
2. Update tenant middleware to load roles from relations
3. Add permission checks for tenant operations

### Phase 4: Frontend
1. Create platform admin dashboard
2. Update roles/permissions/relations pages to use /platform/* endpoints
3. Add relation-to-role mapping UI
4. Add platform admin management UI

## Benefits

1. **Clear Separation:** Platform vs Tenant entities
2. **Better Security:** Only platform admins can modify core entities
3. **Flexible RBAC:** Relations define default roles, with ability to add more
4. **Scalability:** Easy to add new platform-level features
5. **Multi-tenancy:** Tenants can't interfere with each other's data

## Migration Strategy

1. **Backward Compatible:** Existing APIs continue to work
2. **Gradual Migration:** Move to /platform/* endpoints over time
3. **Default Platform Admin:** First user becomes platform admin
4. **Seed Data:** Pre-populate common relations and roles

## Example Workflow

### Creating a New Tenant Admin

```
1. Platform Admin creates "Admin" relation with roles:
   - Relation: "Admin"
   - Roles: ["Tenant Manager", "User Manager", "Content Admin"]

2. User signs up and creates tenant
   - User becomes member of tenant with "Admin" relation
   - Automatically gets all 3 roles from the relation

3. Tenant Admin adds more users:
   - User A: "Writer" relation → gets "Content Creator" role
   - User B: "Viewer" relation → gets "Read Only" role
   - User C: "Writer" relation + "Analytics" role (additional)
```

## Security Considerations

1. **Platform Admin Bootstrap:** First admin must be created via migration or CLI
2. **Immutability:** Relations and core roles should be system-protected
3. **Audit Logging:** All platform admin actions should be logged
4. **No Cross-over:** Platform admins shouldn't be tenant members (or clearly separated)

## Next Steps

1. Review and approve this design
2. Create database migrations
3. Implement backend changes
4. Update frontend
5. Write tests
6. Document API changes

