import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Plus, X, Loader2, Shield, AlertTriangle, Lock } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { Textarea } from '../ui/textarea';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../ui/tabs';

export function RoleDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  console.log('[RoleDetailsPage] Component mounted with role ID:', id);
  
  const [role, setRole] = useState(null);
  const [allPermissions, setAllPermissions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showAddPermissionDialog, setShowAddPermissionDialog] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [permissionSearch, setPermissionSearch] = useState('');

  const [editForm, setEditForm] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadRoleDetails();
    loadAllPermissions();
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
      console.log('[RoleDetailsPage] Role permissions:', data.data?.permissions);
      console.log('[RoleDetailsPage] Permissions count:', data.data?.permissions?.length || 0);
      
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

  const loadAllPermissions = async () => {
    console.log('[RoleDetailsPage] Loading all permissions...');
    try {
      const response = await fetch('/api/v1/platform/permissions', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[RoleDetailsPage] Permissions loaded:', data);
        setAllPermissions(data.data || []);
      }
    } catch (err) {
      console.error('[RoleDetailsPage] Error loading permissions:', err);
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
      
      const response = await fetch(`/api/v1/platform/roles/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete role');
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

  const handleAddPermission = async (permissionId) => {
    console.log('[RoleDetailsPage] Adding permission to role:', permissionId);

    try {
      const response = await fetch(`/api/v1/platform/roles/${id}/permissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ permission_id: permissionId })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to add permission');
      }

      console.log('[RoleDetailsPage] Permission added successfully');
      loadRoleDetails();
    } catch (err) {
      console.error('[RoleDetailsPage] Error adding permission:', err);
      setError(err.message);
    }
  };

  const handleRemovePermission = async (permissionId) => {
    console.log('[RoleDetailsPage] Removing permission from role:', permissionId);

    try {
      const response = await fetch(`/api/v1/platform/roles/${id}/permissions/${permissionId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to remove permission');
      }

      console.log('[RoleDetailsPage] Permission removed successfully');
      loadRoleDetails();
    } catch (err) {
      console.error('[RoleDetailsPage] Error removing permission:', err);
      setError(err.message);
    }
  };

  const getAvailablePermissions = () => {
    const rolePermissionIds = new Set((role?.permissions || []).map(p => p.id));
    return allPermissions.filter(p => !rolePermissionIds.has(p.id));
  };

  const filteredAvailablePermissions = getAvailablePermissions().filter(p =>
    (p.key || `${p.service}:${p.entity}:${p.action}`)?.toLowerCase().includes(permissionSearch.toLowerCase())
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

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-sm text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <Tabs defaultValue="permissions">
        <TabsList>
          <TabsTrigger value="permissions">
            Permissions ({role.permissions?.length || 0})
          </TabsTrigger>
          <TabsTrigger value="relations">
            Relations ({role.relations_count || 0})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Assigned Permissions</CardTitle>
                  <CardDescription>
                    Permissions granted by this role
                  </CardDescription>
                </div>
                <Button onClick={() => setShowAddPermissionDialog(true)} size="sm" className="gap-2">
                  <Plus className="h-4 w-4" />
                  Add Permission
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {!role.permissions || !Array.isArray(role.permissions) || role.permissions.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Lock className="mx-auto h-12 w-12 mb-4 opacity-50" />
                  <p className="text-sm">No permissions assigned yet</p>
                  <p className="text-xs mt-1">Add permissions to define what this role can do</p>
                  {role.permissions && !Array.isArray(role.permissions) && (
                    <p className="text-xs text-destructive mt-2">
                      Debug: permissions is not an array - {typeof role.permissions}
                    </p>
                  )}
                </div>
              ) : (
                <div className="space-y-2">
                  {role.permissions.map((permission) => (
                    <div
                      key={permission.id}
                      className="flex items-center justify-between p-3 border rounded-lg"
                    >
                      <div className="flex-1">
                        <p className="font-mono text-sm">
                          {permission.key || `${permission.service}:${permission.entity}:${permission.action}`}
                        </p>
                        {permission.description && (
                          <p className="text-xs text-muted-foreground mt-1">{permission.description}</p>
                        )}
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemovePermission(permission.id)}
                      >
                        <X className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="relations">
          <Card>
            <CardHeader>
              <CardTitle>Mapped Relations</CardTitle>
              <CardDescription>
                Relations that automatically grant this role
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12 text-muted-foreground">
                <Lock className="mx-auto h-12 w-12 mb-4 opacity-50" />
                <p className="text-sm">Relation mapping managed from Relations page</p>
                <p className="text-xs mt-1">
                  {role.relations_count || 0} {role.relations_count === 1 ? 'relation' : 'relations'} mapped
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent onClose={() => setShowEditDialog(false)}>
          <DialogHeader>
            <DialogTitle>Edit Role</DialogTitle>
            <DialogDescription>
              Update role information
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Role Name <span className="text-red-500">*</span></Label>
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
            <Button onClick={handleUpdateRole} disabled={updating}>
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
            <DialogTitle className="text-center">Delete Role?</DialogTitle>
            <DialogDescription className="text-center">
              Are you sure you want to delete <strong>{role.name}</strong>?
              <br />
              <span className="text-red-600 dark:text-red-400 font-semibold mt-2 block">
                This will remove it from all relations.
              </span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteRole} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Permission Dialog */}
      <Dialog open={showAddPermissionDialog} onOpenChange={setShowAddPermissionDialog}>
        <DialogContent onClose={() => setShowAddPermissionDialog(false)} className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add Permission</DialogTitle>
            <DialogDescription>
              Select permissions to add to this role
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <Input
              placeholder="Search permissions..."
              value={permissionSearch}
              onChange={(e) => setPermissionSearch(e.target.value)}
            />
            <div className="max-h-96 overflow-y-auto space-y-2">
              {filteredAvailablePermissions.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">
                  {getAvailablePermissions().length === 0
                    ? 'All available permissions are already assigned'
                    : 'No permissions found matching your search'}
                </div>
              ) : (
                filteredAvailablePermissions.map((permission) => (
                  <div
                    key={permission.id}
                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50"
                  >
                    <div className="flex-1">
                      <span className="font-mono text-sm">
                        {permission.key || `${permission.service}:${permission.entity}:${permission.action}`}
                      </span>
                      {permission.description && (
                        <p className="text-xs text-muted-foreground mt-1">{permission.description}</p>
                      )}
                    </div>
                    <Button
                      size="sm"
                      onClick={() => {
                        handleAddPermission(permission.id);
                        setShowAddPermissionDialog(false);
                        setPermissionSearch('');
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
              setShowAddPermissionDialog(false);
              setPermissionSearch('');
            }}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

