# Node.js Middleware Library

Node.js/Express middleware for RBAC authorization.

## Overview

This library provides Express.js middleware for integrating RBAC authorization checks in Node.js applications.

**Status**: 🚧 Planned - Not yet implemented

**When available**, this library will provide:
- Express middleware
- Permission checking functions
- Tenant access verification
- Platform admin checks
- System user authentication
- TypeScript support

## Planned Features

### Express Middleware

```typescript
// Future implementation
import express from 'express';
import { requirePermission, requireTenantAccess } from 'utm-rbac';

const app = express();

// Protect a route with permission check
app.post('/api/posts',
  requirePermission('blog-api', 'post', 'create'),
  (req, res) => {
    // Permission verified, user and tenant available in req
    const { userId, tenantId } = req.auth;
    
    // Your handler logic
    res.json({ success: true });
  }
);

// Verify tenant membership
app.get('/api/tenants/:tenantId/data',
  requireTenantAccess(),
  (req, res) => {
    // User is verified member of tenant
    const { tenantId } = req.params;
    
    res.json({ data: getTenantData(tenantId) });
  }
);
```

### TypeScript Support

```typescript
// Future implementation
import { Request, Response, NextFunction } from 'express';
import { 
  RBACMiddleware, 
  PermissionCheck,
  AuthContext 
} from 'utm-rbac';

// Augmented request type
interface AuthenticatedRequest extends Request {
  auth: AuthContext;
}

// Custom permission middleware
const checkPermission = (check: PermissionCheck) => {
  return async (
    req: AuthenticatedRequest,
    res: Response,
    next: NextFunction
  ) => {
    const { userId, tenantId } = req.auth;
    
    const authorized = await rbac.checkPermission({
      userId,
      tenantId,
      ...check
    });
    
    if (!authorized) {
      return res.status(403).json({ error: 'Permission denied' });
    }
    
    next();
  };
};

// Usage
app.post('/api/posts',
  checkPermission({ service: 'blog-api', entity: 'post', action: 'create' }),
  (req: AuthenticatedRequest, res: Response) => {
    res.json({ success: true });
  }
);
```

### Multiple Permission Checks

```typescript
// Future implementation
import { requireAnyPermission, requireAllPermissions } from 'utm-rbac';

// User needs ANY of these permissions
app.get('/api/posts/:id',
  requireAnyPermission([
    { service: 'blog-api', entity: 'post', action: 'read' },
    { service: 'blog-api', entity: 'post', action: 'update' },
    { service: 'blog-api', entity: 'post', action: 'delete' }
  ]),
  (req, res) => {
    res.json({ post: getPost(req.params.id) });
  }
);

// User needs ALL of these permissions
app.post('/api/posts/:id/publish',
  requireAllPermissions([
    { service: 'blog-api', entity: 'post', action: 'update' },
    { service: 'blog-api', entity: 'post', action: 'publish' }
  ]),
  (req, res) => {
    res.json({ success: publishPost(req.params.id) });
  }
);
```

## Current Workaround

Until the Node.js library is available, use direct API calls:

```typescript
import express, { Request, Response, NextFunction } from 'express';
import axios from 'axios';

interface AuthRequest extends Request {
  userId?: string;
  tenantId?: string;
}

// Simple permission checker
const checkPermission = (
  service: string,
  entity: string,
  action: string
) => {
  return async (req: AuthRequest, res: Response, next: NextFunction) => {
    const { userId, tenantId } = req;
    
    if (!userId || !tenantId) {
      return res.status(401).json({ error: 'Unauthorized' });
    }
    
    try {
      const response = await axios.get('http://localhost:8080/api/v1/authorize', {
        params: {
          tenant_id: tenantId,
          user_id: userId,
          service,
          entity,
          action
        }
      });
      
      if (!response.data.data.authorized) {
        return res.status(403).json({ error: 'Permission denied' });
      }
      
      next();
    } catch (error) {
      console.error('Authorization check failed:', error);
      return res.status(500).json({ error: 'Authorization failed' });
    }
  };
};

// Usage
app.post('/api/posts',
  checkPermission('blog-api', 'post', 'create'),
  (req, res) => {
    res.json({ success: true });
  }
);
```

### System User Client

