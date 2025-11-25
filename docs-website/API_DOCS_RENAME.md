# API Documentation Folder Rename - Conflict Resolution

**Date**: November 25, 2025  
**Issue**: Nginx route conflict between backend API and documentation  
**Solution**: Renamed docs API folder from `/api` to `/x-api`

## Problem

The documentation had an `/api` folder containing API reference documentation, which conflicted with the nginx route that proxies `/api/*` requests to the backend API server.

### Conflict Scenario

```
User Request: http://localhost/api/overview
                     ↓
              Nginx sees /api
                     ↓
         Proxies to backend API
                     ↓
        Backend returns 404 (not found)
                     ↓
     Documentation never served!
```

## Solution

Renamed the documentation API folder to avoid the conflict:

```bash
# Renamed folder
docs/api/ → docs/x-api/

# Updated all references in VitePress config
/api/overview → /x-api/overview
/api/authentication → /x-api/authentication
/api/tenants → /x-api/tenants
# ... etc
```

## Changes Made

### 1. Folder Rename
```bash
cd docs-website/docs
mv api x-api
```

### 2. VitePress Config Update

**File**: `docs-website/docs/.vitepress/config.mts`

**Navigation Updated**:
```typescript
// Before
{ text: 'API Reference', link: '/api/overview' }

// After
{ text: 'API Reference', link: '/x-api/overview' }
```

**Sidebar Updated**:
```typescript
{
  text: 'API Reference',
  items: [
    { text: 'Overview', link: '/x-api/overview' },
    { text: 'Authentication', link: '/x-api/authentication' },
    { text: 'Tenants', link: '/x-api/tenants' },
    { text: 'Members', link: '/x-api/members' },
    { text: 'Invitations', link: '/x-api/invitations' },
    { text: 'RBAC', link: '/x-api/rbac' },
    { text: 'System Users', link: '/x-api/system-users' },
    { text: 'Platform Admin', link: '/x-api/platform-admin' },
    { text: 'Users', link: '/x-api/users' }
  ]
}
```

## URL Mapping

### Documentation URLs (Now Working)
| Documentation Page | URL |
|-------------------|-----|
| API Overview | `http://localhost/x-api/overview` |
| Authentication API | `http://localhost/x-api/authentication` |
| Tenants API | `http://localhost/x-api/tenants` |
| Members API | `http://localhost/x-api/members` |
| Invitations API | `http://localhost/x-api/invitations` |
| RBAC API | `http://localhost/x-api/rbac` |
| System Users API | `http://localhost/x-api/system-users` |
| Platform Admin API | `http://localhost/x-api/platform-admin` |
| Users API | `http://localhost/x-api/users` |

### Backend API URLs (Still Working)
| Backend Endpoint | URL |
|-----------------|-----|
| Authentication | `http://localhost/api/auth/*` |
| Tenants | `http://localhost/api/v1/tenants` |
| Members | `http://localhost/api/v1/members` |
| Platform Admin | `http://localhost/api/v1/platform/*` |

**No conflicts!** Documentation and backend API coexist peacefully.

## Nginx Configuration

**No changes needed** to nginx.conf:

```nginx
# Backend API proxy (unchanged)
location /api {
    proxy_pass http://api;
    # ... headers, CORS, etc.
}

# Documentation (unchanged)
location / {
    proxy_pass http://docs;
    # ... serves all docs including /x-api/*
}
```

## Files Changed

1. ✅ `docs-website/docs/api/` → `docs-website/docs/x-api/` (folder renamed)
2. ✅ `docs-website/docs/.vitepress/config.mts` (9 link references updated)

## Testing

All API documentation links now work:

```bash
✅ http://localhost/x-api/overview → 200 OK
✅ http://localhost/x-api/authentication → 200 OK
✅ http://localhost/x-api/tenants → 200 OK
✅ http://localhost/x-api/members → 200 OK
✅ http://localhost/x-api/invitations → 200 OK
✅ http://localhost/x-api/rbac → 200 OK
✅ http://localhost/x-api/system-users → 200 OK
✅ http://localhost/x-api/platform-admin → 200 OK
✅ http://localhost/x-api/users → 200 OK
```

Backend API routes still work:
```bash
✅ http://localhost/api/auth/* → Backend API
✅ http://localhost/api/v1/* → Backend API
```

## Why `/x-api`?

- **`x-` prefix** is a common convention for custom/extended resources
- **Short and clear** - obvious it's API documentation
- **No conflicts** - won't clash with standard paths
- **Easy to remember** - similar to original `/api` path

## Alternative Prefixes (Not Used)

We could have used:
- `/api-docs` - longer but very descriptive
- `/reference` - generic but less clear
- `/docs-api` - redundant with docs context
- `/apidocs` - no separator, less readable

Chose `/x-api` for brevity and convention.

## Navigation Flow

### From Documentation Homepage

```
Homepage → Click "API Reference" → /x-api/overview
```

### From Sidebar

```
Sidebar → API Reference section → Click any API doc → /x-api/{page}
```

### Direct Access

```
http://localhost/x-api/tenants
```

All paths work correctly!

## Future Considerations

### Adding More API Documentation

When adding new API endpoint documentation:

1. Create markdown file in `docs-website/docs/x-api/`
2. Update `config.mts` sidebar under "API Reference"
3. Use path `/x-api/your-new-page`

Example:
```typescript
{
  text: 'API Reference',
  items: [
    // ... existing items
    { text: 'Your New API', link: '/x-api/your-new-api' }
  ]
}
```

### Other Potential Conflicts

Be mindful of these reserved paths:
- `/api` - Backend API routes
- `/demo` - Admin demo app
- `/inbox` - MailHog email testing
- `/@vite` - Vite HMR client
- `/@fs` - Vite filesystem access

## Benefits

✅ **No nginx changes** - Kept configuration simple  
✅ **Clear separation** - Docs and API clearly different  
✅ **No conflicts** - Both systems work independently  
✅ **Easy to maintain** - Simple folder structure  
✅ **Consistent URLs** - All docs at `/x-api/*`  

## Rollback (If Needed)

To revert this change:

```bash
# 1. Rename folder back
cd docs-website/docs
mv x-api api

# 2. Update config.mts
# Replace all /x-api/ with /api/

# 3. Restart docs
docker-compose restart docs
```

But this would bring back the conflict!

## Documentation

This change is transparent to users:
- Navigation works seamlessly
- Links are updated automatically
- No broken links
- Search still works

## Status

✅ **Implementation**: Complete  
✅ **Testing**: Verified all links work  
✅ **Conflicts**: Resolved  
✅ **Documentation**: Updated  

---

**Summary**: API documentation successfully moved from `/api` to `/x-api` to avoid nginx routing conflicts with backend API. All 9 API reference pages now load correctly without modifying nginx configuration.

