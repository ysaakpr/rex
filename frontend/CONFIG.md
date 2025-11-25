# Frontend Configuration Guide

## Centralized Base Path Configuration

All base path configuration is centralized in a single environment variable: **`VITE_BASE_PATH`**

### Quick Start

The app currently runs at `/demo`. To change this:

1. **Edit `.env` file:**
   ```bash
   # Run at /demo (current)
   VITE_BASE_PATH=/demo
   
   # OR run at root
   VITE_BASE_PATH=/
   
   # OR run at /admin
   VITE_BASE_PATH=/admin
   ```

2. **Restart the frontend container:**
   ```bash
   docker-compose restart frontend
   ```

That's it! No other changes needed.

## How It Works

### Single Source of Truth

The `VITE_BASE_PATH` environment variable controls ALL path-related configuration:

```
.env file
   ↓
VITE_BASE_PATH=/demo
   ↓
┌─────────────────────┬──────────────────────┐
│                     │                      │
vite.config.js    src/config.js         App.jsx
│                     │                      │
base: '/demo/'    BASENAME='/demo'    BrowserRouter
                  AUTH_PATH='/demo/auth'  SuperTokens
```

### Files Involved

1. **`.env` or `.env.example`**
   - Defines `VITE_BASE_PATH=/demo`
   - Only place you need to change!

2. **`src/config.js`** (centralized config)
   - Reads `import.meta.env.VITE_BASE_PATH`
   - Exports normalized paths for all components
   - Handles edge cases (trailing slashes, root path, etc.)

3. **`vite.config.js`**
   - Loads env with `loadEnv(mode, process.cwd(), '')`
   - Sets `base` for asset paths
   - Configures HMR path

4. **`src/App.jsx`**
   - Imports config from `src/config.js`
   - Uses `config.basename` for React Router
   - Uses `config.authPath` for SuperTokens

## Configuration Options

### Option 1: Run at Root (/)

```bash
# .env
VITE_BASE_PATH=/
```

**Result:**
- App accessible at: `http://localhost/`
- Auth pages at: `http://localhost/auth`
- Routes: `/tenants`, `/roles`, etc.

### Option 2: Run at /demo (Current)

```bash
# .env
VITE_BASE_PATH=/demo
```

**Result:**
- App accessible at: `http://localhost/demo/`
- Auth pages at: `http://localhost/demo/auth`
- Routes: `/demo/tenants`, `/demo/roles`, etc.

### Option 3: Run at /admin

```bash
# .env
VITE_BASE_PATH=/admin
```

**Result:**
- App accessible at: `http://localhost/admin/`
- Auth pages at: `http://localhost/admin/auth`
- Routes: `/admin/tenants`, `/admin/roles`, etc.

## Production Deployment

### For Production Build

1. **Set base path in .env:**
   ```bash
   VITE_BASE_PATH=/demo  # or / or /admin
   ```

2. **Build the app:**
   ```bash
   npm run build
   ```

3. **Nginx configuration:**
   ```nginx
   # If base path is /demo
   location /demo {
       proxy_pass http://frontend;
   }
   
   # If base path is / (root)
   location / {
       proxy_pass http://frontend;
   }
   ```

### Docker Compose

The `.env` file is automatically picked up by Vite in the container because we mount the entire `frontend` directory:

```yaml
frontend:
  volumes:
    - ./frontend:/app  # Mounts .env file
```

## Troubleshooting

### Issue: Changes not taking effect

**Solution:** Restart the frontend container
```bash
docker-compose restart frontend
```

### Issue: Assets not loading (404)

**Check:**
1. Verify `.env` has `VITE_BASE_PATH` set correctly
2. Check browser dev tools → Network tab
3. Assets should load from the base path (e.g., `/demo/assets/...`)

### Issue: Routing not working

**Check:**
1. React Router `basename` should match your base path
2. Check browser console for configuration logs (development mode)
3. Verify `src/config.js` is imported in `App.jsx`

