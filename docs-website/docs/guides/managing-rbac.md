# Managing RBAC

Complete administrative guide for managing the Role-Based Access Control system.

## Overview

This guide is for platform administrators who need to:
- Set up and configure RBAC for the entire platform
- Create custom roles and policies for specific use cases
- Manage permissions across services
- Audit and maintain the authorization system
- Troubleshoot permission issues

## RBAC Architecture Review

### 3-Tier Model

```
Users → Roles → Policies → Permissions
```

**Example**:
```
User: john@example.com
  ↓
Role: "Content Editor"
  ↓
Policies: ["Content Management Policy", "Media Policy"]
  ↓
Permissions: [
  "blog-api:post:create",
  "blog-api:post:update",
  "media-api:image:upload"
]
```

## Initial RBAC Setup

### Step 1: Define Your Services and Resources

Map out your application's services and resources:

```javascript
const serviceMap = {
  "tenant-api": {
    description: "Tenant management",
    entities: ["tenant", "member", "settings"]
  },
  "blog-api": {
    description: "Blog and content management",
    entities: ["post", "comment", "category"]
  },
  "media-api": {
    description: "Media management",
    entities: ["image", "video", "file"]
  },
  "analytics-api": {
    description: "Analytics and reporting",
    entities: ["report", "dashboard", "metric"]
  }
};
```

### Step 2: Create Permissions

Create all necessary permissions systematically:

```javascript
// create-permissions.js
const services = {
  "tenant-api": {
    entities: {
      "tenant": ["create", "read", "update", "delete"],
      "member": ["invite", "read", "update", "remove"],
      "settings": ["read", "update"]
    }
  },
  "blog-api": {
    entities: {
      "post": ["create", "read", "update", "delete", "publish"],
      "comment": ["create", "read", "update", "delete", "moderate"],
      "category": ["create", "read", "update", "delete"]
    }
  }
};

async function createAllPermissions() {
  const createdPermissions = {};
  
  for (const [service, config] of Object.entries(services)) {
    createdPermissions[service] = {};
    
    for (const [entity, actions] of Object.entries(config.entities)) {
      createdPermissions[service][entity] = {};
      
      for (const action of actions) {
        const response = await fetch('/api/v1/platform/permissions', {
          method: 'POST',
          credentials: 'include',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({
            service,
            entity,
            action,
            description: `${action} ${entity} in ${service}`
          })
        });
        
        const {data} = await response.json();
        createdPermissions[service][entity][action] = data.id;
        
        console.log(`✓ Created: ${data.key}`);
      }
    }
  }
  
  // Save for reference
  localStorage.setItem('permissions', JSON.stringify(createdPermissions));
  return createdPermissions;
}

// Run
createAllPermissions().then(() => {
  console.log('All permissions created successfully');
});
```

### Step 3: Design Policies

Group permissions into logical policies:

```javascript
// Policy Design Document
const policyDesign = {
  "Tenant Admin Policy": {
    description: "Full tenant management",
    permissions: [
      "tenant-api:tenant:read",
      "tenant-api:tenant:update",
      "tenant-api:member:invite",
      "tenant-api:member:read",
      "tenant-api:member:update",
      "tenant-api:member:remove",
      "tenant-api:settings:read",
      "tenant-api:settings:update"
    ]
  },
  
  "Content Writer Policy": {
    description: "Create and edit content",
    permissions: [
      "blog-api:post:create",
      "blog-api:post:read",
      "blog-api:post:update",
      "blog-api:category:read"
    ]
  },
  
  "Content Publisher Policy": {
    description: "Publish content",
    permissions: [
      "blog-api:post:read",
      "blog-api:post:publish",
      "blog-api:comment:moderate"
    ]
  },
  
  "Media Manager Policy": {
    description: "Manage media assets",
    permissions: [
      "media-api:image:upload",
      "media-api:image:read",
      "media-api:image:delete",
      "media-api:video:upload",
      "media-api:video:read"
    ]
  },
  
  "Read Only Policy": {
    description: "View-only access",
    permissions: [
      "tenant-api:tenant:read",
      "tenant-api:member:read",
      "blog-api:post:read",
      "blog-api:category:read"
    ]
  }
};
```

