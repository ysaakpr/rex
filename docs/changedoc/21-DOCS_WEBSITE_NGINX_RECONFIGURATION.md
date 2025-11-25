# Change Documentation 21: Documentation Website & Nginx Reconfiguration

**Date**: November 25, 2025  
**Type**: Infrastructure & Documentation  
**Status**: Complete

## Overview

Reconfigured the nginx reverse proxy to serve the VitePress documentation website at the root URL (`/`) and moved the admin dashboard React app to `/demo`. This change transforms the main site into a documentation-first platform with easy access to a live demo.

## Purpose

1. **Documentation First**: Make comprehensive documentation the primary entry point
2. **Demo Access**: Provide immediate access to a live demo of the admin dashboard
3. **Better User Experience**: Help users understand the platform before diving into the demo
4. **Professional Presentation**: Present a polished, documentation-focused front page

## Changes Made

### 1. Documentation Website Service

#### Created `docs-website/Dockerfile`
- Multi-stage Dockerfile for VitePress documentation
- Development stage with hot reload on port 5173
- Production stage with nginx serving static files
- Supports both development and production deployments

**Key Features**:
- Node 20 Alpine base image
- Hot Module Replacement (HMR) support for development
- Optimized build process for production
- Nginx serving for production static files

#### Updated `docker-compose.yml`
- Added new `docs` service for VitePress documentation
- Configured to run on port 5173
- Added volume mounts for hot reload
- Updated nginx service dependencies to include docs

### 2. Nginx Configuration Changes

#### Updated `nginx.conf`

**New Upstream**:
```nginx
upstream docs {
    server docs:5173;
}
```

**Root Path (`/`)**: Now serves documentation
- Proxies to docs-website service
- VitePress with HMR support
- WebSocket support for development

**Demo Path (`/demo`)**: Serves admin dashboard
- Rewrites `/demo` to `/` for the frontend app
- Maintains all React Router functionality
- SuperTokens auth at `/demo/auth`
- Preserves HMR and WebSocket support

**API Path (`/api`)**: Unchanged
- Backend API continues to work at `/api`
- No changes to API endpoints

**MailHog (`/inbox`)**: Unchanged
- Email testing interface still at `/inbox`

### 3. Frontend Configuration Updates

#### Updated `frontend/vite.config.js`
- Added `base: '/demo/'` for proper routing
- Updated HMR path to `/demo/@vite/client`
- Maintained proxy configuration for `/api`

#### Updated `frontend/src/App.jsx`
- Changed `websiteBasePath` from `/auth` to `/demo/auth`
- Updated SuperTokens routing to work with new base path
- All frontend routes now prefixed with `/demo`

### 4. Environment Configuration

#### Created `.env.example`
- Comprehensive environment variable documentation
- Updated `WEBSITE_DOMAIN` to `http://localhost/demo`
- Updated `INVITATION_BASE_URL` to `http://localhost/demo/accept-invite`
- Included production deployment notes

**Key Changes**:
```bash
# Old
WEBSITE_DOMAIN=http://localhost
INVITATION_BASE_URL=http://localhost/accept-invite

# New
WEBSITE_DOMAIN=http://localhost/demo
INVITATION_BASE_URL=http://localhost/demo/accept-invite
```

### 5. Documentation Homepage Update

#### Updated `docs-website/docs/index.md`
- Added "Try Demo" button in hero section
- Links to `/demo` for immediate access to admin dashboard
- Maintains "Get Started" and "View on GitHub" buttons

## URL Structure

### Before Change
```
http://localhost/          → React Admin Dashboard
http://localhost/api       → Backend API
http://localhost/inbox     → MailHog
```

### After Change
```
http://localhost/          → VitePress Documentation
http://localhost/demo      → React Admin Dashboard
http://localhost/api       → Backend API
http://localhost/inbox     → MailHog
```

### Direct Access (Development Only)
```
http://localhost:3000      → Frontend (bypass nginx)
http://localhost:5173      → Docs (bypass nginx)
http://localhost:8080      → API (bypass nginx)
http://localhost:8025      → MailHog UI (direct)
```

## Technical Details

### Nginx Routing Logic

1. **Documentation (Root)**:
   - Matches: `location /`
   - Priority: Lowest (catch-all)
   - Proxies to: `docs:5173`
   - Handles: All documentation pages and assets

