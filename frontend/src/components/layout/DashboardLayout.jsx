import React, { useState, useEffect } from 'react';
import { Sidebar } from './Sidebar';
import { TopNav } from './TopNav';
import { signOut } from "supertokens-auth-react/recipe/emailpassword";
import { useNavigate } from 'react-router-dom';

export function DashboardLayout({ children }) {
  const navigate = useNavigate();
  const [userInfo, setUserInfo] = useState(null);
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    loadUserInfo();
    checkPlatformAdmin();
  }, []);

  const loadUserInfo = async () => {
    try {
      // Get current user info from backend
      const response = await fetch('/api/v1/users/me', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success && data.data) {
          setUserInfo({ 
            userId: data.data.user_id || data.data.id,
            email: data.data.email,
            name: data.data.name,
          });
        }
      }
    } catch (err) {
      console.error('Error loading user info:', err);
    }
  };

  const checkPlatformAdmin = async () => {
    try {
      const response = await fetch('/api/v1/platform/admins/check', {
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setIsPlatformAdmin(data.data?.is_platform_admin || false);
      }
    } catch (err) {
      console.error('Error checking platform admin:', err);
    }
  };

  const handleSignOut = async () => {
    await signOut();
    // Use window.location to force full page reload and go to auth page
    window.location.href = '/auth';
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <Sidebar onSignOut={handleSignOut} />

      {/* Main Content Area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Top Navigation */}
        <TopNav 
          userInfo={userInfo} 
          isPlatformAdmin={isPlatformAdmin} 
          onSignOut={handleSignOut} 
        />

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>
    </div>
  );
}

