# Example: Custom RBAC Setup

Complete example of setting up a custom RBAC configuration for a specific use case.

## Scenario

You're building a blog platform with these requirements:

**User Types**:
- **Blog Admin**: Full control over blog settings and all content
- **Editor**: Can publish and manage all posts
- **Author**: Can create and edit their own posts
- **Contributor**: Can create drafts but can't publish
- **Subscriber**: Can only read content

**Services**:
- `blog-api`: Blog posts, comments, categories
- `media-api`: Images and files
- `analytics-api`: View statistics

## Solution: Custom RBAC Configuration

### Step 1: Define Permissions

```javascript
// scripts/setup-blog-rbac.js

const permissions = {
  'blog-api': {
    post: ['create', 'read', 'update', 'delete', 'publish'],
    comment: ['create', 'read', 'update', 'delete', 'moderate'],
    category: ['create', 'read', 'update', 'delete']
  },
  'media-api': {
    image: ['upload', 'read', 'delete'],
    file: ['upload', 'read', 'delete']
  },
  'analytics-api': {
    report: ['view', 'export'],
    dashboard: ['view']
  }
};

async function createPermissions() {
  const created = {};
  
  for (const [service, entities] of Object.entries(permissions)) {
    created[service] = {};
    
    for (const [entity, actions] of Object.entries(entities)) {
      created[service][entity] = {};
      
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
        created[service][entity][action] = data.id;
        console.log(`✓ Created: ${data.key}`);
      }
    }
  }
  
  return created;
}
```

### Step 2: Design Policies

```javascript
const policyDesign = {
  'Blog Full Control': {
    description: 'Complete blog management',
    permissions: [
      'blog-api:post:create',
      'blog-api:post:read',
      'blog-api:post:update',
      'blog-api:post:delete',
      'blog-api:post:publish',
      'blog-api:comment:moderate',
      'blog-api:category:create',
      'blog-api:category:update',
      'blog-api:category:delete'
    ]
  },
  
  'Content Publishing': {
    description: 'Publish and manage all content',
    permissions: [
      'blog-api:post:read',
      'blog-api:post:update',
      'blog-api:post:publish',
      'blog-api:comment:moderate'
    ]
  },
  
  'Content Creation': {
    description: 'Create and edit own content',
    permissions: [
      'blog-api:post:create',
      'blog-api:post:read',
      'blog-api:post:update'
    ]
  },
  
  'Media Management': {
    description: 'Upload and manage media',
    permissions: [
      'media-api:image:upload',
      'media-api:image:read',
      'media-api:image:delete',
      'media-api:file:upload',
      'media-api:file:read'
    ]
  },
  
  'Analytics Viewer': {
    description: 'View analytics and reports',
    permissions: [
      'analytics-api:report:view',
      'analytics-api:dashboard:view'
    ]
  },
  
  'Content Reader': {
    description: 'Read-only content access',
    permissions: [
      'blog-api:post:read',
      'blog-api:comment:read'
    ]
  }
};

async function createPolicies(permissionsMap) {
  const created = {};
  
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
    created[name] = policy.data.id;
    
    // Get permission IDs
    const permissionIds = config.permissions.map(key => {
      const [service, entity, action] = key.split(':');
      return permissionsMap[service][entity][action];
    });
    
    // Assign permissions to policy
    await fetch(`/api/v1/platform/policies/${policy.data.id}/permissions`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({permission_ids: permissionIds})
    });
    
    console.log(`✓ Created policy: ${name}`);
  }
  
  return created;
}
```

### Step 3: Create Roles

```javascript
const roleDesign = {
  'Blog Admin': {
    description: 'Full blog administration',
    type: 'tenant',
    policies: [
      'Blog Full Control',
      'Media Management',
      'Analytics Viewer'
    ]
  },
  
  'Editor': {
    description: 'Publish and manage content',
    type: 'tenant',
    policies: [
      'Content Publishing',
      'Media Management',
      'Analytics Viewer'
    ]
  },
  
  'Author': {
    description: 'Create and edit content',
    type: 'tenant',
    policies: [
      'Content Creation',
      'Media Management'
    ]
  },
  
  'Contributor': {
    description: 'Create draft content',
    type: 'tenant',
    policies: [
      'Content Creation'
    ]
  },
  
  'Subscriber': {
    description: 'Read-only access',
    type: 'tenant',
    policies: [
      'Content Reader'
    ]
  }
};

async function createRoles(policiesMap) {
  const created = {};
  
  for (const [name, config] of Object.entries(roleDesign)) {
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
    created[name] = role.data.id;
    
    const policyIds = config.policies.map(policyName => policiesMap[policyName]);
    
    await fetch(`/api/v1/platform/roles/${role.data.id}/policies`, {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({policy_ids: policyIds})
    });
    
    console.log(`✓ Created role: ${name}`);
  }
  
  return created;
}
```