### Issue: SuperTokens auth not working

**Check:**
1. `websiteBasePath` should be `{basePath}/auth`
2. Backend `.env` should have matching `WEBSITE_DOMAIN`
3. Example: 
   - Frontend: `VITE_BASE_PATH=/demo`
   - Backend: `WEBSITE_DOMAIN=http://localhost/demo`

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_BASE_PATH` | `/` | Base path for the application |
| `VITE_API_DOMAIN` | `window.location.origin` | Backend API domain |

### Exported Config (`src/config.js`)

| Export | Example (`/demo`) | Description |
|--------|-------------------|-------------|
| `BASENAME` | `/demo` | React Router basename (no trailing slash) |
| `AUTH_PATH` | `/demo/auth` | SuperTokens websiteBasePath |
| `API_DOMAIN` | `http://localhost` | Backend API domain |
| `API_BASE_PATH` | `/api/auth` | SuperTokens API endpoints |

### Path Normalization Rules

```javascript
// Input          → BASENAME     → AUTH_PATH
'/demo/'          → '/demo'      → '/demo/auth'
'/demo'           → '/demo'      → '/demo/auth'
'/'               → ''           → '/auth'
'/admin/'         → '/admin'     → '/admin/auth'
```

## Testing Different Configurations

### Test Script

```bash
# Test at /demo
echo "VITE_BASE_PATH=/demo" > frontend/.env
docker-compose restart frontend
sleep 5
curl http://localhost/demo/

# Test at root
echo "VITE_BASE_PATH=/" > frontend/.env
docker-compose restart frontend
sleep 5
curl http://localhost/
```

### Verification Checklist

- [ ] App loads at correct URL
- [ ] Assets load correctly (check Network tab)
- [ ] Internal navigation works (click links)
- [ ] Auth pages load (`/auth` or `/demo/auth`)
- [ ] API calls work (check console)
- [ ] SuperTokens session works

## Best Practices

1. **Always use trailing slash in `.env`**: `VITE_BASE_PATH=/demo/` or `/`
2. **Keep config centralized**: Only change `.env`, never hardcode paths
3. **Test before deploying**: Verify all routes work with new base path
4. **Document your choice**: Add comment in `.env` explaining why you chose that path
5. **Match nginx config**: Ensure nginx location matches base path

## Examples

### Example: Marketing Site at Root, App at /demo

**Nginx:**
```nginx
location / {
    # Marketing site or docs
    proxy_pass http://marketing;
}

location /demo {
    # Admin app
    proxy_pass http://frontend;
}
```

**Frontend `.env`:**
```bash
VITE_BASE_PATH=/demo
```

### Example: App at Root

**Nginx:**
```nginx
location / {
    # Admin app at root
    proxy_pass http://frontend;
}
```

**Frontend `.env`:**
```bash
VITE_BASE_PATH=/
```

## Migration Guide

### Moving from /demo to Root

1. Update `.env`:
   ```bash
   # OLD
   VITE_BASE_PATH=/demo
   
   # NEW
   VITE_BASE_PATH=/
   ```

2. Update backend `.env`:
   ```bash
   # OLD
   WEBSITE_DOMAIN=http://localhost/demo
   
   # NEW
   WEBSITE_DOMAIN=http://localhost
   ```

3. Update nginx.conf:
   ```nginx
   # OLD
   location /demo {
       proxy_pass http://frontend;
   }
   
   # NEW
   location / {
       proxy_pass http://frontend;
   }
   ```

4. Restart all services:
   ```bash
   docker-compose restart frontend api nginx
   ```

## Support

For issues or questions:
- Check `src/config.js` for configuration logic
- Review `vite.config.js` for build configuration
- Check browser console in development mode for config logs

---

**Last Updated:** November 25, 2025  
**Version:** 1.0  
**Configuration:** Centralized via `VITE_BASE_PATH`

