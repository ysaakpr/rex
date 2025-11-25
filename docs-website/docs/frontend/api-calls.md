# API Calls

Complete guide to making API calls from React frontend.

## Overview

This guide covers:
- API helper functions
- Authentication with SuperTokens
- Error handling
- Loading states
- Custom hooks
- Request/response patterns

## Basic API Call

### Always Include Credentials

```javascript
// ✅ Correct: Include credentials for cookies
fetch('/api/v1/tenants', {
  credentials: 'include',  // Send cookies
  headers: {
    'Content-Type': 'application/json'
  }
});

// ❌ Wrong: Cookies not sent
fetch('/api/v1/tenants');
```

## API Helper Function

### Create Reusable Helper

```javascript
// src/lib/api.js

export class APIError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }
}

export async function apiCall(endpoint, options = {}) {
  const config = {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  };
  
  try {
    const response = await fetch(
      `${import.meta.env.VITE_API_URL || ''}${endpoint}`,
      config
    );
    
    // Handle authentication errors
    if (response.status === 401) {
      window.location.href = '/auth?redirect=' + window.location.pathname;
      throw new APIError('Session expired', 401);
    }
    
    // Parse response
    const data = await response.json();
    
    // Handle API errors
    if (!response.ok) {
      throw new APIError(
        data.error || 'API error',
        response.status,
        data
      );
    }
    
    return data;
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    throw new APIError('Network error', 0, error);
  }
}

// Convenience methods
export const api = {
  get: (endpoint) => apiCall(endpoint, {method: 'GET'}),
  
  post: (endpoint, body) => apiCall(endpoint, {
    method: 'POST',
    body: JSON.stringify(body)
  }),
  
  put: (endpoint, body) => apiCall(endpoint, {
    method: 'PUT',
    body: JSON.stringify(body)
  }),
  
  patch: (endpoint, body) => apiCall(endpoint, {
    method: 'PATCH',
    body: JSON.stringify(body)
  }),
  
  delete: (endpoint) => apiCall(endpoint, {method: 'DELETE'})
};
```

### Usage Examples

```javascript
import {api} from './lib/api';

// GET request
const tenants = await api.get('/api/v1/tenants');

// POST request
const newTenant = await api.post('/api/v1/tenants', {
  name: 'My Tenant',
  slug: 'my-tenant'
});

// PATCH request
const updated = await api.patch(`/api/v1/tenants/${id}`, {
  name: 'Updated Name'
});

// DELETE request
await api.delete(`/api/v1/tenants/${id}`);
```

## Custom Hooks

### useAPI Hook

```javascript
// src/hooks/useAPI.js
import {useState, useEffect} from 'react';
import {apiCall} from '../lib/api';

export function useAPI(endpoint, options = {}) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    let cancelled = false;
    
    apiCall(endpoint, options)
      .then(result => {
        if (!cancelled) {
          setData(result.data);
          setLoading(false);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setError(err.message);
          setLoading(false);
        }
      });
    
    return () => {
      cancelled = true;
    };
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

### useMutation Hook

```javascript
// src/hooks/useMutation.js
import {useState} from 'react';
import {apiCall} from '../lib/api';

export function useMutation(endpoint, options = {}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [data, setData] = useState(null);
  
  const mutate = async (body) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await apiCall(endpoint, {
        ...options,
        body: JSON.stringify(body)
      });
      setData(result.data);
      setLoading(false);
      return result.data;
    } catch (err) {
      setError(err.message);
      setLoading(false);
      throw err;
    }
  };
  
  return {mutate, loading, error, data};
}