### Step 4: Complete Setup Script

```javascript
// scripts/setup-blog-rbac.js - Complete script

async function setupBlogRBAC() {
  console.log('=== Setting up Blog RBAC ===\n');
  
  try {
    // Step 1: Create permissions
    console.log('Creating permissions...');
    const permissions = await createPermissions();
    console.log(`✓ Created ${countPermissions(permissions)} permissions\n`);
    
    // Step 2: Create policies
    console.log('Creating policies...');
    const policies = await createPolicies(permissions);
    console.log(`✓ Created ${Object.keys(policies).length} policies\n`);
    
    // Step 3: Create roles
    console.log('Creating roles...');
    const roles = await createRoles(policies);
    console.log(`✓ Created ${Object.keys(roles).length} roles\n`);
    
    console.log('=== RBAC Setup Complete! ===\n');
    console.log('Available Roles:');
    Object.entries(roles).forEach(([name, id]) => {
      console.log(`  - ${name}: ${id}`);
    });
    
    // Save configuration
    const config = {permissions, policies, roles, timestamp: new Date().toISOString()};
    localStorage.setItem('blog_rbac_config', JSON.stringify(config));
    
    return config;
  } catch (error) {
    console.error('Setup failed:', error);
    throw error;
  }
}

function countPermissions(perms) {
  let count = 0;
  for (const service of Object.values(perms)) {
    for (const entity of Object.values(service)) {
      count += Object.keys(entity).length;
    }
  }
  return count;
}

// Run setup
setupBlogRBAC().then(() => {
  console.log('\nRBAC configuration saved to localStorage');
}).catch(err => {
  console.error('Setup failed:', err);
});
```

## Using Custom RBAC in Backend

### Middleware for Owner-Only Access

```go
// internal/api/middleware/post_ownership.go
package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

func RequirePostOwnership(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		userID := c.GetString("user_id")
		
		var post models.Post
		if err := db.Where("id = ?", postID).First(&post).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			c.Abort()
			return
		}
		
		// Allow if user owns the post OR has publish permission
		if post.AuthorID == userID {
			c.Set("post", post)
			c.Next()
			return
		}
		
		// Check if user can manage all posts
		tenantID := c.GetString("tenant_id")
		authorized, _ := checkPermission(userID, tenantID, "blog-api", "post", "publish")
		
		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only edit your own posts",
			})
			c.Abort()
			return
		}
		
		c.Set("post", post)
		c.Next()
	}
}
```

### Handler with Custom Logic

```go
// internal/api/handlers/post_handler.go
package handlers

func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	
	var input UpdatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}
	
	// Get post
	var post models.Post
	if err := h.db.Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	
	// Check ownership OR publish permission
	canPublish, _ := checkPermission(userID, tenantID, "blog-api", "post", "publish")
	
	if post.AuthorID != userID && !canPublish {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}
	
	// Update post
	post.Title = input.Title
	post.Content = input.Content
	
	// Only editors can change status
	if input.Status != "" && input.Status != post.Status {
		if !canPublish {
			c.JSON(403, gin.H{"error": "Only editors can change post status"})
			return
		}
		post.Status = input.Status
	}
	
	h.db.Save(&post)
	
	c.JSON(200, gin.H{"data": post})
}
```

## Frontend Integration

### Role-Based UI

