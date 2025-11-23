import React, { useState, useEffect } from 'react';
import { 
  Plus, Package, Loader2, Key, MoreVertical, 
  Trash2, Power, Copy, CheckCircle, XCircle, Calendar, Download,
  Clock, Shield, AlertTriangle, X, Info
} from 'lucide-react';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';

export function ApplicationsPage() {
  console.log('[ApplicationsPage] Component mounted');

  const [systemUsers, setSystemUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Fixed sidebar for viewing application credentials
  const [selectedApp, setSelectedApp] = useState(null);
  const [siderLoading, setSiderLoading] = useState(false);
  const [siderCredentials, setSiderCredentials] = useState([]);
  
  // Create App Dialog
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState('');
  const [newAppName, setNewAppName] = useState('');
  const [newAppDescription, setNewAppDescription] = useState('');
  const [newAppServiceType, setNewAppServiceType] = useState('api');
  
  // Created App Credentials Dialog
  const [showCredentialsDialog, setShowCredentialsDialog] = useState(false);
  const [createdCredentials, setCreatedCredentials] = useState(null);
  const [copiedEmail, setCopiedEmail] = useState(false);
  const [copiedPassword, setCopiedPassword] = useState(false);

  // Rotate Credentials Dialog
  const [showRotateDialog, setShowRotateDialog] = useState(false);
  const [rotateAppId, setRotateAppId] = useState(null);
  const [rotateAppName, setRotateAppName] = useState('');
  const [rotateLoading, setRotateLoading] = useState(false);
  const [rotateError, setRotateError] = useState('');
  const [gracePeriodDays, setGracePeriodDays] = useState(7);
  const [rotatedCredentials, setRotatedCredentials] = useState(null);

  // Revoke Dialog
  const [showRevokeDialog, setShowRevokeDialog] = useState(false);
  const [revokeAppName, setRevokeAppName] = useState('');
  const [revokeLoading, setRevokeLoading] = useState(false);

  useEffect(() => {
    loadSystemUsers();
  }, []);

  const loadSystemUsers = async () => {
    console.log('[ApplicationsPage] Loading system users...');
    try {
      setLoading(true);
      setError('');

      const response = await fetch('/api/v1/platform/system-users', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load system users: ${response.status}`);
      }

      const data = await response.json();
      console.log('[ApplicationsPage] System users loaded:', data);

      setSystemUsers(data.data || []);
    } catch (err) {
      console.error('[ApplicationsPage] Error loading system users:', err);
      setError(err.message);
      setSystemUsers([]);
    } finally {
      setLoading(false);
    }
  };

  // Group system users by application_name for display
  const getApplications = () => {
    const appMap = new Map();
    
    systemUsers.forEach(user => {
      const appName = user.application_name;
      if (!appMap.has(appName)) {
        appMap.set(appName, {
          name: appName,
          primary: null,
          credentialCount: 0,
          hasExpiring: false,
          serviceType: user.service_type
        });
      }
      
      const app = appMap.get(appName);
      app.credentialCount++;
      
      if (user.is_primary) {
        app.primary = user;
      }
      
      if (user.expires_at && new Date(user.expires_at) > new Date()) {
        app.hasExpiring = true;
      }
    });
    
    return Array.from(appMap.values());
  };

  const handleSelectApp = async (applicationName) => {
    console.log('[ApplicationsPage] Selecting application:', applicationName);
    setSelectedApp(applicationName);
    setSiderLoading(true);
    setSiderCredentials([]);
    
    try {
      const response = await fetch(`/api/v1/platform/applications/${encodeURIComponent(applicationName)}/credentials`, {
        credentials: 'include'
      });
      
      if (!response.ok) {
        throw new Error(`Failed to load credentials: ${response.status}`);
      }
      
      const data = await response.json();
      console.log('[ApplicationsPage] Credentials loaded:', data);
      
      setSiderCredentials(data.data?.credentials || []);
    } catch (err) {
      console.error('[ApplicationsPage] Error loading credentials:', err);
      setSiderCredentials([]);
    } finally {
      setSiderLoading(false);
    }
  };

  const handleCreateApp = async (e) => {
    e.preventDefault();
    console.log('[ApplicationsPage] Creating application:', newAppName);

    try {
      setCreateLoading(true);
      setCreateError('');

      const response = await fetch('/api/v1/platform/system-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          name: newAppName,
          description: newAppDescription,
          service_type: newAppServiceType
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to create application: ${response.status}`);
      }

      const data = await response.json();
      console.log('[ApplicationsPage] Application created:', data);

      // Show credentials dialog
      setCreatedCredentials(data.data);
      setShowCreateDialog(false);
      setShowCredentialsDialog(true);

      // Reset form
      setNewAppName('');
      setNewAppDescription('');
      setNewAppServiceType('api');

      // Reload applications
      loadSystemUsers();
    } catch (err) {
      console.error('[ApplicationsPage] Error creating application:', err);
      setCreateError(err.message);
    } finally {
      setCreateLoading(false);
    }
  };

  const handleRotate = async () => {
    console.log('[ApplicationsPage] Rotating credentials:', rotateAppId);

    try {
      setRotateLoading(true);
      setRotateError('');

      const response = await fetch(`/api/v1/platform/system-users/${rotateAppId}/rotate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          grace_period_days: gracePeriodDays
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to rotate credentials: ${response.status}`);
      }

      const data = await response.json();
      console.log('[ApplicationsPage] Credentials rotated:', data);

      // Show rotated credentials
      setRotatedCredentials(data.data);
      setShowRotateDialog(false);

      // Reload
      loadSystemUsers();
      if (selectedApp) {
        handleSelectApp(selectedApp);
      }
    } catch (err) {
      console.error('[ApplicationsPage] Error rotating credentials:', err);
      setRotateError(err.message);
    } finally {
      setRotateLoading(false);
    }
  };

  const handleRevokeOld = async () => {
    console.log('[ApplicationsPage] Revoking old credentials for:', revokeAppName);

    try {
      setRevokeLoading(true);

      const response = await fetch(`/api/v1/platform/applications/${encodeURIComponent(revokeAppName)}/revoke-old`, {
        method: 'POST',
        credentials: 'include'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to revoke old credentials: ${response.status}`);
      }

      const data = await response.json();
      console.log('[ApplicationsPage] Old credentials revoked:', data);

      setShowRevokeDialog(false);

      // Reload
      loadSystemUsers();
      if (selectedApp) {
        handleSelectApp(selectedApp);
      }
    } catch (err) {
      console.error('[ApplicationsPage] Error revoking old credentials:', err);
    } finally {
      setRevokeLoading(false);
    }
  };

  const handleCopyEmail = (email) => {
    navigator.clipboard.writeText(email);
    setCopiedEmail(true);
    setTimeout(() => setCopiedEmail(false), 2000);
  };

  const handleCopyPassword = (password) => {
    navigator.clipboard.writeText(password);
    setCopiedPassword(true);
    setTimeout(() => setCopiedPassword(false), 2000);
  };

  const downloadCredentials = (credentials, filename) => {
    const jsonData = {
      application_name: credentials.application_name || credentials.name,
      email: credentials.email,
      password: credentials.password,
      user_id: credentials.user_id,
      service_type: credentials.service_type,
      created_at: credentials.created_at,
      message: credentials.message
    };

    if (credentials.old_credentials && credentials.old_credentials.length > 0) {
      jsonData.old_credentials = credentials.old_credentials;
    }

    const blob = new Blob([JSON.stringify(jsonData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename || `${credentials.name || credentials.application_name}-credentials.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  const isExpiringSoon = (expiresAt) => {
    if (!expiresAt) return false;
    const expiry = new Date(expiresAt);
    const now = new Date();
    const daysUntilExpiry = (expiry - now) / (1000 * 60 * 60 * 24);
    return daysUntilExpiry > 0 && daysUntilExpiry <= 3;
  };

  const isExpired = (expiresAt) => {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  };

  const applications = getApplications();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
        <span className="ml-3 text-gray-600">Loading applications...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <XCircle className="h-12 w-12 text-red-500 mx-auto mb-3" />
          <p className="text-red-600 font-medium">Failed to load applications</p>
          <p className="text-gray-500 text-sm mt-1">{error}</p>
          <Button onClick={loadSystemUsers} className="mt-4">
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full gap-6">
      {/* Main Content - Applications List */}
      <div className="flex-1 space-y-6 overflow-y-auto">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Applications</h1>
            <p className="text-gray-600 mt-1">
              Manage system users and M2M authentication credentials
            </p>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Application
          </Button>
        </div>

        {/* Applications Grid */}
        {applications.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <Package className="h-12 w-12 text-gray-400 mx-auto mb-3" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No applications</h3>
            <p className="text-gray-500 mb-4">Create your first application to get started</p>
            <Button onClick={() => setShowCreateDialog(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Create Application
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {applications.map((app) => (
              <div
                key={app.name}
                className={`bg-white rounded-lg border-2 p-6 hover:shadow-md transition-all cursor-pointer ${
                  selectedApp === app.name
                    ? 'border-blue-500 shadow-md'
                    : 'border-gray-200'
                }`}
                onClick={() => handleSelectApp(app.name)}
              >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <Package className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">{app.name}</h3>
                    <Badge variant="outline" className="mt-1">
                      {app.serviceType}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-gray-500">Credentials:</span>
                  <span className="font-medium">{app.credentialCount}</span>
                </div>
                
                {app.primary && (
                  <>
                    <div className="flex items-center justify-between">
                      <span className="text-gray-500">Status:</span>
                      <Badge variant={app.primary.is_active ? 'success' : 'secondary'}>
                        {app.primary.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </div>
                    
                    {app.primary.last_used_at && (
                      <div className="flex items-center justify-between">
                        <span className="text-gray-500">Last used:</span>
                        <span className="text-xs text-gray-600">
                          {formatDate(app.primary.last_used_at)}
                        </span>
                      </div>
                    )}
                  </>
                )}

                {app.hasExpiring && (
                  <div className="mt-3 pt-3 border-t border-gray-200">
                    <div className="flex items-center gap-2 text-orange-600">
                      <AlertTriangle className="h-4 w-4" />
                      <span className="text-xs font-medium">Has expiring credentials</span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
        )}
      </div>

      {/* Fixed Sidebar - Credential Details */}
      <div className="w-96 bg-white border-l-2 border-gray-200 overflow-y-auto flex-shrink-0">
        {!selectedApp ? (
          /* Empty State */
          <div className="flex flex-col items-center justify-center h-full p-8 text-center">
            <div className="p-4 bg-blue-50 rounded-full mb-4">
              <Info className="h-8 w-8 text-blue-600" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              No Application Selected
            </h3>
            <p className="text-sm text-gray-600">
              Select an application from the left to view its credentials and manage rotation.
            </p>
          </div>
        ) : (
          /* Credential Details */
          <div className="p-6 space-y-6">
            {/* Header */}
            <div className="pb-4 border-b border-gray-200">
              <div className="flex items-center gap-3 mb-2">
                <Package className="h-5 w-5 text-gray-700" />
                <h2 className="text-lg font-semibold text-gray-900">{selectedApp}</h2>
              </div>
              <p className="text-sm text-gray-600">
                Manage credentials and rotation for this application
              </p>
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-2">
              <Button
                onClick={() => {
                  const primary = siderCredentials.find(c => c.is_primary);
                  if (primary) {
                    setRotateAppId(primary.id);
                    setRotateAppName(selectedApp);
                    setShowRotateDialog(true);
                  }
                }}
                disabled={!siderCredentials.find(c => c.is_primary)}
                className="flex-1"
              >
                <Key className="h-4 w-4 mr-2" />
                Rotate Credentials
              </Button>
              
              <Button
                variant="outline"
                onClick={() => {
                  setRevokeAppName(selectedApp);
                  setShowRevokeDialog(true);
                }}
                disabled={!siderCredentials.some(c => !c.is_primary && c.is_active)}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Revoke Old
              </Button>
            </div>

            {/* Credentials List */}
            {siderLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-6 w-6 animate-spin text-gray-500" />
              </div>
            ) : (
              <div className="space-y-4 pt-2">
                <h3 className="font-semibold text-gray-900 flex items-center gap-2 pb-2 border-b border-gray-200">
                  <Shield className="h-4 w-4" />
                  Credentials ({siderCredentials.length})
                </h3>

                {siderCredentials.map((cred) => (
                  <div
                    key={cred.id}
                    className={`p-4 rounded-lg border-2 ${
                      cred.is_primary
                        ? 'border-green-500 bg-green-50'
                        : isExpired(cred.expires_at)
                        ? 'border-red-300 bg-red-50 opacity-60'
                        : isExpiringSoon(cred.expires_at)
                        ? 'border-orange-300 bg-orange-50'
                        : 'border-gray-200 bg-gray-50'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-gray-900">{cred.name}</h4>
                        {cred.is_primary && (
                          <Badge variant="success">Primary</Badge>
                        )}
                        {isExpired(cred.expires_at) && (
                          <Badge variant="destructive">Expired</Badge>
                        )}
                        {isExpiringSoon(cred.expires_at) && (
                          <Badge variant="warning">Expiring Soon</Badge>
                        )}
                      </div>
                      <Badge variant={cred.is_active ? 'default' : 'secondary'}>
                        {cred.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </div>

                    <div className="space-y-2 text-sm">
                      <div className="flex items-center justify-between">
                        <span className="text-gray-600">Email:</span>
                        <code className="text-xs bg-white px-2 py-1 rounded border">
                          {cred.email}
                        </code>
                      </div>

                      <div className="flex items-center justify-between">
                        <span className="text-gray-600">User ID:</span>
                        <code className="text-xs bg-white px-2 py-1 rounded border">
                          {cred.user_id.substring(0, 16)}...
                        </code>
                      </div>

                      <div className="flex items-center justify-between">
                        <span className="text-gray-600">Created:</span>
                        <span className="text-xs text-gray-700">
                          {formatDate(cred.created_at)}
                        </span>
                      </div>

                      {cred.last_used_at && (
                        <div className="flex items-center justify-between">
                          <span className="text-gray-600">Last used:</span>
                          <span className="text-xs text-gray-700">
                            {formatDate(cred.last_used_at)}
                          </span>
                        </div>
                      )}

                      {cred.expires_at && (
                        <div className="flex items-center justify-between">
                          <span className="text-gray-600 flex items-center gap-1">
                            <AlertTriangle className="h-3 w-3" />
                            Expires:
                          </span>
                          <span className={`text-xs font-medium ${
                            isExpired(cred.expires_at)
                              ? 'text-red-600'
                              : isExpiringSoon(cred.expires_at)
                              ? 'text-orange-600'
                              : 'text-gray-700'
                          }`}>
                            {formatDate(cred.expires_at)}
                          </span>
                        </div>
                      )}

                      {cred.description && (
                        <div className="pt-2 border-t border-gray-200">
                          <p className="text-xs text-gray-600">{cred.description}</p>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Create App Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Application</DialogTitle>
            <DialogDescription>
              Create a new system user for machine-to-machine authentication
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreateApp} className="space-y-4 mb-6">
            <div>
              <Label htmlFor="name">Application Name *</Label>
              <Input
                id="name"
                value={newAppName}
                onChange={(e) => setNewAppName(e.target.value)}
                placeholder="my-background-worker"
                required
                minLength={3}
              />
              <p className="text-xs text-gray-500 mt-1">
                Unique identifier for your application
              </p>
            </div>

            <div>
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={newAppDescription}
                onChange={(e) => setNewAppDescription(e.target.value)}
                placeholder="Background job processor"
              />
            </div>

            <div>
              <Label htmlFor="service_type">Service Type *</Label>
              <select
                id="service_type"
                value={newAppServiceType}
                onChange={(e) => setNewAppServiceType(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              >
                <option value="api">API Service</option>
                <option value="worker">Background Worker</option>
                <option value="integration">Integration</option>
                <option value="cron">Cron Job</option>
              </select>
            </div>

            {createError && (
              <div className="text-sm text-red-600 bg-red-50 p-3 rounded">
                {createError}
              </div>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowCreateDialog(false)}
                disabled={createLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createLoading}>
                {createLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Creating...
                  </>
                ) : (
                  <>Create Application</>
                )}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Credentials Dialog (after creation) */}
      <Dialog open={showCredentialsDialog} onOpenChange={setShowCredentialsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-green-600">
              <CheckCircle className="h-5 w-5" />
              Application Created Successfully
            </DialogTitle>
            <DialogDescription>
              Save these credentials securely. The password will not be shown again.
            </DialogDescription>
          </DialogHeader>

          {createdCredentials && (
            <div className="space-y-4 mb-6">
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <p className="text-sm font-medium text-yellow-800 mb-1">⚠️ Important</p>
                <p className="text-xs text-yellow-700">
                  {createdCredentials.message}
                </p>
              </div>

              <div className="space-y-3">
                <div>
                  <Label className="text-xs text-gray-600">Email</Label>
                  <div className="flex gap-2 mt-1">
                    <code className="flex-1 text-sm bg-gray-100 px-3 py-2 rounded border">
                      {createdCredentials.email}
                    </code>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopyEmail(createdCredentials.email)}
                    >
                      {copiedEmail ? (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                <div>
                  <Label className="text-xs text-gray-600">Password</Label>
                  <div className="flex gap-2 mt-1">
                    <code className="flex-1 text-sm bg-gray-100 px-3 py-2 rounded border font-mono break-all">
                      {createdCredentials.password}
                    </code>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopyPassword(createdCredentials.password)}
                    >
                      {copiedPassword ? (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                <div className="text-xs text-gray-600 space-y-1 pt-2 border-t">
                  <p><strong>User ID:</strong> {createdCredentials.user_id}</p>
                  <p><strong>Service Type:</strong> {createdCredentials.service_type}</p>
                  <p><strong>Created:</strong> {formatDate(createdCredentials.created_at)}</p>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => downloadCredentials(createdCredentials, `${createdCredentials.name}-credentials.json`)}
            >
              <Download className="h-4 w-4 mr-2" />
              Download JSON
            </Button>
            <Button onClick={() => setShowCredentialsDialog(false)}>
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rotate Credentials Dialog */}
      <Dialog open={showRotateDialog} onOpenChange={setShowRotateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rotate Credentials</DialogTitle>
            <DialogDescription>
              Create new credential with a grace period. Both old and new credentials will work during the transition.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 mb-6">
            <div>
              <Label htmlFor="grace_period">Grace Period (days)</Label>
              <Input
                id="grace_period"
                type="number"
                min="1"
                max="30"
                value={gracePeriodDays}
                onChange={(e) => setGracePeriodDays(parseInt(e.target.value))}
              />
              <p className="text-xs text-gray-500 mt-1">
                Old credentials will remain active for this many days
              </p>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-blue-900 mb-2">What happens:</h4>
              <ul className="text-xs text-blue-800 space-y-1 list-disc list-inside">
                <li>New credential is created and becomes primary</li>
                <li>Old credential remains active for {gracePeriodDays} days</li>
                <li>Both credentials work simultaneously</li>
                <li>Update your services gradually during the grace period</li>
                <li>Old credential will auto-expire after grace period</li>
              </ul>
            </div>

            {rotateError && (
              <div className="text-sm text-red-600 bg-red-50 p-3 rounded">
                {rotateError}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowRotateDialog(false)}
              disabled={rotateLoading}
            >
              Cancel
            </Button>
            <Button onClick={handleRotate} disabled={rotateLoading}>
              {rotateLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Rotating...
                </>
              ) : (
                <>
                  <Key className="h-4 w-4 mr-2" />
                  Rotate Credentials
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rotated Credentials Dialog */}
      <Dialog open={!!rotatedCredentials} onOpenChange={() => setRotatedCredentials(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-green-600">
              <CheckCircle className="h-5 w-5" />
              Credentials Rotated Successfully
            </DialogTitle>
            <DialogDescription>
              New credential created. Both old and new credentials are now active.
            </DialogDescription>
          </DialogHeader>

          {rotatedCredentials && (
            <div className="space-y-4 mb-6">
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <p className="text-sm font-medium text-blue-800 mb-1">ℹ️ Grace Period Active</p>
                <p className="text-xs text-blue-700">
                  {rotatedCredentials.message}
                </p>
              </div>

              {/* New Credentials */}
              <div className="space-y-3">
                <h4 className="font-medium text-gray-900">New Credentials (Primary)</h4>
                
                <div>
                  <Label className="text-xs text-gray-600">Email</Label>
                  <div className="flex gap-2 mt-1">
                    <code className="flex-1 text-sm bg-gray-100 px-3 py-2 rounded border">
                      {rotatedCredentials.email}
                    </code>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopyEmail(rotatedCredentials.email)}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                <div>
                  <Label className="text-xs text-gray-600">Password</Label>
                  <div className="flex gap-2 mt-1">
                    <code className="flex-1 text-sm bg-gray-100 px-3 py-2 rounded border font-mono break-all">
                      {rotatedCredentials.password}
                    </code>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopyPassword(rotatedCredentials.password)}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>

              {/* Old Credentials Info */}
              {rotatedCredentials.old_credentials && rotatedCredentials.old_credentials.length > 0 && (
                <div className="space-y-3 pt-4 border-t">
                  <h4 className="font-medium text-gray-900 flex items-center gap-2">
                    <Clock className="h-4 w-4" />
                    Old Credentials (Still Active)
                  </h4>
                  
                  {rotatedCredentials.old_credentials.map((oldCred, index) => (
                    <div key={index} className="bg-orange-50 border border-orange-200 rounded-lg p-3">
                      <div className="text-sm space-y-1">
                        <p className="font-medium text-orange-900">Email: {oldCred.email}</p>
                        <p className="text-xs text-orange-700">
                          <strong>Expires:</strong> {formatDate(oldCred.expires_at)}
                        </p>
                        <p className="text-xs text-orange-600 mt-2">
                          ⚠️ {oldCred.message}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => downloadCredentials(rotatedCredentials, `${rotatedCredentials.name}-rotated-credentials.json`)}
            >
              <Download className="h-4 w-4 mr-2" />
              Download JSON
            </Button>
            <Button onClick={() => setRotatedCredentials(null)}>
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Revoke Old Credentials Dialog */}
      <Dialog open={showRevokeDialog} onOpenChange={setShowRevokeDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <AlertTriangle className="h-5 w-5" />
              Revoke Old Credentials
            </DialogTitle>
            <DialogDescription>
              This will immediately deactivate all non-primary credentials for "{revokeAppName}".
            </DialogDescription>
          </DialogHeader>

          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <p className="text-sm font-medium text-red-800 mb-2">⚠️ Warning</p>
            <p className="text-xs text-red-700">
              This action will immediately revoke all old credentials. Services using old credentials will stop working.
              Make sure all services have been updated to use the new credentials before proceeding.
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowRevokeDialog(false)}
              disabled={revokeLoading}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleRevokeOld}
              disabled={revokeLoading}
            >
              {revokeLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Revoking...
                </>
              ) : (
                <>
                  <Trash2 className="h-4 w-4 mr-2" />
                  Revoke Old Credentials
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