### Step 4: Create Policies

```javascript
async function createPolicies(permissionsMap) {
  const createdPolicies = {};
  
  for (const [name, config] of Object.entries(policyDesign)) {
    // Create policy
    const policyResponse = await fetch('/api/v1/platform/policies', {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        name,
        description: config.description
      })
    });
    
    const policy = await policyResponse.json();
    createdPolicies[name] = policy.data.id;
    
    // Assign permissions to policy
    const permissionIds = config.permissions.map(key => {
      const [service, entity, action] = key.split(':');
      return permissionsMap[service][entity][action];
    });
    
    await fetch(`/api/v1/platform/policies/${policy.data.id}/permissions`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({permission_ids: permissionIds})
    });
    
    console.log(`✓ Created policy: ${name} with ${permissionIds.length} permissions`);
  }
  
  return createdPolicies;
}
```

### Step 5: Design Roles

Map roles to policies:

```javascript
const roleDesign = {
  "Admin": {
    description: "Full tenant administration",
    type: "tenant",
    policies: [
      "Tenant Admin Policy",
      "Content Writer Policy",
      "Content Publisher Policy",
      "Media Manager Policy"
    ]
  },
  
  "Editor": {
    description: "Edit and publish content",
    type: "tenant",
    policies: [
      "Content Writer Policy",
      "Content Publisher Policy",
      "Media Manager Policy"
    ]
  },
  
  "Writer": {
    description: "Create content",
    type: "tenant",
    policies: [
      "Content Writer Policy",
      "Media Manager Policy"
    ]
  },
  
  "Viewer": {
    description: "Read-only access",
    type: "tenant",
    policies: [
      "Read Only Policy"
    ]
  }
};
```

### Step 6: Create Roles

```javascript
async function createRoles(policiesMap) {
  const createdRoles = {};
  
  for (const [name, config] of Object.entries(roleDesign)) {
    // Create role
    const roleResponse = await fetch('/api/v1/platform/roles', {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        name,
        type: config.type,
        description: config.description
      })
    });
    
    const role = await roleResponse.json();
    createdRoles[name] = role.data.id;
    
    // Assign policies to role
    const policyIds = config.policies.map(policyName => policiesMap[policyName]);
    
    await fetch(`/api/v1/platform/roles/${role.data.id}/policies`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({policy_ids: policyIds})
    });
    
    console.log(`✓ Created role: ${name} with ${policyIds.length} policies`);
  }
  
  return createdRoles;
}
```

## Complete Setup Script

```javascript
// setup-rbac.js - Complete RBAC initialization script

async function setupCompleteRBAC() {
  console.log('=== Starting RBAC Setup ===\n');
  
  try {
    // Step 1: Create all permissions
    console.log('Step 1: Creating permissions...');
    const permissions = await createAllPermissions();
    console.log(`✓ Created ${Object.keys(permissions).length} permission groups\n`);
    
    // Step 2: Create policies
    console.log('Step 2: Creating policies...');
    const policies = await createPolicies(permissions);
    console.log(`✓ Created ${Object.keys(policies).length} policies\n`);
    
    // Step 3: Create roles
    console.log('Step 3: Creating roles...');
    const roles = await createRoles(policies);
    console.log(`✓ Created ${Object.keys(roles).length} roles\n`);
    
    console.log('=== RBAC Setup Complete! ===');
    console.log('\nRole IDs:');
    Object.entries(roles).forEach(([name, id]) => {
      console.log(`  ${name}: ${id}`);
    });
    
    // Save configuration
    const config = {permissions, policies, roles, timestamp: new Date().toISOString()};
    localStorage.setItem('rbac_config', JSON.stringify(config));
    
    return config;
  } catch (error) {
    console.error('Setup failed:', error);
    throw error;
  }
}

// Run setup
setupCompleteRBAC();
```

## Managing Existing RBAC

### View Complete RBAC Structure

