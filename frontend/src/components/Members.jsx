import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

function Members() {
  const { tenantId } = useParams();
  const navigate = useNavigate();
  const [members, setMembers] = useState([]);
  const [relations, setRelations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedMember, setSelectedMember] = useState(null);
  const [newMember, setNewMember] = useState({ user_id: '', role_id: '' });

  useEffect(() => {
    loadMembers();
    loadRelations();
  }, [tenantId]);

  const loadMembers = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/tenants/${tenantId}/members`, {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load members');
      
      const data = await response.json();
      // API returns { success: true, data: { data: [...], page, total_count, ... } }
      setMembers(data.data?.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadRelations = async () => {
    try {
      const response = await fetch(`/api/v1/roles`, {
        credentials: 'include'
      });
      
      if (!response.ok) throw new Error('Failed to load relations');
      
      const data = await response.json();
      // API returns { success: true, data: [...] }
      setRelations(data.data || []);
    } catch (err) {
      console.error('Failed to load relations:', err);
    }
  };

  const handleAddMember = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(newMember)
      });

      if (!response.ok) throw new Error('Failed to add member');

      setShowAddModal(false);
      setNewMember({ user_id: '', role_id: '' });
      loadMembers();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleUpdateMember = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members/${selectedMember.user_id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ role_id: selectedMember.role_id })
      });

      if (!response.ok) throw new Error('Failed to update member');

      setShowEditModal(false);
      setSelectedMember(null);
      loadMembers();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteMember = async (memberUserId) => {
    if (!confirm('Are you sure you want to remove this member?')) return;

    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members/${memberUserId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) throw new Error('Failed to delete member');

      loadMembers();
    } catch (err) {
      setError(err.message);
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'active': return 'status-badge status-active';
      case 'pending': return 'status-badge status-pending';
      case 'inactive': return 'status-badge status-inactive';
      default: return 'status-badge';
    }
  };

  if (loading) return <div className="loading">Loading members...</div>;

  return (
    <div className="members-container">
      <div className="page-header">
        <div>
          <button onClick={() => navigate('/dashboard')} className="btn-back">
            ← Back to Dashboard
          </button>
          <h1>Team Members</h1>
          <p className="subtitle">Manage users and their access levels</p>
        </div>
        <button onClick={() => setShowAddModal(true)} className="btn-primary">
          + Add Member
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      <div className="members-table">
        <table>
          <thead>
            <tr>
              <th>User ID</th>
              <th>Relation</th>
              <th>Status</th>
              <th>Roles</th>
              <th>Joined</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {members.length === 0 ? (
              <tr>
                <td colSpan="6" className="text-center">No members found</td>
              </tr>
            ) : (
              members.map(member => (
                <tr key={member.id}>
                  <td>
                    <code>{member.user_id.substring(0, 8)}...</code>
                  </td>
                  <td>
                    <span className="relation-badge">
                      {member.relation?.name || 'N/A'}
                    </span>
                  </td>
                  <td>
                    <span className={getStatusBadgeClass(member.status)}>
                      {member.status}
                    </span>
                  </td>
                  <td>
                    {member.roles?.length > 0 ? (
                      <div className="roles-list">
                        {member.roles.map(role => (
                          <span key={role.id} className="role-tag">{role.name}</span>
                        ))}
                      </div>
                    ) : (
                      <span className="text-muted">No roles</span>
                    )}
                  </td>
                  <td>{new Date(member.joined_at).toLocaleDateString()}</td>
                  <td>
                    <div className="action-buttons">
                      <button
                        onClick={() => {
                          setSelectedMember(member);
                          setShowEditModal(true);
                        }}
                        className="btn-small btn-edit"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteMember(member.user_id)}
                        className="btn-small btn-delete"
                      >
                        Remove
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Add Member Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Add Team Member</h2>
            <form onSubmit={handleAddMember}>
              <div className="form-group">
                <label>User ID *</label>
                <input
                  type="text"
                  value={newMember.user_id}
                  onChange={(e) => setNewMember({ ...newMember, user_id: e.target.value })}
                  placeholder="Enter SuperTokens user ID"
                  required
                />
                <small>The user must already exist in SuperTokens</small>
              </div>
              <div className="form-group">
                <label>Relation *</label>
                <select
                  value={newMember.role_id}
                  onChange={(e) => setNewMember({ ...newMember, role_id: e.target.value })}
                  required
                >
                  <option value="">Select relation...</option>
                  {relations.map(relation => (
                    <option key={relation.id} value={relation.id}>
                      {relation.name} - {relation.description}
                    </option>
                  ))}
                </select>
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowAddModal(false)} className="btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Add Member
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Member Modal */}
      {showEditModal && selectedMember && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Edit Member</h2>
            <form onSubmit={handleUpdateMember}>
              <div className="form-group">
                <label>User ID</label>
                <input
                  type="text"
                  value={selectedMember.user_id}
                  disabled
                  className="input-disabled"
                />
              </div>
              <div className="form-group">
                <label>Relation *</label>
                <select
                  value={selectedMember.role_id}
                  onChange={(e) => setSelectedMember({ ...selectedMember, role_id: e.target.value })}
                  required
                >
                  {relations.map(relation => (
                    <option key={relation.id} value={relation.id}>
                      {relation.name} - {relation.description}
                    </option>
                  ))}
                </select>
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowEditModal(false)} className="btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Update Member
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default Members;

