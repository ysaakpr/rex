# Protected Routes

Complete guide to implementing route guards and protected pages in React.

## Overview

Protected routes ensure users are authenticated and authorized before accessing specific pages. This guide covers:
- Session-based protection with SuperTokens
- Permission-based access control
- Tenant-specific routes
- Loading states and redirects
- Custom guards

## Basic Session Protection

### Using SessionAuth Component

```jsx
import {SessionAuth} from 'supertokens-auth-react/recipe/session';
import Dashboard from './pages/Dashboard';

function App() {
  return (
    <Route path="/dashboard" element={
      <SessionAuth>
        <Dashboard />
      </SessionAuth>
    } />
  );
}
```

**What it does**:
- Checks if user has valid session
- Shows loading state while verifying
- Redirects to `/auth` if not authenticated

### Custom Redirect

```jsx
<SessionAuth
  redirectToLogin={() => {
    window.location.href = '/auth?redirect=' + window.location.pathname;
  }}
>
  <Dashboard />
</SessionAuth>
```

## Permission-Based Routes

### Permission Guard Component

```jsx
// src/components/guards/PermissionGuard.jsx
import {useState, useEffect} from 'react';
import {useParams, Navigate} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function PermissionGuard({
  service,
  entity,
  action,
  children,
  fallback = '/access-denied'
}) {
  const {tenant_id} = useParams();
  const session = useSessionContext();
  const [hasPermission, setHasPermission] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist && tenant_id) {
      fetch(
        `/api/v1/authorize?` +
        `tenant_id=${tenant_id}&` +
        `service=${service}&` +
        `entity=${entity}&` +
        `action=${action}`,
        {credentials: 'include'}
      )
        .then(res => res.json())
        .then(data => {
          setHasPermission(data.data.authorized);
          setLoading(false);
        })
        .catch(() => {
          setHasPermission(false);
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, [tenant_id, service, entity, action, session]);
  
  if (loading) {
    return <div>Checking permissions...</div>;
  }
  
  if (!hasPermission) {
    return <Navigate to={fallback} replace />;
  }
  
  return children;
}

// Usage
<Route path="/tenant/:tenant_id/posts/new" element={
  <SessionAuth>
    <PermissionGuard
      service="blog-api"
      entity="post"
      action="create"
    >
      <CreatePost />
    </PermissionGuard>
  </SessionAuth>
} />
```

### Multiple Permission Check

```jsx
// src/components/guards/RequireAnyPermission.jsx
export default function RequireAnyPermission({
  permissions, // [{service, entity, action}, ...]
  children,
  fallback = '/access-denied'
}) {
  const {tenant_id} = useParams();
  const [hasAnyPermission, setHasAnyPermission] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (!tenant_id) return;
    
    // Check all permissions in parallel
    Promise.all(
      permissions.map(p =>
        fetch(
          `/api/v1/authorize?` +
          `tenant_id=${tenant_id}&` +
          `service=${p.service}&` +
          `entity=${p.entity}&` +
          `action=${p.action}`,
          {credentials: 'include'}
        ).then(res => res.json())
      )
    )
      .then(results => {
        const hasAny = results.some(r => r.data.authorized);
        setHasAnyPermission(hasAny);
        setLoading(false);
      })
      .catch(() => {
        setHasAnyPermission(false);
        setLoading(false);
      });
  }, [tenant_id, permissions]);
  
  if (loading) return <div>Loading...</div>;
  if (!hasAnyPermission) return <Navigate to={fallback} replace />;
  
  return children;
}

// Usage
<Route path="/tenant/:tenant_id/admin" element={
  <RequireAnyPermission
    permissions={[
      {service: 'tenant-api', entity: 'admin', action: 'access'},
      {service: 'tenant-api', entity: '*', action: 'admin'}
    ]}
  >
    <AdminPanel />
  </RequireAnyPermission>
} />
```

## Tenant Access Guard

### Tenant Member Guard

```jsx
// src/components/guards/TenantMemberGuard.jsx
import {useState, useEffect} from 'react';
import {useParams, Navigate} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function TenantMemberGuard({children}) {
  const {tenant_id} = useParams();
  const session = useSessionContext();
  const [isMember, setIsMember] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist && tenant_id) {
      fetch(
        `/api/v1/tenants/${tenant_id}/members/${session.userId}`,
        {credentials: 'include'}
      )
        .then(res => {
          if (res.ok) {
            setIsMember(true);
          } else {
            setIsMember(false);
          }
          setLoading(false);
        })
        .catch(() => {
          setIsMember(false);
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, [tenant_id, session]);
  
  if (loading) return <div>Loading...</div>;
  if (!isMember) return <Navigate to="/dashboard" replace />;
  
  return children;
}

// Usage
<Route path="/tenant/:tenant_id/*" element={
  <SessionAuth>
    <TenantMemberGuard>
      <TenantLayout />
    </TenantMemberGuard>
  </SessionAuth>
} />
```

