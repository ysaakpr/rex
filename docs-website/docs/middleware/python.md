# Python Middleware Library

Python client library for RBAC authorization middleware.

## Overview

This library provides Python middleware for integrating RBAC authorization checks in Python applications and services.

**Status**: 🚧 Planned - Not yet implemented

**When available**, this library will provide:
- Flask and FastAPI middleware
- Permission checking decorators
- Tenant access verification
- Platform admin checks
- System user authentication

## Planned Features

### Flask Integration

```python
# Future implementation
from utm_rbac import RBACMiddleware, require_permission
from flask import Flask

app = Flask(__name__)
rbac = RBACMiddleware(api_url="http://localhost:8080")

@app.route('/api/posts', methods=['POST'])
@require_permission(service='blog-api', entity='post', action='create')
def create_post():
    # Your handler logic
    return {"success": True}

@app.route('/api/posts', methods=['GET'])
@require_permission(service='blog-api', entity='post', action='read')
def list_posts():
    # Your handler logic
    return {"posts": []}
```

### FastAPI Integration

```python
# Future implementation
from utm_rbac import RBACMiddleware, RequirePermission
from fastapi import FastAPI, Depends

app = FastAPI()
rbac = RBACMiddleware(api_url="http://localhost:8080")

@app.post("/api/posts")
async def create_post(
    authorized: bool = Depends(
        RequirePermission(service='blog-api', entity='post', action='create')
    )
):
    # Your handler logic
    return {"success": True}

@app.get("/api/posts")
async def list_posts(
    authorized: bool = Depends(
        RequirePermission(service='blog-api', entity='post', action='read')
    )
):
    # Your handler logic
    return {"posts": []}
```

### Decorator-Based Authorization

```python
# Future implementation
from utm_rbac import check_permission, require_tenant_access

@check_permission(service='blog-api', entity='post', action='delete')
def delete_post(tenant_id: str, post_id: str, user_id: str):
    """Delete a post if user has permission"""
    # Permission already verified by decorator
    # Your deletion logic
    pass

@require_tenant_access()
def get_tenant_data(tenant_id: str, user_id: str):
    """Get tenant data if user is a member"""
    # Tenant membership already verified
    # Your logic
    pass
```

## Current Workaround

Until the Python library is available, use direct API calls:

```python
import requests
from functools import wraps
from flask import request, jsonify

def check_permission_api(service: str, entity: str, action: str):
    """Check permission via API"""
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            # Extract from request
            tenant_id = request.view_args.get('tenant_id')
            user_id = request.headers.get('X-User-Id')  # From your auth
            
            # Check permission
            response = requests.get(
                f"http://localhost:8080/api/v1/authorize",
                params={
                    'tenant_id': tenant_id,
                    'user_id': user_id,
                    'service': service,
                    'entity': entity,
                    'action': action
                }
            )
            
            if not response.ok or not response.json()['data']['authorized']:
                return jsonify({'error': 'Permission denied'}), 403
            
            return f(*args, **kwargs)
        return decorated_function
    return decorator

# Usage
@app.route('/api/tenants/<tenant_id>/posts', methods=['POST'])
@check_permission_api(service='blog-api', entity='post', action='create')
def create_post(tenant_id):
    # Your handler
    return jsonify({'success': True})
```

### Using System User Client

For service-to-service calls:

```python
import os
import requests

class SystemUserClient:
    """Simple System User client for Python"""
    
    def __init__(self, token: str, base_url: str = "http://localhost:8080"):
        self.token = token
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        })
    
    def request(self, method: str, endpoint: str, **kwargs):
        """Make authenticated request"""
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response.json()
    
    def get_tenants(self):
        """Get all tenants"""
        result = self.request("GET", "/api/v1/tenants")
        return result["data"]["data"]
    
    def get_members(self, tenant_id: str):
        """Get tenant members"""
        result = self.request("GET", f"/api/v1/tenants/{tenant_id}/members")
        return result["data"]["data"]

# Usage
token = os.getenv("SYSTEM_USER_TOKEN")
client = SystemUserClient(token)

tenants = client.get_tenants()
for tenant in tenants:
    members = client.get_members(tenant["id"])
    print(f"{tenant['name']}: {len(members)} members")
```

## Request for Implementation

If you need this library, please:
1. Star the repository
2. Open a GitHub issue describing your use case
3. Contribute to the implementation

**Implementation Priority**: Based on community demand

## Alternative Approaches

### 1. Use Go Service as Gateway
Route Python requests through Go service with middleware

### 2. Direct API Integration
Make authorization API calls directly (as shown above)

### 3. Implement Custom Middleware
Use the patterns above as a starting point

## Planned API

When implemented, the API will follow this structure:

```python
# Configuration
from utm_rbac import configure

configure(
    api_url="http://localhost:8080",
    timeout=30,
    cache_ttl=300  # Cache permissions for 5 minutes
)

# Middleware
from utm_rbac import RBACMiddleware

middleware = RBACMiddleware()
app.add_middleware(middleware)

# Decorators
from utm_rbac import (
    require_permission,
    require_any_permission,
    require_tenant_access,
    require_platform_admin
)

# Permission checking
from utm_rbac import check_permission

authorized = check_permission(
    user_id="user-id",
    tenant_id="tenant-id",
    service="blog-api",
    entity="post",
    action="create"
)
```

## Contributing

Interested in implementing this library? See:
- [Backend Integration Guide](/guides/backend-integration)
- [Go Middleware Implementation](/middleware/go)
- [RBAC API Reference](/x-api/rbac)

## Related Documentation

- [Go Middleware](/middleware/go) - Reference implementation
- [Node.js Middleware](/middleware/nodejs) - Alternative language
- [System Auth Usage](/system-auth/usage) - System User examples
- [RBAC API](/x-api/rbac) - API reference
