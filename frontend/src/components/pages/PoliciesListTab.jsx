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

export function PoliciesListTab() {
  console.log('[PoliciesListTab] Component mounted');
  
  const navigate = useNavigate();
  const [policies, setPolicies] = useState([]);
  const [filteredPolicies, setFilteredPolicies] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [creating, setCreating] = useState(false);

  const [newPolicy, setNewPolicy] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadPolicies();
  }, []);

  useEffect(() => {
    filterPolicies();
  }, [policies, searchQuery]);

  const loadPolicies = async () => {
    console.log('[PoliciesListTab] Loading policies...');
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch('/api/v1/platform/policies', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load policies: ${response.status}`);
      }

      const data = await response.json();
      console.log('[PoliciesListTab] Policies loaded:', data);
      
      const policiesArray = data.data || [];
      setPolicies(Array.isArray(policiesArray) ? policiesArray : []);
    } catch (err) {
      console.error('[PoliciesListTab] Error loading policies:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filterPolicies = () => {
    let filtered = [...policies];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(r =>
        r.name?.toLowerCase().includes(query) ||
        r.description?.toLowerCase().includes(query)
      );
    }

    console.log('[PoliciesListTab] Filtered policies:', filtered.length, 'from', policies.length);
    setFilteredPolicies(filtered);
  };

  const handleCreatePolicy = async () => {
    console.log('[PoliciesListTab] Creating policy:', newPolicy);
    
    if (!newPolicy.name) {
      setError('Policy name is required');
      return;
    }

    try {
      setCreating(true);
      setError('');

      const response = await fetch('/api/v1/platform/policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(newPolicy)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create policy');
      }

      const result = await response.json();
      console.log('[PoliciesListTab] Policy created successfully:', result);
      
      setShowCreateDialog(false);
      setNewPolicy({ name: '', description: '' });
      
      // Navigate to the new policy details page
      if (result.data?.id) {
        navigate(`/policies/${result.data.id}`);
      } else {
        loadPolicies();
      }
    } catch (err) {
      console.error('[PoliciesListTab] Error creating policy:', err);
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading policies...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Policies</h1>
          <p className="text-muted-foreground mt-2">
            Manage policies and their permissions
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search policies..."
              value={searchQuery}
              onChange={(e) => {
                console.log('[PoliciesListTab] Search query changed:', e.target.value);
                setSearchQuery(e.target.value);
              }}
              className="pl-9"
            />
          </div>
          <Button onClick={() => setShowCreateDialog(true)} className="gap-2">
            <Plus className="h-4 w-4" />
            Create Policy
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

      {/* Policies Grid */}
      {filteredPolicies.length === 0 ? (
        <Card>
          <CardContent className="pt-12 pb-12">
            <div className="text-center text-muted-foreground">
              <Shield className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p className="text-sm">No policies found</p>
              <p className="text-xs mt-1">
                {searchQuery ? 'Try adjusting your search' : 'Create your first policy to get started'}
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredPolicies.map((policy) => (
            <Card
              key={policy.id}
              className="cursor-pointer hover:border-primary/50 transition-all hover:shadow-md"
              onClick={() => {
                console.log('[PoliciesListTab] Policy card clicked:', policy.id);
                navigate(`/policies/${policy.id}`);
              }}
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <Shield className="h-8 w-8 text-primary" />
                  <Badge variant="outline">
                    {policy.permissions?.length || 0} permissions
                  </Badge>
                </div>
                <CardTitle className="mt-4">{policy.name}</CardTitle>
                <CardDescription className="line-clamp-2">
                  {policy.description || 'No description'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center text-sm text-muted-foreground gap-2">
                  <Lock className="h-4 w-4" />
                  <span>
                    {policy.relations_count || 0} {policy.relations_count === 1 ? 'relation' : 'relations'}
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Policy Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent onClose={() => setShowCreateDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create Policy</DialogTitle>
            <DialogDescription>
              Define a new policy with permissions
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="policy-name">Policy Name <span className="text-red-500">*</span></Label>
              <Input
                id="policy-name"
                placeholder="e.g., Content Editor"
                value={newPolicy.name}
                onChange={(e) => setNewPolicy(prev => ({ ...prev, name: e.target.value }))}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="policy-description">Description</Label>
              <Textarea
                id="policy-description"
                placeholder="Describe what this policy can do..."
                value={newPolicy.description}
                onChange={(e) => setNewPolicy(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={handleCreatePolicy} disabled={creating}>
              {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create & Configure
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

