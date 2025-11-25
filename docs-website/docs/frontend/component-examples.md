# Component Examples

Reusable React component library for common use cases.

## Overview

This page provides ready-to-use React components for common features:
- Tenant management
- Member management
- User profile
- Navigation
- Forms
- Loading states
- Error handling

## Layout Components

### Main Layout

```jsx
// src/components/layout/MainLayout.jsx
import {Outlet, Link, useNavigate} from 'react-router-dom';
import {signOut} from 'supertokens-auth-react/recipe/session';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function MainLayout() {
  const navigate = useNavigate();
  const session = useSessionContext();
  
  const handleSignOut = async () => {
    await signOut();
    navigate('/');
  };
  
  if (session.loading) {
    return <div className="min-h-screen flex items-center justify-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>;
  }
  
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <Link to="/" className="flex items-center">
                <span className="text-xl font-bold text-blue-600">Rex</span>
              </Link>
              {session.doesSessionExist && (
                <div className="ml-10 flex space-x-4">
                  <Link to="/dashboard" className="text-gray-700 hover:text-blue-600 px-3 py-2">
                    Dashboard
                  </Link>
                  <Link to="/tenants" className="text-gray-700 hover:text-blue-600 px-3 py-2">
                    Tenants
                  </Link>
                </div>
              )}
            </div>
            <div className="flex items-center">
              {session.doesSessionExist ? (
                <button
                  onClick={handleSignOut}
                  className="bg-gray-200 hover:bg-gray-300 px-4 py-2 rounded text-sm font-medium"
                >
                  Sign Out
                </button>
              ) : (
                <Link to="/auth" className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded text-sm font-medium">
                  Sign In
                </Link>
              )}
            </div>
          </div>
        </div>
      </nav>
      
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}
```

### Sidebar Layout

```jsx
// src/components/layout/SidebarLayout.jsx
import {NavLink, Outlet} from 'react-router-dom';
import {useParams} from 'react-router-dom';

export default function SidebarLayout() {
  const {tenant_id} = useParams();
  
  const navigation = [
    {name: 'Overview', href: `/tenant/${tenant_id}`, icon: '📊'},
    {name: 'Members', href: `/tenant/${tenant_id}/members`, icon: '👥'},
    {name: 'Settings', href: `/tenant/${tenant_id}/settings`, icon: '⚙️'},
  ];
  
  return (
    <div className="flex h-screen bg-gray-100">
      <aside className="w-64 bg-white shadow">
        <div className="p-4">
          <h2 className="text-lg font-semibold">Tenant Menu</h2>
        </div>
        <nav className="mt-4">
          {navigation.map(item => (
            <NavLink
              key={item.name}
              to={item.href}
              className={({isActive}) =>
                `flex items-center px-4 py-3 text-sm font-medium ${
                  isActive
                    ? 'bg-blue-50 text-blue-700 border-r-4 border-blue-700'
                    : 'text-gray-700 hover:bg-gray-50'
                }`
              }
            >
              <span className="mr-3">{item.icon}</span>
              {item.name}
            </NavLink>
          ))}
        </nav>
      </aside>
      
      <main className="flex-1 overflow-auto p-8">
        <Outlet />
      </main>
    </div>
  );
}
```

## Tenant Components

### Tenant Card

```jsx
// src/components/tenant/TenantCard.jsx
import {Link} from 'react-router-dom';

export default function TenantCard({tenant}) {
  return (
    <Link
      to={`/tenant/${tenant.slug}`}
      className="block p-6 bg-white rounded-lg shadow hover:shadow-lg transition"
    >
      <h3 className="text-xl font-semibold text-gray-900 mb-2">
        {tenant.name}
      </h3>
      {tenant.description && (
        <p className="text-gray-600 mb-4">{tenant.description}</p>
      )}
      <div className="flex items-center justify-between text-sm text-gray-500">
        <span>Role: {tenant.role_name}</span>
        <span className={`px-2 py-1 rounded ${
          tenant.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
        }`}>
          {tenant.status}
        </span>
      </div>
    </Link>
  );
}
```

### Tenant List

```jsx
// src/components/tenant/TenantList.jsx
import {useState, useEffect} from 'react';
import {api} from '../../lib/api';
import TenantCard from './TenantCard';

export default function TenantList() {
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    api.get('/api/v1/tenants')
      .then(response => {
        setTenants(response.data.data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);
  
  if (loading) {
    return <div className="text-center py-12">Loading tenants...</div>;
  }
  
  if (tenants.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">No tenants yet</p>
        <button className="btn-primary">Create Your First Tenant</button>
      </div>
    );
  }
  
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {tenants.map(tenant => (
        <TenantCard key={tenant.id} tenant={tenant} />
      ))}
    </div>
  );
}
```

### Create Tenant Form

