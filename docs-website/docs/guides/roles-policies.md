# Roles & Policies

Complete guide to managing roles and policies in the 3-tier RBAC system.

## Overview

The RBAC system uses a 3-tier architecture:

```
User → Role → Policy → Permissions
```

- **Role**: User's position (e.g., "Admin", "Editor")
- **Policy**: Group of related permissions
- **Permission**: Atomic action (e.g., "blog-api:post:create")

This guide focuses on **Roles** and **Policies**.

## Understanding Roles

### What is a Role?

A Role represents a user's position or responsibility within a tenant. Examples:
- **Admin** - Full tenant management
- **Editor** - Content management
- **Writer** - Content creation
- **Viewer** - Read-only access
- **Reviewer** - Approval workflows

### Role Types

1. **Platform Roles** (`type: platform`)
   - Available across all tenants
   - Managed by Platform Admins
   - Examples: "Platform Admin", "Support Agent"

2. **Tenant Roles** (`type: tenant`)
   - Default roles for tenants
   - Examples: "Admin", "Writer", "Viewer", "Basic"

### System Roles

**System roles** are predefined and cannot be deleted:
- `Admin` - Full tenant control
- `Writer` - Create and manage content
- `Viewer` - Read-only access
- `Basic` - Minimal access

These are seeded during setup and serve as the foundation for your RBAC system.

## Creating Roles

### Platform Admin: Create Role

```javascript
const response = await fetch('/api/v1/platform/roles', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Content Editor',
    type: 'tenant',
    description: 'Can create, edit, and publish content'
  })
});

const {data} = await response.json();
console.log('Role created:', data.id);
```

### Role Properties

- `name` (required): Display name, 2-100 characters
- `type` (required): "platform" or "tenant"
- `description` (optional): Human-readable description, max 500 chars
- `tenant_id` (optional): For tenant-specific roles (advanced use case)

## Managing Policies

### What is a Policy?

A Policy groups related permissions together. This provides:
- **Organization**: Group permissions logically
- **Reusability**: Assign same policy to multiple roles
- **Flexibility**: Update permissions without changing roles

**Example Policy Structure**:

```
Policy: "Content Writer Policy"
├── blog-api:post:create
├── blog-api:post:update
├── blog-api:post:read
└── media-api:image:upload
```

### Creating Policies

```javascript
const response = await fetch('/api/v1/platform/policies', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Content Writer Policy',
    description: 'Create and edit blog posts',
    tenant_id: null  // null for system-wide policy
  })
});

const {data} = await response.json();
const policyId = data.id;
```

### Assigning Permissions to Policy

```javascript
// First, get available permissions
const permsResp = await fetch('/api/v1/platform/permissions?service=blog-api', {
  credentials: 'include'
});
const permissions = await permsResp.json();

// Select permissions to assign
const permissionIds = permissions.data
  .filter(p => ['create', 'update', 'read'].includes(p.action))
  .map(p => p.id);

// Assign to policy
await fetch(`/api/v1/platform/policies/${policyId}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({permission_ids: permissionIds})
});
```

### Listing Policies

```javascript
const response = await fetch('/api/v1/platform/policies', {
  credentials: 'include'
});

const {data} = await response.json();
data.forEach(policy => {
  console.log(`${policy.name}: ${policy.description}`);
});
```

### Getting Policy Details

```javascript
const response = await fetch(`/api/v1/platform/policies/${policyId}`, {
  credentials: 'include'
});

const {data} = await response.json();
console.log('Policy:', data.name);
console.log('Permissions:', data.permissions.length);
data.permissions.forEach(perm => {
  console.log(`  - ${perm.key}`);
});
```

## Connecting Roles to Policies

### Assign Policies to Role

```javascript
await fetch(`/api/v1/platform/roles/${roleId}/policies`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    policy_ids: [
      'writer-policy-uuid',
      'media-policy-uuid'
    ]
  })
});
```

### Get Role's Policies

```javascript
const response = await fetch(`/api/v1/platform/roles/${roleId}/policies`, {
  credentials: 'include'
});

