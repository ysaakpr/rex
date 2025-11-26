import React, { useState, useEffect } from 'react';
import { Users, Plus, Trash2, UserPlus, Mail, Copy, Check, X, Clock } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../ui/tabs';
import appConfig from '../../config';

export function TenantUserManagement({ tenantId, onMembersUpdate }) {
  const [members, setMembers] = useState([]);
  const [invitations, setInvitations] = useState([]);
  const [roles, setRoles] = useState([]);
  const [userDetails, setUserDetails] = useState({}); // Map of userId -> user details
  const [allUsers, setAllUsers] = useState([]); // All platform users for search
  const [userSearchQuery, setUserSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [loadingInvitations, setLoadingInvitations] = useState(false);
  const [searchingUsers, setSearchingUsers] = useState(false);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [addMemberData, setAddMemberData] = useState({ userId: '', roleId: '' });
  const [inviteData, setInviteData] = useState({ email: '', roleId: '' });
  const [copiedText, setCopiedText] = useState(''); // Track which text was copied
  const [activeTab, setActiveTab] = useState('members');

  useEffect(() => {
    loadData();
  }, [tenantId]);

  useEffect(() => {
    if (activeTab === 'invitations') {
      loadInvitations();
    }
  }, [activeTab, tenantId]);

  const loadData = async () => {
    try {
      setLoading(true);
      
      // Load members
      const membersResponse = await fetch(`/api/v1/tenants/${tenantId}/members`, {
        credentials: 'include'
      });
      
      let membersArray = [];
      if (membersResponse.ok) {
        const membersData = await membersResponse.json();
        console.log('[TenantUserManagement] Members response:', membersData);
        // Handle both direct array and nested data structure
        membersArray = membersData.data?.data || membersData.data || [];
        setMembers(Array.isArray(membersArray) ? membersArray : []);
      } else {
        console.error('[TenantUserManagement] Failed to load members:', membersResponse.status);
        setMembers([]);
      }

      // Load roles
      const rolesResponse = await fetch(`/api/v1/roles`, {
        credentials: 'include'
      });
      
      if (rolesResponse.ok) {
        const rolesData = await rolesResponse.json();
        console.log('[TenantUserManagement] Roles response:', rolesData);
        const rolesArray = rolesData.data || [];
        setRoles(Array.isArray(rolesArray) ? rolesArray : []);
      } else {
        console.error('[TenantUserManagement] Failed to load roles:', rolesResponse.status);
        setRoles([]);
      }

      // Load user details for all members
      if (Array.isArray(membersArray) && membersArray.length > 0) {
        const userIds = membersArray.map(m => m.user_id).filter(Boolean);
        if (userIds.length > 0) {
          await fetchUserDetails(userIds);
        }
      }
    } catch (err) {
      console.error('[TenantUserManagement] Error loading data:', err);
      setMembers([]);
      setRoles([]);
    } finally {
      setLoading(false);
    }
  };

  const loadInvitations = async () => {
    try {
      setLoadingInvitations(true);
      
      const response = await fetch(`/api/v1/tenants/${tenantId}/invitations`, {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        console.log('[TenantUserManagement] Invitations response:', data);
        // Handle paginated response
        const invitationsArray = data.data?.data || data.data || [];
        setInvitations(Array.isArray(invitationsArray) ? invitationsArray : []);
      } else {
        console.error('[TenantUserManagement] Failed to load invitations:', response.status);
        setInvitations([]);
      }
    } catch (err) {
      console.error('[TenantUserManagement] Error loading invitations:', err);
      setInvitations([]);
    } finally {
      setLoadingInvitations(false);
    }
  };

  const searchUsers = async (query) => {
    if (!query || query.length < 2) {
      setAllUsers([]);
      return;
    }

    try {
      setSearchingUsers(true);
      const response = await fetch(`/api/v1/users/search?q=${encodeURIComponent(query)}`, {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[TenantUserManagement] User search response:', data);
        setAllUsers(data.data || []);
      } else {
        console.error('[TenantUserManagement] Failed to search users:', response.status);
        setAllUsers([]);
      }
    } catch (err) {
      console.error('[TenantUserManagement] Error searching users:', err);
      setAllUsers([]);
    } finally {
      setSearchingUsers(false);
    }
  };

  const handleUserSearchChange = (query) => {
    setUserSearchQuery(query);
    searchUsers(query);
  };

  const selectUser = (user) => {
    setAddMemberData(prev => ({ ...prev, userId: user.id }));
    setUserSearchQuery(`${user.name || user.email} (${user.email})`);
    setAllUsers([]);
  };

  const fetchUserDetails = async (userIds) => {
    try {
      const response = await fetch('/api/v1/users/batch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(userIds)
      });

      if (response.ok) {
        const data = await response.json();
        console.log('[TenantUserManagement] User details response:', data);
        setUserDetails(data.data || {});
      } else {
        console.error('[TenantUserManagement] Failed to fetch user details:', response.status);
      }
    } catch (err) {
      console.error('[TenantUserManagement] Error fetching user details:', err);
    }
  };

  const handleAddMember = async () => {
    if (!addMemberData.userId || !addMemberData.roleId) {
      alert('Please fill in all fields');
      return;
    }

    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          user_id: addMemberData.userId,
          role_id: addMemberData.roleId
        })
      });

      if (!response.ok) {
        throw new Error('Failed to add member');
      }

      setShowAddDialog(false);
      setAddMemberData({ userId: '', roleId: '' });
      loadData();
      if (onMembersUpdate) onMembersUpdate();
    } catch (err) {
      console.error('Error adding member:', err);
      alert('Failed to add member: ' + err.message);
    }
  };

  const handleSendInvite = async () => {
    if (!inviteData.email || !inviteData.roleId) {
      alert('Please fill in all fields');
      return;
    }

    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/invitations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          email: inviteData.email,
          role_id: inviteData.roleId
        })
      });

      if (!response.ok) {
        throw new Error('Failed to send invitation');
      }

      setShowInviteDialog(false);
      setInviteData({ email: '', roleId: '' });
      alert('Invitation sent successfully!');
      // Reload invitations if on that tab
      if (activeTab === 'invitations') {
        loadInvitations();
      }
    } catch (err) {
      console.error('Error sending invitation:', err);
      alert('Failed to send invitation: ' + err.message);
    }
  };

  const handleCancelInvitation = async (invitationId) => {
    try {
      const response = await fetch(`/api/v1/invitations/${invitationId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to cancel invitation');
      }

      // Reload invitations
      loadInvitations();
    } catch (err) {
      console.error('Error canceling invitation:', err);
      alert('Failed to cancel invitation: ' + err.message);
    }
  };

  const handleRemoveMember = async (userId) => {
    try {
      const response = await fetch(`/api/v1/tenants/${tenantId}/members/${userId}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to remove member');
      }

      loadData();
      if (onMembersUpdate) onMembersUpdate();
    } catch (err) {
      console.error('Error removing member:', err);
      alert('Failed to remove member: ' + err.message);
    }
  };

  const copyToClipboard = async (text, label) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedText(label);
      setTimeout(() => setCopiedText(''), 2000); // Reset after 2 seconds
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center py-8">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Team Management</CardTitle>
              <CardDescription>Manage members and invitations for this tenant</CardDescription>
            </div>
            <div className="flex gap-2">
              <Button onClick={() => setShowInviteDialog(true)} variant="outline" size="sm" className="gap-2">
                <Mail className="h-4 w-4" />
                Invite
              </Button>
              <Button onClick={() => setShowAddDialog(true)} size="sm" className="gap-2">
                <Plus className="h-4 w-4" />
                Add Member
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="members" value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full max-w-[400px] grid-cols-2">
              <TabsTrigger value="members">
                <Users className="h-4 w-4 mr-2" />
                Members
              </TabsTrigger>
              <TabsTrigger value="invitations">
                <Mail className="h-4 w-4 mr-2" />
                Invitations
              </TabsTrigger>
            </TabsList>

            <TabsContent value="members" className="mt-4">
              {!Array.isArray(members) || members.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Users className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p className="text-sm">No members yet</p>
                  <p className="text-xs mt-1">Add members or send invitations to get started</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {members.map((member) => {
                    const userInfo = userDetails[member.user_id] || {};
                    const userName = userInfo.name || userInfo.email?.split('@')[0] || 'Unknown User';
                    const userEmail = userInfo.email || 'No email';
                    
                    return (
                      <div key={member.id} className="flex items-start justify-between p-4 border rounded-lg hover:bg-muted/50">
                        <div className="flex items-start gap-3 flex-1">
                          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 flex-shrink-0 mt-1">
                            <Users className="h-5 w-5 text-primary" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="font-semibold text-sm mb-1">{userName}</div>
                            
                            {/* Email with copy button */}
                            <div className="flex items-center gap-2 mb-1">
                              <span className="text-xs text-muted-foreground truncate">{userEmail}</span>
                              <button
                                onClick={() => copyToClipboard(userEmail, `email-${member.user_id}`)}
                                className="flex-shrink-0 p-1 hover:bg-muted rounded transition-colors"
                                title="Copy email"
                              >
                                {copiedText === `email-${member.user_id}` ? (
                                  <Check className="h-3 w-3 text-green-600" />
                                ) : (
                                  <Copy className="h-3 w-3 text-muted-foreground" />
                                )}
                              </button>
                            </div>
                            
                            {/* User ID with copy button */}
                            <div className="flex items-center gap-2">
                              <span className="text-xs text-muted-foreground font-mono truncate">
                                ID: {member.user_id}
                              </span>
                              <button
                                onClick={() => copyToClipboard(member.user_id, `id-${member.user_id}`)}
                                className="flex-shrink-0 p-1 hover:bg-muted rounded transition-colors"
                                title="Copy user ID"
                              >
                                {copiedText === `id-${member.user_id}` ? (
                                  <Check className="h-3 w-3 text-green-600" />
                                ) : (
                                  <Copy className="h-3 w-3 text-muted-foreground" />
                                )}
                              </button>
                            </div>
                            
                            {/* Role badge */}
                            <div className="mt-2">
                              <Badge variant="secondary" className="text-xs">
                                {member.role?.name || 'Member'}
                              </Badge>
                            </div>
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-2 flex-shrink-0 ml-4">
                          {Array.isArray(member.roles) && member.roles.map((role) => (
                            <Badge key={role.id} variant="outline" className="text-xs">
                              {role.name}
                            </Badge>
                          ))}
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                              if (confirm(`Remove ${userName} (${userEmail}) from this tenant?`)) {
                                handleRemoveMember(member.user_id);
                              }
                            }}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </TabsContent>

            <TabsContent value="invitations" className="mt-4">
              {loadingInvitations ? (
                <div className="flex items-center justify-center py-8">
                  <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
                </div>
              ) : !Array.isArray(invitations) || invitations.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <Mail className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p className="text-sm">No pending invitations</p>
                  <p className="text-xs mt-1">Send invitations to add new members</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {invitations.map((invitation) => (
                    <div key={invitation.id} className="flex items-start justify-between p-4 border rounded-lg hover:bg-muted/50">
                      <div className="flex items-start gap-3 flex-1">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/30 flex-shrink-0 mt-1">
                          <Mail className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="font-semibold text-sm mb-1">{invitation.email}</div>
                          
                          {/* Status and Role */}
                          <div className="flex items-center gap-2 mb-2">
                            <Badge 
                              variant={invitation.status === 'pending' ? 'default' : 'secondary'} 
                              className="text-xs"
                            >
                              {invitation.status === 'pending' && <Clock className="h-3 w-3 mr-1" />}
                              {invitation.status}
                            </Badge>
                            <Badge variant="outline" className="text-xs">
                              {invitation.role?.name || 'Member'}
                            </Badge>
                          </div>
                          
                          {/* Invitation link with copy button */}
                          {invitation.token && (
                            <div className="flex items-center gap-2">
                              <span className="text-xs text-muted-foreground font-mono truncate">
                                Link: ...{invitation.token.slice(-8)}
                              </span>
                              <button
                                onClick={() => {
                                  // Use invitation URL from backend (single source of truth)
                                  const inviteLink = invitation.invitation_url || `${window.location.origin}${appConfig.basename || ''}/accept-invite?token=${invitation.token}`;
                                  copyToClipboard(inviteLink, `invite-${invitation.id}`);
                                }}
                                className="flex-shrink-0 p-1 hover:bg-muted rounded transition-colors"
                                title="Copy invitation link"
                              >
                                {copiedText === `invite-${invitation.id}` ? (
                                  <Check className="h-3 w-3 text-green-600" />
                                ) : (
                                  <Copy className="h-3 w-3 text-muted-foreground" />
                                )}
                              </button>
                            </div>
                          )}
                          
                          {/* Created date */}
                          <div className="text-xs text-muted-foreground mt-1">
                            Sent: {new Date(invitation.created_at).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-2 flex-shrink-0 ml-4">
                        {invitation.status === 'pending' && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                              if (confirm(`Cancel invitation for ${invitation.email}?`)) {
                                handleCancelInvitation(invitation.id);
                              }
                            }}
                            title="Cancel invitation"
                          >
                            <X className="h-4 w-4 text-destructive" />
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Add Member Dialog */}
      <Dialog open={showAddDialog} onOpenChange={(open) => {
        setShowAddDialog(open);
        if (!open) {
          setUserSearchQuery('');
          setAllUsers([]);
          setAddMemberData({ userId: '', roleId: '' });
        }
      }}>
        <DialogContent onClose={() => {
          setShowAddDialog(false);
          setUserSearchQuery('');
          setAllUsers([]);
          setAddMemberData({ userId: '', roleId: '' });
        }}>
          <DialogHeader>
            <DialogTitle>Add Member</DialogTitle>
            <DialogDescription>
              Search for an existing user and assign them a role
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="userSearch">Search User</Label>
              <div className="relative">
                <Input
                  id="userSearch"
                  placeholder="Search by name or email..."
                  value={userSearchQuery}
                  onChange={(e) => handleUserSearchChange(e.target.value)}
                  autoComplete="off"
                />
                {searchingUsers && (
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
                  </div>
                )}
                {allUsers.length > 0 && (
                  <div className="absolute z-10 w-full mt-1 max-h-60 overflow-auto bg-background border rounded-md shadow-lg">
                    {allUsers.map((user) => (
                      <button
                        key={user.id}
                        onClick={() => selectUser(user)}
                        className="w-full text-left px-4 py-3 hover:bg-muted border-b last:border-b-0 transition-colors"
                      >
                        <div className="font-medium text-sm">{user.name || 'No name'}</div>
                        <div className="text-xs text-muted-foreground">{user.email}</div>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              {addMemberData.userId && (
                <p className="text-xs text-muted-foreground">
                  Selected user ID: {addMemberData.userId}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="roleId">Role</Label>
              <select
                id="roleId"
                value={addMemberData.roleId}
                onChange={(e) => setAddMemberData(prev => ({ ...prev, roleId: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="">Select a role</option>
                {roles.map(role => (
                  <option key={role.id} value={role.id}>
                    {role.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setShowAddDialog(false);
              setUserSearchQuery('');
              setAllUsers([]);
              setAddMemberData({ userId: '', roleId: '' });
            }}>
              Cancel
            </Button>
            <Button onClick={handleAddMember} disabled={!addMemberData.userId || !addMemberData.roleId}>
              Add Member
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Invite Dialog */}
      <Dialog open={showInviteDialog} onOpenChange={setShowInviteDialog}>
        <DialogContent onClose={() => setShowInviteDialog(false)}>
          <DialogHeader>
            <DialogTitle>Invite User</DialogTitle>
            <DialogDescription>
              Send an invitation email to a new user
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                placeholder="user@example.com"
                value={inviteData.email}
                onChange={(e) => setInviteData(prev => ({ ...prev, email: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="inviteRoleId">Role</Label>
              <select
                id="inviteRoleId"
                value={inviteData.roleId}
                onChange={(e) => setInviteData(prev => ({ ...prev, roleId: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="">Select a role</option>
                {roles.map(role => (
                  <option key={role.id} value={role.id}>
                    {role.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowInviteDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleSendInvite}>
              Send Invitation
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

