import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, User, Loader2, Filter, CheckCircle, XCircle, X, Cpu, Eye, EyeOff } from 'lucide-react';
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
  const [initialLoading, setInitialLoading] = useState(true);
  const [filterLoading, setFilterLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [emailFilter, setEmailFilter] = useState('');
  const [nameFilter, setNameFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [showSystemUsers, setShowSystemUsers] = useState(false);
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);
  const filterDropdownRef = useRef(null);
  const searchTimeoutRef = useRef(null);
  const isFirstLoad = useRef(true);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  useEffect(() => {
    // Debounce search query
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }
    
    searchTimeoutRef.current = setTimeout(() => {
      loadUsers();
    }, 300); // 300ms debounce
    
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [currentPage, pageSize, searchQuery, showSystemUsers]);

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

  // Helper function to check if string is a valid UUID
  const isUUID = (str) => {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(str);
  };

  // Helper function to check if string is an email
  const isEmail = (str) => {
    return str.includes('@');
  };

  const loadUsers = async () => {
    console.log('[UsersPage] Loading users...', { currentPage, pageSize, searchQuery, showSystemUsers });
    try {
      // Show appropriate loading indicator
      if (isFirstLoad.current) {
        setInitialLoading(true);
      } else {
        setFilterLoading(true);
      }
      setError('');
      
      // Build query parameters
      const params = new URLSearchParams({
        page: currentPage.toString(),
        page_size: pageSize.toString(),
      });
      
      // Smart search logic
      if (searchQuery.trim()) {
        const query = searchQuery.trim();
        
        if (isEmail(query)) {
          // Email search
          params.append('email', query);
          console.log('[UsersPage] Searching by email:', query);
        } else if (isUUID(query)) {
          // UUID/ID search
          params.append('user_id', query);
          console.log('[UsersPage] Searching by ID:', query);
        } else {
          // Name search (fallback)
          params.append('name', query);
          console.log('[UsersPage] Searching by name:', query);
        }
      }
      
      // System user filter
      if (!showSystemUsers) {
        params.append('exclude_system', 'true');
      }
      
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
      
      // Mark first load as complete
      if (isFirstLoad.current) {
        isFirstLoad.current = false;
      }
    } catch (err) {
      console.error('[UsersPage] Error loading users:', err);
      setError(err.message);
      // Set empty array on error so UI doesn't break
      setUsers([]);
      setTotalCount(0);
      setTotalPages(0);
    } finally {
      setInitialLoading(false);
      setFilterLoading(false);
    }
  };

  const hasActiveFilters = () => {
    return searchQuery !== '' || !showSystemUsers;
  };

  const clearAllFilters = () => {
    setSearchQuery('');
    setShowSystemUsers(false);
    setCurrentPage(1); // Reset to first page
  };

  const handleSearchChange = (value) => {
    setSearchQuery(value);
    setCurrentPage(1); // Reset to first page on search
  };

  const handlePageChange = (page) => {
    console.log('[UsersPage] Page changed to:', page);
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  if (initialLoading) {
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
            {!showSystemUsers && ' (excluding system users)'}
          </p>
        </div>
        
        {/* Search and Filters */}
        <div className="flex items-center gap-3">
          {/* Smart Search Input */}
          <div className="relative w-96">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search by name, email, or ID..."
              value={searchQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
              className="pl-10 pr-10"
            />
            {filterLoading ? (
              <Loader2 className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 animate-spin text-primary" />
            ) : searchQuery ? (
              <X 
                className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 cursor-pointer text-muted-foreground hover:text-destructive" 
                onClick={() => handleSearchChange('')}
              />
            ) : null}
          </div>

          {/* System Users Toggle */}
          <Button
            variant={showSystemUsers ? "default" : "outline"}
            size="default"
            onClick={() => {
              setShowSystemUsers(!showSystemUsers);
              setCurrentPage(1);
            }}
            className="gap-2 min-w-[180px]"
          >
            {showSystemUsers ? (
              <>
                <Eye className="h-4 w-4" />
                Hide System Users
              </>
            ) : (
              <>
                <EyeOff className="h-4 w-4" />
                Show System Users
              </>
            )}
          </Button>
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
          {searchQuery && (
            <Badge variant="secondary" className="gap-1 px-3 py-1">
              <Search className="h-3 w-3" />
              Search: {searchQuery}
              <X 
                className="h-3 w-3 cursor-pointer hover:text-destructive" 
                onClick={() => handleSearchChange('')}
              />
            </Badge>
          )}
          {!showSystemUsers && (
            <Badge variant="secondary" className="gap-1 px-3 py-1">
              <EyeOff className="h-3 w-3" />
              System users hidden
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
      {!initialLoading && totalCount > 0 && (
        <PaginationInfo
          currentPage={currentPage}
          pageSize={pageSize}
          totalCount={totalCount}
        />
      )}

      {/* Users List */}
      <div className="space-y-3 relative">
        {/* Subtle loading overlay when filtering */}
        {filterLoading && (
          <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 rounded-lg flex items-center justify-center">
            <div className="flex items-center gap-2 bg-background border rounded-lg px-4 py-2 shadow-lg">
              <Loader2 className="h-4 w-4 animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">Updating results...</span>
            </div>
          </div>
        )}
        
        {users.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground border rounded-lg bg-muted/20">
            <User className="mx-auto h-12 w-12 mb-4 opacity-50" />
            <p className="text-sm">No users found</p>
            <p className="text-xs mt-1">
              {searchQuery ? 'Try adjusting your search query' : 'No users registered yet'}
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
      {!initialLoading && totalPages > 0 && (
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

