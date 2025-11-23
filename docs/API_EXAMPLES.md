# API Examples

Complete examples for all API endpoints.

## Authentication

All protected endpoints require a SuperTokens session. Include the session cookies or Authorization header in your requests.

## Tenant Management

### Create Tenant (Self-Onboarding)

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "metadata": {
      "industry": "technology",
      "size": "50-100"
    }
  }'
```

Response:
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "status": "pending",
    "metadata": {
      "industry": "technology",
      "size": "50-100"
    },
    "created_by": "user_123",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Create Managed Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants/managed \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Beta Company",
    "slug": "beta-company",
    "admin_email": "admin@beta.com",
    "metadata": {
      "tier": "enterprise"
    }
  }'
```

### List Tenants

```bash
curl http://localhost:8080/api/v1/tenants?page=1&page_size=20 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Acme Corporation",
        "slug": "acme-corp",
        "status": "active",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 1,
    "total_pages": 1
  }
}
```

### Get Tenant

```bash
curl http://localhost:8080/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Update Tenant

```bash
curl -X PATCH http://localhost:8080/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Acme Corp Inc",
    "metadata": {
      "industry": "fintech"
    }
  }'
```

### Check Tenant Status

```bash
curl http://localhost:8080/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000/status \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "active"
  }
}
```

## Member Management

### Add Member to Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants/TENANT_ID/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "user_id": "user_456",
    "relation_id": "relation_uuid_for_writer"
  }'
```

Response:
```json
{
  "success": true,
  "message": "Member added successfully",
  "data": {
    "id": "member_uuid",
    "tenant_id": "TENANT_ID",
    "user_id": "user_456",
    "relation_id": "relation_uuid",
    "relation": {
      "id": "relation_uuid",
      "name": "Writer",
      "description": "Can create and edit content"
    },
    "status": "active",
    "joined_at": "2024-01-01T00:00:00Z"
  }
}
```

### List Tenant Members

```bash
curl http://localhost:8080/api/v1/tenants/TENANT_ID/members?page=1&page_size=20 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Get Member Details

```bash
curl http://localhost:8080/api/v1/tenants/TENANT_ID/members/user_456 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Update Member

```bash
curl -X PATCH http://localhost:8080/api/v1/tenants/TENANT_ID/members/user_456 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "relation_id": "new_relation_uuid",
    "status": "active"
  }'
```

### Remove Member

```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/TENANT_ID/members/user_456 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Assign Roles to Member

```bash
curl -X POST http://localhost:8080/api/v1/tenants/TENANT_ID/members/user_456/roles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "role_ids": [
      "role_uuid_1",
      "role_uuid_2"
    ]
  }'
```

### Remove Role from Member

```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/TENANT_ID/members/user_456/roles/role_uuid_1 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## Invitations

### Create Invitation

```bash
curl -X POST http://localhost:8080/api/v1/tenants/TENANT_ID/invitations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "email": "newuser@example.com",
    "relation_id": "relation_uuid_for_basic"
  }'
```

Response:
```json
{
  "success": true,
  "message": "Invitation sent successfully",
  "data": {
    "id": "invitation_uuid",
    "tenant_id": "TENANT_ID",
    "email": "newuser@example.com",
    "invited_by": "user_123",
    "relation_id": "relation_uuid",
    "relation": {
      "id": "relation_uuid",
      "name": "Basic"
    },
    "status": "pending",
    "expires_at": "2024-01-04T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### List Tenant Invitations

```bash
curl http://localhost:8080/api/v1/tenants/TENANT_ID/invitations \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Accept Invitation

```bash
curl -X POST http://localhost:8080/api/v1/invitations/INVITATION_TOKEN/accept \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Cancel Invitation

```bash
curl -X DELETE http://localhost:8080/api/v1/invitations/invitation_uuid \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## RBAC - Relations

### List Relations

```bash
curl http://localhost:8080/api/v1/relations \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "relation_uuid_1",
      "name": "Admin",
      "description": "Full administrative access to tenant",
      "tenant_id": null,
      "is_system": true,
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "relation_uuid_2",
      "name": "Writer",
      "description": "Can create and edit content",
      "tenant_id": null,
      "is_system": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Create Custom Relation

```bash
curl -X POST http://localhost:8080/api/v1/relations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Manager",
    "description": "Can manage team members",
    "tenant_id": "TENANT_ID"
  }'
```

### Update Relation

```bash
curl -X PATCH http://localhost:8080/api/v1/relations/relation_uuid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Senior Manager",
    "description": "Senior management role"
  }'
