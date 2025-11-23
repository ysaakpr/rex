import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Building2, Shield, Key, Users, UserCircle, LogOut, Package } from 'lucide-react';
import { cn } from '../../lib/utils';
import { Button } from '../ui/button';
import { Separator } from '../ui/separator';

const sidebarItems = [
  {
    title: 'Tenants',
    icon: Building2,
    path: '/tenants',
  },
  {
    title: 'Users',
    icon: UserCircle,
    path: '/users',
  },
  {
    title: 'Applications',
    icon: Package,
    path: '/applications',
  },
  {
    title: 'Policies',
    icon: Key,
    path: '/permissions',
  },
  {
    title: 'Roles',
    icon: Shield,
    path: '/roles',
  },
];

export function Sidebar({ onSignOut }) {
  const navigate = useNavigate();
  const location = useLocation();

  const isActive = (path) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  return (
    <div className="flex h-full w-64 flex-col border-r bg-card">
      {/* Logo/Brand */}
      <div className="p-6">
        <div className="flex items-center gap-2">
          <span className="text-3xl">🦖</span>
          <h1 className="text-2xl font-bold text-primary">Rex</h1>
        </div>
        <p className="text-sm text-muted-foreground mt-1">Admin Dashboard</p>
      </div>

      <Separator />

      {/* Navigation Items */}
      <nav className="flex-1 space-y-1 p-4">
        {sidebarItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.path);
          
          return (
            <Button
              key={item.path}
              variant={active ? "secondary" : "ghost"}
              className={cn(
                "w-full justify-start",
                active && "bg-secondary font-semibold"
              )}
              onClick={() => navigate(item.path)}
            >
              <Icon className="mr-3 h-5 w-5" />
              {item.title}
            </Button>
          );
        })}
      </nav>

      <Separator />

      {/* Sign Out */}
      <div className="p-4">
        <Button
          variant="ghost"
          className="w-full justify-start text-destructive hover:text-destructive hover:bg-destructive/10"
          onClick={onSignOut}
        >
          <LogOut className="mr-3 h-5 w-5" />
          Sign Out
        </Button>
      </div>
    </div>
  );
}