```javascript
async function auditRBACStructure() {
  // Get all roles
  const rolesResponse = await fetch('/api/v1/platform/roles', {
    credentials: 'include'
  });
  const roles = await rolesResponse.json();
  
  console.log('=== RBAC Structure Audit ===\n');
  
  for (const role of roles.data) {
    console.log(`\n📋 Role: ${role.name}`);
    console.log(`   Type: ${role.type}`);
    console.log(`   Description: ${role.description}`);
    
    // Get role's policies
    const policiesResponse = await fetch(
      `/api/v1/platform/roles/${role.id}/policies`,
      {credentials: 'include'}
    );
    const policies = await policiesResponse.json();
    
    console.log(`   Policies (${policies.data.length}):`);
    
    for (const policy of policies.data) {
      console.log(`     • ${policy.name}`);
      
      // Get policy's permissions
      const policyDetails = await fetch(
        `/api/v1/platform/policies/${policy.id}`,
        {credentials: 'include'}
      );
      const policyData = await policyDetails.json();
      
      console.log(`       Permissions (${policyData.data.permissions.length}):`);
      policyData.data.permissions.forEach(perm => {
        console.log(`         - ${perm.key}`);
      });
    }
  }
}
```

### Add Permission to Existing Role

```javascript
async function addPermissionToRole(roleName, permissionKey) {
  // Find role
  const roles = await fetch('/api/v1/platform/roles', {
    credentials: 'include'
  }).then(r => r.json());
  
  const role = roles.data.find(r => r.name === roleName);
  if (!role) throw new Error(`Role not found: ${roleName}`);
  
  // Find or create permission
  const [service, entity, action] = permissionKey.split(':');
  let permission;
  
  try {
    const permissions = await fetch(
      `/api/v1/platform/permissions?service=${service}`,
      {credentials: 'include'}
    ).then(r => r.json());
    
    permission = permissions.data.find(p => p.key === permissionKey);
    
    if (!permission) {
      // Create permission
      const created = await fetch('/api/v1/platform/permissions', {
        method: 'POST',
        credentials: 'include',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({service, entity, action})
      }).then(r => r.json());
      
      permission = created.data;
    }
  } catch (err) {
    throw new Error(`Failed to get/create permission: ${err.message}`);
  }
  
  // Create or update policy for this role
  const policyName = `${roleName} Custom Policy`;
  
  // Check if custom policy exists
  const policies = await fetch('/api/v1/platform/policies', {
    credentials: 'include'
  }).then(r => r.json());
  
  let policy = policies.data.find(p => p.name === policyName);
  
  if (!policy) {
    // Create new policy
    const created = await fetch('/api/v1/platform/policies', {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        name: policyName,
        description: `Custom permissions for ${roleName}`
      })
    }).then(r => r.json());
    
    policy = created.data;
    
    // Assign policy to role
    await fetch(`/api/v1/platform/roles/${role.id}/policies`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({policy_ids: [policy.id]})
    });
  }
  
  // Add permission to policy
  await fetch(`/api/v1/platform/policies/${policy.id}/permissions`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({permission_ids: [permission.id]})
  });
  
  console.log(`✓ Added ${permissionKey} to role ${roleName}`);
}

// Usage
await addPermissionToRole('Editor', 'analytics-api:report:view');
```

## Troubleshooting RBAC

### Debug Permission Issues

