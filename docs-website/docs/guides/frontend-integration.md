# Frontend Integration

Complete guide to integrating your frontend application with the multi-tenant backend.

## Overview

This guide covers:
- SuperTokens React SDK setup
- Making authenticated API calls
- Multi-tenant navigation
- Permission-based UI
- Error handling
- Best practices

## Quick Start

### 1. Install Dependencies

```bash
npm install supertokens-auth-react react-router-dom
```

### 2. Initialize SuperTokens

```jsx
// src/main.jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import SuperTokens from 'supertokens-auth-react';
import EmailPassword from 'supertokens-auth-react/recipe/emailpassword';
import ThirdParty from 'supertokens-auth-react/recipe/thirdparty';
import Session from 'supertokens-auth-react/recipe/session';
import App from './App';

SuperTokens.init({
  appInfo: {
    appName: "Your App Name",
    apiDomain: "http://localhost:8080",
    websiteDomain: "http://localhost:3000",
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    ThirdParty.init({
      signInAndUpFeature: {
        providers: [
          ThirdParty.Google.init()
        ]
      }
    }),
    Session.init()
  ]
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

### 3. Setup Routing

```jsx
// src/App.jsx
import {BrowserRouter, Routes, Route} from 'react-router-dom';
import {SessionAuth} from 'supertokens-auth-react/recipe/session';
import * as SuperTokensConfig from 'supertokens-auth-react/ui';
import EmailPassword from 'supertokens-auth-react/recipe/emailpassword/prebuiltui';
import ThirdParty from 'supertokens-auth-react/recipe/thirdparty/prebuiltui';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Auth routes */}
        {SuperTokensConfig.getSuperTokensRoutesForReactRouterDom(require("react-router-dom"), [
          EmailPassword,
          ThirdParty
        ])}
        
        {/* Protected routes */}
        <Route path="/dashboard" element={
          <SessionAuth>
            <Dashboard />
          </SessionAuth>
        } />
        
        <Route path="/tenant/:slug" element={
          <SessionAuth>
            <TenantView />
          </SessionAuth>
        } />
        
        {/* Public routes */}
        <Route path="/" element={<Home />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
```

## Making API Calls

### Always Include Credentials

```javascript
const response = await fetch('/api/v1/tenants', {
  method: 'GET',
  credentials: 'include',  // ← CRITICAL: Sends cookies
  headers: {
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
```

### API Helper Function

```javascript
// src/lib/api.js
export async function apiCall(endpoint, options = {}) {
  const defaultOptions = {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    }
  };
  
  const response = await fetch(
    `${import.meta.env.VITE_API_URL || ''}${endpoint}`,
    {...defaultOptions, ...options}
  );
  
  // Handle 401 - redirect to login
  if (response.status === 401) {
    window.location.href = '/auth';
    throw new Error('Session expired');
  }
  
  // Handle errors
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'API error');
  }
  
  return response.json();
}

// Usage
const tenants = await apiCall('/api/v1/tenants');
const newTenant = await apiCall('/api/v1/tenants', {
  method: 'POST',
  body: JSON.stringify({name: 'New Tenant', slug: 'new-tenant'})
});
```

### React Hook: useAPI

```jsx
import {useState, useEffect} from 'react';
import {apiCall} from '../lib/api';

export function useAPI(endpoint, options = {}) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    apiCall(endpoint, options)
      .then(result => {
        setData(result.data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [endpoint]);
  
  return {data, loading, error};
}

// Usage
function TenantList() {
  const {data: tenants, loading, error} = useAPI('/api/v1/tenants');
  
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  
  return (
    <ul>
      {tenants.data.map(tenant => (
        <li key={tenant.id}>{tenant.name}</li>
      ))}
    </ul>
  );
}
```

## User Session Management

### Get Current User

```jsx
import {useSessionContext} from 'supertokens-auth-react/recipe/session';
import {useEffect, useState} from 'react';

function UserProfile() {
  const session = useSessionContext();
  const [user, setUser] = useState(null);
  
  useEffect(() => {
    if (session.doesSessionExist) {
      fetch('/api/v1/users/me', {credentials: 'include'})
        .then(res => res.json())
        .then(data => setUser(data.data));
    }
  }, [session]);
  
  if (session.loading) return <div>Loading...</div>;
  if (!session.doesSessionExist) return <div>Not logged in</div>;
  
  return (
    <div>
      <h2>Profile</h2>
      {user && (
        <>
          <p>Email: {user.email}</p>
          <p>User ID: {session.userId}</p>
          <p>Verified: {user.email_verified ? 'Yes' : 'No'}</p>
        </>
      )}
    </div>
  );
}
```

### Sign Out

```jsx
import {signOut} from 'supertokens-auth-react/recipe/session';

function SignOutButton() {
  const handleSignOut = async () => {
    await signOut();
    window.location.href = '/';
  };
  
  return <button onClick={handleSignOut}>Sign Out</button>;
}
```

## Multi-Tenant Navigation

### Tenant Switcher

```jsx
import {useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

function TenantSwitcher() {
  const session = useSessionContext();
  const navigate = useNavigate();
  const [tenants, setTenants] = useState([]);
  const [currentTenant, setCurrentTenant] = useState(null);
  
  useEffect(() => {
    if (session.doesSessionExist) {
      fetch(`/api/v1/users/${session.userId}/tenants`, {
        credentials: 'include'
      })
        .then(res => res.json())
        .then(data => setTenants(data.data));
    }
  }, [session]);
  
  const switchTenant = (slug) => {
    setCurrentTenant(slug);
    navigate(`/tenant/${slug}`);
  };
  
  return (
    <div className="tenant-switcher">
      <select 
        value={currentTenant || ''}
        onChange={(e) => switchTenant(e.target.value)}
      >
        <option value="">Select Tenant...</option>
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

### Tenant Context Provider

```jsx
import {createContext, useContext, useState, useEffect} from 'react';
import {useParams} from 'react-router-dom';

const TenantContext = createContext(null);

export function TenantProvider({children}) {
  const {slug} = useParams();
  const [tenant, setTenant] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (slug) {
      fetch(`/api/v1/tenants/slug/${slug}`, {credentials: 'include'})
        .then(res => res.json())
        .then(data => {
          setTenant(data.data);
          setLoading(false);
        })
        .catch(() => setLoading(false));
    }
  }, [slug]);
  
  return (
    <TenantContext.Provider value={{tenant, loading}}>
      {children}
    </TenantContext.Provider>
  );
}

export function useTenant() {
  return useContext(TenantContext);
}

// Usage
function TenantView() {
  return (
    <TenantProvider>
      <TenantDashboard />
    </TenantProvider>
  );
}

function TenantDashboard() {
  const {tenant, loading} = useTenant();
  
  if (loading) return <div>Loading...</div>;
  if (!tenant) return <div>Tenant not found</div>;
  
  return (
    <div>
      <h1>{tenant.name}</h1>
      <p>{tenant.description}</p>
    </div>
  );
}
```

## Permission-Based UI

### Permission Hook

```jsx
import {useState, useEffect} from 'react';

export function usePermission(tenantId, service, entity, action) {
  const [hasPermission, setHasPermission] = useState(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (!tenantId) {
      setLoading(false);
      return;
    }
    
    fetch(
      `/api/v1/authorize?` +
      `tenant_id=${tenantId}&` +
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
  }, [tenantId, service, entity, action]);
  
  return {hasPermission, loading};
}

// Usage
function CreatePostButton({tenantId}) {
  const {hasPermission, loading} = usePermission(
    tenantId,
    'blog-api',
    'post',
    'create'
  );
  
  if (loading) return <button disabled>Loading...</button>;
  if (!hasPermission) return null;  // Hide button
  
  return <button onClick={handleCreate}>Create Post</button>;
}
```

### Permission Context (Optimized)

```jsx
import {createContext, useContext, useEffect, useState} from 'react';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

const PermissionContext = createContext(null);

export function PermissionProvider({children, tenantId}) {
  const session = useSessionContext();
  const [permissions, setPermissions] = useState(new Set());
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist && tenantId) {
      fetch(
        `/api/v1/permissions/user?tenant_id=${tenantId}&user_id=${session.userId}`,
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
  }, [session, tenantId]);
  
  const can = (service, entity, action) => {
    return permissions.has(`${service}:${entity}:${action}`);
  };
  
  return (
    <PermissionContext.Provider value={{can, loading, permissions}}>
      {children}
    </PermissionContext.Provider>
  );
}

export function usePermissions() {
  const context = useContext(PermissionContext);
  if (!context) {
    throw new Error('usePermissions must be used within PermissionProvider');
  }
  return context;
}

// Usage
function TenantApp({tenantId}) {
  return (
    <PermissionProvider tenantId={tenantId}>
      <Dashboard />
    </PermissionProvider>
  );
}

function Dashboard() {
  const {can, loading} = usePermissions();
  
  if (loading) return <div>Loading permissions...</div>;
  
  return (
    <div>
      {can('blog-api', 'post', 'create') && (
        <button>Create Post</button>
      )}
      {can('tenant-api', 'member', 'invite') && (
        <button>Invite Member</button>
      )}
      {can('blog-api', 'post', 'delete') && (
        <button>Delete Posts</button>
      )}
    </div>
  );
}
```

### Can Component

```jsx
function Can({service, entity, action, children, fallback = null}) {
  const {can, loading} = usePermissions();
  
  if (loading) return fallback;
  if (!can(service, entity, action)) return fallback;
  
  return children;
}

// Usage
<Can service="blog-api" entity="post" action="create">
  <CreatePostButton />
</Can>

<Can
  service="tenant-api"
  entity="member"
  action="invite"
  fallback={<div>You don't have permission to invite members</div>}
>
  <InviteMemberForm />
</Can>
```

## Error Handling

### Error Boundary

```jsx
import React from 'react';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {hasError: false, error: null};
  }
  
  static getDerivedStateFromError(error) {
    return {hasError: true, error};
  }
  
  componentDidCatch(error, errorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
    // Log to error tracking service (e.g., Sentry)
  }
  
  render() {
    if (this.state.hasError) {
      return (
        <div className="error-screen">
          <h2>Something went wrong</h2>
          <p>{this.state.error?.message}</p>
          <button onClick={() => window.location.reload()}>
            Reload Page
          </button>
        </div>
      );
    }
    
    return this.props.children;
  }
}

// Usage
<ErrorBoundary>
  <App />
</ErrorBoundary>
```

### API Error Handler

```javascript
export class APIError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }
}

export async function apiCall(endpoint, options = {}) {
  try {
    const response = await fetch(endpoint, {
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      ...options
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new APIError(
        error.error || 'API error',
        response.status,
        error
      );
    }
    
    return response.json();
  } catch (error) {
    if (error instanceof APIError) {
      // Handle specific API errors
      if (error.status === 401) {
        window.location.href = '/auth';
      } else if (error.status === 403) {
        alert('Permission denied');
      } else if (error.status === 404) {
        alert('Resource not found');
      }
    }
    throw error;
  }
}
```

### Toast Notifications

```jsx
import {useState, createContext, useContext} from 'react';

const ToastContext = createContext(null);

export function ToastProvider({children}) {
  const [toasts, setToasts] = useState([]);
  
  const addToast = (message, type = 'info') => {
    const id = Date.now();
    setToasts(prev => [...prev, {id, message, type}]);
    setTimeout(() => removeToast(id), 5000);
  };
  
  const removeToast = (id) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };
  
  return (
    <ToastContext.Provider value={{addToast}}>
      {children}
      <div className="toast-container">
        {toasts.map(toast => (
          <div key={toast.id} className={`toast toast-${toast.type}`}>
            {toast.message}
            <button onClick={() => removeToast(toast.id)}>×</button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  return useContext(ToastContext);
}

// Usage
function CreateTenantForm() {
  const {addToast} = useToast();
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await apiCall('/api/v1/tenants', {
        method: 'POST',
        body: JSON.stringify(formData)
      });
      addToast('Tenant created successfully!', 'success');
    } catch (error) {
      addToast(error.message, 'error');
    }
  };
  
  return <form onSubmit={handleSubmit}>...</form>;
}
```

## Loading States

### Skeleton Loader

```jsx
function Skeleton({width = '100%', height = '20px', className = ''}) {
  return (
    <div
      className={`skeleton ${className}`}
      style={{width, height}}
    />
  );
}

// Usage
function TenantListSkeleton() {
  return (
    <div>
      {[1, 2, 3].map(i => (
        <div key={i} className="tenant-card">
          <Skeleton width="200px" height="24px" />
          <Skeleton width="100%" height="16px" />
          <Skeleton width="80%" height="16px" />
        </div>
      ))}
    </div>
  );
}

function TenantList() {
  const {data, loading} = useAPI('/api/v1/tenants');
  
  if (loading) return <TenantListSkeleton />;
  
  return (
    <div>
      {data.data.map(tenant => (
        <TenantCard key={tenant.id} tenant={tenant} />
      ))}
    </div>
  );
}
```

## Best Practices

### 1. Always Use credentials: 'include'

```javascript
// ✅ Correct
fetch('/api/v1/tenants', {credentials: 'include'})

// ❌ Wrong
fetch('/api/v1/tenants')  // Cookies not sent!
```

### 2. Handle Session Expiry

```javascript
// Interceptor pattern
const originalFetch = window.fetch;
window.fetch = async (...args) => {
  const response = await originalFetch(...args);
  
  if (response.status === 401 && !args[0].includes('/auth')) {
    window.location.href = '/auth';
    throw new Error('Session expired');
  }
  
  return response;
};
```

### 3. Cache User Permissions

```javascript
// Don't check permission on every render
const {hasPermission} = usePermission(tenantId, 'blog-api', 'post', 'create');

// Do fetch all permissions once and cache
const {can} = usePermissions();  // Fetches once, caches
if (can('blog-api', 'post', 'create')) {
  // ...
}
```

### 4. Optimistic UI Updates

```javascript
async function deleteTenant(tenantId) {
  // Optimistically remove from UI
  setTenants(prev => prev.filter(t => t.id !== tenantId));
  
  try {
    await apiCall(`/api/v1/tenants/${tenantId}`, {method: 'DELETE'});
    addToast('Tenant deleted', 'success');
  } catch (error) {
    // Revert on error
    loadTenants();
    addToast('Failed to delete tenant', 'error');
  }
}
```

### 5. Environment Variables

```javascript
// .env
VITE_API_URL=http://localhost:8080
VITE_APP_NAME=My App

// Usage
const API_URL = import.meta.env.VITE_API_URL;
```

## Complete Example App

```jsx
// src/App.jsx
import {BrowserRouter, Routes, Route, Navigate} from 'react-router-dom';
import {SessionAuth} from 'supertokens-auth-react/recipe/session';
import {ToastProvider} from './components/ToastProvider';
import {PermissionProvider} from './components/PermissionProvider';
import Dashboard from './pages/Dashboard';
import TenantView from './pages/TenantView';
import NotFound from './pages/NotFound';

function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
        <Routes>
          {/* Auth routes handled by SuperTokens */}
          
          {/* Protected routes */}
          <Route path="/dashboard" element={
            <SessionAuth>
              <Dashboard />
            </SessionAuth>
          } />
          
          <Route path="/tenant/:slug/*" element={
            <SessionAuth>
              <TenantView />
            </SessionAuth>
          } />
          
          {/* Public routes */}
          <Route path="/" element={<Navigate to="/dashboard" />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </ToastProvider>
    </BrowserRouter>
  );
}

export default App;
```

## Next Steps

- [User Authentication](/guides/user-authentication) - SuperTokens details
- [Session Management](/guides/session-management) - Session handling
- [Backend Integration](/guides/backend-integration) - API implementation
- [Permissions Guide](/guides/permissions) - RBAC system

