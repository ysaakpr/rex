# API Overview

Complete REST API reference for Rex.

## Base URL

```
Development:  http://localhost:8080/api/v1
Production:   https://api.yourdomain.com/api/v1
```

## Authentication

All endpoints (except public ones) require authentication using one of these methods:

### Cookie-Based (Web)

Cookies are automatically sent by the browser:

```javascript
fetch('/api/v1/tenants', {
  credentials: 'include'  // Required!
})
```

### Header-Based (API/Mobile)

Send access token in Authorization header:

```bash
Authorization: Bearer <access-token>
st-auth-mode: header
```

Example:
```bash
curl https://api.yourdomain.com/api/v1/tenants \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "st-auth-mode: header"
```

## Response Format

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // Response data here
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

### Paginated Response

```json
{
  "success": true,
  "data": {
    "data": [
      // Array of items
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 150,
    "total_pages": 8
  }
}
```

## Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| **200** | OK | Successful GET, PATCH, POST (non-creation) |
| **201** | Created | Successful resource creation |
| **204** | No Content | Successful DELETE |
| **400** | Bad Request | Invalid input, validation error |
| **401** | Unauthorized | Not authenticated (no/invalid session) |
| **403** | Forbidden | Authenticated but not authorized |
| **404** | Not Found | Resource doesn't exist |
| **500** | Internal Server Error | Unexpected server error |

## Rate Limiting

| User Type | Limit | Window |
|-----------|-------|--------|
| Regular User | 1000 requests | 1 hour |
| System User | 5000 requests | 1 hour |
| Platform Admin | 10000 requests | 1 hour |

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1732358400
```

## Pagination

List endpoints support pagination:

**Query Parameters**:
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

**Example**:
```bash
GET /api/v1/tenants?page=2&page_size=50
```

**Response**:
```json
{
  "success": true,
  "data": {
    "data": [ /* items */ ],
    "page": 2,
    "page_size": 50,
    "total_count": 250,
    "total_pages": 5
  }
}
```

## API Endpoints by Category

### Authentication

SuperTokens handles authentication endpoints:

- `POST /api/auth/signup` - Create account
- `POST /api/auth/signin` - Sign in
- `POST /api/auth/signout` - Sign out
- `POST /api/auth/session/refresh` - Refresh token
- `GET /api/auth/session` - Get session info

**[Full Authentication API →](/x-api/authentication)**

### Tenants

Manage tenant organizations:

- `POST /api/v1/tenants` - Create tenant (self-service)
- `POST /api/v1/tenants/managed` - Create managed tenant
- `GET /api/v1/tenants` - List user's tenants
- `GET /api/v1/tenants/:id` - Get tenant details
- `PATCH /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant
- `GET /api/v1/tenants/:id/status` - Get tenant status

**[Full Tenants API →](/x-api/tenants)**

### Members

Manage tenant members:

- `POST /api/v1/tenants/:id/members` - Add member
- `GET /api/v1/tenants/:id/members` - List members
- `GET /api/v1/tenants/:id/members/:user_id` - Get member
- `PATCH /api/v1/tenants/:id/members/:user_id` - Update member
- `DELETE /api/v1/tenants/:id/members/:user_id` - Remove member
- `POST /api/v1/tenants/:id/members/:user_id/roles` - Assign roles
- `DELETE /api/v1/tenants/:id/members/:user_id/roles/:role_id` - Remove role

**[Full Members API →](/x-api/members)**

### Invitations

Manage user invitations:

- `POST /api/v1/tenants/:id/invitations` - Create invitation
- `GET /api/v1/tenants/:id/invitations` - List invitations
- `GET /api/v1/invitations/:token` - Get invitation (public)
- `POST /api/v1/invitations/:token/accept` - Accept invitation
- `POST /api/v1/invitations/check-pending` - Auto-accept pending
- `DELETE /api/v1/invitations/:id` - Cancel invitation

**[Full Invitations API →](/x-api/invitations)**

### RBAC (Roles, Policies, Permissions)

Manage authorization (Platform Admin only):

**Roles**:
- `POST /api/v1/platform/roles` - Create role
- `GET /api/v1/platform/roles` - List roles
- `GET /api/v1/platform/roles/:id` - Get role
- `PATCH /api/v1/platform/roles/:id` - Update role
- `DELETE /api/v1/platform/roles/:id` - Delete role

**Policies**:
- `POST /api/v1/platform/policies` - Create policy
- `GET /api/v1/platform/policies` - List policies
- `GET /api/v1/platform/policies/:id` - Get policy
- `PATCH /api/v1/platform/policies/:id` - Update policy
- `DELETE /api/v1/platform/policies/:id` - Delete policy

**Permissions**:
- `POST /api/v1/platform/permissions` - Create permission
- `GET /api/v1/platform/permissions` - List permissions
- `GET /api/v1/platform/permissions/:id` - Get permission
- `DELETE /api/v1/platform/permissions/:id` - Delete permission

**[Full RBAC API →](/x-api/rbac)**

### System Users

Manage system users for M2M auth (Platform Admin only):