```javascript
async function debugUserPermissions(userId, tenantId) {
  console.log('=== Permission Debug ===\n');
  console.log(`User: ${userId}`);
  console.log(`Tenant: ${tenantId}\n`);
  
  // 1. Check tenant membership
  const memberResponse = await fetch(
    `/api/v1/tenants/${tenantId}/members/${userId}`,
    {credentials: 'include'}
  );
  
  if (!memberResponse.ok) {
    console.error('❌ User is not a member of this tenant');
    return;
  }
  
  const member = await memberResponse.json();
  console.log(`✓ Member found`);
  console.log(`  Role: ${member.data.role_name} (${member.data.role_id})`);
  console.log(`  Active: ${member.data.is_active}\n`);
  
  // 2. Get role details
  const roleResponse = await fetch(
    `/api/v1/platform/roles/${member.data.role_id}`,
    {credentials: 'include'}
  );
  const role = await roleResponse.json();
  
  console.log(`Role Details:`);
  console.log(`  Policies: ${role.data.policies.length}\n`);
  
  // 3. Get all user permissions
  const permsResponse = await fetch(
    `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${userId}`,
    {credentials: 'include'}
  );
  const perms = await permsResponse.json();
  
  console.log(`Total Permissions: ${perms.data.length}`);
  console.log('\nPermissions by service:');
  
  const byService = {};
  perms.data.forEach(p => {
    if (!byService[p.service]) byService[p.service] = [];
    byService[p.service].push(p.key);
  });
  
  Object.entries(byService).forEach(([service, permissions]) => {
    console.log(`\n  ${service}:`);
    permissions.forEach(p => console.log(`    - ${p}`));
  });
}

// Usage
await debugUserPermissions('user-id', 'tenant-id');
```

### Find Missing Permissions

```javascript
async function findMissingPermissions(roleName, requiredPermissions) {
  // Get role
  const roles = await fetch('/api/v1/platform/roles', {
    credentials: 'include'
  }).then(r => r.json());
  
  const role = roles.data.find(r => r.name === roleName);
  if (!role) throw new Error('Role not found');
  
  // Get role's policies
  const policies = await fetch(
    `/api/v1/platform/roles/${role.id}/policies`,
    {credentials: 'include'}
  ).then(r => r.json());
  
  // Get all permissions for this role
  const allPermissions = new Set();
  
  for (const policy of policies.data) {
    const details = await fetch(
      `/api/v1/platform/policies/${policy.id}`,
      {credentials: 'include'}
    ).then(r => r.json());
    
    details.data.permissions.forEach(p => allPermissions.add(p.key));
  }
  
  // Find missing
  const missing = requiredPermissions.filter(p => !allPermissions.has(p));
  
  if (missing.length > 0) {
    console.log(`❌ Role "${roleName}" is missing ${missing.length} permissions:`);
    missing.forEach(p => console.log(`  - ${p}`));
  } else {
    console.log(`✓ Role "${roleName}" has all required permissions`);
  }
  
  return missing;
}

// Usage
const required = [
  'blog-api:post:create',
  'blog-api:post:update',
  'blog-api:post:delete',
  'media-api:image:upload'
];

await findMissingPermissions('Editor', required);
```

## Best Practices

### 1. Use Descriptive Names
```
✅ Good: "Content Management Policy", "Media Upload Policy"
❌ Bad: "Policy1", "custom_policy", "temp"
```

### 2. Keep Policies Focused
```
✅ Good: Separate policies for different concerns
  - Content Writer Policy (create, edit)
  - Content Publisher Policy (publish, moderate)

❌ Bad: One policy with everything
  - All Permissions Policy (everything mixed)
```

### 3. Regular Audits
```javascript
// Schedule regular audits
async function monthlyRBACAudit() {
  // 1. List all roles and their usage
  // 2. Check for unused permissions
  // 3. Review policy assignments
  // 4. Verify role memberships
}
```

### 4. Version Control RBAC Config
```javascript
// Export RBAC configuration
async function exportRBACConfig() {
  const config = {
    version: '1.0',
    exported_at: new Date().toISOString(),
    roles: await fetchAllRoles(),
    policies: await fetchAllPolicies(),
    permissions: await fetchAllPermissions()
  };
  
  // Save to file
  const blob = new Blob([JSON.stringify(config, null, 2)], {
    type: 'application/json'
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `rbac-config-${new Date().toISOString()}.json`;
  a.click();
}
```

## Next Steps

- [RBAC Overview](/guides/rbac-overview) - Architecture overview
- [Roles & Policies](/guides/roles-policies) - Detailed role management
- [Permissions](/guides/permissions) - Permission system
- [RBAC API](/x-api/rbac) - API reference