```jsx
// src/components/tenant/CreateTenantForm.jsx
import {useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {api} from '../../lib/api';

export default function CreateTenantForm({onSuccess}) {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    description: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const generateSlug = (name) => {
    return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
  };
  
  const handleNameChange = (e) => {
    const name = e.target.value;
    setFormData({
      ...formData,
      name,
      slug: generateSlug(name)
    });
  };
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.post('/api/v1/tenants', formData);
      onSuccess?.(response.data);
      navigate(`/tenant/${response.data.slug}`);
    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-2xl">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded">
          {error}
        </div>
      )}
      
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Tenant Name *
        </label>
        <input
          type="text"
          value={formData.name}
          onChange={handleNameChange}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
          placeholder="Acme Corporation"
        />
      </div>
      
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Slug * (auto-generated)
        </label>
        <input
          type="text"
          value={formData.slug}
          onChange={(e) => setFormData({...formData, slug: e.target.value})}
          required
          pattern="[a-z0-9-]+"
          className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
          placeholder="acme-corporation"
        />
        <p className="mt-1 text-sm text-gray-500">
          Lowercase letters, numbers, and hyphens only
        </p>
      </div>
      
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Description
        </label>
        <textarea
          value={formData.description}
          onChange={(e) => setFormData({...formData, description: e.target.value})}
          rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
          placeholder="Brief description of your organization..."
        />
      </div>
      
      <div className="flex gap-3">
        <button
          type="submit"
          disabled={loading}
          className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? 'Creating...' : 'Create Tenant'}
        </button>
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="px-6 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
```

## Member Components

### Member List

```jsx
// src/components/member/MemberList.jsx
import {useState, useEffect} from 'react';
import {useParams} from 'react-router-dom';
import {api} from '../../lib/api';

export default function MemberList() {
  const {tenant_id} = useParams();
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    loadMembers();
  }, [tenant_id]);
  
  const loadMembers = async () => {
    try {
      const response = await api.get(`/api/v1/tenants/${tenant_id}/members`);
      setMembers(response.data.data);
      setLoading(false);
    } catch (err) {
      setLoading(false);
    }
  };
  
  const handleRoleChange = async (userId, roleId) => {
    try {
      await api.patch(`/api/v1/tenants/${tenant_id}/members/${userId}`, {
        role_id: roleId
      });
      loadMembers();
    } catch (err) {
      alert('Failed to update role');
    }
  };
  
  const handleRemove = async (userId) => {
    if (!confirm('Remove this member?')) return;
    
    try {
      await api.delete(`/api/v1/tenants/${tenant_id}/members/${userId}`);
      setMembers(members.filter(m => m.user_id !== userId));
    } catch (err) {
      alert('Failed to remove member');
    }
  };
  
  if (loading) return <div>Loading members...</div>;
  
  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
              Email
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
              Role
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
              Joined
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {members.map(member => (
            <tr key={member.user_id}>
              <td className="px-6 py-4 whitespace-nowrap">
                {member.email}
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <select
                  value={member.role_id}
                  onChange={(e) => handleRoleChange(member.user_id, e.target.value)}
                  className="border rounded px-2 py-1"
                >
                  <option value="admin-role-id">Admin</option>
                  <option value="writer-role-id">Writer</option>
                  <option value="viewer-role-id">Viewer</option>
                </select>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {new Date(member.created_at).toLocaleDateString()}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                <button
                  onClick={() => handleRemove(member.user_id)}
                  className="text-red-600 hover:text-red-900"
                >
                  Remove
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Invite Member Form

```jsx
// src/components/member/InviteMemberForm.jsx
import {useState, useEffect} from 'react';
import {useParams} from 'react-router-dom';
import {api} from '../../lib/api';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function InviteMemberForm({onSuccess}) {
  const {tenant_id} = useParams();
  const session = useSessionContext();
  const [email, setEmail] = useState('');
  const [roleId, setRoleId] = useState('');
  const [roles, setRoles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    api.get('/api/v1/platform/roles')
      .then(response => setRoles(response.data))
      .catch(() => {});
  }, []);
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    
    try {
      await api.post('/api/v1/invitations', {
        tenant_id,
        email,
        role_id: roleId,
        invited_by: session.userId
      });
      setEmail('');
      setRoleId('');
      onSuccess?.();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded">
          {error}
        </div>
      )}
      
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Email Address
        </label>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
          placeholder="user@example.com"
        />
      </div>
      
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Role
        </label>
        <select
          value={roleId}
          onChange={(e) => setRoleId(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
        >
          <option value="">Select role...</option>
          {roles.map(role => (
            <option key={role.id} value={role.id}>
              {role.name} - {role.description}
            </option>
          ))}
        </select>
      </div>
      
      <button
        type="submit"
        disabled={loading}
        className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
      >
        {loading ? 'Sending Invitation...' : 'Send Invitation'}
      </button>
    </form>
  );
}
```

## User Profile Component

```jsx
// src/components/user/UserProfile.jsx
import {useState, useEffect} from 'react';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';
import {api} from '../../lib/api';

