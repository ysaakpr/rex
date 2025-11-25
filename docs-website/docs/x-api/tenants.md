# Tenants API

Complete API reference for tenant management endpoints.

## Base URL

```
/api/v1/tenants
```

## Endpoints Overview

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/tenants` | Create tenant (self-service) | Yes |
| POST | `/tenants/managed` | Create managed tenant | Platform Admin |
| GET | `/tenants` | List user's tenants | Yes |
| GET | `/tenants/:id` | Get tenant details | Tenant Member |
| PATCH | `/tenants/:id` | Update tenant | Tenant Member |
| DELETE | `/tenants/:id` | Delete tenant | Tenant Member |
| GET | `/tenants/:id/status` | Get tenant status | Tenant Member |
| GET | `/platform/tenants` | List all tenants | Platform Admin |
| GET | `/platform/tenants/:id` | Get any tenant | Platform Admin |

## Create Tenant (Self-Service)

Create a new tenant with the authenticated user as admin.

### Request

```http
POST /api/v1/tenants
Content-Type: application/json
Authorization: Bearer <token> (or cookies)
```

**Body**:
```json
{
  "name": "My Company",
  "slug": "my-company",
  "metadata": {
    "industry": "Technology",
    "size": "10-50"
  }
}
```

**Fields**:
- `name` (required): Display name, 2-255 characters
- `slug` (required): URL-friendly identifier, 3-255 characters, alphanumeric + hyphens
- `metadata` (optional): Custom JSON object

### Response

**Status**: `201 Created`

```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Company",
    "slug": "my-company",
    "status": "pending",
    "metadata": {
      "industry": "Technology",
      "size": "10-50"
    },
    "member_count": 1,
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

### Behavior

1. Tenant created with status `pending`
2. Creator automatically added as Admin member
3. Background job enqueued for initialization
4. Status changes to `active` after initialization

### Examples

**cURL**:
```bash
curl -X POST https://api.yourdomain.com/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "st-auth-mode: header" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "slug": "my-company"
  }'
```

**JavaScript**:
```javascript
const response = await fetch('/api/v1/tenants', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'My Company',
    slug: 'my-company'
  })
});

const data = await response.json();
console.log(data);
```

**Go**:
```go
input := models.CreateTenantInput{
    Name: "My Company",
    Slug: "my-company",
}
jsonData, _ := json.Marshal(input)

req, _ := http.NewRequest("POST", apiURL+"/api/v1/tenants", 
    bytes.NewBuffer(jsonData))
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("st-auth-mode", "header")
req.Header.Set("Content-Type", "application/json")

resp, _ := client.Do(req)
```

## Create Managed Tenant

Create a tenant on behalf of a customer (Platform Admin only).

### Request

```http
POST /api/v1/tenants/managed
Content-Type: application/json
Authorization: Bearer <admin-token>
```

**Body**:
```json
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

**Fields**:
- `name` (required): Company name
- `slug` (required): URL-friendly identifier
- `admin_email` (required): Email of the customer admin
- `metadata` (optional): Custom data

### Response

**Status**: `201 Created`

```json
{
  "success": true,
  "message": "Managed tenant created successfully",
  "data": {
    "id": "uuid",
    "name": "Enterprise Customer Inc",
    "slug": "enterprise-customer",
    "status": "pending",
    "member_count": 0,
    "created_at": "2024-11-20T10:00:00Z"
  }
}
```

### Behavior

1. Tenant created
2. Invitation sent to `admin_email`
3. When customer accepts invitation, they become Admin
4. `member_count` is 0 until invitation accepted

## List User's Tenants

Get all tenants where the user is a member.

### Request

```http
GET /api/v1/tenants?page=1&page_size=20
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
        "id": "uuid-1",
        "name": "Company A",
        "slug": "company-a",
        "status": "active",
        "member_count": 5,
        "created_at": "2024-11-15T10:00:00Z"
      },
      {
        "id": "uuid-2",
        "name": "Company B",
        "slug": "company-b",
        "status": "active",
        "member_count": 12,
        "created_at": "2024-11-18T10:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 2,
    "total_pages": 1
  }
}
```

## Get Tenant Details

Get detailed information about a specific tenant.

### Request

```http
GET /api/v1/tenants/:id
Authorization: Bearer <token>
```

**Path Parameters**:
- `id`: Tenant UUID

**Requirements**:
- User must be a member of the tenant OR
- User must be a platform admin

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Company",
    "slug": "my-company",
    "status": "active",
    "metadata": {
      "industry": "Technology",
      "size": "10-50"
    },
    "member_count": 5,
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T10:00:00Z"
  }
}
```

