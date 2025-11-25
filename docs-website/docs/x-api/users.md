# Users API

Complete API reference for user management and retrieval.

## Overview

The Users API provides endpoints to:
- Get current authenticated user details
- Retrieve user information
- List all users (Platform Admin only)
- Get user's tenant memberships
- Search users across the system
- Batch retrieve user details

**Base URL**: `/api/v1/users`

**Authentication**: User session required (SuperTokens)

## Get Current User

Get the authenticated user's profile information.

**Request**:
```http
GET /api/v1/users/me
Authorization: Bearer <token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "auth0|user123",
    "email": "john.doe@example.com",
    "time_joined": 1700000000000,
    "email_verified": true
  }
}
```

**Fields**:
- `id`: SuperTokens user ID
- `email`: User's email address
- `time_joined`: Unix timestamp (milliseconds) of account creation
- `email_verified`: Email verification status

:::tip Use Case
Use this endpoint on app load to display user profile, check verification status, and personalize UI.
:::

## Get User Details

Get detailed information about a specific user.

**Request**:
```http
GET /api/v1/users/:id
Authorization: Bearer <token>
```

**Response** (200):
```json
{
  "success": true,
  "data": {
    "id": "auth0|user456",
    "email": "jane.smith@example.com",
    "time_joined": 1699000000000,
    "email_verified": true
  }
}
```

**Error** (404):
```json
{
  "success": false,
  "error": "User not found"
}
```

## List All Users

Get all registered users (Platform Admin only).

**Request**:
```http
GET /api/v1/users?page=1&page_size=50
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 50, max: 100)

**Response** (200):
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "auth0|user123",
        "email": "john.doe@example.com",
        "time_joined": 1700000000000,
        "email_verified": true
      },
      {
        "id": "auth0|user456",
        "email": "jane.smith@example.com",
        "time_joined": 1699000000000,
        "email_verified": false
      }
    ],
    "total_count": 150
  }
}
```

:::warning Platform Admin Only
This endpoint requires Platform Admin privileges. Regular users receive a 403 Forbidden error.
:::

## Get User Tenants

Get all tenants a user is a member of.

**Request**:
```http
GET /api/v1/users/:id/tenants
Authorization: Bearer <token>
```

**Response** (200):
```json
{
  "success": true,
  "data": [
    {
      "tenant_id": "tenant-uuid-1",
      "tenant_name": "Acme Corp",
      "tenant_slug": "acme-corp",
      "role_id": "admin-role-uuid",
      "role_name": "Admin",
      "joined_at": "2024-11-01T10:00:00Z",
      "is_active": true
    },
    {
      "tenant_id": "tenant-uuid-2",
      "tenant_name": "Beta Inc",
      "tenant_slug": "beta-inc",
      "role_id": "viewer-role-uuid",
      "role_name": "Viewer",
      "joined_at": "2024-11-15T14:30:00Z",
      "is_active": true
    }
  ]
}
```

**Response Fields**:
- `tenant_id`: Tenant UUID
- `tenant_name`: Display name
- `tenant_slug`: URL-safe identifier
- `role_id`: User's role UUID in this tenant
- `role_name`: Role display name (e.g., "Admin", "Writer")
- `joined_at`: Membership creation timestamp
- `is_active`: Membership status

**Use Cases**:
- Display tenant switcher in UI
- Check which tenants user can access
- Show user's roles across tenants

## Search Users

Search for users by email or name (partial match).

**Request**:
```http
GET /api/v1/users/search?q=john&page=1&page_size=20
Authorization: Bearer <token>
```

**Query Parameters**:
- `q` (required): Search query (min 2 characters)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Response** (200):
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "auth0|user123",
        "email": "john.doe@example.com",
        "time_joined": 1700000000000,
        "email_verified": true
      },
      {
        "id": "auth0|user789",
        "email": "johnny.bravo@example.com",
        "time_joined": 1698000000000,
        "email_verified": true
      }
    ],
    "total_count": 2,
    "page": 1,
    "page_size": 20
  }
}
```

:::info Search Behavior
- Case-insensitive search
- Matches email addresses (partial)
- Returns results from SuperTokens user database
- Minimum 2 characters required
:::

## Get Batch User Details

Retrieve details for multiple users in a single request.

**Request**:
```http
POST /api/v1/users/batch
Content-Type: application/json
Authorization: Bearer <token>
```

**Body**:
```json
{
  "user_ids": [
    "auth0|user123",
    "auth0|user456",
    "auth0|user789"
  ]
}
```

**Fields**:
- `user_ids` (required): Array of SuperTokens user IDs (max 100)

**Response** (200):
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "auth0|user123",
        "email": "john.doe@example.com",
        "time_joined": 1700000000000,
        "email_verified": true
      },
      {
        "id": "auth0|user456",
        "email": "jane.smith@example.com",
        "time_joined": 1699000000000,
        "email_verified": false
      }
    ],
    "not_found": ["auth0|user789"]
  }
}
```

**Response Fields**:
- `users`: Array of found user objects
- `not_found`: Array of user IDs that don't exist

:::tip Use Case
Efficiently retrieve user details for member lists, activity feeds, or reports without multiple API calls.
:::

## Complete Examples

### React: User Profile Component

