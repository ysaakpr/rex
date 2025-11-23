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

export function RelationsPage() {
  console.log('[RelationsPage] Component mounted');
  
  const navigate = useNavigate();
  const [relations, setRelations] = useState([]);
  const [filteredRelations, setFilteredRelations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [creating, setCreating] = useState(false);

  const [newRelation, setNewRelation] = useState({
    name: '',
    description: ''
  });

  useEffect(() => {
    loadRelations();
  }, []);

  useEffect(() => {
    filterRelations();
  }, [relations, searchQuery]);

  const loadRelations = async () => {
    console.log('[RelationsPage] Loading relations...');
    try {
      setLoading(true);
      setError('');
      
      const response = await fetch('/api/v1/platform/relations', {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load relations: ${response.status}`);
      }

      const data = await response.json();
      console.log('[RelationsPage] Relations loaded:', data);
      
      const relationsArray = data.data || [];
      setRelations(Array.isArray(relationsArray) ? relationsArray : []);
    } catch (err) {
      console.error('[RelationsPage] Error loading relations:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filterRelations = () => {
    let filtered = [...relations];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(r =>
        r.name?.toLowerCase().includes(query) ||
        r.description?.toLowerCase().includes(query)
      );
    }

    console.log('[RelationsPage] Filtered relations:', filtered.length, 'from', relations.length);
    setFilteredRelations(filtered);
  };

  const handleCreateRelation = async () => {
    console.log('[RelationsPage] Creating relation:', newRelation);
    
    if (!newRelation.name) {
      setError('Relation name is required');
      return;
    }

    try {
      setCreating(true);
      setError('');

      const response = await fetch('/api/v1/platform/relations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(newRelation)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create relation');
      }

      const result = await response.json();
      console.log('[RelationsPage] Relation created successfully:', result);
      
      setShowCreateDialog(false);
      setNewRelation({ name: '', description: '' });
      
      // Navigate to the new relation details page
      if (result.data?.id) {
        navigate(`/relations/${result.data.id}`);
      } else {
        loadRelations();
      }
    } catch (err) {
      console.error('[RelationsPage] Error creating relation:', err);
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading relations...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Relations</h1>
          <p className="text-muted-foreground mt-2">
            Manage tenant member relations and role mappings
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          Create Relation
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
              placeholder="Search relations..."
              value={searchQuery}
              onChange={(e) => {
                console.log('[RelationsPage] Search query changed:', e.target.value);
                setSearchQuery(e.target.value);
              }}
              className="pl-9"
            />
          </div>
        </CardContent>
      </Card>

      {/* Relations Grid */}
      {filteredRelations.length === 0 ? (
        <Card>
          <CardContent className="pt-12 pb-12">
            <div className="text-center text-muted-foreground">
              <Users className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p className="text-sm">No relations found</p>
              <p className="text-xs mt-1">
                {searchQuery ? 'Try adjusting your search' : 'Create your first relation to get started'}
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredRelations.map((relation) => (
            <Card
              key={relation.id}
              className="cursor-pointer hover:border-primary/50 transition-all hover:shadow-md"
              onClick={() => {
                console.log('[RelationsPage] Relation card clicked:', relation.id);
                navigate(`/relations/${relation.id}`);
              }}
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <Users className="h-8 w-8 text-primary" />
                  <Badge variant="outline">
                    {relation.roles?.length || 0} roles
                  </Badge>
                </div>
                <CardTitle className="mt-4">{relation.name}</CardTitle>
                <CardDescription className="line-clamp-2">
                  {relation.description || 'No description'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center text-sm text-muted-foreground gap-2">
                  <Shield className="h-4 w-4" />
                  <span>
                    {relation.roles?.length || 0} {relation.roles?.length === 1 ? 'role' : 'roles'} assigned
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Relation Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent onClose={() => setShowCreateDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create Relation</DialogTitle>
            <DialogDescription>
              Define a new tenant member relation type
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="relation-name">Relation Name <span className="text-red-500">*</span></Label>
              <Input
                id="relation-name"
                placeholder="e.g., Contributor"
                value={newRelation.name}
                onChange={(e) => setNewRelation(prev => ({ ...prev, name: e.target.value }))}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="relation-description">Description</Label>
              <Textarea
                id="relation-description"
                placeholder="Describe this relation type..."
                value={newRelation.description}
                onChange={(e) => setNewRelation(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={handleCreateRelation} disabled={creating}>
              {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create & Configure
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

