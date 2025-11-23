import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Plus, X, Loader2, Shield, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { Textarea } from '../ui/textarea';

export function RoleDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  console.log('[RoleDetailsPage] Component mounted with role ID:', id);
  
  const [role, setRole] = useState(null);
  const [allPolicies, setAllPolicies] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showAttachPolicyDialog, setShowAttachPolicyDialog] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [policySearch, setPolicySearch] = useState('');

  const [editForm, setEditForm] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadRoleDetails();
    loadAllPolicies();
  }, [id]);

  const loadRoleDetails = async () => {
    console.log('[RoleDetailsPage] Loading role details for:', id);
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch(`/api/v1/platform/roles/${id}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load role: ${response.status}`);
      }

      const data = await response.json();
      console.log('[RoleDetailsPage] Role loaded:', data);
      
      setRole(data.data);
      setEditForm({
        name: data.data.name || '',
        description: data.data.description || ''
      });
    } catch (err) {
      console.error('[RoleDetailsPage] Error loading role:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadAllPolicies = async () => {
    console.log('[RoleDetailsPage] Loading all policies...');
    try {
      const response = await fetch('/api/v1/platform/policies', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[RoleDetailsPage] Policies loaded:', data);
        setAllPolicies(data.data || []);
      }
    } catch (err) {
      console.error('[RoleDetailsPage] Error loading policies:', err);
    }
  };

  const handleUpdateRole = async () => {
    console.log('[RoleDetailsPage] Updating role:', editForm);
    
    if (!editForm.name) {
      setError('Role name is required');
      return;
    }

    try {
      setUpdating(true);
      setError('');

      const response = await fetch(`/api/v1/platform/roles/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(editForm)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update role');
      }

      console.log('[RoleDetailsPage] Role updated successfully');
      setShowEditDialog(false);
      loadRoleDetails();
    } catch (err) {
      console.error('[RoleDetailsPage] Error updating role:', err);
      setError(err.message);
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteRole = async () => {
    console.log('[RoleDetailsPage] Deleting role:', id);
    
    try {
      setDeleting(true);
      setError('');

      const response = await fetch(`/api/v1/platform/roles/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete role');
      }

      console.log('[RoleDetailsPage] Role deleted successfully');
      navigate('/roles');
    } catch (err) {
      console.error('[RoleDetailsPage] Error deleting role:', err);
      setError(err.message);
    } finally {
      setDeleting(false);
    }
  };

  const handleAttachPolicy = async (policyId) => {
    console.log('[RoleDetailsPage] Attaching policy to role:', policyId);

    try {
      // Backend expects policy_ids as an array
      const response = await fetch(`/api/v1/platform/roles/${id}/policies`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ policy_ids: [policyId] })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to attach policy');
      }

      console.log('[RoleDetailsPage] Policy attached successfully');
      setShowAttachPolicyDialog(false);
      setPolicySearch('');
      loadRoleDetails();
    } catch (err) {
      console.error('[RoleDetailsPage] Error attaching policy:', err);
      setError(err.message);
    }
  };

  const handleDetachPolicy = async (policyId) => {
    console.log('[RoleDetailsPage] Detaching policy from role:', policyId);

    try {
      const response = await fetch(`/api/v1/platform/roles/${id}/policies/${policyId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to detach policy');
      }

      console.log('[RoleDetailsPage] Policy detached successfully');
      loadRoleDetails();
    } catch (err) {
      console.error('[RoleDetailsPage] Error detaching policy:', err);
      setError(err.message);
    }
  };

  const getAvailablePolicies = () => {
    const rolePolicyIds = new Set((role?.policies || []).map(p => p.id));
    return allPolicies.filter(p => !rolePolicyIds.has(p.id));
  };

  const filteredAvailablePolicies = getAvailablePolicies().filter(p =>
    p.name?.toLowerCase().includes(policySearch.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading role details...</p>
      </div>
    );
  }

  if (!role) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-destructive">Role not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/roles')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2">
              <Shield className="h-6 w-6 text-primary" />
              <h1 className="text-3xl font-bold tracking-tight">{role.name}</h1>
              {role.is_system && (
                <Badge variant="secondary">System</Badge>
              )}
            </div>
            <p className="text-muted-foreground mt-2">
              {role.description || 'No description'}
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

      {/* Error Display */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <AlertTriangle className="h-5 w-5 text-destructive mt-0.5" />
              <div>
                <p className="font-medium text-destructive">Error</p>
                <p className="text-sm text-destructive/80">{error}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Attached Policies Section */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Attached Policies</CardTitle>
              <CardDescription>
                Policies that grant permissions to this role
              </CardDescription>
            </div>
            <Button onClick={() => setShowAttachPolicyDialog(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Attach Policy
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {role.policies && role.policies.length > 0 ? (
            <div className="space-y-2">
              {role.policies.map(policy => (
                <div
                  key={policy.id}
                  className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Shield className="h-4 w-4 text-muted-foreground" />
                      <h3 className="font-medium">{policy.name}</h3>
                      {policy.is_system && (
                        <Badge variant="secondary" className="text-xs">System</Badge>
                      )}
                    </div>
                    {policy.description && (
                      <p className="text-sm text-muted-foreground mt-1 ml-6">
                        {policy.description}
                      </p>
                    )}
                    {policy.permissions && policy.permissions.length > 0 && (
                      <p className="text-xs text-muted-foreground mt-2 ml-6">
                        {policy.permissions.length} permission(s)
                      </p>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDetachPolicy(policy.id)}
                    className="text-destructive hover:text-destructive hover:bg-destructive/10"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Shield className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>No policies attached</p>
              <p className="text-sm mt-1">Click "Attach Policy" to add policies to this role</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Role</DialogTitle>
            <DialogDescription>
              Update the role name and description
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Role Name</Label>
              <Input
                id="name"
                value={editForm.name}
                onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                placeholder="Enter role name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={editForm.description}
                onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                placeholder="Enter role description"
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateRole} disabled={updating}>
              {updating ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Updating...
                </>
              ) : (
                'Update Role'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Role</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this role? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 flex items-start gap-3">
            <AlertTriangle className="h-5 w-5 text-destructive mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-destructive">Warning</p>
              <p className="text-sm text-muted-foreground mt-1">
                Deleting this role will affect all users currently assigned to it.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteRole} disabled={deleting}>
              {deleting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Role'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Attach Policy Dialog */}
      <Dialog open={showAttachPolicyDialog} onOpenChange={setShowAttachPolicyDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Attach Policy</DialogTitle>
            <DialogDescription>
              Select a policy to attach to this role
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="policy-search">Search Policies</Label>
              <Input
                id="policy-search"
                placeholder="Search by policy name..."
                value={policySearch}
                onChange={(e) => setPolicySearch(e.target.value)}
              />
            </div>
            <div className="border rounded-lg max-h-[400px] overflow-y-auto">
              {filteredAvailablePolicies.length > 0 ? (
                <div className="divide-y">
                  {filteredAvailablePolicies.map(policy => (
                    <button
                      key={policy.id}
                      onClick={() => handleAttachPolicy(policy.id)}
                      className="w-full p-4 text-left hover:bg-accent transition-colors"
                    >
                      <div className="flex items-center gap-2">
                        <Shield className="h-4 w-4 text-muted-foreground" />
                        <h3 className="font-medium">{policy.name}</h3>
                        {policy.is_system && (
                          <Badge variant="secondary" className="text-xs">System</Badge>
                        )}
                      </div>
                      {policy.description && (
                        <p className="text-sm text-muted-foreground mt-1 ml-6">
                          {policy.description}
                        </p>
                      )}
                    </button>
                  ))}
                </div>
              ) : (
                <div className="p-8 text-center text-muted-foreground">
                  <Shield className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p>No available policies found</p>
                  <p className="text-sm mt-1">All policies are already attached to this role</p>
                </div>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setShowAttachPolicyDialog(false);
              setPolicySearch('');
            }}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