```

### Delete Relation

```bash
curl -X DELETE http://localhost:8080/api/v1/relations/relation_uuid \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## RBAC - Roles

### List Roles

```bash
curl http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Create Role

```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "name": "Data Analyst",
    "description": "Access to analytics and reports",
    "tenant_id": null
  }'
```

### Get Role with Permissions

```bash
curl http://localhost:8080/api/v1/roles/role_uuid \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "role_uuid",
    "name": "Content Manager",
    "description": "Manage content across services",
    "is_system": true,
    "permissions": [
      {
        "id": "perm_uuid_1",
        "service": "content-api",
        "entity": "content",
        "action": "create",
        "key": "content-api:content:create"
      },
      {
        "id": "perm_uuid_2",
        "service": "content-api",
        "entity": "content",
        "action": "update",
        "key": "content-api:content:update"
      }
    ]
  }
}
```

### Assign Permissions to Role

```bash
curl -X POST http://localhost:8080/api/v1/roles/role_uuid/permissions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "permission_ids": [
      "perm_uuid_1",
      "perm_uuid_2",
      "perm_uuid_3"
    ]
  }'
```

### Revoke Permission from Role

```bash
curl -X DELETE http://localhost:8080/api/v1/roles/role_uuid/permissions/perm_uuid \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## RBAC - Permissions

### List All Permissions

```bash
curl http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### List Permissions by Service

```bash
curl "http://localhost:8080/api/v1/permissions?service=tenant-api" \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Create Permission

```bash
curl -X POST http://localhost:8080/api/v1/permissions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "service": "billing-api",
    "entity": "invoice",
    "action": "read",
    "description": "View invoices"
  }'
```

Response:
```json
{
  "success": true,
  "message": "Permission created successfully",
  "data": {
    "id": "perm_uuid",
    "service": "billing-api",
    "entity": "invoice",
    "action": "read",
    "description": "View invoices",
    "key": "billing-api:invoice:read",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## Authorization

### Check User Permission

```bash
curl -X POST http://localhost:8080/api/v1/authorize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "user_id": "user_123",
    "tenant_id": "TENANT_ID",
    "service": "content-api",
    "entity": "content",
    "action": "create"
  }'
```

Response (Allowed):
```json
{
  "success": true,
  "data": {
    "allowed": true
  }
}
```

Response (Denied):
```json
{
  "success": true,
  "data": {
    "allowed": false,
    "reason": "User does not have the required permission"
  }
}
```

### Get All User Permissions

```bash
curl "http://localhost:8080/api/v1/permissions/user?tenant_id=TENANT_ID&user_id=user_123" \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "perm_uuid_1",
      "service": "tenant-api",
      "entity": "tenant",
      "action": "read",
      "key": "tenant-api:tenant:read"
    },
    {
      "id": "perm_uuid_2",
      "service": "content-api",
      "entity": "content",
      "action": "create",
      "key": "content-api:content:create"
    }
  ]
}
```

## Error Responses

### Validation Error

```json
{
  "success": false,
  "error": "Validation failed",
  "details": {
    "name": "field is required",
    "slug": "must be at least 3 characters"
  }
}
```

### Not Found

```json
{
  "success": false,
  "error": "tenant not found"
}
```

### Forbidden

```json
{
  "success": false,
  "error": "Access denied: You are not a member of this tenant"
}
```

### Unauthorized

```json
{
  "success": false,
  "error": "User not authenticated"
}
```

## Postman Collection

You can import these examples into Postman:

1. Create a new collection
2. Add environment variables:
   - `BASE_URL`: http://localhost:8080
   - `ACCESS_TOKEN`: Your SuperTokens access token
   - `TENANT_ID`: Your tenant UUID
3. Import the requests above

## Testing Flow

### Complete User Journey

1. **Sign up** (via SuperTokens)
2. **Create tenant** → POST /api/v1/tenants
3. **List relations** → GET /api/v1/relations
4. **Invite user** → POST /api/v1/tenants/:id/invitations
5. **User accepts** → POST /api/v1/invitations/:token/accept
6. **List members** → GET /api/v1/tenants/:id/members
7. **Create custom role** → POST /api/v1/roles
8. **Assign permissions** → POST /api/v1/roles/:id/permissions
9. **Assign role to member** → POST /api/v1/tenants/:id/members/:user_id/roles
10. **Check authorization** → POST /api/v1/authorize

---

For more examples and detailed documentation, see the [README](./README.md).

