# Members API

Complete API reference for tenant member management.

## Base URL

```
/api/v1/tenants/:tenant_id/members
```

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/tenants/:id/members` | Add member to tenant |
| GET | `/tenants/:id/members` | List tenant members |
| GET | `/tenants/:id/members/:user_id` | Get member details |
| PATCH | `/tenants/:id/members/:user_id` | Update member |
| DELETE | `/tenants/:id/members/:user_id` | Remove member |
| POST | `/tenants/:id/members/:user_id/roles` | Assign roles |
| DELETE | `/tenants/:id/members/:user_id/roles/:role_id` | Remove role |

## Add Member

Add an existing user as a member of the tenant.

### Request

```http
POST /api/v1/tenants/:id/members
Content-Type: application/json
Authorization: Bearer <token>
```

**Body**:
```json
{
  "user_id": "existing-user-id",
  "role_id": "admin-role-id"
}
```

**Fields**:
- `user_id` (required): SuperTokens user ID
- `role_id` (required): Role UUID to assign

### Response

**Status**: `201 Created`

```json
{
  "success": true,
  "message": "Member added successfully",
  "data": {
    "id": "member-uuid",
    "tenant_id": "tenant-uuid",
    "user_id": "user-id",
    "role_id": "role-uuid",
    "status": "active",
    "joined_at": "2024-11-20T10:00:00Z"
  }
}
```

### Notes

- User must already exist in SuperTokens
- Use Invitations API for new users
- Requires permission to add members

## List Members

Get all members of a tenant with pagination.

### Request

```http
GET /api/v1/tenants/:id/members?page=1&page_size=20
Authorization: Bearer <token>
```

**Query Parameters**:
- `page` (optional): Page number, default 1
- `page_size` (optional): Items per page, default 20, max 100

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "member-uuid-1",
        "tenant_id": "tenant-uuid",
        "user_id": "user-id-1",
        "role_id": "admin-role-id",
        "status": "active",
        "invited_by": "inviter-user-id",
        "joined_at": "2024-11-15T10:00:00Z"
      },
      {
        "id": "member-uuid-2",
        "tenant_id": "tenant-uuid",
        "user_id": "user-id-2",
        "role_id": "writer-role-id",
        "status": "active",
        "invited_by": "inviter-user-id",
        "joined_at": "2024-11-18T10:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 2,
    "total_pages": 1
  }
}
```

## Get Member

Get details of a specific member.

### Request

```http
GET /api/v1/tenants/:id/members/:user_id
Authorization: Bearer <token>
```

**Path Parameters**:
- `id`: Tenant UUID
- `user_id`: SuperTokens user ID

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "member-uuid",
    "tenant_id": "tenant-uuid",
    "user_id": "user-id",
    "role_id": "admin-role-id",
    "status": "active",
    "invited_by": "inviter-user-id",
    "joined_at": "2024-11-20T10:00:00Z"
  }
}
```

## Update Member

Update member status or role.

### Request

```http
PATCH /api/v1/tenants/:id/members/:user_id
Content-Type: application/json
Authorization: Bearer <token>
```

**Body** (all fields optional):
```json
{
  "status": "inactive",
  "role_id": "new-role-uuid"
}
```

**Status Values**:
- `active` - Full access
- `inactive` - Suspended, no access
- `pending` - Invitation not accepted

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "member-uuid",
    "tenant_id": "tenant-uuid",
    "user_id": "user-id",
    "role_id": "new-role-uuid",
    "status": "inactive",
    "joined_at": "2024-11-20T10:00:00Z"
  }
}
```

## Remove Member

Remove a member from the tenant.

### Request

```http
DELETE /api/v1/tenants/:id/members/:user_id
Authorization: Bearer <token>
```

### Response

**Status**: `204 No Content`

### Behavior

- Member record deleted
- User loses access immediately
- User can be re-invited later
- Cannot remove last admin

## Assign Roles

Assign one or more roles to a member.

### Request

```http
POST /api/v1/tenants/:id/members/:user_id/roles
Content-Type: application/json
Authorization: Bearer <token>
```

**Body**:
```json
{
  "role_ids": [
    "role-uuid-1",
    "role-uuid-2"
  ]
}
```

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "message": "Roles assigned successfully"
  }
}
```

::: warning
Currently, members have a single `role_id`. This endpoint may be used for future multi-role support.
:::

## Remove Role

Remove a specific role from a member.

### Request

```http
DELETE /api/v1/tenants/:id/members/:user_id/roles/:role_id
Authorization: Bearer <token>
```

### Response

**Status**: `204 No Content`

## Examples

### Add Member

**cURL**:
```bash
curl -X POST https://api.yourdomain.com/api/v1/tenants/$TENANT_ID/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "st-auth-mode: header" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "existing-user-id",
    "role_id": "writer-role-id"
  }'
```

**JavaScript**:
```javascript
const response = await fetch(`/api/v1/tenants/${tenantId}/members`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    user_id: 'existing-user-id',
    role_id: 'writer-role-id'
  })
});
```

### List Members

**JavaScript**:
```javascript
const response = await fetch(
  `/api/v1/tenants/${tenantId}/members?page=1&page_size=50`,
  {credentials: 'include'}
);
const data = await response.json();
console.log(data.data.data); // Array of members
```

### Update Member Status

**cURL**:
```bash
curl -X PATCH https://api.yourdomain.com/api/v1/tenants/$TENANT_ID/members/$USER_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "st-auth-mode: header" \
  -H "Content-Type: application/json" \
  -d '{"status": "inactive"}'
```

## Error Responses

### 400 Bad Request

```json
{
  "success": false,
  "error": "Invalid user_id format"
}
```

### 403 Forbidden

```json
{
  "success": false,
  "error": "Permission denied: Cannot add members"
}
```

### 404 Not Found

```json
{
  "success": false,
  "error": "Member not found"
}
```

## Best Practices

### 1. Use Invitations for New Users

Don't add members directly if they haven't signed up:

```javascript
// ❌ Wrong: User doesn't exist yet
POST /api/v1/tenants/:id/members
{"user_id": "nonexistent-id", "role_id": "..."}

// ✅ Correct: Use invitations
POST /api/v1/tenants/:id/invitations
{"email": "newuser@example.com", "role_id": "..."}
```

### 2. Check Member Status

Before allowing operations:

```javascript
const member = await getMember(tenantId, userId);
if (member.status !== 'active') {
  // Handle inactive member
}
```

### 3. Pagination for Large Teams

Always use pagination:

```javascript
// ✅ Good
GET /api/v1/tenants/:id/members?page=1&page_size=50

// ❌ Bad (might time out for large teams)
GET /api/v1/tenants/:id/members
```

## Next Steps

- **[Invitations API](/x-api/invitations)** - Invite new users
- **[RBAC API](/x-api/rbac)** - Manage roles and permissions
- **[Member Management Guide](/guides/member-management)** - Best practices
