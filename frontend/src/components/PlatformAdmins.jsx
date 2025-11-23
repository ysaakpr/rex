import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function PlatformAdmins() {
  const navigate = useNavigate();
  const [admins, setAdmins] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);
  const [checkingAdmin, setCheckingAdmin] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    user_id: ''
  });

  useEffect(() => {
    checkPlatformAdmin();
  }, []);

  useEffect(() => {
    if (isPlatformAdmin) {
      loadAdmins();
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

  const loadAdmins = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/platform/admins', {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load platform admins');
      
      const data = await response.json();
      setAdmins(data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAdmin = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/v1/platform/admins', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(formData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create platform admin');
      }

      setShowModal(false);
      setFormData({ user_id: '' });
      loadAdmins();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteAdmin = async (userId) => {
    if (!confirm('Are you sure you want to remove this platform admin?')) return;

    try {
      const response = await fetch(`/api/v1/platform/admins/${userId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) throw new Error('Failed to delete platform admin');

      loadAdmins();
    } catch (err) {
      setError(err.message);
    }
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
            <p>You need platform administrator privileges to manage platform admins.</p>
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
        <h1>👑 Platform Admins Management</h1>
        <div>
          <button onClick={() => navigate('/')}>← Back</button>
          <button onClick={() => setShowModal(true)}>+ Add Platform Admin</button>
        </div>
      </div>

      {error && (
        <div style={{ background: '#fee', color: '#c00', padding: '1rem', borderRadius: '4px', margin: '1rem 0' }}>
          {error}
        </div>
      )}

      <div className="section">
        <h2>Platform Administrators</h2>
        <p style={{ color: '#666', marginBottom: '1rem' }}>
          Platform admins have full access to manage roles, permissions, relations, and other platform-level entities.
        </p>
        {loading ? (
          <div className="loading">Loading platform admins...</div>
        ) : admins.length === 0 ? (
          <p>No platform admins found.</p>
        ) : (
          <div className="tenant-list">
            {admins.map(admin => (
              <div key={admin.id} className="tenant-card">
                <h3>User ID: {admin.user_id.substring(0, 16)}...</h3>
                <div className="tenant-meta">
                  <span className="status-badge status-active">ADMIN</span>
                  <span>ID: {admin.id.substring(0, 8)}...</span>
                </div>
                <div style={{ marginTop: '0.5rem', fontSize: '0.9rem', color: '#666' }}>
                  <div>Created: {new Date(admin.created_at).toLocaleString()}</div>
                  {admin.created_by && <div>Created by: {admin.created_by.substring(0, 16)}...</div>}
                </div>
                <div style={{ marginTop: '1rem' }}>
                  <button 
                    onClick={() => handleDeleteAdmin(admin.user_id)} 
                    className="button secondary" 
                    style={{ color: '#d32f2f' }}
                  >
                    🗑️ Remove Admin
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Admin Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h2>Add Platform Admin</h2>
            <form onSubmit={handleCreateAdmin}>
              <div className="form-group">
                <label>User ID:</label>
                <input
                  type="text"
                  placeholder="Enter SuperTokens user ID"
                  value={formData.user_id}
                  onChange={(e) => setFormData({ ...formData, user_id: e.target.value })}
                  required
                />
                <small style={{ color: '#666', display: 'block', marginTop: '0.25rem' }}>
                  Enter the SuperTokens user ID of the user you want to make a platform admin.
                </small>
              </div>
              <div style={{ 
                background: '#fff3cd', 
                padding: '0.75rem', 
                borderRadius: '4px',
                marginBottom: '1rem',
                fontSize: '0.9rem',
                border: '1px solid #ffc107'
              }}>
                <strong>⚠️ Warning:</strong> Platform admins have full access to manage all platform entities.
              </div>
              <div className="form-actions">
                <button type="button" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit">Add Admin</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default PlatformAdmins;

