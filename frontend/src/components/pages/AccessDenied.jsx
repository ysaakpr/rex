import React from 'react';
import { ShieldAlert, Home, LogOut } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { signOut } from "supertokens-auth-react/recipe/emailpassword";
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '../ui/card';

export function AccessDenied() {
  const navigate = useNavigate();

  const handleSignOut = async () => {
    await signOut();
    window.location.href = '/auth';
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 p-4">
      <Card className="w-full max-w-md text-center shadow-2xl">
        <CardHeader className="space-y-4">
          {/* Fun Animated Icon */}
          <div className="flex justify-center">
            <div className="relative">
              <div className="absolute inset-0 animate-ping rounded-full bg-red-400 opacity-20"></div>
              <div className="relative flex h-24 w-24 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
                <ShieldAlert className="h-12 w-12 text-red-600 dark:text-red-400" />
              </div>
            </div>
          </div>
          
          <div>
            <CardTitle className="text-3xl font-bold">
              Oops! Platform Admins Only! 
            </CardTitle>
            <div className="mt-2 text-6xl">
              😅🚫🎭
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="space-y-4">
          <CardDescription className="text-base leading-relaxed">
            This area is exclusively for platform administrators. 
            It seems you've wandered into the VIP lounge!
          </CardDescription>
          
          <div className="rounded-lg bg-muted p-4 text-sm">
            <p className="font-semibold text-foreground mb-2">
              Need access?
            </p>
            <p className="text-muted-foreground">
              Contact your system administrator if you believe you should have access to this area. 
              They can grant you the necessary permissions.
            </p>
          </div>
        </CardContent>
        
        <CardFooter className="flex flex-col gap-3">
          <div className="flex gap-3">
            <Button onClick={handleSignOut} variant="default" className="gap-2">
              <LogOut className="h-4 w-4" />
              Sign Out
            </Button>
            <Button 
              variant="outline" 
              onClick={() => window.location.href = 'mailto:admin@example.com?subject=Platform Admin Access Request'}
            >
              Request Access
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Sign out to try a different account or request admin access
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}

