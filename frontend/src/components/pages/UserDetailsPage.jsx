import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, User, Copy, Check, Loader2, Building2, Shield, Calendar, ExternalLink } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../ui/tabs';

export function UserDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  console.log('[UserDetailsPage] Component mounted with user ID:', id);
  
  const [user, setUser] = useState(null);
  const [tenantMemberships, setTenantMemberships] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copiedField, setCopiedField] = useState('');

  useEffect(() => {
    loadUserDetails();
    loadUserTenants();
  }, [id]);

  const loadUserDetails = async () => {
    console.log('[UserDetailsPage] Loading user details for:', id);
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch(`/api/v1/users/${id}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load user: ${response.status}`);
      }

      const data = await response.json();
      console.log('[UserDetailsPage] User loaded:', data);
      
      setUser(data.data);
    } catch (err) {
      console.error('[UserDetailsPage] Error loading user:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadUserTenants = async () => {
    console.log('[UserDetailsPage] Loading user tenants for:', id);
    try {
      const response = await fetch(`/api/v1/users/${id}/tenants`, {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[UserDetailsPage] User tenants loaded:', data);
        setTenantMemberships(data.data || []);
      }
    } catch (err) {
      console.error('[UserDetailsPage] Error loading user tenants:', err);
    }
  };

  const handleCopy = (text, field) => {
    console.log('[UserDetailsPage] Copying to clipboard:', field);
    navigator.clipboard.writeText(text).then(() => {
      setCopiedField(field);
      setTimeout(() => setCopiedField(''), 2000);
    });
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    } catch {
      return dateString;
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading user details...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-destructive">User not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/users')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                <User className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-3xl font-bold tracking-tight">
                  {user.name || user.email?.split('@')[0] || 'User'}
                </h1>
                <p className="text-muted-foreground">{user.email}</p>
              </div>
            </div>
          </div>
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
      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="tenants">
            Tenants ({tenantMemberships.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile">
          <Card>
            <CardHeader>
              <CardTitle>User Profile</CardTitle>
              <CardDescription>
                Basic user information and account details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* User ID */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">User ID</label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 rounded bg-muted px-3 py-2 text-sm font-mono">
                    {user.user_id || id}
                  </code>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => handleCopy(user.user_id || id, 'user_id')}
                  >
                    {copiedField === 'user_id' ? (
                      <Check className="h-4 w-4 text-green-600" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              {/* Email */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">Email</label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 rounded bg-muted px-3 py-2 text-sm">
                    {user.email}
                  </code>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => handleCopy(user.email, 'email')}
                  >
                    {copiedField === 'email' ? (
                      <Check className="h-4 w-4 text-green-600" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              {/* Name (if available) */}
              {user.name && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-muted-foreground">Name</label>
                  <div className="rounded bg-muted px-3 py-2 text-sm">
                    {user.name}
                  </div>
                </div>
              )}

              {/* Account Status */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">Account Status</label>
                <div>
                  <Badge variant={user.is_active !== false ? 'default' : 'destructive'}>
                    {user.is_active !== false ? 'Active' : 'Inactive'}
                  </Badge>
                </div>
              </div>

              {/* Created Date */}
              {user.created_at && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-muted-foreground">Created</label>
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {formatDate(user.created_at)}
                  </div>
                </div>
              )}

              {/* Last Login (if available) */}
              {user.last_login && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-muted-foreground">Last Login</label>
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {formatDate(user.last_login)}
                  </div>
                </div>
              )}

              {/* Platform Admin Badge */}
              {user.is_platform_admin && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-muted-foreground">Special Roles</label>
                  <div>
                    <Badge variant="default" className="gap-1">
                      <Shield className="h-3 w-3" />
                      Platform Administrator
                    </Badge>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="tenants">
          <Card>
            <CardHeader>
              <CardTitle>Tenant Memberships</CardTitle>
              <CardDescription>
                Tenants this user belongs to and their relations
              </CardDescription>
            </CardHeader>
            <CardContent>
              {tenantMemberships.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Building2 className="mx-auto h-12 w-12 mb-4 opacity-50" />
                  <p className="text-sm">Not a member of any tenants yet</p>
                  <p className="text-xs mt-1">User can be invited to tenants by administrators</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {tenantMemberships.map((membership) => (
                    <div
                      key={membership.tenant_id}
                      className="p-4 border rounded-lg hover:bg-muted/50 transition-all"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <Building2 className="h-5 w-5 text-primary" />
                            <h3 className="font-semibold">{membership.tenant_name || 'Unknown Tenant'}</h3>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6"
                              onClick={() => navigate(`/tenants/${membership.tenant_id}`)}
                            >
                              <ExternalLink className="h-3 w-3" />
                            </Button>
                          </div>
                          
                          <div className="mt-2 flex flex-wrap gap-2">
                            {/* Role Badge */}
                            {membership.role_name && (
                              <Badge variant="secondary" className="gap-1">
                                <Shield className="h-3 w-3" />
                                {membership.role_name}
                              </Badge>
                            )}
                            
                            {/* Status Badge */}
                            <Badge variant={membership.status === 'active' ? 'default' : 'secondary'} className="text-xs">
                              {membership.status}
                            </Badge>
                          </div>
                          
                          {membership.joined_at && (
                            <p className="text-xs text-muted-foreground mt-2">
                              Joined {formatDate(membership.joined_at)}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