```typescript
import axios, { AxiosInstance } from 'axios';

class SystemUserClient {
  private client: AxiosInstance;
  
  constructor(token: string, baseURL: string = 'http://localhost:8080') {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
  }
  
  async getTenants() {
    const { data } = await this.client.get('/api/v1/tenants');
    return data.data.data;
  }
  
  async getMembers(tenantId: string) {
    const { data } = await this.client.get(`/api/v1/tenants/${tenantId}/members`);
    return data.data.data;
  }
  
  async createTenant(input: any) {
    const { data } = await this.client.post('/api/v1/tenants', input);
    return data.data;
  }
  
  async inviteUser(tenantId: string, email: string, roleId: string) {
    const { data } = await this.client.post('/api/v1/invitations', {
      tenant_id: tenantId,
      email,
      role_id: roleId
    });
    return data.data;
  }
}

// Usage
const token = process.env.SYSTEM_USER_TOKEN!;
const client = new SystemUserClient(token);

// Service-to-service operations
async function syncTenants() {
  const tenants = await client.getTenants();
  
  for (const tenant of tenants) {
    const members = await client.getMembers(tenant.id);
    console.log(`${tenant.name}: ${members.length} members`);
  }
}
```

### Reusable Permission Module

```typescript
// rbac-helper.ts
import axios from 'axios';

export interface PermissionCheck {
  service: string;
  entity: string;
  action: string;
}

export class RBACHelper {
  constructor(private apiUrl: string = 'http://localhost:8080') {}
  
  async checkPermission(
    userId: string,
    tenantId: string,
    check: PermissionCheck
  ): Promise<boolean> {
    try {
      const { data } = await axios.get(`${this.apiUrl}/api/v1/authorize`, {
        params: {
          user_id: userId,
          tenant_id: tenantId,
          ...check
        }
      });
      
      return data.data.authorized;
    } catch (error) {
      console.error('Permission check failed:', error);
      return false;
    }
  }
  
  async checkAnyPermission(
    userId: string,
    tenantId: string,
    checks: PermissionCheck[]
  ): Promise<boolean> {
    const results = await Promise.all(
      checks.map(check => this.checkPermission(userId, tenantId, check))
    );
    
    return results.some(authorized => authorized);
  }
  
  async checkAllPermissions(
    userId: string,
    tenantId: string,
    checks: PermissionCheck[]
  ): Promise<boolean> {
    const results = await Promise.all(
      checks.map(check => this.checkPermission(userId, tenantId, check))
    );
    
    return results.every(authorized => authorized);
  }
  
  async getUserPermissions(
    userId: string,
    tenantId: string
  ): Promise<PermissionCheck[]> {
    const { data } = await axios.get(`${this.apiUrl}/api/v1/permissions/user`, {
      params: { user_id: userId, tenant_id: tenantId }
    });
    
    return data.data.map((p: any) => ({
      service: p.service,
      entity: p.entity,
      action: p.action
    }));
  }
}

// Usage
const rbac = new RBACHelper();

const canCreate = await rbac.checkPermission(
  'user-id',
  'tenant-id',
  { service: 'blog-api', entity: 'post', action: 'create' }
);

const permissions = await rbac.getUserPermissions('user-id', 'tenant-id');
```

## Planned API

When implemented, the API will follow this structure:

```typescript
// Configuration
import { configure } from 'utm-rbac';

configure({
  apiUrl: 'http://localhost:8080',
  timeout: 30000,
  cacheTTL: 300, // Cache permissions for 5 minutes
  cacheEnabled: true
});

// Middleware
import { 
  requirePermission,
  requireAnyPermission,
  requireTenantAccess,
  requirePlatformAdmin
} from 'utm-rbac';

// Direct checks
import { checkPermission, getUserPermissions } from 'utm-rbac';

const authorized = await checkPermission({
  userId: 'user-id',
  tenantId: 'tenant-id',
  service: 'blog-api',
  entity: 'post',
  action: 'create'
});

const permissions = await getUserPermissions('user-id', 'tenant-id');
```

## Package Structure

```
utm-rbac/
├── src/
│   ├── middleware/
│   │   ├── permission.ts       # Permission middleware
│   │   ├── tenant.ts           # Tenant access middleware
│   │   └── platform-admin.ts  # Platform admin middleware
│   ├── client/
│   │   ├── rbac.ts            # RBAC API client
│   │   └── system-user.ts     # System User client
│   ├── cache/
│   │   └── memory.ts          # In-memory cache
│   ├── types.ts               # TypeScript types
│   └── index.ts               # Main export
├── tests/
│   └── ...
├── package.json
├── tsconfig.json
└── README.md
```

## Contributing

Interested in implementing this library? See:
- [Backend Integration Guide](/guides/backend-integration)
- [Go Middleware Implementation](/middleware/go)
- [RBAC API Reference](/x-api/rbac)

## Request for Implementation

If you need this library, please:
1. Star the repository
2. Open a GitHub issue with your use case
3. Consider contributing to implementation

**Implementation Priority**: Based on community demand

## Related Documentation

- [Go Middleware](/middleware/go) - Reference implementation
- [Python Middleware](/middleware/python) - Alternative language
- [System Auth Usage](/system-auth/usage) - System User examples
- [RBAC API](/x-api/rbac) - API reference