export default function UserProfile() {
  const session = useSessionContext();
  const [user, setUser] = useState(null);
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session.doesSessionExist) {
      Promise.all([
        api.get('/api/v1/users/me'),
        api.get(`/api/v1/users/${session.userId}/tenants`)
      ])
        .then(([userResponse, tenantsResponse]) => {
          setUser(userResponse.data);
          setTenants(tenantsResponse.data);
          setLoading(false);
        })
        .catch(() => setLoading(false));
    }
  }, [session]);
  
  if (loading) return <div>Loading...</div>;
  if (!user) return null;
  
  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4">Profile</h2>
        <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-sm font-medium text-gray-500">Email</dt>
            <dd className="mt-1 text-sm text-gray-900">{user.email}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500">Status</dt>
            <dd className="mt-1">
              <span className={`px-2 py-1 text-xs rounded ${
                user.email_verified
                  ? 'bg-green-100 text-green-800'
                  : 'bg-yellow-100 text-yellow-800'
              }`}>
                {user.email_verified ? 'Verified' : 'Not Verified'}
              </span>
            </dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500">Member Since</dt>
            <dd className="mt-1 text-sm text-gray-900">
              {new Date(user.time_joined).toLocaleDateString()}
            </dd>
          </div>
        </dl>
      </div>
      
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-2xl font-bold mb-4">Your Tenants</h2>
        {tenants.length === 0 ? (
          <p className="text-gray-500">No tenants yet</p>
        ) : (
          <div className="space-y-3">
            {tenants.map(tenant => (
              <div
                key={tenant.tenant_id}
                className="flex items-center justify-between p-4 border rounded hover:bg-gray-50"
              >
                <div>
                  <h3 className="font-medium">{tenant.tenant_name}</h3>
                  <p className="text-sm text-gray-500">Role: {tenant.role_name}</p>
                </div>
                <a
                  href={`/tenant/${tenant.tenant_slug}`}
                  className="text-blue-600 hover:text-blue-800"
                >
                  View →
                </a>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
```

## Loading Components

### Spinner

```jsx
// src/components/common/Spinner.jsx
export default function Spinner({size = 'md'}) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12'
  };
  
  return (
    <div className={`animate-spin rounded-full border-b-2 border-blue-600 ${sizeClasses[size]}`}></div>
  );
}
```

### Skeleton

```jsx
// src/components/common/Skeleton.jsx
export default function Skeleton({width = '100%', height = '20px', className = ''}) {
  return (
    <div
      className={`animate-pulse bg-gray-300 rounded ${className}`}
      style={{width, height}}
    />
  );
}

// Usage
<Skeleton width="200px" height="24px" />
<Skeleton width="100%" height="16px" className="mb-2" />
```

## Error Components

### Error Alert

```jsx
// src/components/common/ErrorAlert.jsx
export default function ErrorAlert({message, onClose}) {
  if (!message) return null;
  
  return (
    <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded relative mb-4">
      <span className="block sm:inline">{message}</span>
      {onClose && (
        <button
          onClick={onClose}
          className="absolute top-0 right-0 px-4 py-3"
        >
          <span className="text-2xl">&times;</span>
        </button>
      )}
    </div>
  );
}
```

### Success Alert

```jsx
// src/components/common/SuccessAlert.jsx
export default function SuccessAlert({message, onClose}) {
  if (!message) return null;
  
  return (
    <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded relative mb-4">
      <span className="block sm:inline">{message}</span>
      {onClose && (
        <button
          onClick={onClose}
          className="absolute top-0 right-0 px-4 py-3"
        >
          <span className="text-2xl">&times;</span>
        </button>
      )}
    </div>
  );
}
```

## Modal Component

```jsx
// src/components/common/Modal.jsx
import {useEffect} from 'react';

export default function Modal({isOpen, onClose, title, children}) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4">
        <div
          className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
          onClick={onClose}
        />
        
        <div className="relative bg-white rounded-lg max-w-lg w-full p-6 z-10">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">{title}</h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <span className="text-2xl">&times;</span>
            </button>
          </div>
          
          <div>{children}</div>
        </div>
      </div>
    </div>
  );
}

// Usage
const [isOpen, setIsOpen] = useState(false);

<Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Create Tenant">
  <CreateTenantForm onSuccess={() => setIsOpen(false)} />
</Modal>
```

## Next Steps

- [React Setup](/frontend/react-setup) - Initial setup
- [Protected Routes](/frontend/protected-routes) - Route guards
- [API Calls](/frontend/api-calls) - API integration
- [Frontend Integration](/guides/frontend-integration) - Complete guide
