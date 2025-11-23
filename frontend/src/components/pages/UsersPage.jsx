import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, User, Loader2, Filter, CheckCircle, XCircle, X, Cpu } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Badge } from '../ui/badge';
import { Pagination, PaginationInfo } from '../ui/pagination';

export function UsersPage() {
  console.log('[UsersPage] Component mounted');
  
  const navigate = useNavigate();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [emailFilter, setEmailFilter] = useState('');
  const [nameFilter, setNameFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);
  const filterDropdownRef = useRef(null);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  useEffect(() => {
    loadUsers();
  }, [currentPage, pageSize, emailFilter, statusFilter]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (filterDropdownRef.current && !filterDropdownRef.current.contains(event.target)) {
        setShowFilterDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const loadUsers = async () => {
    console.log('[UsersPage] Loading users...', { currentPage, pageSize, emailFilter, statusFilter });
    try {
      setLoading(true);
      setError('');
      
      // Build query parameters
      const params = new URLSearchParams({
        page: currentPage.toString(),
        page_size: pageSize.toString(),
      });
      
      if (emailFilter) {
        params.append('email', emailFilter);
      }
      
      // Note: Status filter would need backend support
      // For now, we'll handle it client-side if needed
      
      const response = await fetch(`/api/v1/users?${params.toString()}`, {
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Failed to load users: ${response.status}`);
      }

      const result = await response.json();
      console.log('[UsersPage] Users loaded:', result);
      
      // Extract pagination data
      const data = result.data || {};
      const usersArray = data.data || [];
      
      setUsers(Array.isArray(usersArray) ? usersArray : []);
      setTotalCount(data.total_count || 0);
      setTotalPages(data.total_pages || 0);
      setCurrentPage(data.page || 1);
    } catch (err) {
      console.error('[UsersPage] Error loading users:', err);
      setError(err.message);
      // Set empty array on error so UI doesn't break
      setUsers([]);
      setTotalCount(0);
      setTotalPages(0);
    } finally {
      setLoading(false);
    }
  };

  const hasActiveFilters = () => {
    return emailFilter !== '' || nameFilter !== '' || statusFilter !== 'all';
  };

  const clearAllFilters = () => {
    setEmailFilter('');
    setNameFilter('');
    setStatusFilter('all');
    setCurrentPage(1); // Reset to first page
  };

  const removeFilter = (filterType) => {
    switch(filterType) {
      case 'email':
        setEmailFilter('');
        break;
      case 'name':
        setNameFilter('');
        break;
      case 'status':
        setStatusFilter('all');
        break;
    }
    setCurrentPage(1); // Reset to first page when removing filter
  };

  const handlePageChange = (page) => {
    console.log('[UsersPage] Page changed to:', page);
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleFilterChange = (filterType, value) => {
    switch(filterType) {
      case 'email':
        setEmailFilter(value);
        break;
      case 'name':
        setNameFilter(value);
        break;
      case 'status':
        setStatusFilter(value);
        break;
    }
    setCurrentPage(1); // Reset to first page on filter change
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="ml-3 text-muted-foreground">Loading users...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground mt-2">
            {totalCount} total {totalCount === 1 ? 'user' : 'users'}
          </p>
        </div>
        
        {/* Filter Button - Moved to top right */}
        <div className="relative" ref={filterDropdownRef}>
          <Button
            variant="outline"
            size="default"
            onClick={() => setShowFilterDropdown(!showFilterDropdown)}
            className="relative"
          >
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {hasActiveFilters() && (
              <span className="absolute -top-1 -right-1 h-3 w-3 bg-red-500 rounded-full border-2 border-background"></span>
            )}
          </Button>

          {/* Dropdown Menu */}
          {showFilterDropdown && (
            <div className="absolute right-0 top-full mt-2 w-80 bg-background border rounded-lg shadow-lg z-50 p-4">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email-filter">Filter by Email</Label>
                  <div className="relative">
                    <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="email-filter"
                      placeholder="user@example.com"
                      value={emailFilter}
                      onChange={(e) => handleFilterChange('email', e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="name-filter">Filter by Name</Label>
                  <Input
                    id="name-filter"
                    placeholder="Search names..."
                    value={nameFilter}
                    onChange={(e) => handleFilterChange('name', e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="status-filter">Status</Label>
                  <select
                    id="status-filter"
                    value={statusFilter}
                    onChange={(e) => handleFilterChange('status', e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="all">All Users</option>
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-sm text-destructive">{error}</p>
            <p className="text-xs text-muted-foreground mt-2">
              Note: User listing requires SuperTokens Core API integration
            </p>
          </CardContent>
        </Card>
      )}

      {/* Active Filter Chips */}
      {hasActiveFilters() && (
      <div className="flex items-center gap-3 flex-wrap">
          {emailFilter && (
            <Badge variant="secondary" className="gap-1 px-3 py-1">
              Email: {emailFilter}
              <X 
                className="h-3 w-3 cursor-pointer hover:text-destructive" 
                onClick={() => removeFilter('email')}
              />
            </Badge>
          )}
          {nameFilter && (
            <Badge variant="secondary" className="gap-1 px-3 py-1">
              Name: {nameFilter}
              <X 
                className="h-3 w-3 cursor-pointer hover:text-destructive" 
                onClick={() => removeFilter('name')}
              />
            </Badge>
          )}
          {statusFilter !== 'all' && (
            <Badge variant="secondary" className="gap-1 px-3 py-1">
              Status: {statusFilter}
              <X 
                className="h-3 w-3 cursor-pointer hover:text-destructive" 
                onClick={() => removeFilter('status')}
              />
            </Badge>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={clearAllFilters}
            className="h-7 text-xs"
          >
            Clear all
          </Button>
      </div>
      )}

      {/* Pagination Info */}
      {!loading && totalCount > 0 && (
        <PaginationInfo
          currentPage={currentPage}
          pageSize={pageSize}
          totalCount={totalCount}
        />
      )}

      {/* Users List */}
      <div className="space-y-3">
        {users.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground border rounded-lg bg-muted/20">
            <User className="mx-auto h-12 w-12 mb-4 opacity-50" />
            <p className="text-sm">No users found</p>
            <p className="text-xs mt-1">
              {emailFilter || nameFilter ? 'Try adjusting your filters' : 'No users registered yet'}
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {users.map((user) => (
              <div
                key={user.user_id || user.id}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 cursor-pointer transition-all bg-background"
                onClick={() => {
                  console.log('[UsersPage] User clicked:', user.user_id || user.id);
                  navigate(`/users/${user.user_id || user.id}`);
                }}
              >
                <div className="flex items-center gap-4 flex-1">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                    <User className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium">
                        {user.name || user.email?.split('@')[0] || 'Unknown User'}
                      </span>
                      {user.is_system && (
                        <Badge variant="secondary" className="gap-1 bg-purple-100 text-purple-700 border-purple-200">
                          <Cpu className="h-3 w-3" />
                          System
                        </Badge>
                      )}
                      {user.is_active !== false ? (
                        <Badge variant="outline" className="gap-1">
                          <CheckCircle className="h-3 w-3" />
                          Active
                        </Badge>
                      ) : (
                        <Badge variant="destructive" className="gap-1">
                          <XCircle className="h-3 w-3" />
                          Inactive
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">{user.email}</p>
                  </div>
                </div>
                <div className="text-sm text-muted-foreground">
                  {user.tenant_count > 0 && (
                    <span>{user.tenant_count} {user.tenant_count === 1 ? 'tenant' : 'tenants'}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Pagination Controls */}
      {!loading && totalPages > 0 && (
        <div className="flex flex-col items-center gap-4 pt-4">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
          <PaginationInfo
            currentPage={currentPage}
            pageSize={pageSize}
            totalCount={totalCount}
          />
        </div>
      )}
    </div>
  );
}

