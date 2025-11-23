import React from 'react';
import { User, Settings, LogOut, Crown } from 'lucide-react';
import { Badge } from '../ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';
import { Button } from '../ui/button';

export function TopNav({ userInfo, isPlatformAdmin, onSignOut }) {
  return (
    <div className="flex h-16 items-center justify-between border-b bg-background px-6">
      {/* Left side - can add breadcrumbs or page title later */}
      <div className="flex items-center gap-4">
        {isPlatformAdmin && (
          <Badge variant="default" className="gap-1">
            <Crown className="h-3 w-3" />
            Platform Admin
          </Badge>
        )}
      </div>

      {/* Right side - User profile */}
      <div className="flex items-center gap-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-10 gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
                <User className="h-4 w-4 text-primary" />
              </div>
              <div className="flex flex-col items-start text-left">
                <span className="text-sm font-medium">
                  {userInfo?.email || 'User'}
                </span>
                <span className="text-xs text-muted-foreground">
                  {userInfo?.userId ? `ID: ${userInfo.userId.substring(0, 8)}...` : ''}
                </span>
              </div>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem disabled>
              <User className="mr-2 h-4 w-4" />
              <span>Profile</span>
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onSignOut} className="text-destructive focus:text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              <span>Sign out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

