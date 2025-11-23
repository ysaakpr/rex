import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function Permissions() {
  const navigate = useNavigate();
  const [permissions, setPermissions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
  const [checkingAdmin, setCheckingAdmin] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    service: '',
    entity: '',
    action: '',
    description: ''
  });

  useEffect(() => {
    checkPlatformAdmin();
  }, []);

  useEffect(() => {
    if (isPlatformAdmin) {
      loadPermissions();
    }
  }, [isPlatformAdmin]);

  const checkPlatformAdmin = async () => {
    try {
      const response = await fetch('/api/v1/platform/admins/check', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setIsPlatformAdmin(data.data?.is_platform_admin || false);
      } else {
        setIsPlatformAdmin(false);
      }
    } catch (err) {
      console.error('Error checking platform admin:', err);
      setIsPlatformAdmin(false);
    } finally {
      setCheckingAdmin(false);
    }
  };

  const loadPermissions = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/platform/permissions', {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load permissions');
      
      const data = await response.json();
      setPermissions(data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePermission = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/v1/platform/permissions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(formData)
      });

      if (!response.ok) throw new Error('Failed to create permission');

      setShowModal(false);
      setFormData({ service: '', entity: '', action: '', description: '' });
      loadPermissions();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeletePermission = async (permissionId) => {
    if (!confirm('Are you sure you want to delete this permission? This will affect all roles using it.')) return;

    try {
      const response = await fetch(`/api/v1/platform/permissions/${permissionId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) throw new Error('Failed to delete permission');

      loadPermissions();
    } catch (err) {
      setError(err.message);
    }
  };

  // Group permissions by service
  const groupedPermissions = permissions.reduce((acc, perm) => {
    if (!acc[perm.service]) {
      acc[perm.service] = [];
    }
    acc[perm.service].push(perm);
    return acc;
  }, {});

  if (checkingAdmin) {
    return (
      <div className="dashboard">
        <div className="loading">🔐 Checking permissions...</div>
      </div>
    );
  }

  if (!isPlatformAdmin) {
    return (
      <div className="dashboard">
        <div className="header">
          <h1>🔐 Access Denied</h1>
        </div>
        <div className="section">
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <h2>Platform Admin Access Required</h2>
            <p>You need platform administrator privileges to manage permissions.</p>
            <button onClick={() => navigate('/')} style={{ marginTop: '1rem' }}>
              Back to Dashboard
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="header">
        <h1>🔑 Permissions Management</h1>
        <div>
          <button onClick={() => navigate('/')}>← Back</button>
          <button onClick={() => setShowModal(true)}>+ Create Permission</button>
        </div>
      </div>

      {error && (
        <div style={{ background: '#fee', color: '#c00', padding: '1rem', borderRadius: '4px', margin: '1rem 0' }}>
          {error}
        </div>
      )}

      <div className="section">
        <h2>Permissions</h2>
        {loading ? (
          <div className="loading">Loading permissions...</div>
        ) : Object.keys(groupedPermissions).length === 0 ? (
          <p>No permissions found. Create your first permission!</p>
        ) : (
          <div>
            {Object.keys(groupedPermissions).sort().map(service => (
              <div key={service} style={{ marginBottom: '2rem' }}>
                <h3 style={{ 
                  background: '#f5f5f5', 
                  padding: '0.75rem', 
                  borderRadius: '4px',
                  marginBottom: '1rem'
                }}>
                  📦 {service}
                </h3>
                <div style={{ 
                  display: 'grid', 
                  gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', 
                  gap: '1rem',
                  paddingLeft: '1rem'
                }}>
                  {groupedPermissions[service].map(permission => (
                    <div 
                      key={permission.id} 
                      style={{ 
                        border: '1px solid #e0e0e0', 
                        padding: '1rem', 
                        borderRadius: '4px',
                        background: '#fff'
                      }}
                    >
                      <div style={{ marginBottom: '0.5rem' }}>
                        <strong style={{ fontFamily: 'monospace', fontSize: '0.9rem' }}>
                          {permission.entity}:{permission.action}
                        </strong>
                      </div>
                      <p style={{ fontSize: '0.85rem', color: '#666', margin: '0.5rem 0' }}>
                        {permission.description}
                      </p>
                      <div style={{ 
                        fontSize: '0.75rem', 
                        color: '#999',
                        marginTop: '0.5rem',
                        paddingTop: '0.5rem',
                        borderTop: '1px solid #f0f0f0'
                      }}>
                        ID: {permission.id.substring(0, 8)}...
                      </div>
                      <div style={{ marginTop: '0.5rem' }}>
                        <button 
                          onClick={() => handleDeletePermission(permission.id)} 
                          className="button secondary" 
                          style={{ 
                            fontSize: '0.85rem',
                            padding: '0.25rem 0.5rem',
                            color: '#d32f2f'
                          }}
                        >
                          🗑️ Delete
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Permission Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>Create New Permission</h2>
            <form onSubmit={handleCreatePermission}>
              <div className="form-group">
                <label>Service:</label>
                <input
                  type="text"
                  placeholder="e.g., tenant-api, platform-api"
                  value={formData.service}
                  onChange={(e) => setFormData({ ...formData, service: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Entity:</label>
                <input
                  type="text"
                  placeholder="e.g., tenant, member, role"
                  value={formData.entity}
                  onChange={(e) => setFormData({ ...formData, entity: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Action:</label>
                <input
                  type="text"
                  placeholder="e.g., create, read, update, delete"
                  value={formData.action}
                  onChange={(e) => setFormData({ ...formData, action: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description:</label>
                <textarea
                  placeholder="Describe what this permission allows"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                  required
                />
              </div>
              <div style={{ 
                background: '#f0f0f0', 
                padding: '0.75rem', 
                borderRadius: '4px',
                marginBottom: '1rem',
                fontSize: '0.9rem'
              }}>
                <strong>Preview:</strong> {formData.service}:{formData.entity}:{formData.action}
              </div>
              <div className="form-actions">
                <button type="button" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit">Create</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default Permissions;