## Platform Admin Routes

### Platform Admin Guard

```jsx
// src/components/guards/PlatformAdminGuard.jsx
import {useState, useEffect} from 'react';
import {Navigate} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function PlatformAdminGuard({children}) {
  const session = useSessionContext();
  const [isAdmin, setIsAdmin] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist) {
      fetch('/api/v1/platform-admins/check', {credentials: 'include'})
        .then(res => res.json())
        .then(data => {
          setIsAdmin(data.data.is_platform_admin);
          setLoading(false);
        })
        .catch(() => {
          setIsAdmin(false);
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, [session]);
  
  if (loading) return <div>Verifying admin status...</div>;
  if (!isAdmin) return <Navigate to="/dashboard" replace />;
  
  return children;
}

// Usage
<Route path="/platform/*" element={
  <SessionAuth>
    <PlatformAdminGuard>
      <PlatformAdminPanel />
    </PlatformAdminGuard>
  </SessionAuth>
} />
```

## Complete Route Structure

### App Router with All Guards

```jsx
// src/App.jsx
import {BrowserRouter, Routes, Route, Navigate} from 'react-router-dom';
import {SessionAuth} from 'supertokens-auth-react/recipe/session';
import * as SuperTokensConfig from 'supertokens-auth-react/ui';

// Guards
import TenantMemberGuard from './components/guards/TenantMemberGuard';
import PermissionGuard from './components/guards/PermissionGuard';
import PlatformAdminGuard from './components/guards/PlatformAdminGuard';

// Pages
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import TenantView from './pages/TenantView';
import CreatePost from './pages/CreatePost';
import AdminPanel from './pages/AdminPanel';
import PlatformAdmin from './pages/PlatformAdmin';
import AccessDenied from './pages/AccessDenied';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Auth routes */}
        {SuperTokensConfig.getSuperTokensRoutesForReactRouterDom(
          require("react-router-dom")
        )}
        
        {/* Public routes */}
        <Route path="/" element={<Home />} />
        <Route path="/access-denied" element={<AccessDenied />} />
        
        {/* Protected: User Dashboard */}
        <Route path="/dashboard" element={
          <SessionAuth>
            <Dashboard />
          </SessionAuth>
        } />
        
        {/* Protected: Tenant Routes */}
        <Route path="/tenant/:tenant_id" element={
          <SessionAuth>
            <TenantMemberGuard>
              <TenantView />
            </TenantMemberGuard>
          </SessionAuth>
        } />
        
        {/* Protected: Create Post (requires permission) */}
        <Route path="/tenant/:tenant_id/posts/new" element={
          <SessionAuth>
            <TenantMemberGuard>
              <PermissionGuard
                service="blog-api"
                entity="post"
                action="create"
              >
                <CreatePost />
              </PermissionGuard>
            </TenantMemberGuard>
          </SessionAuth>
        } />
        
        {/* Protected: Tenant Admin Panel */}
        <Route path="/tenant/:tenant_id/admin" element={
          <SessionAuth>
            <TenantMemberGuard>
              <PermissionGuard
                service="tenant-api"
                entity="admin"
                action="access"
              >
                <AdminPanel />
              </PermissionGuard>
            </TenantMemberGuard>
          </SessionAuth>
        } />
        
        {/* Protected: Platform Admin */}
        <Route path="/platform/*" element={
          <SessionAuth>
            <PlatformAdminGuard>
              <PlatformAdmin />
            </PlatformAdminGuard>
          </SessionAuth>
        } />
        
        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
```

## Loading States

### Custom Loading Component

```jsx
// src/components/common/LoadingGuard.jsx
export default function LoadingGuard({children}) {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">Loading...</p>
      </div>
    </div>
  );
}
```

### Skeleton Loading

```jsx
// src/components/common/SkeletonRoute.jsx
export default function SkeletonRoute() {
  return (
    <div className="max-w-7xl mx-auto py-6 px-4">
      <div className="animate-pulse">
        <div className="h-8 bg-gray-300 rounded w-1/4 mb-4"></div>
        <div className="h-4 bg-gray-300 rounded w-1/2 mb-8"></div>
        <div className="space-y-3">
          <div className="h-20 bg-gray-300 rounded"></div>
          <div className="h-20 bg-gray-300 rounded"></div>
          <div className="h-20 bg-gray-300 rounded"></div>
        </div>
      </div>
    </div>
  );
}
```

## Error States

### Access Denied Page

