import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function Roles() {
  const navigate = useNavigate();
  const [roles, setRoles] = useState([]);
  const [permissions, setPermissions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
  const [checkingAdmin, setCheckingAdmin] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingRole, setEditingRole] = useState(null);
  const [showPermissionsModal, setShowPermissionsModal] = useState(false);
  const [selectedRole, setSelectedRole] = useState(null);
  const [selectedPermissions, setSelectedPermissions] = useState([]);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    permission_ids: []
  });

  useEffect(() => {
    checkPlatformAdmin();
  }, []);

  useEffect(() => {
    if (isPlatformAdmin) {
      loadRoles();
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

  const loadRoles = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/platform/policies', {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load roles');
      
      const data = await response.json();
      setRoles(data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadPermissions = async () => {
    try {
      const response = await fetch('/api/v1/platform/permissions', {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load permissions');
      
      const data = await response.json();
      setPermissions(data.data || []);
    } catch (err) {
      console.error('Error loading permissions:', err);
    }
  };

  const handleCreateRole = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/v1/platform/policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(formData)
      });

      if (!response.ok) throw new Error('Failed to create role');

      setShowModal(false);
      setFormData({ name: '', description: '', permission_ids: [] });
      loadRoles();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleUpdateRole = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`/api/v1/platform/policies/${editingRole.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          name: formData.name,
          description: formData.description
        })
      });

      if (!response.ok) throw new Error('Failed to update role');

      setShowModal(false);
      setEditingRole(null);
      setFormData({ name: '', description: '', permission_ids: [] });
      loadRoles();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteRole = async (roleId) => {
    if (!confirm('Are you sure you want to delete this role?')) return;

    try {
      const response = await fetch(`/api/v1/platform/policies/${roleId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) throw new Error('Failed to delete role');

      loadRoles();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleAssignPermissions = async () => {
    if (!selectedRole) return;

    try {
      const response = await fetch(`/api/v1/platform/policies/${selectedRole.id}/permissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ permission_ids: selectedPermissions })
      });

      if (!response.ok) throw new Error('Failed to assign permissions');

      setShowPermissionsModal(false);
      setSelectedRole(null);
      setSelectedPermissions([]);
      loadRoles();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleRevokePermission = async (roleId, permissionId) => {
    if (!confirm('Are you sure you want to revoke this permission?')) return;

    try {
      const response = await fetch(`/api/v1/platform/policies/${roleId}/permissions/${permissionId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) throw new Error('Failed to revoke permission');

      loadRoles();
    } catch (err) {
      setError(err.message);
    }
  };

  const openEditModal = (role) => {
    setEditingRole(role);
    setFormData({
      name: role.name,
      description: role.description,
      permission_ids: []
    });
    setShowModal(true);
  };

  const openPermissionsModal = (role) => {
    setSelectedRole(role);
    setSelectedPermissions(role.permissions?.map(p => p.id) || []);
    setShowPermissionsModal(true);
  };

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
            <p>You need platform administrator privileges to manage roles.</p>
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
        <h1>🔐 Roles Management</h1>
        <div>
          <button onClick={() => navigate('/')}>← Back</button>
          <button onClick={() => { setShowModal(true); setEditingRole(null); setFormData({ name: '', description: '', permission_ids: [] }); }}>
            + Create Role
          </button>
        </div>
      </div>

      {error && (
        <div style={{ background: '#fee', color: '#c00', padding: '1rem', borderRadius: '4px', margin: '1rem 0' }}>
          {error}
        </div>
      )}

      <div className="section">
        <h2>Roles</h2>
        {loading ? (
          <div className="loading">Loading roles...</div>
        ) : roles.length === 0 ? (
          <p>No roles found. Create your first role!</p>
        ) : (
          <div className="tenant-list">
            {roles.map(role => (
              <div key={role.id} className="tenant-card">
                <h3>{role.name}</h3>
                <p>{role.description}</p>
                <div className="tenant-meta">
                  <span className={role.is_system ? 'status-badge status-active' : 'status-badge status-pending'}>
                    {role.is_system ? 'SYSTEM' : 'CUSTOM'}
                  </span>
                  <span>ID: {role.id.substring(0, 8)}...</span>
                </div>
                {role.permissions && role.permissions.length > 0 && (
                  <div style={{ marginTop: '0.5rem' }}>
                    <strong>Permissions ({role.permissions.length}):</strong>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.25rem', marginTop: '0.25rem' }}>
                      {role.permissions.map(perm => (
                        <span key={perm.id} style={{ 
                          background: '#e3f2fd', 
                          padding: '0.25rem 0.5rem', 
                          borderRadius: '4px', 
                          fontSize: '0.85rem',
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.25rem'
                        }}>
                          {perm.service}:{perm.entity}:{perm.action}
                          <button 
                            onClick={() => handleRevokePermission(role.id, perm.id)}
                            style={{ 
                              background: 'transparent', 
                              border: 'none', 
                              color: '#d32f2f', 
                              cursor: 'pointer',
                              padding: '0 0.25rem',
                              fontSize: '1rem'
                            }}
                            title="Revoke permission"
                          >
                            ×
                          </button>
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
                  <button onClick={() => openPermissionsModal(role)} className="button secondary">
                    🔑 Manage Permissions
                  </button>
                  {!role.is_system && (
                    <>
                      <button onClick={() => openEditModal(role)} className="button secondary">
                        ✏️ Edit
                      </button>
                      <button onClick={() => handleDeleteRole(role.id)} className="button secondary" style={{ color: '#d32f2f' }}>
                        🗑️ Delete
                      </button>
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create/Edit Role Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>{editingRole ? 'Edit Role' : 'Create New Role'}</h2>
            <form onSubmit={editingRole ? handleUpdateRole : handleCreateRole}>
              <div className="form-group">
                <label>Name:</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description:</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="form-actions">
                <button type="button" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit">{editingRole ? 'Update' : 'Create'}</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Assign Permissions Modal */}
      {showPermissionsModal && selectedRole && (
        <div className="modal-overlay" onClick={() => setShowPermissionsModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '600px', maxHeight: '80vh', overflow: 'auto' }}>
            <h2>Assign Permissions to {selectedRole.name}</h2>
            <div style={{ marginBottom: '1rem' }}>
              <strong>Select permissions:</strong>
              <div style={{ maxHeight: '400px', overflow: 'auto', marginTop: '0.5rem' }}>
                {permissions.map(permission => (
                  <label key={permission.id} style={{ display: 'flex', alignItems: 'center', padding: '0.5rem', cursor: 'pointer', borderBottom: '1px solid #eee' }}>
                    <input
                      type="checkbox"
                      checked={selectedPermissions.includes(permission.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedPermissions([...selectedPermissions, permission.id]);
                        } else {
                          setSelectedPermissions(selectedPermissions.filter(id => id !== permission.id));
                        }
                      }}
                      style={{ marginRight: '0.5rem' }}
                    />
                    <span>
                      <strong>{permission.service}:{permission.entity}:{permission.action}</strong>
                      <br />
                      <small style={{ color: '#666' }}>{permission.description}</small>
                    </span>
                  </label>
                ))}
              </div>
            </div>
            <div className="form-actions">
              <button type="button" onClick={() => setShowPermissionsModal(false)}>Cancel</button>
              <button onClick={handleAssignPermissions}>Assign Permissions</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Roles;