const {data} = await response.json();
data.forEach(policy => {
  console.log(`Policy: ${policy.name}`);
});
```

### Revoke Policy from Role

```javascript
await fetch(`/api/v1/platform/roles/${roleId}/policies/${policyId}`, {
  method: 'DELETE',
  credentials: 'include'
});
```

## Complete RBAC Setup Example

### Scenario: Blog Platform

Create roles and policies for a blog platform with these roles:
- **Publisher** - Full control
- **Editor** - Edit all content
- **Author** - Create own content
- **Contributor** - Submit drafts

```javascript
async function setupBlogRBAC() {
  // Step 1: Create Permissions (if not exist)
  const permissions = [
    {service: 'blog-api', entity: 'post', action: 'create'},
    {service: 'blog-api', entity: 'post', action: 'read'},
    {service: 'blog-api', entity: 'post', action: 'update'},
    {service: 'blog-api', entity: 'post', action: 'delete'},
    {service: 'blog-api', entity: 'post', action: 'publish'},
    {service: 'blog-api', entity: 'comment', action: 'moderate'},
    {service: 'media-api', entity: 'image', action: 'upload'}
  ];
  
  const permIds = {};
  for (const perm of permissions) {
    const resp = await fetch('/api/v1/platform/permissions', {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        ...perm,
        description: `${perm.action} ${perm.entity}`
      })
    });
    const {data} = await resp.json();
    permIds[`${perm.entity}:${perm.action}`] = data.id;
  }
  
  // Step 2: Create Policies
  
  // Publisher Policy
  const publisherPolicyResp = await fetch('/api/v1/platform/policies', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Publisher Policy',
      description: 'Full blog management'
    })
  });
  const publisherPolicy = await publisherPolicyResp.json();
  
  await fetch(`/api/v1/platform/policies/${publisherPolicy.data.id}/permissions`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      permission_ids: Object.values(permIds)  // All permissions
    })
  });
  
  // Editor Policy
  const editorPolicyResp = await fetch('/api/v1/platform/policies', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Editor Policy',
      description: 'Edit and publish posts'
    })
  });
  const editorPolicy = await editorPolicyResp.json();
  
  await fetch(`/api/v1/platform/policies/${editorPolicy.data.id}/permissions`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      permission_ids: [
        permIds['post:read'],
        permIds['post:update'],
        permIds['post:publish'],
        permIds['image:upload']
      ]
    })
  });
  
  // Author Policy
  const authorPolicyResp = await fetch('/api/v1/platform/policies', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Author Policy',
      description: 'Create and edit own posts'
    })
  });
  const authorPolicy = await authorPolicyResp.json();
  
  await fetch(`/api/v1/platform/policies/${authorPolicy.data.id}/permissions`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      permission_ids: [
        permIds['post:create'],
        permIds['post:read'],
        permIds['post:update'],
        permIds['image:upload']
      ]
    })
  });
  
  // Step 3: Create Roles
  
  const publisherRoleResp = await fetch('/api/v1/platform/roles', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Publisher',
      type: 'tenant',
      description: 'Full blog management'
    })
  });
  const publisherRole = await publisherRoleResp.json();
  
  const editorRoleResp = await fetch('/api/v1/platform/roles', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Editor',
      type: 'tenant',
      description: 'Edit and publish content'
    })
  });
  const editorRole = await editorRoleResp.json();
  
  const authorRoleResp = await fetch('/api/v1/platform/roles', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      name: 'Author',
      type: 'tenant',
      description: 'Create own content'
    })
  });
  const authorRole = await authorRoleResp.json();
  
  // Step 4: Assign Policies to Roles
  
  await fetch(`/api/v1/platform/roles/${publisherRole.data.id}/policies`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      policy_ids: [publisherPolicy.data.id]
    })
  });
  
  await fetch(`/api/v1/platform/roles/${editorRole.data.id}/policies`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      policy_ids: [editorPolicy.data.id]
    })
  });
  
  await fetch(`/api/v1/platform/roles/${authorRole.data.id}/policies`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      policy_ids: [authorPolicy.data.id]
    })
  });
  
  console.log('RBAC setup complete!');
  return {
    roles: {
      publisher: publisherRole.data.id,
      editor: editorRole.data.id,
      author: authorRole.data.id
    },
    policies: {
      publisher: publisherPolicy.data.id,
      editor: editorPolicy.data.id,
      author: authorPolicy.data.id
    }
  };
}
```

## Updating Roles and Policies

### Update Role Name/Description

```javascript
await fetch(`/api/v1/platform/roles/${roleId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Senior Editor',
    description: 'Lead content editor with approval rights'
  })
});
```

### Update Policy

```javascript
await fetch(`/api/v1/platform/policies/${policyId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    description: 'Updated: Can create, edit, and publish content'
  })
});
```

### Add Permission to Existing Policy

```javascript
// Get current permissions
const policyResp = await fetch(`/api/v1/platform/policies/${policyId}`, {
  credentials: 'include'
});
const policy = await policyResp.json();
const currentPermIds = policy.data.permissions.map(p => p.id);