## Update Tenant

Update tenant information.

### Request

```http
PATCH /api/v1/tenants/:id
Content-Type: application/json
Authorization: Bearer <token>
```

**Body** (all fields optional):
```json
{
  "name": "Updated Company Name",
  "status": "active",
  "metadata": {
    "industry": "FinTech",
    "size": "50-100"
  }
}
```

**Allowed Status Transitions**:
- `pending` → `active`
- `active` → `suspended`
- `suspended` → `active`
- Any → `deleted`

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Updated Company Name",
    "slug": "my-company",
    "status": "active",
    "metadata": {...},
    "member_count": 5,
    "created_at": "2024-11-20T10:00:00Z",
    "updated_at": "2024-11-20T11:00:00Z"
  }
}
```

## Delete Tenant

Soft delete a tenant (status changed to `deleted`).

### Request

```http
DELETE /api/v1/tenants/:id
Authorization: Bearer <token>
```

### Response

**Status**: `204 No Content`

### Behavior

- Tenant status changed to `deleted`
- Data retained in database
- Can be restored by platform admin
- Members lose access immediately

## Get Tenant Status

Check tenant initialization status.

### Request

```http
GET /api/v1/tenants/:id/status
Authorization: Bearer <token>
```

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "status": "active"
  }
}
```

**Possible Values**:
- `pending` - Initialization in progress
- `active` - Ready to use
- `suspended` - Temporarily disabled
- `deleted` - Soft deleted

## List All Tenants (Platform Admin)

Get all tenants in the system (platform admin only).

### Request

```http
GET /api/v1/platform/tenants?page=1&page_size=20
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `page` (optional): Page number
- `page_size` (optional): Items per page

### Response

Same format as "List User's Tenants" but includes all tenants.

## Get Any Tenant (Platform Admin)

Platform admins can access any tenant without being a member.

### Request

```http
GET /api/v1/platform/tenants/:id
Authorization: Bearer <admin-token>
```

### Response

Same format as "Get Tenant Details".

## Error Responses

### 400 Bad Request

```json
{
  "success": false,
  "error": "Key: 'CreateTenantInput.Slug' Error:Field validation for 'Slug' failed on the 'alphanum' tag"
}
```

**Common Causes**:
- Invalid input format
- Slug already taken
- Invalid UUID format

### 401 Unauthorized

```json
{
  "success": false,
  "error": "User not authenticated"
}
```

### 403 Forbidden

```json
{
  "success": false,
  "error": "Access denied: You are not a member of this tenant"
}
```

### 404 Not Found

```json
{
  "success": false,
  "error": "Tenant not found"
}
```

## Validation Rules

**Name**:
- Required
- 2-255 characters
- Any characters allowed

**Slug**:
- Required
- 3-255 characters
- Alphanumeric + hyphens only
- Must be unique
- Lowercase recommended

**Metadata**:
- Optional
- Valid JSON object
- No size limit (reasonable use)

## Next Steps

- **[Members API](/x-api/members)** - Manage tenant members
- **[Invitations API](/x-api/invitations)** - Invite users
- **[Multi-Tenancy Guide](/guides/multi-tenancy)** - Understand multi-tenancy
