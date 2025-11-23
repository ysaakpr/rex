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
import { Select } from '../ui/select';
import { Alert, AlertDescription, AlertTitle } from '../ui/alert';

export function PolicyDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  console.log('[PolicyDetailsPage] Component mounted with policy ID:', id);
  
  const [policy, setPolicy] = useState(null);
  const [allPermissions, setAllPermissions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showAddPermissionDialog, setShowAddPermissionDialog] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [permissionSearch, setPermissionSearch] = useState('');
  const [selectedService, setSelectedService] = useState('all');
  const [showBulkAddWarning, setShowBulkAddWarning] = useState(false);
  const [addingBulk, setAddingBulk] = useState(false);

  const [editForm, setEditForm] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadPolicyDetails();
    loadAllPermissions();
  }, [id]);

  const loadPolicyDetails = async () => {
    console.log('[PolicyDetailsPage] Loading policy details for:', id);
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch(`/api/v1/platform/policies/${id}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load policy: ${response.status}`);
      }

      const data = await response.json();
      console.log('[PolicyDetailsPage] Policy loaded:', data);
      console.log('[PolicyDetailsPage] Policy permissions:', data.data?.permissions);
      console.log('[PolicyDetailsPage] Permissions count:', data.data?.permissions?.length || 0);
      
      setPolicy(data.data);
      setEditForm({
        name: data.data.name || '',
        description: data.data.description || ''
      });
    } catch (err) {
      console.error('[PolicyDetailsPage] Error loading policy:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadAllPermissions = async () => {
    console.log('[PolicyDetailsPage] Loading all permissions...');
    try {
      const response = await fetch('/api/v1/platform/permissions', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[PolicyDetailsPage] Permissions loaded:', data);
        setAllPermissions(data.data || []);
      }
    } catch (err) {
      console.error('[PolicyDetailsPage] Error loading permissions:', err);
    }
  };

  const handleUpdatePolicy = async () => {
    console.log('[PolicyDetailsPage] Updating policy:', editForm);
    
    if (!editForm.name) {
      setError('Policy name is required');
      return;
    }

    try {
      setUpdating(true);
      setError('');

      const response = await fetch(`/api/v1/platform/policies/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(editForm)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update policy');
      }

      console.log('[PolicyDetailsPage] Policy updated successfully');
      
      setShowEditDialog(false);
      loadPolicyDetails();
    } catch (err) {
      console.error('[PolicyDetailsPage] Error updating policy:', err);
      setError(err.message);
    } finally {
      setUpdating(false);
    }
  };

  const handleDeletePolicy = async () => {
    console.log('[PolicyDetailsPage] Deleting policy:', id);

    try {
      setDeleting(true);
      
      const response = await fetch(`/api/v1/platform/policies/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete policy');
      }

      console.log('[PolicyDetailsPage] Policy deleted successfully');
      navigate('/permissions');
    } catch (err) {
      console.error('[PolicyDetailsPage] Error deleting policy:', err);
      setError(err.message);
    } finally {
      setDeleting(false);
    }
  };

  const handleAddPermission = async (permissionId) => {
    console.log('[PolicyDetailsPage] Adding permission to policy:', permissionId);

    try {
      const response = await fetch(`/api/v1/platform/policies/${id}/permissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ permission_ids: [permissionId] })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to add permission');
      }

      console.log('[PolicyDetailsPage] Permission added successfully');
      
      // Optimistic UI update - add permission to local state instead of full reload
      const addedPermission = allPermissions.find(p => p.id === permissionId);
      if (addedPermission) {
        setPolicy(prev => ({
          ...prev,
          permissions: [...(prev.permissions || []), addedPermission]
        }));
      }
    } catch (err) {
      console.error('[PolicyDetailsPage] Error adding permission:', err);
      setError(err.message);
    }
  };

  const handleRemovePermission = async (permissionId) => {
    console.log('[PolicyDetailsPage] Removing permission from policy:', permissionId);

    try {
      const response = await fetch(`/api/v1/platform/policies/${id}/permissions/${permissionId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to remove permission');
      }

      console.log('[PolicyDetailsPage] Permission removed successfully');
      
      // Optimistic UI update - remove permission from local state
      setPolicy(prev => ({
        ...prev,
        permissions: (prev.permissions || []).filter(p => p.id !== permissionId)
      }));
    } catch (err) {
      console.error('[PolicyDetailsPage] Error removing permission:', err);
      setError(err.message);
    }
  };

  const getAvailablePermissions = () => {
    const policyPermissionIds = new Set((policy?.permissions || []).map(p => p.id));
    const available = allPermissions.filter(p => !policyPermissionIds.has(p.id));
    console.log('[PolicyDetailsPage] Available permissions:', available.length);
    return available;
  };

  const getUniqueServices = () => {
    const available = getAvailablePermissions();
    const services = new Set(available.map(p => p.service).filter(Boolean));
    const serviceList = Array.from(services).sort();
    console.log('[PolicyDetailsPage] Unique services:', serviceList);
    return serviceList;
  };

  const getFilteredPermissions = () => {
    const available = getAvailablePermissions();
    
    const filtered = available.filter(p => {
      // Filter by service
      if (selectedService !== 'all' && p.service !== selectedService) {
        return false;
      }
      // Filter by search
      if (permissionSearch && permissionSearch.trim()) {
        const searchText = (p.key || `${p.service}:${p.entity}:${p.action}`).toLowerCase();
        return searchText.includes(permissionSearch.toLowerCase());
      }
      return true;
    });
    
    console.log('[PolicyDetailsPage] Filtered permissions:', {
      selectedService,
      searchTerm: permissionSearch,
      totalAvailable: available.length,
      filtered: filtered.length
    });
    
    return filtered;
  };

  const filteredAvailablePermissions = getFilteredPermissions();

  const groupPermissionsByService = (permissions) => {
    const grouped = {};
    permissions.forEach(permission => {
      const service = permission.service || 'Unknown';
      if (!grouped[service]) {
        grouped[service] = [];
      }
      grouped[service].push(permission);
    });
    // Sort services alphabetically
    return Object.keys(grouped).sort().reduce((acc, key) => {
      acc[key] = grouped[key];
      return acc;
    }, {});
  };

  const handleBulkAddPermissions = async () => {
    console.log('[PolicyDetailsPage] Bulk adding permissions');
    
    try {
      setAddingBulk(true);
      const permissionIds = filteredAvailablePermissions.map(p => p.id);
      
      const response = await fetch(`/api/v1/platform/policies/${id}/permissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ permission_ids: permissionIds })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to add permissions');
      }

      console.log('[PolicyDetailsPage] Permissions added successfully');
      setShowBulkAddWarning(false);
      
      // Optimistic UI update - add all permissions to local state
      setPolicy(prev => ({
        ...prev,
        permissions: [...(prev.permissions || []), ...filteredAvailablePermissions]
      }));
      
      // Clear search and reset filters
      setPermissionSearch('');
      setSelectedService('all');
    } catch (err) {
      console.error('[PolicyDetailsPage] Error adding permissions:', err);
      setError(err.message);
    } finally {
      setAddingBulk(false);
    }
  };

  const handleAddAllClick = () => {
    // Show warning if adding all permissions without any filter
    if (selectedService === 'all' && !permissionSearch) {
      setShowBulkAddWarning(true);
    } else {
      handleBulkAddPermissions();
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading policy details...</p>
      </div>
    );
  }

  if (!policy) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-destructive">Policy not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/permissions')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2">
              <Shield className="h-6 w-6 text-primary" />
              <h1 className="text-3xl font-bold tracking-tight">{policy.name}</h1>
            </div>
            <p className="text-muted-foreground mt-2">
              {policy.description || 'No description'}
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
            Permissions ({policy.permissions?.length || 0})
          </TabsTrigger>
          <TabsTrigger value="roles">
            Roles ({policy.roles?.length || 0})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Assigned Permissions</CardTitle>
                  <CardDescription>
                    Permissions granted by this policy
                  </CardDescription>
                </div>
                <Button onClick={() => setShowAddPermissionDialog(true)} size="sm" className="gap-2">
                  <Plus className="h-4 w-4" />
                  Add Permission
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {!policy.permissions || !Array.isArray(policy.permissions) || policy.permissions.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Lock className="mx-auto h-12 w-12 mb-4 opacity-50" />
                  <p className="text-sm">No permissions assigned yet</p>
                  <p className="text-xs mt-1">Add permissions to define what this policy can do</p>
                  {policy.permissions && !Array.isArray(policy.permissions) && (
                    <p className="text-xs text-destructive mt-2">
                      Debug: permissions is not an array - {typeof policy.permissions}
                    </p>
                  )}
                </div>
              ) : (
                <div className="space-y-6">
                  {Object.entries(groupPermissionsByService(policy.permissions)).map(([service, permissions]) => (
                    <div key={service}>
                      {/* Service Header */}
                      <div className="flex items-center gap-2 mb-3">
                        <Badge variant="outline" className="font-mono">
                          {service}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {permissions.length} permission{permissions.length !== 1 ? 's' : ''}
                        </span>
                      </div>
                      
                      {/* Permissions in this service - 3 column grid */}
                      <div className="grid grid-cols-3 gap-3">
                        {permissions.map((permission) => {
                          const fullKey = permission.key || `${permission.service}:${permission.entity}:${permission.action}`;
                          return (
                            <div
                              key={permission.id}
                              className="flex items-start justify-between p-3 border rounded-lg bg-card hover:bg-muted/50 transition-colors group"
                            >
                              <div className="flex-1 min-w-0">
                                <p className="font-mono text-sm truncate" title={fullKey}>
                                  {fullKey}
                                </p>
                                {permission.description && (
                                  <p className="text-xs text-muted-foreground mt-1 line-clamp-2" title={permission.description}>
                                    {permission.description}
                                  </p>
                                )}
                              </div>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 ml-2 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                                onClick={() => handleRemovePermission(permission.id)}
                              >
                                <X className="h-4 w-4 text-destructive" />
                              </Button>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="roles">
          <Card>
            <CardHeader>
              <CardTitle>Attached Roles</CardTitle>
              <CardDescription>
                Roles that include this policy
              </CardDescription>
            </CardHeader>
            <CardContent>
              {!policy.roles || policy.roles.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Shield className="mx-auto h-12 w-12 mb-4 opacity-50" />
                  <p className="text-sm">No roles attached yet</p>
                  <p className="text-xs mt-1">Roles are managed from the Roles page</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {policy.roles.map((role) => (
                    <div
                      key={role.id}
                      className="flex items-center justify-between p-3 border rounded-lg hover:bg-accent cursor-pointer"
                      onClick={() => navigate(`/roles/${role.id}`)}
                    >
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <Shield className="h-4 w-4 text-muted-foreground" />
                          <p className="font-medium">{role.name}</p>
                          {role.is_system && (
                            <Badge variant="secondary" className="text-xs">
                              System
                            </Badge>
                          )}
                        </div>
                        {role.description && (
                          <p className="text-xs text-muted-foreground mt-1">{role.description}</p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent onClose={() => setShowEditDialog(false)}>
          <DialogHeader>
            <DialogTitle>Edit Policy</DialogTitle>
            <DialogDescription>
              Update policy information
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Policy Name <span className="text-red-500">*</span></Label>
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
            <Button onClick={handleUpdatePolicy} disabled={updating}>
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
            <DialogTitle className="text-center">Delete Policy?</DialogTitle>
            <DialogDescription className="text-center">
              Are you sure you want to delete <strong>{policy.name}</strong>?
              <br />
              <span className="text-red-600 dark:text-red-400 font-semibold mt-2 block">
                This will remove it from all roles.
              </span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeletePolicy} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Permission Dialog */}
      <Dialog open={showAddPermissionDialog} onOpenChange={setShowAddPermissionDialog}>
        <DialogContent onClose={() => {
          setShowAddPermissionDialog(false);
          setPermissionSearch('');
          setSelectedService('all');
        }} className="max-w-6xl w-[90vw]">
          <DialogHeader>
            <DialogTitle>Add Permissions</DialogTitle>
            <DialogDescription>
              Select permissions to add to this policy
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {/* Filters */}
            <div className="flex gap-3">
              <div className="flex-1">
                <Label>Service / App</Label>
                <Select 
                  value={selectedService} 
                  onChange={(e) => {
                    console.log('[PolicyDetailsPage] Service selected:', e.target.value);
                    setSelectedService(e.target.value);
                  }}
                >
                  <option value="all">All Services</option>
                  {getUniqueServices().map(service => (
                    <option key={service} value={service}>
                      {service}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="flex-1">
                <Label>Search</Label>
                <Input
                  placeholder="Search permissions..."
                  value={permissionSearch}
                  onChange={(e) => setPermissionSearch(e.target.value)}
                />
              </div>
            </div>

            {/* Bulk Add Button */}
            {filteredAvailablePermissions.length > 0 && (
              <div className="flex items-center justify-between p-3 bg-muted/50 rounded-lg">
                <span className="text-sm">
                  <strong>{filteredAvailablePermissions.length}</strong> permission{filteredAvailablePermissions.length !== 1 ? 's' : ''} available
                </span>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={handleAddAllClick}
                  disabled={addingBulk}
                >
                  {addingBulk && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Add All {filteredAvailablePermissions.length > 0 && `(${filteredAvailablePermissions.length})`}
                </Button>
              </div>
            )}

            {/* Warning for adding all permissions without filter */}
            {showBulkAddWarning && (
              <Alert variant="warning">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>Full Access Warning</AlertTitle>
                <AlertDescription>
                  You are about to give full access to this policy by adding all {filteredAvailablePermissions.length} permissions.
                  This grants complete control across all services.
                  <div className="flex gap-2 mt-3">
                    <Button 
                      variant="destructive" 
                      size="sm"
                      onClick={handleBulkAddPermissions}
                      disabled={addingBulk}
                    >
                      {addingBulk && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Yes, Add All
                    </Button>
                    <Button 
                      variant="outline" 
                      size="sm"
                      onClick={() => setShowBulkAddWarning(false)}
                    >
                      Cancel
                    </Button>
                  </div>
                </AlertDescription>
              </Alert>
            )}

            {/* Permission List */}
            <div className="max-h-96 overflow-y-auto space-y-2">
              {filteredAvailablePermissions.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">
                  {getAvailablePermissions().length === 0
                    ? 'All available permissions are already assigned'
                    : 'No permissions found matching your filters'}
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
              setSelectedService('all');
            }}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

