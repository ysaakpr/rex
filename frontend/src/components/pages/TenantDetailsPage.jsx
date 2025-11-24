import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Building2, Users, Calendar, Edit, Trash2, AlertTriangle, Loader2 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Separator } from '../ui/separator';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { TenantUserManagement } from './TenantUserManagement';

export function TenantDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [tenant, setTenant] = useState(null);
  const [memberCount, setMemberCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    checkPlatformAdmin();
  }, []);

  useEffect(() => {
    if (isPlatformAdmin !== null) {
      loadTenantDetails();
    }
  }, [id, isPlatformAdmin]);

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
      console.error('Error checking platform admin status:', err);
      setIsPlatformAdmin(false);
    }
  };

  const loadTenantDetails = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Use platform admin endpoint if user is a platform admin
      const endpoint = isPlatformAdmin 
        ? `/api/v1/platform/tenants/${id}`
        : `/api/v1/tenants/${id}`;
      
      const response = await fetch(endpoint, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to load tenant details');
      }

      const data = await response.json();
      setTenant(data.data);
      setMemberCount(data.data.member_count || 0);
    } catch (err) {
      console.error('Error loading tenant:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      setDeleting(true);
      
      const response = await fetch(`/api/v1/tenants/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to delete tenant');
      }

      // Success - navigate back to tenants list
      navigate('/tenants');
    } catch (err) {
      console.error('Error deleting tenant:', err);
      setError(err.message);
      setShowDeleteDialog(false);
    } finally {
      setDeleting(false);
    }
  };

  const handleMembersUpdate = () => {
    // Reload tenant to get updated member count
    loadTenantDetails();
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading tenant details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-8">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-red-500">
              <p className="font-semibold">Error loading tenant</p>
              <p className="text-sm mt-2">{error}</p>
              <Button className="mt-4" onClick={() => navigate('/tenants')}>
                Back to Tenants
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!tenant) {
    return (
      <div className="container mx-auto py-8">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-muted-foreground">
              <Building2 className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p className="font-semibold">Tenant not found</p>
              <Button className="mt-4" onClick={() => navigate('/tenants')}>
                Back to Tenants
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      {/* Header */}
      <div className="mb-6">
        <Button 
          variant="ghost" 
          className="mb-4" 
          onClick={() => navigate('/tenants')}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Tenants
        </Button>
        
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl font-bold">{tenant.name}</h1>
            <p className="text-muted-foreground mt-1">Slug: {tenant.slug}</p>
          </div>
          <div className="flex gap-2">
            <Button 
              variant="outline" 
              onClick={() => navigate(`/tenants/${id}/edit`)}
            >
              <Edit className="mr-2 h-4 w-4" />
              Edit
            </Button>
            <Button 
              variant="destructive" 
              onClick={() => setShowDeleteDialog(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          </div>
        </div>
      </div>

      <Separator className="mb-6" />

      {/* Overview Cards */}
      <div className="grid gap-6 md:grid-cols-3 mb-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Members</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{memberCount}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Active users in this tenant
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Created</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {new Date(tenant.created_at).toLocaleDateString()}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {new Date(tenant.created_at).toLocaleDateString('en-US', { 
                month: 'long', 
                day: 'numeric', 
                year: 'numeric' 
              })}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Industry</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {tenant.metadata?.industry ? (
                <Badge variant="secondary">{tenant.metadata.industry}</Badge>
              ) : (
                <span className="text-muted-foreground text-base">Not set</span>
              )}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Company Size: {tenant.metadata?.companySize || 'Not set'}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Tenant Details Card */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Tenant Information</CardTitle>
          <CardDescription>Detailed information about this tenant</CardDescription>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Tenant ID</dt>
              <dd className="mt-1 text-sm font-mono">{tenant.id}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Tenant Name</dt>
              <dd className="mt-1 text-sm">{tenant.name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Slug</dt>
              <dd className="mt-1 text-sm font-mono">{tenant.slug}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Created At</dt>
              <dd className="mt-1 text-sm">
                {new Date(tenant.created_at).toLocaleString()}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Last Updated</dt>
              <dd className="mt-1 text-sm">
                {new Date(tenant.updated_at).toLocaleString()}
              </dd>
            </div>
            {tenant.metadata?.notes && (
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-muted-foreground">Internal Notes</dt>
                <dd className="mt-1 text-sm">{tenant.metadata.notes}</dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>

      {/* User Management Widget */}
      <TenantUserManagement tenantId={id} onMembersUpdate={handleMembersUpdate} />

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
              Are you sure you want to delete <strong>{tenant.name}</strong>?
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
              <li>Remove all {memberCount} member(s)</li>
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
              onClick={handleDelete}
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
