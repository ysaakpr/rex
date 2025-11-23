import React, { useState, useEffect } from 'react';
import { Plus, Search, Trash2, Filter, Loader2, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';

export function PermissionsPage() {
  console.log('[PermissionsPage] Component mounted');
  
  const [permissions, setPermissions] = useState([]);
  const [filteredPermissions, setFilteredPermissions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedApp, setSelectedApp] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [permissionToDelete, setPermissionToDelete] = useState(null);
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const [newPermission, setNewPermission] = useState({
    service: '',
    entity: '',
    action: 'read',
    description: ''
  });

  const apps = ['all', 'platform-api', 'tenant-api', 'user-api'];
  const actions = ['read', 'write', 'delete', 'manage', 'create', 'update'];

  useEffect(() => {
    loadPermissions();
  }, []);

  useEffect(() => {
    filterPermissions();
  }, [permissions, selectedApp, searchQuery]);

  const loadPermissions = async () => {
    console.log('[PermissionsPage] Loading permissions...');
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch('/api/v1/platform/permissions', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load permissions: ${response.status}`);
      }

      const data = await response.json();
      console.log('[PermissionsPage] Permissions loaded:', data);
      
      const permsArray = data.data || [];
      setPermissions(Array.isArray(permsArray) ? permsArray : []);
    } catch (err) {
      console.error('[PermissionsPage] Error loading permissions:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filterPermissions = () => {
    let filtered = [...permissions];

    // Filter by service (what we're calling "app" in the UI)
    if (selectedApp !== 'all') {
      filtered = filtered.filter(p => {
        // Check if service starts with the selected app
        // (e.g., "tenant" service for "tenant-api" app)
        return p.service?.toLowerCase().includes(selectedApp.replace('-api', ''));
      });
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(p =>
        p.key?.toLowerCase().includes(query) ||
        p.service?.toLowerCase().includes(query) ||
        p.entity?.toLowerCase().includes(query) ||
        p.action?.toLowerCase().includes(query)
      );
    }

    console.log('[PermissionsPage] Filtered permissions:', filtered.length, 'from', permissions.length);
    setFilteredPermissions(filtered);
  };

  const parsePermissionName = (name) => {
    if (!name) return { app: '', service: '', entity: '', action: '' };
    const parts = name.split(':');
    return {
      app: parts[0] || '',
      service: parts[1] || '',
      entity: parts[2] || '',
      action: parts[3] || ''
    };
  };

  const handleCreatePermission = async () => {
    console.log('[PermissionsPage] Creating permission:', newPermission);
    
    if (!newPermission.service || !newPermission.entity || !newPermission.action) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      setCreating(true);
      setError('');

      // Backend expects service, entity, action fields (not a name field)
      const response = await fetch('/api/v1/platform/permissions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          service: newPermission.service,
          entity: newPermission.entity,
          action: newPermission.action,
          description: newPermission.description || ''
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create permission');
      }

      console.log('[PermissionsPage] Permission created successfully');
      
      setShowCreateDialog(false);
      setNewPermission({
        service: '',
        entity: '',
        action: 'read',
        description: ''
      });
      loadPermissions();
    } catch (err) {
      console.error('[PermissionsPage] Error creating permission:', err);
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  const handleDeletePermission = async () => {
    if (!permissionToDelete) return;

    console.log('[PermissionsPage] Deleting permission:', permissionToDelete.id);

    try {
      setDeleting(true);
      
      const response = await fetch(`/api/v1/platform/permissions/${permissionToDelete.id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete permission');
      }

      console.log('[PermissionsPage] Permission deleted successfully');
      
      setShowDeleteDialog(false);
      setPermissionToDelete(null);
      loadPermissions();
    } catch (err) {
      console.error('[PermissionsPage] Error deleting permission:', err);
      setError(err.message);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading permissions...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Permissions</h1>
          <p className="text-muted-foreground mt-2">
            Manage platform and tenant permissions
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          Create Permission
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-sm text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="app-filter">Filter by App</Label>
              <select
                id="app-filter"
                value={selectedApp}
                onChange={(e) => {
                  console.log('[PermissionsPage] App filter changed:', e.target.value);
                  setSelectedApp(e.target.value);
                }}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {apps.map(app => (
                  <option key={app} value={app}>
                    {app === 'all' ? 'All Apps' : app}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="search">Search</Label>
              <div className="relative">
                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="search"
                  placeholder="Search permissions..."
                  value={searchQuery}
                  onChange={(e) => {
                    console.log('[PermissionsPage] Search query changed:', e.target.value);
                    setSearchQuery(e.target.value);
                  }}
                  className="pl-9"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Permissions List */}
      <Card>
        <CardHeader>
          <CardTitle>
            Permissions ({filteredPermissions.length})
          </CardTitle>
          <CardDescription>
            {selectedApp !== 'all' && `Showing permissions for ${selectedApp}`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredPermissions.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <p className="text-sm">No permissions found</p>
              <p className="text-xs mt-1">Try adjusting your filters or create a new permission</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {filteredPermissions.map((permission) => {
                return (
                  <div
                    key={permission.id}
                    className="flex flex-col p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex-1 space-y-2">
                      <div className="flex items-start justify-between gap-2">
                        <p className="font-mono text-sm font-medium break-all">
                          {permission.key || `${permission.service}:${permission.entity}:${permission.action}`}
                        </p>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 flex-shrink-0"
                          onClick={() => {
                            console.log('[PermissionsPage] Delete clicked for:', permission.id);
                            setPermissionToDelete(permission);
                            setShowDeleteDialog(true);
                          }}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                      {permission.description && (
                        <p className="text-xs text-muted-foreground">{permission.description}</p>
                      )}
                      <div className="flex flex-wrap gap-2 mt-2">
                        <Badge variant="outline" className="text-xs">Service: {permission.service}</Badge>
                        <Badge variant="outline" className="text-xs">Entity: {permission.entity}</Badge>
                        <Badge className="text-xs">{permission.action}</Badge>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Permission Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent onClose={() => setShowCreateDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create Permission</DialogTitle>
            <DialogDescription>
              Define a new permission for your platform
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="service">Service <span className="text-red-500">*</span></Label>
              <Input
                id="service"
                placeholder="e.g., tenant, user, rbac"
                value={newPermission.service}
                onChange={(e) => setNewPermission(prev => ({ ...prev, service: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">The service this permission belongs to</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="entity">Entity <span className="text-red-500">*</span></Label>
              <Input
                id="entity"
                placeholder="e.g., member, role, permission"
                value={newPermission.entity}
                onChange={(e) => setNewPermission(prev => ({ ...prev, entity: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">The entity being acted upon</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="action">Action <span className="text-red-500">*</span></Label>
              <select
                id="action"
                value={newPermission.action}
                onChange={(e) => setNewPermission(prev => ({ ...prev, action: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {actions.map(action => (
                  <option key={action} value={action}>{action}</option>
                ))}
              </select>
              <p className="text-xs text-muted-foreground">The action being performed</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="e.g., Allows creating new members"
                value={newPermission.description}
                onChange={(e) => setNewPermission(prev => ({ ...prev, description: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">Optional description of what this permission allows</p>
            </div>
            <div className="rounded-md bg-muted p-3">
              <p className="text-sm font-medium">Permission Key:</p>
              <p className="text-sm font-mono mt-1">
                {newPermission.service || '...'}:{newPermission.entity || '...'}:{newPermission.action}
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={handleCreatePermission} disabled={creating}>
              {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent onClose={() => setShowDeleteDialog(false)}>
          <DialogHeader>
            <div className="flex items-center justify-center mb-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
                <AlertTriangle className="h-8 w-8 text-red-600 dark:text-red-400" />
              </div>
            </div>
            <DialogTitle className="text-center">Delete Permission?</DialogTitle>
            <DialogDescription className="text-center">
              Are you sure you want to delete this permission?
              <br />
              <strong className="font-mono">
                {permissionToDelete?.key || `${permissionToDelete?.service}:${permissionToDelete?.entity}:${permissionToDelete?.action}`}
              </strong>
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
            <Button variant="destructive" onClick={handleDeletePermission} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

