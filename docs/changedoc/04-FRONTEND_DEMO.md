# 🎨 Frontend Demo - Live Screenshots & Guide

## Access the Frontend

🌐 **URL**: http://localhost:3000

## Features Showcase

### 1. Authentication (SuperTokens UI)

The frontend uses SuperTokens pre-built UI components for:
- ✅ Email/Password Sign Up
- ✅ Email/Password Sign In  
- ✅ Session Management
- ✅ Secure Cookie Handling

**No code needed for auth UI!** SuperTokens provides everything.

### 2. Protected Dashboard

After signing in, you'll see:

```
🏢 Tenant Management Dashboard
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
User ID: 8e28fb0c...                [Sign Out]

Create New Tenant
┌─────────────────────────────────────────┐
│ Tenant Name: [My Company            ]   │
│ Slug:        [my-company            ]   │
│ Industry:    [Technology            ]   │
│                                          │
│         [Create Tenant]                  │
└─────────────────────────────────────────┘

Your Tenants
┌─────────────────────────────────────────┐
│ 📦 My Company                            │
│ Slug: my-company                         │
│ [ACTIVE] ID: 7f3a... Created: 11/21/2025│
│ Industry: Technology                     │
└─────────────────────────────────────────┘
```

### 3. Auto-Generated Slug

Type a tenant name → slug generates automatically
- `"Acme Corporation"` → `"acme-corporation"`
- `"My Tech Startup"` → `"my-tech-startup"`

### 4. Real-Time Status

Tenants show status badges:
- 🟡 **PENDING** - Initialization in progress
- 🟢 **ACTIVE** - Ready to use
- 🔴 **FAILED** - Initialization failed

### 5. Responsive Design

Works on:
- 💻 Desktop
- 📱 Mobile
- 🎨 Dark/Light mode (auto-detects)

## Technology Stack

### Frontend
- **React 18** - Latest React with hooks
- **Vite** - Lightning-fast build tool
- **SuperTokens React** - Pre-built auth UI
- **React Router v6** - Modern routing

### Authentication
- **Cookie-Based Sessions** - Secure HTTP-only cookies
- **Protected Routes** - `<SessionAuth>` wrapper
- **Auto Token Refresh** - SuperTokens handles it

### Styling
- **Vanilla CSS** - No heavy frameworks
- **CSS Variables** - Easy theming
- **Responsive Grid** - Mobile-first

## File Structure

```
frontend/
├── src/
│   ├── components/
│   │   └── Dashboard.jsx      # Main dashboard
│   ├── App.jsx                 # App + routing
│   ├── App.css                 # Styles
│   ├── main.jsx                # Entry point
│   └── index.css               # Global styles
├── index.html                  # HTML template
├── vite.config.js             # Vite config
└── package.json               # Dependencies
```

## How It Works

### 1. SuperTokens Initialization

```javascript
SuperTokens.init({
  appInfo: {
    appName: "UTM Backend",
    apiDomain: window.location.origin,
    websiteDomain: window.location.origin,
    apiBasePath: "/auth",
    websiteBasePath: "/auth"
  },
  recipeList: [
    EmailPassword.init(),
    Session.init()
  ]
});
```

### 2. Protected Routes

```javascript
<Route
  path="/"
  element={
    <SessionAuth>
      <Dashboard />
    </SessionAuth>
  }
/>
```

### 3. API Calls with Cookies

```javascript
const response = await fetch('/api/v1/tenants', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include', // 👈 This sends cookies!
  body: JSON.stringify(tenantData)
});
```

## Development

### Hot Module Replacement (HMR)

Edit any file → changes appear instantly in browser!

```bash
# Watch mode (already running in Docker)
docker-compose logs -f frontend
```

### Add New Features

1. Create component in `src/components/`
2. Import in `App.jsx`
3. Add route if needed
4. Style in `App.css`

### Customize Styles

Edit `src/App.css` for:
- Colors
- Layout
- Component styles
- Responsive breakpoints

## API Integration

The dashboard makes these API calls:

1. **On Load**:
   - `GET /api/v1/tenants` - Fetch user's tenants

2. **Create Tenant**:
   - `POST /api/v1/tenants` - Create new tenant

3. **Check Status**:
   - `GET /api/v1/tenants/:id/status` - Get initialization status

All authenticated via cookies automatically! No token management needed.

## Browser DevTools

### Check Cookies

1. Open DevTools (F12)
2. Go to Application tab
3. Cookies → http://localhost:3000
4. You'll see SuperTokens session cookies

### Network Tab

Watch API calls in real-time:
- See request/response
- Verify cookies are sent
- Check status codes

### Console

No errors should appear. If you see auth errors, check:
1. Backend is running
2. SuperTokens is healthy
3. Cookies are enabled

## Deployment

### Production Build

```bash
cd frontend
npm run build
```

Outputs to `dist/` directory.

### Docker Production

```bash
docker build -t utm-frontend:prod --target production ./frontend
docker run -p 80:80 utm-frontend:prod
```

Uses Nginx to serve static files with proper routing.

## Customization Ideas

### 1. Add More Fields

Edit `Dashboard.jsx`:
```javascript
<input 
  name="description"
  placeholder="Company description"
/>
```

### 2. Add Member Management

Create `MembersList.jsx`:
```javascript
const MembersList = ({ tenantId }) => {
  // Fetch and display members
  // Add/remove members
};
```

### 3. Add RBAC UI

Create components for:
- Role assignment
- Permission management  
- Access control visualization

### 4. Add Analytics

Integrate charts/graphs:
- Tenant growth
- Member activity
- Permission usage

## Testing the Frontend

### Manual Testing

1. ✅ Sign up with new email
2. ✅ Sign in with existing account
3. ✅ Create tenant with various names
4. ✅ Check slug generation
5. ✅ Create multiple tenants
6. ✅ Refresh page (session persists)
7. ✅ Sign out
8. ✅ Try accessing without login (redirects)

### Browser Compatibility

Tested on:
- ✅ Chrome/Edge (Chromium)
- ✅ Firefox
- ✅ Safari

### Mobile Testing

Responsive on:
- 📱 iPhone (Safari)
- 📱 Android (Chrome)
- 📱 Tablets

## Common Customizations

### Change Color Scheme

Edit `src/index.css`:
```css
:root {
  --primary-color: #646cff;  /* Your brand color */
  --background: #242424;
  --text-color: #ffffff;
}
```

### Add Logo

Edit `src/App.jsx`:
```javascript
<div className="header">
  <img src="/logo.png" alt="Logo" />
  <h1>Your Company Name</h1>
</div>
```

### Add Footer

Edit `src/App.jsx`:
```javascript
<footer className="footer">
  <p>&copy; 2025 Your Company</p>
</footer>
```

## Support

- **SuperTokens Docs**: https://supertokens.com/docs
- **Vite Docs**: https://vitejs.dev
- **React Docs**: https://react.dev

## 🎉 Enjoy Your Frontend!

The frontend is fully functional and ready for development. Start customizing and building your multi-tenant application!
