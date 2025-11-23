import React, { useState, useEffect } from 'react';
import SuperTokens, { SuperTokensWrapper } from "supertokens-auth-react";
import { getSuperTokensRoutesForReactRouterDom } from "supertokens-auth-react/ui";
import { EmailPasswordPreBuiltUI } from "supertokens-auth-react/recipe/emailpassword/prebuiltui";
import { SessionAuth } from "supertokens-auth-react/recipe/session";
import * as reactRouterDom from "react-router-dom";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import EmailPassword from "supertokens-auth-react/recipe/emailpassword";
import Session from "supertokens-auth-react/recipe/session";
import { DashboardLayout } from './components/layout/DashboardLayout';
import { AccessDenied } from './components/pages/AccessDenied';
import { TenantsPage } from './components/pages/TenantsPage';
import { ManagedTenantOnboarding } from './components/pages/ManagedTenantOnboarding';
import { TenantDetailsPage } from './components/pages/TenantDetailsPage';
import { TenantEditPage } from './components/pages/TenantEditPage';
import { RolesPage } from './components/pages/RolesPage';
import { RoleDetailsPage } from './components/pages/RoleDetailsPage';
import { PermissionsPage } from './components/pages/PermissionsPage';
import { RelationsPage } from './components/pages/RelationsPage';
import { RelationDetailsPage } from './components/pages/RelationDetailsPage';
import { UsersPage } from './components/pages/UsersPage';
import { UserDetailsPage } from './components/pages/UserDetailsPage';
import { ApplicationsPage } from './components/pages/ApplicationsPage';
import './index.css';

// SuperTokens configuration
SuperTokens.init({
  appInfo: {
    appName: "UTM Backend",
    apiDomain: window.location.origin,
    websiteDomain: window.location.origin,
    apiBasePath: "/api/auth",  // API calls go to /api/auth (proxied to backend)
    websiteBasePath: "/auth"   // UI pages stay at /auth (handled by React Router)
  },
  recipeList: [
    EmailPassword.init(),
    Session.init({
      sessionExpiredStatusCode: 401,
    })
  ]
});

// Component that checks platform admin status
function ProtectedDashboard({ children }) {
  const [loading, setLoading] = useState(true);
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    checkPlatformAdmin();
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

function App() {
  return (
    <SuperTokensWrapper>
      <BrowserRouter>
        <Routes>
          {/* SuperTokens auth routes */}
          {getSuperTokensRoutesForReactRouterDom(reactRouterDom, [EmailPasswordPreBuiltUI])}
          
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
            path="/permissions"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <PermissionsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/relations"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <RelationsPage />
                </ProtectedDashboard>
              </SessionAuth>
            }
          />
          
          <Route
            path="/relations/:id"
            element={
              <SessionAuth>
                <ProtectedDashboard>
                  <RelationDetailsPage />
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

export default App;