// Usage
function CreateTenantForm() {
  const {mutate: createTenant, loading, error} = useMutation(
    '/api/v1/tenants',
    {method: 'POST'}
  );
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const tenant = await createTenant({
        name: 'New Tenant',
        slug: 'new-tenant'
      });
      console.log('Created:', tenant);
    } catch (error) {
      console.error('Failed:', error);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <button disabled={loading}>
        {loading ? 'Creating...' : 'Create Tenant'}
      </button>
      {error && <div className="error">{error}</div>}
    </form>
  );
}
```

## Complete CRUD Example

### Tenant Management Component

```jsx
import {useState, useEffect} from 'react';
import {api} from '../lib/api';

export default function TenantManager() {
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // Load tenants
  useEffect(() => {
    loadTenants();
  }, []);
  
  const loadTenants = async () => {
    try {
      const response = await api.get('/api/v1/tenants');
      setTenants(response.data.data);
      setLoading(false);
    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  };
  
  // Create tenant
  const createTenant = async (data) => {
    try {
      await api.post('/api/v1/tenants', data);
      loadTenants(); // Reload list
    } catch (err) {
      alert('Failed to create tenant: ' + err.message);
    }
  };
  
  // Update tenant
  const updateTenant = async (id, data) => {
    try {
      await api.patch(`/api/v1/tenants/${id}`, data);
      loadTenants();
    } catch (err) {
      alert('Failed to update tenant: ' + err.message);
    }
  };
  
  // Delete tenant
  const deleteTenant = async (id) => {
    if (!confirm('Are you sure?')) return;
    
    try {
      await api.delete(`/api/v1/tenants/${id}`);
      setTenants(tenants.filter(t => t.id !== id)); // Optimistic update
    } catch (err) {
      alert('Failed to delete tenant: ' + err.message);
      loadTenants(); // Reload on error
    }
  };
  
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  
  return (
    <div>
      <h2>Tenants</h2>
      <button onClick={() => createTenant({
        name: 'New Tenant',
        slug: 'new-tenant'
      })}>
        Add Tenant
      </button>
      
      <ul>
        {tenants.map(tenant => (
          <li key={tenant.id}>
            {tenant.name}
            <button onClick={() => updateTenant(tenant.id, {
              name: tenant.name + ' (Updated)'
            })}>
              Update
            </button>
            <button onClick={() => deleteTenant(tenant.id)}>
              Delete
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

## Pagination

### Paginated List Component

```jsx
function TenantListPaginated() {
  const [tenants, setTenants] = useState([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    loadPage(page);
  }, [page]);
  
  const loadPage = async (pageNum) => {
    setLoading(true);
    try {
      const response = await api.get(
        `/api/v1/tenants?page=${pageNum}&page_size=20`
      );
      setTenants(response.data.data);
      setTotalPages(Math.ceil(response.data.total_count / 20));
      setLoading(false);
    } catch (err) {
      setLoading(false);
    }
  };
  
  return (
    <div>
      {loading && <div>Loading...</div>}
      
      <ul>
        {tenants.map(tenant => (
          <li key={tenant.id}>{tenant.name}</li>
        ))}
      </ul>
      
      <div className="pagination">
        <button
          disabled={page === 1}
          onClick={() => setPage(page - 1)}
        >
          Previous
        </button>
        <span>Page {page} of {totalPages}</span>
        <button
          disabled={page === totalPages}
          onClick={() => setPage(page + 1)}
        >
          Next
        </button>
      </div>
    </div>
  );
}
```

## File Uploads

### Upload with FormData

```javascript
// Upload file
async function uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch('/api/v1/upload', {
    method: 'POST',
    credentials: 'include',
    body: formData  // Don't set Content-Type header
  });
  
  return response.json();
}

// Usage in component
function FileUpload() {
  const handleFileChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    
    try {
      const result = await uploadFile(file);
      console.log('Uploaded:', result.data.url);
    } catch (error) {
      alert('Upload failed');
    }
  };
  
  return (
    <input
      type="file"
      onChange={handleFileChange}
    />
  );
}
```

## Error Handling

### Global Error Handler

```jsx
// src/contexts/ErrorContext.jsx
import {createContext, useContext, useState} from 'react';

const ErrorContext = createContext(null);

export function ErrorProvider({children}) {
  const [error, setError] = useState(null);
  
  const showError = (message) => {
    setError(message);
    setTimeout(() => setError(null), 5000);
  };
  
  return (
    <ErrorContext.Provider value={{showError}}>
      {children}
      {error && (
        <div className="fixed top-4 right-4 bg-red-500 text-white p-4 rounded">
          {error}
        </div>
      )}
    </ErrorContext.Provider>
  );
}

export function useError() {
  return useContext(ErrorContext);
}

// Usage
function MyComponent() {
  const {showError} = useError();
  
  const handleAction = async () => {
    try {
      await api.post('/api/v1/tenants', data);
    } catch (error) {
      showError(error.message);
    }
  };
}
```

## Request Cancellation

### AbortController

```javascript
function SearchComponent() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  
  useEffect(() => {
    const controller = new AbortController();
    
    if (query.length > 2) {
      fetch(`/api/v1/search?q=${query}`, {
        credentials: 'include',
        signal: controller.signal
      })
        .then(res => res.json())
        .then(data => setResults(data.data))
        .catch(err => {
          if (err.name !== 'AbortError') {
            console.error(err);
          }
        });
    }
    
    return () => controller.abort();
  }, [query]);
  
  return (
    <div>
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search..."
      />
      <ul>
        {results.map(r => <li key={r.id}>{r.name}</li>)}
      </ul>
    </div>
  );
}
```

## Retry Logic

### Auto-Retry Failed Requests

```javascript
async function apiCallWithRetry(endpoint, options = {}, maxRetries = 3) {
  let lastError;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await apiCall(endpoint, options);
    } catch (error) {
      lastError = error;
      
      // Don't retry client errors (4xx)
      if (error.status >= 400 && error.status < 500) {
        throw error;
      }
      
      // Wait before retrying (exponential backoff)
      if (i < maxRetries - 1) {
        await new Promise(resolve =>
          setTimeout(resolve, Math.pow(2, i) * 1000)
        );
      }
    }
  }
  
  throw lastError;
}
```

## Caching

### Simple Cache Implementation

```javascript
const cache = new Map();

export async function apiCallWithCache(endpoint, options = {}, ttl = 60000) {
  const cacheKey = endpoint + JSON.stringify(options);
  const cached = cache.get(cacheKey);
  
  if (cached && Date.now() - cached.timestamp < ttl) {
    return cached.data;
  }
  
  const data = await apiCall(endpoint, options);
  cache.set(cacheKey, {data, timestamp: Date.now()});
  
  return data;
}

// Clear cache
export function clearCache() {
  cache.clear();
}
```

## Best Practices

1. **Always include `credentials: 'include'`** for authenticated requests
2. **Handle errors gracefully** with try/catch
3. **Show loading states** for better UX
4. **Use custom hooks** for reusable logic
5. **Implement retry logic** for transient failures
6. **Cache responses** when appropriate
7. **Cancel pending requests** when component unmounts
8. **Use optimistic updates** for better perceived performance

## Testing

### Mock API Calls

```javascript
// src/__mocks__/api.js
export const api = {
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  patch: jest.fn(),
  delete: jest.fn()
};

// In test
import {api} from '../lib/api';

test('loads tenants', async () => {
  api.get.mockResolvedValue({
    data: {
      data: [{id: '1', name: 'Test Tenant'}]
    }
  });
  
  render(<TenantList />);
  
  await waitFor(() => {
    expect(screen.getByText('Test Tenant')).toBeInTheDocument();
  });
});
```

## Next Steps

- [React Setup](/frontend/react-setup) - Initial setup
- [Protected Routes](/frontend/protected-routes) - Route guards
- [Component Examples](/frontend/component-examples) - UI components
- [Frontend Integration](/guides/frontend-integration) - Complete guide
