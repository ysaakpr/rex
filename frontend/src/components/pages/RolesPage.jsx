import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Search, Shield, Loader2, Lock } from 'lucide-react';
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
            Manage roles and their permissions
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          Create Role
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

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="relative">
            <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
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
        </CardContent>
      </Card>

      {/* Roles Grid */}
      {filteredRoles.length === 0 ? (
        <Card>
          <CardContent className="pt-12 pb-12">
            <div className="text-center text-muted-foreground">
              <Shield className="mx-auto h-12 w-12 mb-4 opacity-50" />
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
                    {role.permissions?.length || 0} permissions
                  </Badge>
                </div>
                <CardTitle className="mt-4">{role.name}</CardTitle>
                <CardDescription className="line-clamp-2">
                  {role.description || 'No description'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center text-sm text-muted-foreground gap-2">
                  <Lock className="h-4 w-4" />
                  <span>
                    {role.relations_count || 0} {role.relations_count === 1 ? 'relation' : 'relations'}
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
              Define a new role with permissions
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="role-name">Role Name <span className="text-red-500">*</span></Label>
              <Input
                id="role-name"
                placeholder="e.g., Content Editor"
                value={newRole.name}
                onChange={(e) => setNewRole(prev => ({ ...prev, name: e.target.value }))}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role-description">Description</Label>
              <Textarea
                id="role-description"
                placeholder="Describe what this role can do..."
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
              Create & Configure
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

