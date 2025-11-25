# React Setup

Complete guide to setting up the React frontend application.

## Prerequisites

- Node.js 18+ installed
- npm or yarn package manager
- Basic React knowledge

## Quick Start

### 1. Install Dependencies

```bash
cd frontend/
npm install
```

### 2. Configure Environment

```bash
# Create .env file
cp .env.example .env
```

Edit `.env`:
```bash
VITE_API_URL=http://localhost:8080
VITE_APP_NAME=Rex
```

### 3. Start Development Server

```bash
npm run dev
```

Application will be available at `http://localhost:3000`

## Project Structure

```
frontend/
├── src/
│   ├── components/        # React components
│   │   ├── auth/         # Authentication components
│   │   ├── tenant/       # Tenant management
│   │   ├── members/      # Member management
│   │   └── common/       # Shared components
│   ├── lib/              # Utilities and helpers
│   ├── styles/           # CSS styles
│   ├── App.jsx           # Main app component
│   └── main.jsx          # Entry point
├── public/               # Static assets
├── index.html            # HTML template
├── vite.config.js       # Vite configuration
└── package.json          # Dependencies
```

## Core Dependencies

### SuperTokens

```json
{
  "supertokens-auth-react": "^0.35.0"
}
```

**Purpose**: Authentication (email/password, Google OAuth, session management)

### React Router

```json
{
  "react-router-dom": "^6.20.0"
}
```

**Purpose**: Client-side routing

### Tailwind CSS

```json
{
  "tailwindcss": "^3.3.0"
}
```

**Purpose**: Utility-first CSS framework

## SuperTokens Setup

### Initialize in Entry Point

```jsx
// src/main.jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import SuperTokens from 'supertokens-auth-react';
import EmailPassword from 'supertokens-auth-react/recipe/emailpassword';
import ThirdParty from 'supertokens-auth-react/recipe/thirdparty';
import Session from 'supertokens-auth-react/recipe/session';
import App from './App';
import './index.css';

SuperTokens.init({
  appInfo: {
    appName: "Rex",
    apiDomain: import.meta.env.VITE_API_URL || "http://localhost:8080",
    websiteDomain: window.location.origin,
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    ThirdParty.init({
      signInAndUpFeature: {
        providers: [
          ThirdParty.Google.init()
        ]
      }
    }),
    Session.init()
  ]
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

## Routing Setup

### Main App Router

```jsx
// src/App.jsx
import {BrowserRouter, Routes, Route, Navigate} from 'react-router-dom';
import {SessionAuth} from 'supertokens-auth-react/recipe/session';
import * as SuperTokensConfig from 'supertokens-auth-react/ui';
import EmailPasswordPrebuiltUI from 'supertokens-auth-react/recipe/emailpassword/prebuiltui';
import ThirdPartyPrebuiltUI from 'supertokens-auth-react/recipe/thirdparty/prebuiltui';

// Pages
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import TenantView from './pages/TenantView';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* SuperTokens auth routes */}
        {SuperTokensConfig.getSuperTokensRoutesForReactRouterDom(require("react-router-dom"), [
          EmailPasswordPrebuiltUI,
          ThirdPartyPrebuiltUI
        ])}
        
        {/* Public routes */}
        <Route path="/" element={<Home />} />
        
        {/* Protected routes */}
        <Route path="/dashboard" element={
          <SessionAuth>
            <Dashboard />
          </SessionAuth>
        } />
        
        <Route path="/tenant/:slug" element={
          <SessionAuth>
            <TenantView />
          </SessionAuth>
        } />
        
        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
```

## Component Organization

### Layout Component

```jsx
// src/components/common/Layout.jsx
import {useNavigate} from 'react-router-dom';
import {signOut} from 'supertokens-auth-react/recipe/session';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';

export default function Layout({children}) {
  const navigate = useNavigate();
  const session = useSessionContext();
  
  const handleSignOut = async () => {
    await signOut();
    navigate('/');
  };
  
  if (session.loading) {
    return <div>Loading...</div>;
  }
  
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <h1 className="text-xl font-bold">Rex</h1>
            </div>
            {session.doesSessionExist && (
              <button onClick={handleSignOut} className="btn-primary">
                Sign Out
              </button>
            )}
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  );
}
```

### Protected Page Example

```jsx
// src/pages/Dashboard.jsx
import {useEffect, useState} from 'react';
import {useSessionContext} from 'supertokens-auth-react/recipe/session';
import Layout from '../components/common/Layout';