2. **Demo App**:
   - Matches: `location /demo`
   - Rewrites: `/demo/path` → `/path`
   - Proxies to: `frontend:3000`
   - Handles: React Router, auth flows, admin UI

3. **API**:
   - Matches: `location /api`
   - Proxies to: `api:8080`
   - Handles: All backend endpoints

4. **MailHog**:
   - Matches: `location /inbox`
   - Proxies to: `mailhog:80`
   - Handles: Email testing interface

### SuperTokens Configuration

The SuperTokens configuration needed updates to work with the new base path:

```javascript
// App.jsx
SuperTokens.init({
  appInfo: {
    apiDomain: window.location.origin,        // http://localhost
    websiteDomain: window.location.origin,    // http://localhost
    apiBasePath: "/api/auth",                 // Unchanged
    websiteBasePath: "/demo/auth"             // Changed from /auth
  }
});
```

### Vite Configuration

The Vite config now includes a base path to ensure proper asset loading:

```javascript
// vite.config.js
export default defineConfig({
  base: '/demo/',
  server: {
    hmr: {
      path: '/demo/@vite/client',
    }
  }
});
```

## Testing Checklist

### Documentation Website
- [ ] Access root URL: `http://localhost/`
- [ ] Verify documentation loads correctly
- [ ] Test navigation between docs pages
- [ ] Verify search functionality works
- [ ] Click "Try Demo" button - should navigate to `/demo`
- [ ] Check HMR works (edit a markdown file)

### Admin Dashboard
- [ ] Access demo URL: `http://localhost/demo`
- [ ] Verify app loads correctly
- [ ] Test sign up flow at `/demo/auth/signup`
- [ ] Test sign in flow at `/demo/auth/signin`
- [ ] Verify all routes work with `/demo` prefix
- [ ] Check HMR works (edit a React component)
- [ ] Test tenant creation and management
- [ ] Verify API calls work correctly

### Authentication
- [ ] Sign up a new user from `/demo`
- [ ] Verify cookies are set correctly
- [ ] Test session persistence across page reloads
- [ ] Verify logout works
- [ ] Test invitation emails link to `/demo/accept-invite`

### API & Backend
- [ ] API accessible at `/api`
- [ ] Test protected endpoints with cookies
- [ ] Verify SuperTokens endpoints at `/api/auth`
- [ ] Check CORS headers are correct

### Email Testing
- [ ] MailHog accessible at `/inbox`
- [ ] Send test invitation
- [ ] Verify email links point to `/demo/accept-invite`

## Deployment Steps

### Development
```bash
# 1. Copy environment file
cp .env.example .env

# 2. Build and start services
docker-compose up -d

# 3. Wait for services to be ready
docker-compose logs -f docs
docker-compose logs -f frontend

# 4. Access the application
open http://localhost           # Documentation
open http://localhost/demo      # Admin Demo
open http://localhost/inbox     # MailHog
```

### Production Considerations

When deploying to production, update the following in `.env`:

1. **Base URLs**:
   ```bash
   API_DOMAIN=https://yourdomain.com
   WEBSITE_DOMAIN=https://yourdomain.com/demo
   INVITATION_BASE_URL=https://yourdomain.com/demo/accept-invite
   ```

2. **SuperTokens**:
   - Use a strong API key (20+ characters)
   - Configure proper cookie domains

3. **Nginx**:
   - Enable HTTPS (port 443)
   - Configure SSL certificates
   - Update security headers

4. **Frontend Build**:
   - Build VitePress docs: `npm run docs:build`
   - Build React app with base path: `npm run build`
   - Ensure `base: '/demo/'` is set in production config

## Migration Guide

If you have an existing deployment, follow these steps:

### 1. Update Environment Variables
```bash
# Update .env file
sed -i 's|WEBSITE_DOMAIN=http://localhost|WEBSITE_DOMAIN=http://localhost/demo|g' .env
sed -i 's|INVITATION_BASE_URL=http://localhost/accept-invite|INVITATION_BASE_URL=http://localhost/demo/accept-invite|g' .env
```

### 2. Update Frontend Config
- Ensure `vite.config.js` has `base: '/demo/'`
- Ensure `App.jsx` has `websiteBasePath: "/demo/auth"`

