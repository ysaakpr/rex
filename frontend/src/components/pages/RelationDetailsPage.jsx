import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Plus, X, Loader2, Users, AlertTriangle, Shield } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { Textarea } from '../ui/textarea';

export function RelationDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  console.log('[RelationDetailsPage] Component mounted with relation ID:', id);
  
  const [relation, setRelation] = useState(null);
  const [allRoles, setAllRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showAddRoleDialog, setShowAddRoleDialog] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [roleSearch, setRoleSearch] = useState('');

  const [editForm, setEditForm] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadRelationDetails();
    loadAllRoles();
  }, [id]);

  const loadRelationDetails = async () => {
    console.log('[RelationDetailsPage] Loading relation details for:', id);
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch(`/api/v1/platform/relations/${id}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load relation: ${response.status}`);
      }

      const data = await response.json();
      console.log('[RelationDetailsPage] Relation loaded:', data);
      
      setRelation(data.data);
      setEditForm({
        name: data.data.name || '',
        description: data.data.description || ''
      });
    } catch (err) {
      console.error('[RelationDetailsPage] Error loading relation:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadAllRoles = async () => {
    console.log('[RelationDetailsPage] Loading all roles...');
    try {
      const response = await fetch('/api/v1/platform/roles', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[RelationDetailsPage] Roles loaded:', data);
        setAllRoles(data.data || []);
      }
    } catch (err) {
      console.error('[RelationDetailsPage] Error loading roles:', err);
    }
  };

  const handleUpdateRelation = async () => {
    console.log('[RelationDetailsPage] Updating relation:', editForm);
    
    if (!editForm.name) {
      setError('Relation name is required');
      return;
    }

    try {
      setUpdating(true);
      setError('');

      const response = await fetch(`/api/v1/platform/relations/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(editForm)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update relation');
      }

      console.log('[RelationDetailsPage] Relation updated successfully');
      
      setShowEditDialog(false);
      loadRelationDetails();
    } catch (err) {
      console.error('[RelationDetailsPage] Error updating relation:', err);
      setError(err.message);
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteRelation = async () => {
    console.log('[RelationDetailsPage] Deleting relation:', id);

    try {
      setDeleting(true);
      
      const response = await fetch(`/api/v1/platform/relations/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete relation');
      }

      console.log('[RelationDetailsPage] Relation deleted successfully');
      navigate('/relations');
    } catch (err) {
      console.error('[RelationDetailsPage] Error deleting relation:', err);
      setError(err.message);
    } finally {
      setDeleting(false);
    }
  };

  const handleAddRole = async (roleId) => {
    console.log('[RelationDetailsPage] Adding role to relation:', roleId);

    try {
      // Backend expects role_ids as an array
      const response = await fetch(`/api/v1/platform/relations/${id}/roles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ role_ids: [roleId] })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to add role');
      }

      console.log('[RelationDetailsPage] Role added successfully');
      loadRelationDetails();
    } catch (err) {
      console.error('[RelationDetailsPage] Error adding role:', err);
      setError(err.message);
    }
  };

  const handleRemoveRole = async (roleId) => {
    console.log('[RelationDetailsPage] Removing role from relation:', roleId);

    try {
      const response = await fetch(`/api/v1/platform/relations/${id}/roles/${roleId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to remove role');
      }

      console.log('[RelationDetailsPage] Role removed successfully');
      loadRelationDetails();
    } catch (err) {
      console.error('[RelationDetailsPage] Error removing role:', err);
      setError(err.message);
    }
  };

  const getAvailableRoles = () => {
    const relationRoleIds = new Set((relation?.roles || []).map(r => r.id));
    return allRoles.filter(r => !relationRoleIds.has(r.id));
  };

  const filteredAvailableRoles = getAvailableRoles().filter(r =>
    r.name?.toLowerCase().includes(roleSearch.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading relation details...</p>
      </div>
    );
  }

  if (!relation) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-destructive">Relation not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/relations')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2">
              <Users className="h-6 w-6 text-primary" />
              <h1 className="text-3xl font-bold tracking-tight">{relation.name}</h1>
            </div>
            <p className="text-muted-foreground mt-2">
              {relation.description || 'No description'}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setShowEditDialog(true)}>
            <Edit className="h-4 w-4 mr-2" />
            Edit
          </Button>
          <Button variant="destructive" onClick={() => setShowDeleteDialog(true)}>
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-sm text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Assigned Roles */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Assigned Roles</CardTitle>
              <CardDescription>
                Roles automatically granted to members with this relation
              </CardDescription>
            </div>
            <Button onClick={() => setShowAddRoleDialog(true)} size="sm" className="gap-2">
              <Plus className="h-4 w-4" />
              Add Role
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {!relation.roles || relation.roles.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Shield className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p className="text-sm">No roles assigned yet</p>
              <p className="text-xs mt-1">Add roles to define permissions for this relation</p>
            </div>
          ) : (
            <div className="space-y-2">
              {relation.roles.map((role) => (
                <div
                  key={role.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Shield className="h-4 w-4 text-primary" />
                      <span className="font-medium">{role.name}</span>
                      <Badge variant="outline">
                        {role.permissions?.length || 0} permissions
                      </Badge>
                    </div>
                    {role.description && (
                      <p className="text-sm text-muted-foreground mt-1 ml-6">
                        {role.description}
                      </p>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveRole(role.id)}
                  >
                    <X className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent onClose={() => setShowEditDialog(false)}>
          <DialogHeader>
            <DialogTitle>Edit Relation</DialogTitle>
            <DialogDescription>
              Update relation information
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Relation Name <span className="text-red-500">*</span></Label>
              <Input
                id="edit-name"
                value={editForm.name}
                onChange={(e) => setEditForm(prev => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-description">Description</Label>
              <Textarea
                id="edit-description"
                value={editForm.description}
                onChange={(e) => setEditForm(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditDialog(false)} disabled={updating}>
              Cancel
            </Button>
            <Button onClick={handleUpdateRelation} disabled={updating}>
              {updating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent onClose={() => setShowDeleteDialog(false)}>
          <DialogHeader>
            <div className="flex items-center justify-center mb-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
                <AlertTriangle className="h-8 w-8 text-red-600 dark:text-red-400" />
              </div>
            </div>
            <DialogTitle className="text-center">Delete Relation?</DialogTitle>
            <DialogDescription className="text-center">
              Are you sure you want to delete <strong>{relation.name}</strong>?
              <br />
              <span className="text-red-600 dark:text-red-400 font-semibold mt-2 block">
                This may affect tenant members using this relation.
              </span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteRelation} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Role Dialog */}
      <Dialog open={showAddRoleDialog} onOpenChange={setShowAddRoleDialog}>
        <DialogContent onClose={() => setShowAddRoleDialog(false)} className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add Role</DialogTitle>
            <DialogDescription>
              Select roles to assign to this relation
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <Input
              placeholder="Search roles..."
              value={roleSearch}
              onChange={(e) => setRoleSearch(e.target.value)}
            />
            <div className="max-h-96 overflow-y-auto space-y-2">
              {filteredAvailableRoles.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">
                  {getAvailableRoles().length === 0
                    ? 'All available roles are already assigned'
                    : 'No roles found matching your search'}
                </div>
              ) : (
                filteredAvailableRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <Shield className="h-4 w-4 text-primary" />
                        <span className="font-medium">{role.name}</span>
                        <Badge variant="outline" className="text-xs">
                          {role.permissions?.length || 0} perms
                        </Badge>
                      </div>
                      {role.description && (
                        <p className="text-xs text-muted-foreground mt-1 ml-6">
                          {role.description}
                        </p>
                      )}
                    </div>
                    <Button
                      size="sm"
                      onClick={() => {
                        handleAddRole(role.id);
                        setShowAddRoleDialog(false);
                        setRoleSearch('');
                      }}
                    >
                      Add
                    </Button>
                  </div>
                ))
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setShowAddRoleDialog(false);
              setRoleSearch('');
            }}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