export default function Dashboard() {
  const session = useSessionContext();
  const [user, setUser] = useState(null);
  const [tenants, setTenants] = useState([]);
  
  useEffect(() => {
    if (session.doesSessionExist) {
      // Fetch user details
      fetch('/api/v1/users/me', {credentials: 'include'})
        .then(res => res.json())
        .then(data => setUser(data.data));
      
      // Fetch user's tenants
      fetch(`/api/v1/users/${session.userId}/tenants`, {credentials: 'include'})
        .then(res => res.json())
        .then(data => setTenants(data.data));
    }
  }, [session]);
  
  return (
    <Layout>
      <div>
        <h1 className="text-3xl font-bold mb-6">Dashboard</h1>
        
        {user && (
          <div className="bg-white p-6 rounded-lg shadow mb-6">
            <h2 className="text-xl font-semibold mb-2">Profile</h2>
            <p>Email: {user.email}</p>
            <p>Verified: {user.email_verified ? 'Yes' : 'No'}</p>
          </div>
        )}
        
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Your Tenants</h2>
          {tenants.length === 0 ? (
            <p>No tenants yet</p>
          ) : (
            <ul className="space-y-2">
              {tenants.map(t => (
                <li key={t.tenant_id} className="p-4 border rounded hover:bg-gray-50">
                  <a href={`/tenant/${t.tenant_slug}`}>
                    <div className="font-medium">{t.tenant_name}</div>
                    <div className="text-sm text-gray-600">Role: {t.role_name}</div>
                  </a>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </Layout>
  );
}
```

## Styling with Tailwind

### Configure Tailwind

```javascript
// tailwind.config.js
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#3B82F6',
        secondary: '#10B981',
      }
    },
  },
  plugins: [],
}
```

### Global Styles

```css
/* src/index.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer components {
  .btn-primary {
    @apply px-4 py-2 bg-primary text-white rounded hover:bg-blue-600 transition;
  }
  
  .btn-secondary {
    @apply px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 transition;
  }
  
  .card {
    @apply bg-white p-6 rounded-lg shadow;
  }
}
```

## API Integration

### API Helper

```javascript
// src/lib/api.js
export async function apiCall(endpoint, options = {}) {
  const response = await fetch(`${import.meta.env.VITE_API_URL}${endpoint}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  });
  
  if (!response.ok) {
    if (response.status === 401) {
      window.location.href = '/auth';
      throw new Error('Session expired');
    }
    const error = await response.json();
    throw new Error(error.error || 'API error');
  }
  
  return response.json();
}
```

### Custom Hook

```jsx
// src/lib/hooks/useAPI.js
import {useState, useEffect} from 'react';
import {apiCall} from '../api';

export function useAPI(endpoint, options = {}) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    apiCall(endpoint, options)
      .then(result => {
        setData(result.data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [endpoint]);
  
  return {data, loading, error};
}

// Usage
function TenantList() {
  const {data: tenants, loading, error} = useAPI('/api/v1/tenants');
  
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  
  return (
    <ul>
      {tenants.data.map(tenant => (
        <li key={tenant.id}>{tenant.name}</li>
      ))}
    </ul>
  );
}
```

## Development Tools

### VS Code Extensions

Recommended extensions:
- ES7+ React/Redux/React-Native snippets
- Tailwind CSS IntelliSense
- ESLint
- Prettier

### ESLint Configuration

```javascript
// .eslintrc.cjs
module.exports = {
  env: {browser: true, es2020: true},
  extends: [
    'eslint:recommended',
    'plugin:react/recommended',
    'plugin:react/jsx-runtime',
    'plugin:react-hooks/recommended',
  ],
  parserOptions: {ecmaVersion: 'latest', sourceType: 'module'},
  settings: {react: {version: '18.2'}},
  plugins: ['react-refresh'],
  rules: {
    'react-refresh/only-export-components': 'warn',
  },
}
```

## Building for Production

### Build Command

```bash
npm run build
```

Output in `dist/` folder.

### Preview Build

```bash
npm run preview
```

### Environment Variables

```bash
# .env.production
VITE_API_URL=https://api.yourdomain.com
VITE_APP_NAME=Rex
```

## Troubleshooting

### Hot Reload Not Working

Solution: Enable polling in Vite config
```javascript
// vite.config.js
export default {
  server: {
    watch: {
      usePolling: true
    }
  }
}
```

### CORS Errors

Solution: Check backend CORS configuration
```go
// Backend must allow your frontend origin
AllowOrigins: []string{"http://localhost:3000"}
```

### Build Failures

```bash
# Clear and reinstall
rm -rf node_modules package-lock.json
npm install
npm run build
```

## Next Steps

- [Protected Routes](/frontend/protected-routes) - Route guards
- [API Calls](/frontend/api-calls) - API integration
- [Component Examples](/frontend/component-examples) - Common components
- [Frontend Integration](/guides/frontend-integration) - Complete guide