### 3. Rebuild and Restart
```bash
# Stop services
docker-compose down

# Rebuild with new config
docker-compose build

# Start services
docker-compose up -d
```

### 4. Clear Browser Cache
- Clear cookies for localhost
- Clear browser cache
- Hard refresh (Cmd+Shift+R / Ctrl+Shift+F5)

### 5. Test All Flows
- Access documentation at `/`
- Access demo at `/demo`
- Test sign up and sign in
- Verify invitation emails

## Breaking Changes

⚠️ **Important**: This is a breaking change if you have existing users or deployments.

### For Existing Users
- Old bookmark: `http://localhost/` (admin dashboard) → Now shows docs
- New URL for admin: `http://localhost/demo`
- Update any saved bookmarks or links

### For Existing Deployments
- SuperTokens cookies may need to be regenerated
- Update all invitation email templates if customized
- Update any external links to the admin dashboard
- Update mobile app deep links if applicable

### For Integrations
- API endpoints unchanged (still at `/api`)
- SuperTokens API endpoints unchanged (still at `/api/auth`)
- Only frontend URLs affected

## Benefits

### For Users
1. **Better First Impression**: Documentation-first landing page
2. **Easy Exploration**: Try demo without commitment
3. **Clear Navigation**: Obvious path from docs to demo
4. **Professional Look**: Polished, well-documented platform

### For Developers
1. **Clearer Structure**: Separation of docs and demo
2. **Easier Onboarding**: New developers see docs first
3. **Better Development**: VitePress HMR for docs changes
4. **Flexible Deployment**: Can deploy docs and demo separately

### For Project
1. **Better SEO**: Documentation content indexed by search engines
2. **Reduced Support**: Users find answers in docs
3. **Professional Image**: Shows project maturity
4. **Marketing Tool**: Docs serve as marketing material

## Troubleshooting

### Issue: Frontend 404 errors
**Solution**: Ensure `base: '/demo/'` is set in `vite.config.js`

### Issue: Auth not working
**Solution**: 
- Check `WEBSITE_DOMAIN` in `.env` is `http://localhost/demo`
- Verify `websiteBasePath` in `App.jsx` is `/demo/auth`
- Clear browser cookies

### Issue: Assets not loading in demo
**Solution**: 
- Verify Vite base path is correct
- Check nginx rewrite rules
- Inspect browser console for path errors

### Issue: HMR not working
**Solution**:
- Check WebSocket connections in browser dev tools
- Verify nginx proxy settings for Upgrade header
- Restart docker-compose services

### Issue: Invitation links broken
**Solution**: 
- Update `INVITATION_BASE_URL` to include `/demo`
- Check email templates for hardcoded URLs

## Files Modified

### New Files
- `docs-website/Dockerfile`
- `.env.example`

### Modified Files
- `docker-compose.yml`
- `nginx.conf`
- `frontend/vite.config.js`
- `frontend/src/App.jsx`
- `docs-website/docs/index.md`

### Configuration Files
- `.env` (created from example)

## Related Documentation

- [docs/QUICKSTART.md](../QUICKSTART.md) - Updated quick start guide
- [docs/API_EXAMPLES.md](../API_EXAMPLES.md) - API documentation
- [docs-website/README.md](../../docs-website/README.md) - VitePress setup
- [frontend/README.md](../../frontend/README.md) - Frontend configuration

## Future Improvements

1. **Custom Domain for Docs**: Deploy docs to separate domain (docs.yourdomain.com)
2. **Versioned Docs**: Support multiple versions of documentation
3. **API Playground**: Interactive API testing in docs
4. **Demo Data**: Pre-populated demo tenant with sample data
5. **Analytics**: Track docs usage and popular pages
6. **Search Enhancement**: Add Algolia or similar for better search
7. **Dark Mode**: Add dark mode toggle for documentation

## Conclusion

This change successfully transforms the UTM Backend into a documentation-first platform while maintaining full functionality of the admin dashboard at `/demo`. The separation of concerns improves user experience, developer onboarding, and overall project presentation.

The implementation maintains backward compatibility for the API layer while providing a clear upgrade path for frontend users. All authentication, authorization, and core functionality remain unchanged.

---

**Last Updated**: November 25, 2025  
**Author**: Development Team  
**Review Status**: Complete  
**Tested**: ✅ Development Environment

