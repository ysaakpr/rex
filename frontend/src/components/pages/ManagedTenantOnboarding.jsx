import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Building2, Loader2, CheckCircle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Textarea } from '../ui/textarea';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';

export function ManagedTenantOnboarding() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [relations, setRelations] = useState([]);
  const [showSuccessDialog, setShowSuccessDialog] = useState(false);
  const [createdTenantId, setCreatedTenantId] = useState(null);
  const [createdTenantName, setCreatedTenantName] = useState('');

  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    ownerEmail: '',
    relationId: '',
    metadata: {
      industry: '',
      companySize: '',
      notes: ''
    }
  });
  const [emailCheckLoading, setEmailCheckLoading] = useState(false);
  const [userExists, setUserExists] = useState(null);
  const [userDetails, setUserDetails] = useState(null);

  useEffect(() => {
    loadRelations();
  }, []);

  // Debounced email check
  useEffect(() => {
    if (!formData.ownerEmail || !formData.ownerEmail.includes('@')) {
      setUserExists(null);
      setUserDetails(null);
      return;
    }

    const timeoutId = setTimeout(() => {
      checkEmailExists(formData.ownerEmail);
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [formData.ownerEmail]);

  const loadRelations = async () => {
    try {
      const response = await fetch('/api/v1/relations', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to load relations');
      }

      const data = await response.json();
      setRelations(data.data || []);

      // Auto-select 'Admin' relation if available
      if (data.data) {
        const adminRelation = data.data.find(r => r.name.toLowerCase() === 'admin');
        if (adminRelation) {
          setFormData(prev => ({ ...prev, relationId: adminRelation.id }));
        }
      }
    } catch (err) {
      console.error('Error loading relations:', err);
      setError('Failed to load relations. Please refresh the page.');
    }
  };

  const checkEmailExists = async (email) => {
    try {
      setEmailCheckLoading(true);
      setUserExists(null);
      setUserDetails(null);

      const response = await fetch(`/api/auth/emailpassword/email/exists?email=${encodeURIComponent(email)}`, {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setUserExists(data.exists);
        
        if (data.exists) {
          // User exists, show their info
          setUserDetails({
            email: email,
            status: 'registered'
          });
        }
      }
    } catch (err) {
      console.error('Error checking email:', err);
    } finally {
      setEmailCheckLoading(false);
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => {
      const newState = { ...prev, [name]: value };
      // Auto-generate slug from name
      if (name === 'name') {
        newState.slug = value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-*|-*$/g, '');
      }
      return newState;
    });
  };

  const handleMetadataChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      metadata: {
        ...prev.metadata,
        [field]: value
      }
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    if (!formData.name || !formData.slug || !formData.ownerEmail) {
      setError('Please fill in all required fields (Tenant Name, Slug, and Owner Email).');
      setLoading(false);
      return;
    }

    // Ensure Admin relation is selected
    if (!formData.relationId) {
      setError('Unable to determine Admin role. Please refresh the page and try again.');
      setLoading(false);
      return;
    }

    try {
      // Step 1: Create Tenant
      const tenantPayload = {
        name: formData.name,
        slug: formData.slug,
        metadata: {}
      };

      if (formData.metadata.industry) tenantPayload.metadata.industry = formData.metadata.industry;
      if (formData.metadata.companySize) tenantPayload.metadata.companySize = formData.metadata.companySize;
      if (formData.metadata.notes) tenantPayload.metadata.notes = formData.metadata.notes;

      const tenantResponse = await fetch('/api/v1/tenants', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(tenantPayload)
      });

      if (!tenantResponse.ok) {
        const errorData = await tenantResponse.json();
        throw new Error(errorData.error || 'Failed to create tenant');
      }

      const tenantData = await tenantResponse.json();
      const tenantId = tenantData.data.id;
      
      // Store tenant info for success dialog
      setCreatedTenantId(tenantId);
      setCreatedTenantName(formData.name);

      // Step 2: Always send invitation (works for both existing and new users)
      const invitationPayload = {
        email: formData.ownerEmail,
        relation_id: formData.relationId
      };

      const inviteResponse = await fetch(`/api/v1/tenants/${tenantId}/invitations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(invitationPayload)
      });

      if (!inviteResponse.ok) {
        console.error('Failed to send invitation');
        setError(`Tenant created but failed to send invitation to ${formData.ownerEmail}`);
      }

      // Show success dialog
      setShowSuccessDialog(true);
    } catch (err) {
      console.error('Error creating tenant:', err);
      setError(err.message || 'Failed to create tenant');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" onClick={() => navigate('/tenants')} size="icon">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Building2 className="h-8 w-8 text-primary" />
            Create New Tenant
          </h1>
          <p className="text-muted-foreground mt-1">
            Set up a new tenant and assign the initial owner
          </p>
        </div>
      </div>

      {/* Form */}
      <Card>
        <CardHeader>
          <CardTitle>Tenant Details</CardTitle>
          <CardDescription>Provide information about the new tenant</CardDescription>
        </CardHeader>
        <CardContent>
          {/* Success/Error Messages */}
          {success && (
            <div className="mb-4 rounded-md bg-green-500/10 border border-green-500/20 p-3 text-sm text-green-600">
              {success}
            </div>
          )}
          {error && (
            <div className="mb-4 rounded-md bg-red-500/10 border border-red-500/20 p-3 text-sm text-red-600">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Basic Info */}
            <div className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="name">
                    Tenant Name <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleInputChange}
                    placeholder="e.g., Acme Corporation"
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="slug">
                    Tenant Slug <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="slug"
                    name="slug"
                    value={formData.slug}
                    onChange={handleInputChange}
                    placeholder="e.g., acme-corporation"
                    pattern="[a-z0-9-]+"
                    required
                  />
                </div>
              </div>
            </div>

            {/* Owner Info */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Owner Information</h3>
              <p className="text-sm text-muted-foreground">
                The owner will be assigned as <strong>Admin</strong> and will have full access to manage the tenant.
              </p>
              
              <div className="space-y-2">
                <Label htmlFor="ownerEmail">
                  Owner Email <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="ownerEmail"
                  name="ownerEmail"
                  type="email"
                  value={formData.ownerEmail}
                  onChange={handleInputChange}
                  placeholder="owner@example.com"
                  required
                />
                {emailCheckLoading && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Loader2 className="h-3 w-3 animate-spin" />
                    <span>Checking email...</span>
                  </div>
                )}
                {!emailCheckLoading && userExists === true && (
                  <div className="rounded-md bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 p-3">
                    <p className="text-sm text-blue-600 dark:text-blue-400 font-medium">
                      ✓ User already registered
                    </p>
                    <p className="text-xs text-blue-600 dark:text-blue-400 mt-1">
                      An invitation will be sent to this existing user
                    </p>
                  </div>
                )}
                {!emailCheckLoading && userExists === false && (
                  <div className="rounded-md bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 p-3">
                    <p className="text-sm text-green-600 dark:text-green-400 font-medium">
                      ✓ Email available
                    </p>
                    <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                      An invitation will be sent to create a new account
                    </p>
                  </div>
                )}
                <p className="text-sm text-muted-foreground">
                  The owner will receive an invitation to join this tenant
                </p>
              </div>
            </div>

            {/* Metadata */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Additional Information (Optional)</h3>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="industry">Industry</Label>
                  <select
                    id="industry"
                    value={formData.metadata.industry}
                    onChange={(e) => handleMetadataChange('industry', e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <option value="">Select industry</option>
                    <option value="Technology">Technology</option>
                    <option value="Finance">Finance</option>
                    <option value="Healthcare">Healthcare</option>
                    <option value="Education">Education</option>
                    <option value="Retail">Retail</option>
                    <option value="Manufacturing">Manufacturing</option>
                    <option value="Other">Other</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="companySize">Company Size</Label>
                  <select
                    id="companySize"
                    value={formData.metadata.companySize}
                    onChange={(e) => handleMetadataChange('companySize', e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <option value="">Select company size</option>
                    <option value="1-10">1-10 employees</option>
                    <option value="11-50">11-50 employees</option>
                    <option value="51-200">51-200 employees</option>
                    <option value="201-500">201-500 employees</option>
                    <option value="501-1000">501-1000 employees</option>
                    <option value="1000+">1000+ employees</option>
                  </select>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="notes">Internal Notes</Label>
                <Textarea
                  id="notes"
                  value={formData.metadata.notes}
                  onChange={(e) => handleMetadataChange('notes', e.target.value)}
                  placeholder="Any internal notes about this tenant..."
                  rows={3}
                />
              </div>
            </div>

            {/* Submit Button */}
            <div className="flex gap-3">
              <Button type="submit" disabled={loading} className="gap-2">
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                {loading ? 'Creating Tenant...' : 'Create Tenant'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate('/tenants')}
                disabled={loading}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Success Dialog */}
      <Dialog open={showSuccessDialog} onOpenChange={setShowSuccessDialog}>
        <DialogContent onClose={() => setShowSuccessDialog(false)}>
          <DialogHeader>
            <div className="flex items-center justify-center mb-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
                <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
            </div>
            <DialogTitle className="text-center">Tenant Created Successfully!</DialogTitle>
            <DialogDescription className="text-center">
              "{createdTenantName}" has been created and is ready to use.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="mt-4">
            <Button variant="outline" onClick={() => {
              setShowSuccessDialog(false);
              navigate('/tenants');
            }}>
              Back to Tenants
            </Button>
            <Button onClick={() => {
              setShowSuccessDialog(false);
              navigate(`/tenants/${createdTenantId}`);
            }}>
              View Tenant Details
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
