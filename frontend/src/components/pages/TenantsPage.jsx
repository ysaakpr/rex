import React, { useState, useEffect } from 'react';
import { Plus, Building2, Users, Calendar, MoreVertical, Edit, Trash2, AlertTriangle, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';

export function TenantsPage() {
  const navigate = useNavigate();
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [tenantToDelete, setTenantToDelete] = useState(null);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadTenants();
  }, []);

  const loadTenants = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await fetch('/api/v1/tenants', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load tenants: ${response.status}`);
      }

      const data = await response.json();
      // Handle both paginated response (data.data.data) and simple array (data.data)
      const tenantsArray = data.data?.data || data.data || [];
      setTenants(Array.isArray(tenantsArray) ? tenantsArray : []);
    } catch (err) {
      console.error('Error loading tenants:', err);
      setError(err.message || 'Failed to load tenants');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTenant = () => {
    navigate('/tenants/create');
  };

  const handleDeleteClick = (tenant) => {
    setTenantToDelete(tenant);
    setShowDeleteDialog(true);
  };

  const handleConfirmDelete = async () => {
    if (!tenantToDelete) return;

    try {
      setDeleting(true);
      
      const response = await fetch(`/api/v1/tenants/${tenantToDelete.id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete tenant');
      }

      // Reload tenants list
      setShowDeleteDialog(false);
      setTenantToDelete(null);
      loadTenants();
    } catch (err) {
      console.error('Error deleting tenant:', err);
      setError('Failed to delete tenant: ' + err.message);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading tenants...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Tenants</h1>
          <p className="text-muted-foreground mt-2">
            Manage and monitor all tenants in your platform
          </p>
        </div>
        <Button onClick={handleCreateTenant} className="gap-2">
          <Plus className="h-4 w-4" />
          Create Tenant
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

      {/* Tenants Grid */}
      {tenants.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted mb-4">
              <Building2 className="h-10 w-10 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-semibold mb-2">No tenants yet</h3>
            <p className="text-sm text-muted-foreground text-center mb-6 max-w-sm">
              Get started by creating your first tenant. Tenants represent organizations or workspaces in your platform.
            </p>
            <Button onClick={handleCreateTenant} className="gap-2">
              <Plus className="h-4 w-4" />
              Create Your First Tenant
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {tenants.map((tenant) => (
            <Card 
              key={tenant.id} 
              className="hover:shadow-lg transition-all cursor-pointer hover:border-primary/50"
              onClick={() => navigate(`/tenants/${tenant.id}`)}
            >
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <CardTitle className="flex items-center gap-2">
                      <Building2 className="h-5 w-5 text-primary" />
                      {tenant.name}
                    </CardTitle>
                    <CardDescription className="mt-1">
                      {tenant.slug}
                    </CardDescription>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button 
                        variant="ghost" 
                        size="icon"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/tenants/${tenant.id}/edit`);
                      }}>
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteClick(tenant);
                        }}
                        className="text-destructive"
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {/* Status Badge */}
                  <div className="flex items-center gap-2">
                    <Badge variant={tenant.status === 'active' ? 'default' : 'secondary'}>
                      {tenant.status || 'active'}
                    </Badge>
                    {tenant.metadata?.industry && (
                      <Badge variant="outline">{tenant.metadata.industry}</Badge>
                    )}
                  </div>

                  {/* Members Count */}
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Users className="h-4 w-4" />
                    <span>{tenant.member_count || 0} members</span>
                  </div>

                  {/* Created Date */}
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Calendar className="h-4 w-4" />
                    <span>Created {new Date(tenant.created_at).toLocaleDateString()}</span>
                  </div>

                  {/* Metadata */}
                  {tenant.metadata?.companySize && (
                    <div className="text-sm text-muted-foreground">
                      Company Size: {tenant.metadata.companySize}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent onClose={() => setShowDeleteDialog(false)}>
          <DialogHeader>
            <div className="flex items-center justify-center mb-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
                <AlertTriangle className="h-8 w-8 text-red-600 dark:text-red-400" />
              </div>
            </div>
            <DialogTitle className="text-center">Delete Tenant?</DialogTitle>
            <DialogDescription className="text-center">
              Are you sure you want to delete <strong>{tenantToDelete?.name}</strong>?
              <br />
              <span className="text-red-600 dark:text-red-400 font-semibold mt-2 block">
                This action cannot be undone.
              </span>
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-md bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-4 my-4">
            <p className="text-sm text-red-800 dark:text-red-200">
              <strong>Warning:</strong> Deleting this tenant will:
            </p>
            <ul className="text-sm text-red-700 dark:text-red-300 mt-2 space-y-1 list-disc list-inside">
              <li>Remove all {tenantToDelete?.member_count || 0} member(s)</li>
              <li>Delete all tenant data</li>
              <li>Revoke all access permissions</li>
            </ul>
          </div>
          <DialogFooter>
            <Button 
              variant="outline" 
              onClick={() => setShowDeleteDialog(false)}
              disabled={deleting}
            >
              Cancel
            </Button>
            <Button 
              variant="destructive" 
              onClick={handleConfirmDelete}
              disabled={deleting}
            >
              {deleting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                <>
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete Tenant
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

