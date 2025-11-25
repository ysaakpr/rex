# Example: Complete User Journey

End-to-end example following a user through registration, tenant creation, inviting members, and managing permissions.

## Journey Overview

Follow **Sarah** as she:
1. Signs up and creates her company's tenant
2. Invites team members with different roles
3. Sets up custom permissions
4. Manages her team
5. Integrates an external service

## Part 1: Registration & Tenant Creation

### Step 1: Sarah Signs Up

```javascript
// Sarah visits the app
// Opens http://yourapp.com

// Clicks "Sign Up"
// Fills form:
// - Email: sarah@acmecorp.com
// - Password: SecurePassword123!

// Frontend sends request
const signupResponse = await fetch('/auth/signup', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    formFields: [
      {id: 'email', value: 'sarah@acmecorp.com'},
      {id: 'password', value: 'SecurePassword123!'}
    ]
  })
});

// Sarah is now registered and logged in
// Session cookie is set automatically
```

### Step 2: Sarah Creates Her Company's Tenant

```javascript
// Sarah sees "Create Your Workspace" page
// Fills tenant creation form:
// - Company Name: Acme Corp
// - Workspace Slug: acme-corp

const createTenantResponse = await fetch('/api/v1/tenants', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Acme Corp',
    slug: 'acme-corp'
  })
});

const tenant = await createTenantResponse.json();
// {
//   "success": true,
//   "data": {
//     "id": "tenant-123",
//     "name": "Acme Corp",
//     "slug": "acme-corp",
//     "status": "pending"
//   }
// }

// Background job starts initializing the tenant
// Sarah is automatically added as Admin member
// Dashboard loads with Sarah as the first admin
```

## Part 2: Inviting Team Members

### Step 3: Sarah Invites Her Team

```javascript
// Sarah navigates to Team Settings
// Clicks "Invite Member"

// Invites her CTO (John) as Admin
await fetch('/api/v1/invitations', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    tenant_id: 'tenant-123',
    email: 'john@acmecorp.com',
    role_id: 'admin-role-id'
  })
});

// Invites content manager (Alice) as Editor
await fetch('/api/v1/invitations', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    tenant_id: 'tenant-123',
    email: 'alice@acmecorp.com',
    role_id: 'editor-role-id'
  })
});

// Invites 3 writers as Authors
const writers = ['bob@acmecorp.com', 'carol@acmecorp.com', 'dave@acmecorp.com'];

for (const email of writers) {
  await fetch('/api/v1/invitations', {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      tenant_id: 'tenant-123',
      email,
      role_id: 'author-role-id'
    })
  });
}

// Everyone receives invitation emails
```

### Step 4: John (CTO) Accepts Invitation

```javascript
// John receives email with link:
// http://yourapp.com/accept-invitation?token=abc123...

// John clicks link and is directed to sign up
// He registers with email: john@acmecorp.com

// After registration, invitation is automatically accepted
const acceptResponse = await fetch('/api/v1/invitations/accept', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    token: 'abc123...'
  })
});

// John is now a member with Admin role
// He can see Acme Corp in his tenant list
```

## Part 3: Setting Up Custom Permissions

### Step 5: Sarah Creates Custom Role for API Access

```javascript
// Sarah wants a special role for developers who need API access
// But shouldn't be able to delete data

// First, create a custom policy
const policyResponse = await fetch('/api/v1/platform/policies', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Developer API Access',
    description: 'Read and update via API, no delete'
  })
});

const policy = await policyResponse.json();

// Add specific permissions to policy
await fetch(`/api/v1/platform/policies/${policy.data.id}/permissions`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    permission_ids: [
      'perm-blog-post-read',
      'perm-blog-post-update',
      'perm-api-access'
    ]
  })
});

// Create custom role
const roleResponse = await fetch('/api/v1/platform/roles', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Developer',
    type: 'tenant',
    description: 'API access with limited permissions'
  })
});

const role = await roleResponse.json();

// Assign policy to role
await fetch(`/api/v1/platform/roles/${role.data.id}/policies`, {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    policy_ids: [policy.data.id]
  })
});

// Sarah can now invite developers with this custom role
```

## Part 4: Daily Operations

### Step 6: Alice (Editor) Publishes a Post

```javascript
// Alice logs in and creates a draft post
const createPost = await fetch('/api/v1/posts', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    tenant_id: 'tenant-123',
    title: 'Welcome to Acme Corp Blog',
    content: 'Our first blog post...',
    status: 'draft'
  })
});

const post = await createPost.json();

// Alice reviews the post and publishes it
await fetch(`/api/v1/posts/${post.data.id}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    status: 'published'
  })
});

// Post is now live!
```

### Step 7: Bob (Author) Tries to Publish

```javascript
// Bob creates a draft
const bobPost = await fetch('/api/v1/posts', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    tenant_id: 'tenant-123',
    title: 'My Article',
    content: 'Article content...',
    status: 'draft'
  })
});

