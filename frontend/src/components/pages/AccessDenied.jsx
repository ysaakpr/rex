import React, { useState, useEffect } from 'react';
import { ShieldAlert, Home, LogOut, User, Mail, Copy, Check } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { signOut } from "supertokens-auth-react/recipe/emailpassword";
import Session from 'supertokens-auth-react/recipe/session';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '../ui/card';

export function AccessDenied() {
  const navigate = useNavigate();
  const [userInfo, setUserInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [copiedField, setCopiedField] = useState(null);

  useEffect(() => {
    loadUserInfo();
  }, []);

  const loadUserInfo = async () => {
    try {
      // Get user ID from session
      const userId = await Session.getUserId();
      
      // Fetch user details from API
      const response = await fetch('/api/v1/users/me', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        setUserInfo({
          userId: userId,
          email: data.data?.email || 'Unknown'
        });
      } else {
        // Fallback to just userId if API fails
        setUserInfo({ userId: userId, email: 'Unknown' });
      }
    } catch (err) {
      console.error('Error loading user info:', err);
      setUserInfo({ userId: 'Unknown', email: 'Unknown' });
    } finally {
      setLoading(false);
    }
  };

  const handleSignOut = async () => {
    await signOut();
    window.location.href = '/auth';
  };

  const handleCopy = async (text, field) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      // Reset copied state after 2 seconds
      setTimeout(() => setCopiedField(null), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
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
          
          {/* Current User Info */}
          <div className="rounded-lg border bg-card p-4 text-sm space-y-3">
            <p className="font-semibold text-foreground mb-2 flex items-center gap-2">
              <User className="h-4 w-4" />
              Your Account Details
            </p>
            
            {loading ? (
              <p className="text-muted-foreground text-center py-2">Loading user info...</p>
            ) : (
              <div className="space-y-3">
                {/* Email Field with Copy */}
                <div className="flex items-start gap-2">
                  <Mail className="h-4 w-4 mt-0.5 text-muted-foreground flex-shrink-0" />
                  <div className="flex-1 min-w-0 text-left">
                    <p className="text-xs text-muted-foreground">Email</p>
                    <div className="flex items-start gap-2">
                      <p className="font-mono text-sm break-all flex-1">{userInfo?.email}</p>
                      <button
                        onClick={() => handleCopy(userInfo?.email, 'email')}
                        className="p-1 hover:bg-muted rounded transition-colors flex-shrink-0"
                        title="Copy email"
                      >
                        {copiedField === 'email' ? (
                          <Check className="h-4 w-4 text-green-600" />
                        ) : (
                          <Copy className="h-4 w-4 text-muted-foreground" />
                        )}
                      </button>
                    </div>
                  </div>
                </div>
                
                {/* User ID Field with Copy */}
                <div className="flex items-start gap-2">
                  <User className="h-4 w-4 mt-0.5 text-muted-foreground flex-shrink-0" />
                  <div className="flex-1 min-w-0 text-left">
                    <p className="text-xs text-muted-foreground">User ID</p>
                    <div className="flex items-start gap-2">
                      <p className="font-mono text-xs break-all text-muted-foreground flex-1">
                        {userInfo?.userId}
                      </p>
                      <button
                        onClick={() => handleCopy(userInfo?.userId, 'userId')}
                        className="p-1 hover:bg-muted rounded transition-colors flex-shrink-0"
                        title="Copy user ID"
                      >
                        {copiedField === 'userId' ? (
                          <Check className="h-4 w-4 text-green-600" />
                        ) : (
                          <Copy className="h-4 w-4 text-muted-foreground" />
                        )}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
          
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