```jsx
// components/PostActions.jsx
import { useState, useEffect } from 'react';

export function PostActions({ post, currentUserId }) {
  const [permissions, setPermissions] = useState({
    canEdit: false,
    canPublish: false,
    canDelete: false
  });
  
  useEffect(() => {
    checkPermissions();
  }, [post]);
  
  const checkPermissions = async () => {
    const tenantId = localStorage.getItem('currentTenantId');
    
    // Check if owner
    const isOwner = post.author_id === currentUserId;
    
    // Check publish permission
    const publishResp = await fetch(
      `/api/v1/authorize?tenant_id=${tenantId}&service=blog-api&entity=post&action=publish`,
      {credentials: 'include'}
    );
    const canPublish = (await publishResp.json()).data.authorized;
    
    // Check delete permission
    const deleteResp = await fetch(
      `/api/v1/authorize?tenant_id=${tenantId}&service=blog-api&entity=post&action=delete`,
      {credentials: 'include'}
    );
    const canDelete = (await deleteResp.json()).data.authorized;
    
    setPermissions({
      canEdit: isOwner || canPublish,
      canPublish: canPublish,
      canDelete: canDelete || (isOwner && post.status === 'draft')
    });
  };
  
  return (
    <div className="post-actions">
      {permissions.canEdit && (
        <button onClick={() => editPost(post.id)}>Edit</button>
      )}
      
      {permissions.canPublish && post.status === 'draft' && (
        <button onClick={() => publishPost(post.id)}>Publish</button>
      )}
      
      {permissions.canDelete && (
        <button onClick={() => deletePost(post.id)}>Delete</button>
      )}
    </div>
  );
}
```

### Permission Hook

```jsx
// hooks/usePermission.js
import { useState, useEffect } from 'react';

export function usePermission(service, entity, action) {
  const [authorized, setAuthorized] = useState(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    checkPermission();
  }, [service, entity, action]);
  
  const checkPermission = async () => {
    const tenantId = localStorage.getItem('currentTenantId');
    
    try {
      const response = await fetch(
        `/api/v1/authorize?tenant_id=${tenantId}&service=${service}&entity=${entity}&action=${action}`,
        {credentials: 'include'}
      );
      
      const {data} = await response.json();
      setAuthorized(data.authorized);
    } catch (err) {
      console.error('Permission check failed:', err);
      setAuthorized(false);
    } finally {
      setLoading(false);
    }
  };
  
  return {authorized, loading};
}

// Usage
function CreatePostButton() {
  const {authorized, loading} = usePermission('blog-api', 'post', 'create');
  
  if (loading) return <div>Loading...</div>;
  if (!authorized) return null;
  
  return <button onClick={createPost}>Create Post</button>;
}
```

## Testing RBAC

### Test Script

```javascript
// test-blog-rbac.js
async function testRBAC() {
  const roles = ['Blog Admin', 'Editor', 'Author', 'Contributor', 'Subscriber'];
  
  for (const roleName of roles) {
    console.log(`\nTesting role: ${roleName}`);
    
    // Get role details
    const roleResp = await fetch(`/api/v1/platform/roles?name=${roleName}`, {
      credentials: 'include'
    });
    const role = (await roleResp.json()).data[0];
    
    // Get role's policies
    const policiesResp = await fetch(`/api/v1/platform/roles/${role.id}/policies`, {
      credentials: 'include'
    });
    const policies = (await policiesResp.json()).data;
    
    console.log(`  Policies: ${policies.map(p => p.name).join(', ')}`);
    
    // Get all permissions
    const permissions = new Set();
    for (const policy of policies) {
      const permResp = await fetch(`/api/v1/platform/policies/${policy.id}`, {
        credentials: 'include'
      });
      const policyData = (await permResp.json()).data;
      policyData.permissions.forEach(p => permissions.add(p.key));
    }
    
    console.log(`  Permissions (${permissions.size}):`);
    Array.from(permissions).forEach(p => console.log(`    - ${p}`));
  }
}

testRBAC();
```

## Complete Example

See the [GitHub repository](https://github.com/ysaakpr/rex) for a complete working example with:
- Backend API implementation
- Frontend React components
- RBAC setup scripts
- Test suites

## Related Documentation

- [RBAC Overview](/guides/rbac-overview) - RBAC architecture
- [Roles & Policies](/guides/roles-policies) - Managing roles
- [Permissions](/guides/permissions) - Permission system
- [RBAC API](/x-api/rbac) - API reference