// Add new permission
const newPermIds = [...currentPermIds, newPermissionId];

await fetch(`/api/v1/platform/policies/${policyId}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({permission_ids: [newPermissionId]})
});
```

## Deleting Roles and Policies

### Delete Role

```javascript
const response = await fetch(`/api/v1/platform/roles/${roleId}`, {
  method: 'DELETE',
  credentials: 'include'
});

if (response.ok) {
  console.log('Role deleted');
}
```

:::warning Cannot Delete if In Use
You cannot delete a role that is assigned to users. Remove all user assignments first.
:::

### Delete Policy

```javascript
await fetch(`/api/v1/platform/policies/${policyId}`, {
  method: 'DELETE',
  credentials: 'include'
});
```

:::warning Cascade Effect
Deleting a policy removes it from all roles. Users with those roles will lose the associated permissions.
:::

## Role and Policy Management UI

```jsx
import {useEffect, useState} from 'react';

function RolePolicyManager() {
  const [roles, setRoles] = useState([]);
  const [policies, setPolicies] = useState([]);
  const [selectedRole, setSelectedRole] = useState(null);
  
  useEffect(() => {
    loadRoles();
    loadPolicies();
  }, []);
  
  const loadRoles = async () => {
    const resp = await fetch('/api/v1/platform/roles', {
      credentials: 'include'
    });
    const {data} = await resp.json();
    setRoles(data);
  };
  
  const loadPolicies = async () => {
    const resp = await fetch('/api/v1/platform/policies', {
      credentials: 'include'
    });
    const {data} = await resp.json();
    setPolicies(data);
  };
  
  const loadRoleDetails = async (roleId) => {
    const resp = await fetch(`/api/v1/platform/roles/${roleId}`, {
      credentials: 'include'
    });
    const {data} = await resp.json();
    setSelectedRole(data);
  };
  
  const assignPolicyToRole = async (roleId, policyId) => {
    await fetch(`/api/v1/platform/roles/${roleId}/policies`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({policy_ids: [policyId]})
    });
    loadRoleDetails(roleId);
  };
  
  const revokePolicyFromRole = async (roleId, policyId) => {
    await fetch(`/api/v1/platform/roles/${roleId}/policies/${policyId}`, {
      method: 'DELETE',
      credentials: 'include'
    });
    loadRoleDetails(roleId);
  };
  
  return (
    <div className="role-policy-manager">
      <div className="roles-list">
        <h2>Roles</h2>
        {roles.map(role => (
          <div
            key={role.id}
            onClick={() => loadRoleDetails(role.id)}
            className={selectedRole?.id === role.id ? 'selected' : ''}
          >
            <strong>{role.name}</strong>
            <p>{role.description}</p>
          </div>
        ))}
      </div>
      
      {selectedRole && (
        <div className="role-details">
          <h2>{selectedRole.name}</h2>
          <p>{selectedRole.description}</p>
          
          <h3>Assigned Policies</h3>
          {selectedRole.policies?.map(policy => (
            <div key={policy.id}>
              {policy.name}
              <button onClick={() => revokePolicyFromRole(selectedRole.id, policy.id)}>
                Remove
              </button>
            </div>
          ))}
          
          <h3>Available Policies</h3>
          <select onChange={(e) => assignPolicyToRole(selectedRole.id, e.target.value)}>
            <option value="">Select policy...</option>
            {policies
              .filter(p => !selectedRole.policies?.find(sp => sp.id === p.id))
              .map(policy => (
                <option key={policy.id} value={policy.id}>
                  {policy.name}
                </option>
              ))
            }
          </select>
        </div>
      )}
    </div>
  );
}
```

## Best Practices

### Role Naming

✅ **Good**:
- Admin, Editor, Author, Viewer
- Content Manager, Media Coordinator
- Support Agent, Developer

❌ **Avoid**:
- admin_role, role_1, temp_role
- user, member (too generic)
- X, Test, Foo

### Policy Organization

Group permissions logically:

```
✅ Good Structure:
- Content Writer Policy (post:create, post:update, post:read)
- Media Manager Policy (image:upload, video:upload)
- Comment Moderator Policy (comment:approve, comment:delete)

❌ Bad Structure:
- Policy 1 (random mix of permissions)
- All Permissions Policy (too broad)
- Single Permission Policy (too granular)
```

### Policy Reusability

Design policies to be reusable across roles:

```javascript
// Good: Composable policies
Role: "Content Manager" = 
  Content Writer Policy + 
  Media Manager Policy + 
  Comment Moderator Policy

Role: "Junior Writer" = 
  Content Writer Policy

Role: "Social Media Manager" = 
  Media Manager Policy + 
  Comment Moderator Policy
```

### Update Strategy

When changing permissions:

1. **Add new policy** instead of modifying existing
2. **Migrate roles** to new policy
3. **Test thoroughly** with test users
4. **Remove old policy** after migration

### Documentation

Document each role and policy:

```javascript
const roleDefinitions = {
  'Publisher': {
    description: 'Full blog management',
    responsibilities: [
      'Publish any content',
      'Moderate comments',
      'Manage media library',
      'Assign roles to authors'
    ],
    policies: ['Publisher Policy']
  },
  'Editor': {
    description: 'Edit and publish content',
    responsibilities: [
      'Edit all posts',
      'Publish posts',
      'Upload media'
    ],
    policies: ['Editor Policy']
  }
};
```

## Troubleshooting

### Role Not Appearing for Users

1. Check role is assigned to user
2. Verify policies are assigned to role
3. Ensure permissions are assigned to policies
4. Check authorization middleware is applied

### Permissions Not Working

```javascript
// Debug: Check user's effective permissions
async function debugUserPermissions(userId, tenantId) {
  const resp = await fetch(
    `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
    {credentials: 'include'}
  );
  const {data} = await resp.json();
  
  console.log('User permissions:');
  data.forEach(perm => console.log(`  - ${perm.key}`));
  
  return data;
}
```

### Cannot Delete Role

If you get "Role is in use" error:

```javascript
// Find users with this role
const membersResp = await fetch(
  `/api/v1/tenants/${tenantId}/members?role_id=${roleId}`,
  {credentials: 'include'}
);
const {data} = await membersResp.json();

console.log(`${data.data.length} users have this role`);

// Re-assign users to different role first
for (const member of data.data) {
  await fetch(`/api/v1/tenants/${tenantId}/members/${member.user_id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({role_id: newRoleId})
  });
}

// Now you can delete the role
await fetch(`/api/v1/platform/roles/${roleId}`, {
  method: 'DELETE',
  credentials: 'include'
});
```

## Next Steps

- [Permissions Guide](/guides/permissions) - Managing permissions
- [RBAC API](/x-api/rbac) - API reference
- [RBAC Overview](/guides/rbac-overview) - System architecture
- [Managing Members](/guides/managing-members) - Assigning roles to users
