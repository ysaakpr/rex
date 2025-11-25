import React, { useState, useEffect } from 'react';
import SuperTokens, { SuperTokensWrapper } from "supertokens-auth-react";
import { getSuperTokensRoutesForReactRouterDom } from "supertokens-auth-react/ui";
import { EmailPasswordPreBuiltUI } from "supertokens-auth-react/recipe/emailpassword/prebuiltui";
import { ThirdPartyPreBuiltUI } from "supertokens-auth-react/recipe/thirdparty/prebuiltui";
import { SessionAuth } from "supertokens-auth-react/recipe/session";
import * as reactRouterDom from "react-router-dom";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";
import ThirdParty from "supertokens-auth-react/recipe/thirdparty";
import Session from "supertokens-auth-react/recipe/session";
import { Loader2 } from 'lucide-react';
import { DashboardLayout } from './components/layout/DashboardLayout';
import { AccessDenied } from './components/pages/AccessDenied';
import { TenantsPage } from './components/pages/TenantsPage';
import { ManagedTenantOnboarding } from './components/pages/ManagedTenantOnboarding';
import { TenantDetailsPage } from './components/pages/TenantDetailsPage';
import { TenantEditPage } from './components/pages/TenantEditPage';
import { PoliciesPage } from './components/pages/PoliciesPage';
import { PolicyDetailsPage } from './components/pages/PolicyDetailsPage';
import { RolesPage } from './components/pages/RolesPage';
import { RoleDetailsPage } from './components/pages/RoleDetailsPage';
import { UsersPage } from './components/pages/UsersPage';
import { UserDetailsPage } from './components/pages/UserDetailsPage';
import { ApplicationsPage } from './components/pages/ApplicationsPage';
import { AcceptInvitationPage } from './components/pages/AcceptInvitationPage';
import './index.css';
import appConfig from './config';

// Initialize SuperTokens with auth config (will be called dynamically)
let authConfig = null;
let superTokensInitialized = false;

function initializeSuperTokens(authProviderConfig) {
  if (superTokensInitialized) return;
  
  console.log('[SuperTokens] Initializing with config:', {
    apiDomain: appConfig.apiDomain,
    websiteDomain: appConfig.websiteDomain,
    authPath: appConfig.authPath,
    apiBasePath: appConfig.apiBasePath,
  });
  
  const recipeList = [];
  
  // Add ThirdParty recipe only if Google OAuth is enabled
  if (authProviderConfig?.providers?.google) {
    console.log('[SuperTokens] Google OAuth enabled - adding ThirdParty recipe');
    recipeList.push(
      ThirdParty.init({
        signInAndUpFeature: {
          providers: [
            ThirdParty.Google.init(),
          ],
        },
      })
    );
  } else {
    console.log('[SuperTokens] Google OAuth disabled - skipping ThirdParty recipe');
  }
  
  // Always add EmailPassword and Session
  recipeList.push(EmailPassword.init());
  recipeList.push(Session.init({
    sessionExpiredStatusCode: 401,
  }));
  
  SuperTokens.init({
    appInfo: {
      appName: appConfig.appName,
      apiDomain: appConfig.apiDomain || window.location.origin,
      websiteDomain: appConfig.websiteDomain || window.location.origin,
      apiBasePath: appConfig.apiBasePath,
      websiteBasePath: appConfig.authPath
    },
    recipeList: recipeList
  });
  
  superTokensInitialized = true;
}

// Component that checks platform admin status
function ProtectedDashboard({ children }) {
  const [loading, setLoading] = useState(true);
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    checkPlatformAdmin();
    checkAndAcceptPendingInvitations();
  }, []);

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
      setIsPlatformAdmin(false);
    } finally {
      setLoading(false);
    }
  };

  const checkAndAcceptPendingInvitations = async () => {
    try {
      // This endpoint will check if the user's email has any pending invitations
      // and automatically accept them
      const response = await fetch('/api/v1/invitations/check-pending', {
        method: 'POST',
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        if (data.data?.accepted_count > 0) {
          console.log(`Auto-accepted ${data.data.accepted_count} pending invitation(s)`);
          // Optionally show a notification to the user
        }
      }
    } catch (err) {
      // Silent fail - this is a background check
      console.error('Error checking pending invitations:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isPlatformAdmin) {
    return <AccessDenied />;
  }

  return <DashboardLayout>{children}</DashboardLayout>;
}

function AppContent({ authConfig }) {
  // Determine which auth UIs to show based on config
  const authUIs = authConfig?.providers?.google 
    ? [ThirdPartyPreBuiltUI, EmailPasswordPreBuiltUI]
    : [EmailPasswordPreBuiltUI];
  
  return (
    <SuperTokensWrapper>
      <BrowserRouter basename={appConfig.basename}>
        <Routes>
          {/* SuperTokens auth routes */}
          {getSuperTokensRoutesForReactRouterDom(reactRouterDom, authUIs)}
          
          {/* Public invitation routes - accessible before authentication */}
          <Route
            path="/invitations/:token/accept"
            element={<AcceptInvitationPage />}
          />
          <Route
            path="/accept-invite"
            element={<AcceptInvitationPage />}
          />
          
          {/* Protected routes - require authentication + platform admin */}
          <Route
            path="/tenants"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <TenantsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/tenants/create"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <ManagedTenantOnboarding />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/tenants/:id/edit"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <TenantEditPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/tenants/:id"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <TenantDetailsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/permissions"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <PoliciesPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/policies/:id"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <PolicyDetailsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/roles"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <RolesPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/roles/:id"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <RoleDetailsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/users"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <UsersPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/users/:id"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <UserDetailsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/applications"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <ApplicationsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          {/* Default route - redirect to tenants if authenticated, otherwise to auth */}
          <Route
            path="/"
            element={<Navigate to="/tenants" replace />}
          />
        </Routes>
      </BrowserRouter>
    </SuperTokensWrapper>
  );
}

// Main App component that fetches auth config before initializing SuperTokens
function App() {
  const [loading, setLoading] = useState(true);
  const [config, setConfig] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchAuthConfig() {
      try {
        console.log('[App] Fetching auth configuration...');
        const response = await fetch('/api/v1/auth/config');
        
        if (!response.ok) {
          throw new Error(`Failed to fetch auth config: ${response.status}`);
        }
        
        const result = await response.json();
        const authConfig = result.data;
        
        console.log('[App] Auth configuration received:', authConfig);
        setConfig(authConfig);
        
        // Initialize SuperTokens with the fetched config
        initializeSuperTokens(authConfig);
        
      } catch (err) {
        console.error('[App] Error fetching auth config:', err);
        setError(err.message);
        
        // Fallback: Initialize SuperTokens without Google OAuth
        console.log('[App] Falling back to email/password only');
        initializeSuperTokens({ providers: { google: false } });
        setConfig({ providers: { google: false } });
        
      } finally {
        setLoading(false);
      }
    }

    fetchAuthConfig();
  }, []);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Loading authentication...</p>
        </div>
      </div>
    );
  }

  if (error && !config) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="text-center max-w-md p-6 border rounded-lg bg-card">
          <h2 className="text-xl font-semibold text-destructive mb-2">Configuration Error</h2>
          <p className="text-sm text-muted-foreground mb-4">{error}</p>
          <button 
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return <AppContent authConfig={config} />;
}

export default App;
