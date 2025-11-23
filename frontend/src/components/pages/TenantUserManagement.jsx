import React, { useState, useEffect } from 'react';
import { Users, Plus, Trash2, UserPlus, Mail, Copy, Check } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '../ui/dialog';

export function TenantUserManagement({ tenantId, onMembersUpdate }) {
  const [members, setMembers] = useState([]);
  const [relations, setRelations] = useState([]);
  const [userDetails, setUserDetails] = useState({}); // Map of userId -> user details
  const [loading, setLoading] = useState(true);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [addMemberData, setAddMemberData] = useState({ userId: '', roleId: '' });
  const [inviteData, setInviteData] = useState({ email: '', roleId: '' });
  const [copiedText, setCopiedText] = useState(''); // Track which text was copied

  useEffect(() => {
    loadData();
  }, [tenantId]);

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

      // Load relations
      const relationsResponse = await fetch(`/api/v1/relations`, {
        credentials: 'include'
      });
      
      if (relationsResponse.ok) {
        const relationsData = await relationsResponse.json();
        console.log('[TenantUserManagement] Relations response:', relationsData);
        const relationsArray = relationsData.data || [];
        setRelations(Array.isArray(relationsArray) ? relationsArray : []);
      } else {
        console.error('[TenantUserManagement] Failed to load relations:', relationsResponse.status);
        setRelations([]);
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
      setRelations([]);
    } finally {
      setLoading(false);
    }
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
    } catch (err) {
      console.error('Error sending invitation:', err);
      alert('Failed to send invitation: ' + err.message);
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
              <CardTitle>Members</CardTitle>
              <CardDescription>Manage users who have access to this tenant</CardDescription>
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
                        
                        {/* Relation badge */}
                        <div className="mt-2">
                          <Badge variant="secondary" className="text-xs">
                            {member.relation?.name || 'Member'}
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
        </CardContent>
      </Card>

      {/* Add Member Dialog */}
      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent onClose={() => setShowAddDialog(false)}>
          <DialogHeader>
            <DialogTitle>Add Member</DialogTitle>
            <DialogDescription>
              Add an existing user to this tenant by their User ID
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="userId">User ID</Label>
              <Input
                id="userId"
                placeholder="e.g., 04413f25-fdfa-..."
                value={addMemberData.userId}
                onChange={(e) => setAddMemberData(prev => ({ ...prev, userId: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="roleId">Relation/Role</Label>
              <select
                id="roleId"
                value={addMemberData.roleId}
                onChange={(e) => setAddMemberData(prev => ({ ...prev, roleId: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="">Select a relation</option>
                {relations.map(relation => (
                  <option key={relation.id} value={relation.id}>
                    {relation.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddMember}>
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
              <Label htmlFor="inviteRelationId">Relation/Role</Label>
              <select
                id="inviteRelationId"
                value={inviteData.roleId}
                onChange={(e) => setInviteData(prev => ({ ...prev, roleId: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="">Select a relation</option>
                {relations.map(relation => (
                  <option key={relation.id} value={relation.id}>
                    {relation.name}
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