- `POST /api/v1/platform/system-users` - Create system user
- `GET /api/v1/platform/system-users` - List system users
- `GET /api/v1/platform/system-users/:id` - Get system user
- `PATCH /api/v1/platform/system-users/:id` - Update system user
- `POST /api/v1/platform/system-users/:id/regenerate-password` - Regenerate password
- `POST /api/v1/platform/system-users/:id/rotate` - Rotate with grace period
- `DELETE /api/v1/platform/system-users/:id` - Deactivate system user

**[Full System Users API →](/x-api/system-users)**

### Platform Admin

Manage platform admins:

- `GET /api/v1/platform/admins/check` - Check admin status
- `POST /api/v1/platform/admins` - Create admin
- `GET /api/v1/platform/admins` - List admins
- `GET /api/v1/platform/admins/:user_id` - Get admin
- `DELETE /api/v1/platform/admins/:user_id` - Delete admin

**[Full Platform Admin API →](/x-api/platform-admin)**

### Users

User information endpoints:

- `GET /api/v1/users/me` - Get current user
- `GET /api/v1/users` - List users
- `GET /api/v1/users/search` - Search users
- `GET /api/v1/users/:user_id` - Get user details
- `GET /api/v1/users/:user_id/tenants` - Get user's tenants
- `POST /api/v1/users/batch` - Get batch user details

**[Full Users API →](/x-api/users)**

## Request Examples

### cURL

```bash
# Create tenant
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -H "st-auth-mode: header" \
  -d '{
    "name": "My Company",
    "slug": "my-company"
  }'
```

### JavaScript (Fetch)

```javascript
// Create tenant
const response = await fetch('/api/v1/tenants', {
  method: 'POST',
  credentials: 'include',  // Include cookies
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: 'My Company',
    slug: 'my-company'
  })
});

const data = await response.json();
console.log(data);
```

### Go

```go
// Create tenant
data := map[string]string{
    "name": "My Company",
    "slug": "my-company",
}
jsonData, _ := json.Marshal(data)

req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/tenants", 
    bytes.NewBuffer(jsonData))
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("st-auth-mode", "header")
req.Header.Set("Content-Type", "application/json")

client := &http.Client{}
resp, err := client.Do(req)
```

### Python

```python
import requests

# Create tenant
response = requests.post(
    'http://localhost:8080/api/v1/tenants',
    headers={
        'Authorization': f'Bearer {token}',
        'st-auth-mode': 'header',
        'Content-Type': 'application/json'
    },
    json={
        'name': 'My Company',
        'slug': 'my-company'
    }
)

data = response.json()
print(data)
```

### Java

```java
// Create tenant
OkHttpClient client = new OkHttpClient();

JSONObject json = new JSONObject();
json.put("name", "My Company");
json.put("slug", "my-company");

RequestBody body = RequestBody.create(
    json.toString(),
    MediaType.parse("application/json")
);

Request request = new Request.Builder()
    .url("http://localhost:8080/api/v1/tenants")
    .addHeader("Authorization", "Bearer " + token)
    .addHeader("st-auth-mode", "header")
    .post(body)
    .build();

Response response = client.newCall(request).execute();
String responseData = response.body().string();
```

## Error Handling

### Common Errors

**401 Unauthorized**:
```json
{
  "success": false,
  "error": "Session not found"
}
```

**Solution**: User needs to log in

**403 Forbidden**:
```json
{
  "success": false,
  "error": "Access denied: You are not a member of this tenant"
}
```

**Solution**: User doesn't have access to this tenant

**400 Bad Request**:
```json
{
  "success": false,
  "error": "Key: 'CreateTenantInput.Slug' Error:Field validation for 'Slug' failed on the 'alphanum' tag"
}
```

**Solution**: Fix invalid input

### Best Practices

1. **Always check response status**:
   ```javascript
   if (!response.ok) {
     const error = await response.json();
     throw new Error(error.error);
   }
   ```

2. **Handle 401 errors** by redirecting to login

3. **Retry on 500 errors** with exponential backoff

4. **Log errors** for debugging

## Postman Collection

Import our Postman collection to test all endpoints:

[Download Postman Collection](https://github.com/yourusername/utm-backend/blob/main/docs/utm-backend.postman_collection.json)

## OpenAPI/Swagger

View interactive API documentation:

- [Swagger UI](http://localhost:8080/swagger/index.html) (when running locally)
- [OpenAPI Spec](https://api.yourdomain.com/swagger/doc.json)

## Next Steps

- **[Authentication API](/x-api/authentication)** - Sign up, sign in, sessions
- **[Tenants API](/x-api/tenants)** - Manage tenants
- **[Members API](/x-api/members)** - Manage tenant members
- **[RBAC API](/x-api/rbac)** - Roles, policies, permissions
- **[Frontend Integration](/frontend/api-calls)** - Call APIs from React

## Support

- **GitHub Issues**: [Report bugs](https://github.com/yourusername/utm-backend/issues)
- **Discussions**: [Ask questions](https://github.com/yourusername/utm-backend/discussions)
- **Email**: support@example.com