```jsx
import {useEffect, useState} from 'react';

function UserProfile() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/v1/users/me', {
      credentials: 'include'
    })
      .then(res => res.json())
      .then(data => {
        setUser(data.data);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to load user:', err);
        setLoading(false);
      });
  }, []);

  if (loading) return <div>Loading...</div>;
  if (!user) return <div>Not logged in</div>;

  return (
    <div>
      <h2>Profile</h2>
      <p>Email: {user.email}</p>
      <p>Verified: {user.email_verified ? 'Yes' : 'No'}</p>
      <p>Member since: {new Date(user.time_joined).toLocaleDateString()}</p>
    </div>
  );
}
```

### React: Tenant Switcher

```jsx
import {useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';

function TenantSwitcher({userId}) {
  const [tenants, setTenants] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetch(`/api/v1/users/${userId}/tenants`, {
      credentials: 'include'
    })
      .then(res => res.json())
      .then(data => setTenants(data.data));
  }, [userId]);

  const switchTenant = (tenantSlug) => {
    navigate(`/tenant/${tenantSlug}`);
  };

  return (
    <div className="tenant-switcher">
      <label>Switch Tenant:</label>
      <select onChange={(e) => switchTenant(e.target.value)}>
        {tenants.map(t => (
          <option key={t.tenant_id} value={t.tenant_slug}>
            {t.tenant_name} ({t.role_name})
          </option>
        ))}
      </select>
    </div>
  );
}
```

### Go: Get User Email

```go
func GetUserEmail(userID string) (string, error) {
    client := &http.Client{}
    
    req, err := http.NewRequest("GET", 
        fmt.Sprintf("http://localhost:8080/api/v1/users/%s", userID), 
        nil)
    if err != nil {
        return "", err
    }
    
    // Add authentication token
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("API error: %d", resp.StatusCode)
    }
    
    var result struct {
        Success bool `json:"success"`
        Data struct {
            Email string `json:"email"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    
    return result.Data.Email, nil
}
```

### JavaScript: Search Users with Debounce

```javascript
// Debounce helper
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// Search component
const searchInput = document.getElementById('user-search');
const resultsDiv = document.getElementById('search-results');

const performSearch = debounce(async (query) => {
  if (query.length < 2) {
    resultsDiv.innerHTML = '';
    return;
  }
  
  const response = await fetch(`/api/v1/users/search?q=${encodeURIComponent(query)}`, {
    credentials: 'include'
  });
  
  const {data} = await response.json();
  
  resultsDiv.innerHTML = data.users.map(user => `
    <div class="user-result">
      <strong>${user.email}</strong>
      ${user.email_verified ? '✓' : '⚠️'}
    </div>
  `).join('');
}, 300);

searchInput.addEventListener('input', (e) => {
  performSearch(e.target.value);
});
```

### Node.js: Batch Get User Details

```javascript
async function getUsersDetails(userIds) {
  const response = await fetch('/api/v1/users/batch', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({user_ids: userIds})
  });
  
  const {data} = await response.json();
  
  // Handle not found users
  if (data.not_found.length > 0) {
    console.warn('Users not found:', data.not_found);
  }
  
  // Create email lookup map
  const emailMap = {};
  data.users.forEach(user => {
    emailMap[user.id] = user.email;
  });
  
  return emailMap;
}

// Usage
const memberIds = ['auth0|user1', 'auth0|user2', 'auth0|user3'];
const emails = await getUsersDetails(memberIds);
console.log('Member emails:', emails);
```

## Integration with Tenant Members

When displaying tenant members, combine Users API with Members API:

```javascript
async function getTenantMembersWithDetails(tenantId) {
  // 1. Get tenant members
  const membersResp = await fetch(`/api/v1/tenants/${tenantId}/members`, {
    credentials: 'include'
  });
  const members = await membersResp.json();
  
  // 2. Extract user IDs
  const userIds = members.data.data.map(m => m.user_id);
  
  // 3. Batch get user details
  const usersResp = await fetch('/api/v1/users/batch', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({user_ids: userIds})
  });
  const users = await usersResp.json();
  
  // 4. Merge data
  const userMap = {};
  users.data.users.forEach(u => userMap[u.id] = u);
  
  return members.data.data.map(member => ({
    ...member,
    email: userMap[member.user_id]?.email,
    verified: userMap[member.user_id]?.email_verified
  }));
}
```

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 400 | Search query too short | Minimum 2 characters required |
| 400 | Too many user IDs | Batch request exceeds 100 users |
| 401 | Unauthorized | No valid session |
| 403 | Forbidden | Insufficient permissions |
| 404 | User not found | Invalid user ID |

## Security Considerations

### Privacy

- Users can only see their own profile details via `/me`
- Viewing other users requires appropriate permissions
- Email addresses are PII - handle according to privacy policy

### Rate Limiting

- Search endpoint: 60 requests/minute per user
- Batch endpoint: 30 requests/minute per user
- List all users: 10 requests/minute (Platform Admin only)

### Data Minimization

- Only request user data when necessary
- Cache user details in frontend to reduce API calls
- Use batch endpoints instead of multiple single requests

## Related Endpoints

- [Members API](/x-api/members) - Tenant member management
- [Platform Admin API](/x-api/platform-admin) - Platform administrator management
- [Authentication](/x-api/authentication) - SuperTokens integration

## Next Steps

- [User Authentication Guide](/guides/user-authentication)
- [Multi-Tenancy Guide](/guides/multi-tenancy)
- [Frontend Integration](/guides/frontend-integration)