// Bob tries to publish
const publishAttempt = await fetch(`/api/v1/posts/${bobPost.data.id}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    status: 'published'
  })
});

// Response: 403 Forbidden
// {
//   "success": false,
//   "error": "Permission denied: You don't have publish permission"
// }

// Bob requests review from Alice instead
```

## Part 5: External Service Integration

### Step 8: Sarah Sets Up Analytics Service

```javascript
// Sarah needs to integrate an analytics service
// that will collect data automatically

// First, Sarah creates a System User
const systemUserResponse = await fetch('/api/v1/system-users', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'analytics-service',
    description: 'Automated analytics data collection',
    expires_in_days: 90
  })
});

const systemUser = await systemUserResponse.json();
// {
//   "success": true,
//   "data": {
//     "id": "sys-user-456",
//     "name": "analytics-service",
//     "token": "sys_abc123...",  // Save this!
//     "expires_at": "2025-02-25T10:00:00Z"
//   }
// }

// Sarah saves the token securely
localStorage.setItem('analytics_token', systemUser.data.token);

// Analytics service can now access the API
const analyticsData = await fetch('/api/v1/tenants/tenant-123/stats', {
  headers: {
    'Authorization': `Bearer ${systemUser.data.token}`
  }
});
```

## Part 6: Team Management

### Step 9: Sarah Changes Carol's Role

```javascript
// Carol has become more experienced
// Sarah promotes her from Author to Editor

// First, get Carol's member ID
const members = await fetch('/api/v1/tenants/tenant-123/members', {
  credentials: 'include'
});

const carolMember = (await members.json()).data.data.find(
  m => m.user_id === 'carol-user-id'
);

// Update Carol's role
await fetch(`/api/v1/tenants/tenant-123/members/${carolMember.id}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    role_id: 'editor-role-id'
  })
});

// Carol now has Editor permissions!
```

### Step 10: John Leaves the Company

```javascript
// John is leaving Acme Corp
// Sarah removes him from the tenant

await fetch(`/api/v1/tenants/tenant-123/members/${johnMemberId}`, {
  method: 'DELETE',
  credentials: 'include'
});

// John no longer has access to Acme Corp tenant
// But his user account still exists for other tenants
```

## Part 7: Monitoring & Analytics

### Step 11: Sarah Checks Team Activity

```javascript
// Sarah wants to see team statistics

// Get all members
const teamMembers = await fetch('/api/v1/tenants/tenant-123/members', {
  credentials: 'include'
});

const members = await teamMembers.json();
console.log(`Team size: ${members.data.total_count}`);

// Get role distribution
const roleCount = {};
members.data.data.forEach(member => {
  roleCount[member.role_name] = (roleCount[member.role_name] || 0) + 1;
});

console.log('Role distribution:', roleCount);
// {
//   "Admin": 2,
//   "Editor": 2,
//   "Author": 2
// }

// Get pending invitations
const invitations = await fetch(
  '/api/v1/invitations?tenant_id=tenant-123&status=pending',
  {credentials: 'include'}
);

console.log('Pending invitations:', (await invitations.json()).data.data.length);
```

## Journey Summary

**What Sarah Accomplished**:
1. ✅ Registered and authenticated
2. ✅ Created company tenant ("Acme Corp")
3. ✅ Became first Admin automatically
4. ✅ Invited 5 team members with appropriate roles
5. ✅ Created custom role for developers
6. ✅ Set up System User for external service
7. ✅ Managed team permissions (promoted Carol)
8. ✅ Removed team member (John)
9. ✅ Monitored team activity

**System Handled Automatically**:
- Tenant initialization in background
- Invitation email sending
- Permission inheritance through roles
- Session management
- Authorization checks
- Audit logging

## Complete Frontend Implementation

### Dashboard Component

```jsx
// Dashboard.jsx
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export function Dashboard() {
  const [tenants, setTenants] = useState([]);
  const [currentTenant, setCurrentTenant] = useState(null);
  const [stats, setStats] = useState(null);
  const navigate = useNavigate();
  
  useEffect(() => {
    loadTenants();
  }, []);
  
  useEffect(() => {
    if (currentTenant) {
      loadTenantStats();
    }
  }, [currentTenant]);
  
  const loadTenants = async () => {
    const response = await fetch('/api/v1/tenants', {
      credentials: 'include'
    });
    
    const {data} = await response.json();
    setTenants(data.data);
    
    if (data.data.length > 0) {
      setCurrentTenant(data.data[0]);
    }
  };
  
  const loadTenantStats = async () => {
    const [membersResp, invitationsResp] = await Promise.all([
      fetch(`/api/v1/tenants/${currentTenant.id}/members`, {credentials: 'include'}),
      fetch(`/api/v1/invitations?tenant_id=${currentTenant.id}&status=pending`, {credentials: 'include'})
    ]);
    
    const members = await membersResp.json();
    const invitations = await invitationsResp.json();
    
    setStats({
      members: members.data.total_count,
      pendingInvitations: invitations.data.total_count
    });
  };
  
  return (
    <div className="dashboard">
      <h1>Welcome to {currentTenant?.name}</h1>
      
      {stats && (
        <div className="stats">
          <div className="stat-card">
            <h3>Team Members</h3>
            <p>{stats.members}</p>
          </div>
          
          <div className="stat-card">
            <h3>Pending Invitations</h3>
            <p>{stats.pendingInvitations}</p>
          </div>
        </div>
      )}
      
      <div className="actions">
        <button onClick={() => navigate('/invite-member')}>
          Invite Team Member
        </button>
        
        <button onClick={() => navigate('/team')}>
          Manage Team
        </button>
      </div>
    </div>
  );
}
```

## Related Documentation

- [Authentication Guide](/guides/authentication) - User authentication
- [Creating Tenants](/guides/creating-tenants) - Tenant creation
- [Managing Members](/guides/managing-members) - Member management
- [RBAC Overview](/guides/rbac-overview) - Authorization
- [Frontend Integration](/guides/frontend-integration) - React components
- [GitHub Repository](https://github.com/ysaakpr/rex) - Complete example code