```jsx
// src/pages/AccessDenied.jsx
import {Link} from 'react-router-dom';

export default function AccessDenied() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full text-center p-8">
        <div className="text-6xl mb-4">🔒</div>
        <h1 className="text-3xl font-bold text-gray-900 mb-4">
          Access Denied
        </h1>
        <p className="text-gray-600 mb-8">
          You don't have permission to access this page.
        </p>
        <Link
          to="/dashboard"
          className="inline-block px-6 py-3 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Go to Dashboard
        </Link>
      </div>
    </div>
  );
}
```

## Route Hooks

### useRequireAuth Hook

```jsx
// src/hooks/useRequireAuth.js
import {useEffect} from 'react';
import {useNavigate} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export function useRequireAuth(redirectTo = '/auth') {
  const navigate = useNavigate();
  const session = useSessionContext();
  
  useEffect(() => {
    if (!session.loading && !session.doesSessionExist) {
      navigate(redirectTo);
    }
  }, [session, navigate, redirectTo]);
  
  return session;
}

// Usage in component
function ProtectedPage() {
  const session = useRequireAuth();
  
  if (session.loading) return <div>Loading...</div>;
  
  return <div>Protected content</div>;
}
```

### usePermission Hook

```jsx
// src/hooks/usePermission.js
import {useState, useEffect} from 'react';
import {useParams} from 'react-router-dom';

export function usePermission(service, entity, action) {
  const {tenant_id} = useParams();
  const [hasPermission, setHasPermission] = useState(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (!tenant_id) {
      setLoading(false);
      return;
    }
    
    fetch(
      `/api/v1/authorize?` +
      `tenant_id=${tenant_id}&` +
      `service=${service}&` +
      `entity=${entity}&` +
      `action=${action}`,
      {credentials: 'include'}
    )
      .then(res => res.json())
      .then(data => {
        setHasPermission(data.data.authorized);
        setLoading(false);
      })
      .catch(() => {
        setHasPermission(false);
        setLoading(false);
      });
  }, [tenant_id, service, entity, action]);
  
  return {hasPermission, loading};
}

// Usage
function PostActions() {
  const {hasPermission: canDelete, loading} = usePermission(
    'blog-api', 'post', 'delete'
  );
  
  if (loading) return null;
  if (!canDelete) return null;
  
  return <button onClick={handleDelete}>Delete</button>;
}
```

## Optimization

### Permission Caching

```jsx
// src/contexts/PermissionContext.jsx
import {createContext, useContext, useState, useEffect} from 'react';
import {useParams} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

const PermissionContext = createContext(null);

export function PermissionProvider({children}) {
  const {tenant_id} = useParams();
  const session = useSessionContext();
  const [permissions, setPermissions] = useState(new Set());
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist && tenant_id) {
      fetch(
        `/api/v1/permissions/user?tenant_id=${tenant_id}&user_id=${session.userId}`,
        {credentials: 'include'}
      )
        .then(res => res.json())
        .then(data => {
          const permSet = new Set(data.data.map(p => p.key));
          setPermissions(permSet);
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, [tenant_id, session]);
  
  const hasPermission = (service, entity, action) => {
    return permissions.has(`${service}:${entity}:${action}`);
  };
  
  return (
    <PermissionContext.Provider value={{hasPermission, loading, permissions}}>
      {children}
    </PermissionContext.Provider>
  );
}

export function usePermissions() {
  return useContext(PermissionContext);
}

// Usage in route
<Route path="/tenant/:tenant_id/*" element={
  <SessionAuth>
    <TenantMemberGuard>
      <PermissionProvider>
        <TenantLayout />
      </PermissionProvider>
    </TenantMemberGuard>
  </SessionAuth>
} />
```

## Best Practices

1. **Nest Guards Properly**: Session → Tenant → Permission
2. **Show Loading States**: Better UX than blank screens
3. **Cache Permissions**: Fetch once per tenant, reuse
4. **Handle Errors Gracefully**: Show friendly error pages
5. **Redirect Smartly**: Save intended URL, redirect after login
6. **Test All Paths**: Verify guards work correctly
7. **Log Access Attempts**: Track unauthorized access for security

## Testing

### Test Protected Routes

```jsx
// src/App.test.jsx
import {render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import App from './App';

test('redirects to login when not authenticated', async () => {
  render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <App />
    </MemoryRouter>
  );
  
  await waitFor(() => {
    expect(window.location.pathname).toBe('/auth');
  });
});

test('shows dashboard when authenticated', async () => {
  // Mock authenticated session
  mockSession({userId: 'user123', doesSessionExist: true});
  
  render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <App />
    </MemoryRouter>
  );
  
  await waitFor(() => {
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });
});
```

## Next Steps

- [React Setup](/frontend/react-setup) - Initial setup
- [Component Examples](/frontend/component-examples) - UI components
- [API Calls](/frontend/api-calls) - Making API requests
- [Frontend Integration](/guides/frontend-integration) - Complete guide
