import React, { useState, useEffect } from 'react';
import { signOut } from "supertokens-auth-react/recipe/emailpassword";
import { useNavigate } from 'react-router-dom';
import Session from "supertokens-auth-react/recipe/session";

function Dashboard() {
  const navigate = useNavigate();
  const [userInfo, setUserInfo] = useState(null);
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showMenu, setShowMenu] = useState(null);
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    industry: ''
  });
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    loadUserInfo();
    loadTenants();
    checkPlatformAdmin();
  }, []);

  const loadUserInfo = async () => {
    try {
      const userId = await Session.getUserId();
      setUserInfo({ userId });
    } catch (err) {
      console.error('Error loading user info:', err);
    }
  };

  const checkPlatformAdmin = async () => {
    try {
      const response = await fetch('/api/v1/platform/admins/check', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setIsPlatformAdmin(data.data?.is_platform_admin || false);
      }
    } catch (err) {
      console.error('Error checking platform admin:', err);
    }
  };

  const loadTenants = async () => {
    try {
      setLoading(true);
      // Wait a moment for session refresh to complete
      await new Promise(resolve => setTimeout(resolve, 100));
      
      const response = await fetch('/api/v1/tenants', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      if (data.success && data.data) {
        setTenants(data.data.data || []);
      }
    } catch (err) {
      console.error('Error loading tenants:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));

    // Auto-generate slug from name
    if (name === 'name') {
      const slug = value
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-|-$/g, '');
      setFormData(prev => ({
        ...prev,
        slug: slug
      }));
    }
  };

  const handleCreateTenant = async (e) => {
    e.preventDefault();
    setCreating(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/tenants', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          name: formData.name,
          slug: formData.slug,
          metadata: {
            industry: formData.industry
          }
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to create tenant');
      }

      if (data.success) {
        // Reset form
        setFormData({ name: '', slug: '', industry: '' });
        // Reload tenants
        await loadTenants();
        alert('Tenant created successfully!');
      }
    } catch (err) {
      console.error('Error creating tenant:', err);
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  const handleSignOut = async () => {
    await signOut();
    window.location.href = "/auth";
  };

  return (
    <div className="dashboard">
      <div className="header">
        <div>
          <h1>🏢 Tenant Management Dashboard</h1>
          <div className="quick-links">
            {isPlatformAdmin && (
              <>
                <span style={{ color: '#ff9800', fontWeight: 'bold', marginRight: '1rem' }}>👑 Platform Admin</span>
                <button onClick={() => navigate('/platform/onboard-tenant')} className="link-button">
                  🏢 Onboard Tenant
                </button>
                <button onClick={() => navigate('/platform/admins')} className="link-button">
                  👥 Platform Admins
                </button>
              </>
            )}
            <button onClick={() => navigate('/roles')} className="link-button">
              🔐 Manage Roles
            </button>
            <button onClick={() => navigate('/permissions')} className="link-button">
              🔑 Manage Permissions
            </button>
          </div>
        </div>
        <div className="user-info">
          {userInfo && <span>User ID: {userInfo.userId.substring(0, 8)}...</span>}
          <button onClick={handleSignOut}>Sign Out</button>
        </div>
      </div>

      {error && (
        <div className="error">
          <strong>Error:</strong> {error}
        </div>
      )}

      <div className="section">
        <h2>Create New Tenant</h2>
        <form className="tenant-form" onSubmit={handleCreateTenant}>
          <div className="form-group">
            <label htmlFor="name">Tenant Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="e.g., Acme Corporation"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="slug">Slug *</label>
            <input
              type="text"
              id="slug"
              name="slug"
              value={formData.slug}
              onChange={handleInputChange}
              placeholder="e.g., acme-corporation"
              pattern="[a-z0-9-]+"
              required
            />
            <small>Lowercase letters, numbers, and hyphens only</small>
          </div>

          <div className="form-group">
            <label htmlFor="industry">Industry</label>
            <input
              type="text"
              id="industry"
              name="industry"
              value={formData.industry}
              onChange={handleInputChange}
              placeholder="e.g., Technology, Healthcare, Finance"
            />
          </div>

          <button type="submit" disabled={creating}>
            {creating ? 'Creating...' : 'Create Tenant'}
          </button>
        </form>
      </div>

      <div className="section">
        <h2>Your Tenants</h2>
        {loading ? (
          <div className="loading">Loading tenants...</div>
        ) : tenants.length === 0 ? (
          <p>No tenants yet. Create your first tenant above!</p>
        ) : (
          <div className="tenant-list">
            {tenants.map(tenant => (
              <div key={tenant.id} className="tenant-card">
                <div className="tenant-card-header">
                  <h3>{tenant.name}</h3>
                  <div className="tenant-menu">
                    <button 
                      className="menu-button"
                      onClick={() => setShowMenu(showMenu === tenant.id ? null : tenant.id)}
                    >
                      ⋮
                    </button>
                    {showMenu === tenant.id && (
                      <div className="dropdown-menu">
                        <button onClick={() => {
                          navigate(`/tenants/${tenant.id}/members`);
                          setShowMenu(null);
                        }}>
                          👥 Manage Members
                        </button>
                        <button onClick={() => setShowMenu(null)}>
                          ⚙️ Settings
                        </button>
                      </div>
                    )}
                  </div>
                </div>
                <p><strong>Slug:</strong> {tenant.slug}</p>
                <div className="tenant-meta">
                  <span className={`status-badge status-${tenant.status}`}>
                    {tenant.status.toUpperCase()}
                  </span>
                  <span>ID: {tenant.id.substring(0, 8)}...</span>
                  <span>Created: {new Date(tenant.created_at).toLocaleDateString()}</span>
                </div>
                {tenant.metadata && tenant.metadata.industry && (
                  <p style={{ marginTop: '0.5rem', color: '#888' }}>
                    Industry: {tenant.metadata.industry}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default Dashboard;

