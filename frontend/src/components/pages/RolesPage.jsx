import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Search, Users, Loader2, Shield } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { Textarea } from '../ui/textarea';

export function RolesPage() {
  console.log('[RolesPage] Component mounted');
  
  const navigate = useNavigate();
  const [roles, setRoles] = useState([]);
  const [filteredRoles, setFilteredRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [creating, setCreating] = useState(false);

  const [newRole, setNewRole] = useState({
    name: '',
    type: 'tenant',
    description: ''
  });

  useEffect(() => {
    loadRoles();
  }, []);

  useEffect(() => {
    filterRoles();
  }, [roles, searchQuery]);

  const loadRoles = async () => {
    console.log('[RolesPage] Loading roles...');
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch('/api/v1/platform/roles', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load roles: ${response.status}`);
      }

      const data = await response.json();
      console.log('[RolesPage] Roles loaded:', data);
      
      const rolesArray = data.data || [];
      setRoles(Array.isArray(rolesArray) ? rolesArray : []);
    } catch (err) {
      console.error('[RolesPage] Error loading roles:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filterRoles = () => {
    let filtered = [...roles];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(r =>
        r.name?.toLowerCase().includes(query) ||
        r.description?.toLowerCase().includes(query)
      );
    }

    console.log('[RolesPage] Filtered roles:', filtered.length, 'from', roles.length);
    setFilteredRoles(filtered);
  };

  const handleCreateRole = async () => {
    console.log('[RolesPage] Creating role:', newRole);
    
    if (!newRole.name) {
      setError('Role name is required');
      return;
    }

    try {
      setCreating(true);
      setError('');

      const response = await fetch('/api/v1/platform/roles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(newRole)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create role');
      }

      const result = await response.json();
      console.log('[RolesPage] Role created successfully:', result);
      
      setShowCreateDialog(false);
      setNewRole({ name: '', description: '' });
      
      // Navigate to the new role details page
      if (result.data?.id) {
        navigate(`/roles/${result.data.id}`);
      } else {
        loadRoles();
      }
    } catch (err) {
      console.error('[RolesPage] Error creating role:', err);
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading roles...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Roles</h1>
          <p className="text-muted-foreground mt-2">
            Manage user roles in tenants (Admin, Writer, Viewer, etc.)
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search roles..."
              value={searchQuery}
              onChange={(e) => {
                console.log('[RolesPage] Search query changed:', e.target.value);
                setSearchQuery(e.target.value);
              }}
              className="pl-9"
            />
          </div>
          <Button onClick={() => setShowCreateDialog(true)} className="gap-2">
            <Plus className="h-4 w-4" />
            Create Role
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

      {/* Roles Grid */}
      {filteredRoles.length === 0 ? (
        <Card>
          <CardContent className="pt-12 pb-12">
            <div className="text-center text-muted-foreground">
              <Users className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p className="text-sm">No roles found</p>
              <p className="text-xs mt-1">
                {searchQuery ? 'Try adjusting your search' : 'Create your first role to get started'}
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredRoles.map((role) => (
            <Card
              key={role.id}
              className="cursor-pointer hover:border-primary/50 transition-all hover:shadow-md"
              onClick={() => {
                console.log('[RolesPage] Role card clicked:', role.id);
                navigate(`/roles/${role.id}`);
              }}
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <Shield className="h-8 w-8 text-primary" />
                  <Badge variant="outline">
                    {role.policies?.length || 0} {role.policies?.length === 1 ? 'policy' : 'policies'}
                  </Badge>
                </div>
                <CardTitle className="mt-4">{role.name}</CardTitle>
                <CardDescription className="line-clamp-2">
                  {role.description || 'No description'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center text-sm text-muted-foreground gap-2">
                  <Shield className="h-4 w-4" />
                  <span>
                    {role.policies?.length || 0} {role.policies?.length === 1 ? 'policy' : 'policies'} attached
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Role Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent onClose={() => setShowCreateDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create Role</DialogTitle>
            <DialogDescription>
              Define a new user role (e.g., Admin, Writer, Viewer)
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="role-name">Role Name <span className="text-red-500">*</span></Label>
              <Input
                id="role-name"
                placeholder="e.g., Contributor, Moderator"
                value={newRole.name}
                onChange={(e) => setNewRole(prev => ({ ...prev, name: e.target.value }))}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role-type">Type <span className="text-red-500">*</span></Label>
              <select
                id="role-type"
                value={newRole.type}
                onChange={(e) => setNewRole(prev => ({ ...prev, type: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="tenant">Tenant (for use within tenants)</option>
                <option value="platform">Platform (system-level)</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="role-description">Description</Label>
              <Textarea
                id="role-description"
                placeholder="Describe this role's purpose..."
                value={newRole.description}
                onChange={(e) => setNewRole(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={handleCreateRole} disabled={creating}>
              {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Role
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

