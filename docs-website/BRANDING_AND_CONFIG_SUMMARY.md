# Docs Website Branding & Configuration Summary

This document summarizes all the branding changes and configurable features implemented for the Rex documentation website.

## Changes Made

### 1. Branding Updates

#### Site Title & Description
- **Title**: Changed from "UTM Backend" to **"Rex"**
- **Description**: Changed from "Multi-Tenant SaaS Platform" to **"Open-Source User Management Platform with Multi-Tenancy and RBAC"**
- **Hero Text**: Updated to emphasize it's an open-source user management platform, not a SaaS product itself

#### Logo & Favicon
- Created custom SVG logo with blue circle and white "R"
- Created matching favicon
- Both stored in `/docs/public/`

#### GitHub Links
- Updated all GitHub repository links from placeholder to `https://github.com/ysaakpr/rex`

#### Content Updates
- Replaced all **41 occurrences** of "UTM Backend" with "Rex" across documentation
- Updated positioning from "SaaS Platform" to "User Management Platform"

### 2. Configurable Demo URL

#### Problem Solved
The documentation site can now be deployed separately from the main application, with the "Try Demo" button pointing to any configured URL.

#### Implementation

**Files Modified:**
1. `.vitepress/config.js` - Added Vite define for demo URL
2. `.vitepress/theme/index.js` - Custom theme setup
3. `.vitepress/theme/components/DemoLink.vue` - Dynamic demo link component
4. `docs/index.md` - Uses the DemoLink component
5. `Dockerfile` - Added build arg for VITE_DEMO_URL
6. `docker-compose.yml` - Added environment variable

**How It Works:**

```
Environment Variable (VITE_DEMO_URL)
    ↓
VitePress Config (config.js)
    ↓
Vite Define (__DEMO_URL__)
    ↓
Vue Component (DemoLink.vue) & Navigation
    ↓
Dynamic "Try Demo" Buttons
```

#### Configuration Options

**Local Development:**
```env
VITE_DEMO_URL=http://localhost/demo
```

**Production - Same Domain:**
```env
VITE_DEMO_URL=https://yourdomain.com
```

**Production - Separate Domains:**
```env
VITE_DEMO_URL=https://demo.yourdomain.com
```

### 3. Custom Theme

Created a custom VitePress theme that:
- Extends the default theme
- Registers custom components globally
- Provides custom styling with brand colors
- Makes the demo URL dynamically available

**Theme Structure:**
```
.vitepress/theme/
├── index.js           # Theme entry point
├── custom.css         # Custom styling (blue brand colors)
└── components/
    └── DemoLink.vue   # Dynamic demo link button
```

### 4. Documentation Files Created

1. **`.env.example`** - Example environment configuration
2. **`.env`** - Local environment configuration (gitignored)
3. **`CONFIGURATION.md`** - Comprehensive configuration guide covering:
   - Local development
   - Production deployments
   - Docker configurations
   - Kubernetes/Cloud deployments
   - Vercel/Netlify deployments
   - Static hosting
   - Troubleshooting

4. **`README.md`** - Updated with:
   - Setup instructions
   - Configuration guide
   - Structure overview
   - Custom components documentation

5. **`BRANDING_AND_CONFIG_SUMMARY.md`** - This document

### 5. File Structure

```
docs-website/
├── .env                           # Local config (gitignored)
├── .env.example                   # Example config
├── .gitignore                     # Ignores .env
├── CONFIGURATION.md               # Config guide
├── README.md                      # Main documentation
├── BRANDING_AND_CONFIG_SUMMARY.md # This file
├── Dockerfile                     # Multi-stage with build args
├── package.json                   # Updated name to "rex-docs"
└── docs/
    ├── .vitepress/
    │   ├── config.js              # VitePress config with demo URL
    │   └── theme/
    │       ├── index.js           # Custom theme
    │       ├── custom.css         # Blue brand colors
    │       └── components/
    │           └── DemoLink.vue   # Dynamic demo button
    ├── index.md                   # Home page with DemoLink
    └── public/
        ├── logo.svg               # Rex logo
        └── favicon.svg            # Rex favicon
```

## Testing

### Test Demo URL Configuration

1. **Local Development:**
   ```bash
   cd docs-website
   cp .env.example .env
   # Verify VITE_DEMO_URL=http://localhost/demo
   npm run docs:dev
   # Open http://localhost:5173
   # Click "Try Demo" - should open http://localhost/demo
   ```

2. **Production Build:**
   ```bash
   VITE_DEMO_URL=https://demo.example.com npm run docs:build
   npm run docs:preview
   # Verify demo button points to demo.example.com
   ```

3. **Docker:**
   ```bash
   docker build --build-arg VITE_DEMO_URL=https://demo.example.com -t rex-docs .
   docker run -p 5173:5173 -e VITE_DEMO_URL=https://demo.example.com rex-docs
   ```

### Verify Branding

1. **Browser Tab**: Should show "Rex" and blue "R" favicon
2. **Navigation Logo**: Should show blue "R" logo next to "Rex"
3. **Home Page Title**: Should say "Rex"
4. **Hero Section**: Should say "Open-Source User Management Platform"
5. **Footer**: Should show Rex branding

## Deployment Examples

### Vercel
```bash
# Set environment variable in Vercel dashboard:
VITE_DEMO_URL=https://demo.yoursite.com

# Or in vercel.json:
{
  "build": {
    "env": {
      "VITE_DEMO_URL": "https://demo.yoursite.com"
    }
  }
}
```

### Netlify
```toml
# netlify.toml
[build.environment]
  VITE_DEMO_URL = "https://demo.yoursite.com"
```

### Docker Compose
```yaml
services:
  docs:
    build:
      args:
        VITE_DEMO_URL: https://demo.yoursite.com
    environment:
      - VITE_DEMO_URL=https://demo.yoursite.com
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rex-docs
spec:
  template:
    spec:
      containers:
      - name: docs
        image: rex-docs:latest
        env:
        - name: VITE_DEMO_URL
          value: "https://demo.yoursite.com"
```

## Brand Colors

The custom theme uses these colors:

**Light Mode:**
- Primary: `#3b82f6` (blue-500)
- Hover: `#2563eb` (blue-600)
- Active: `#1d4ed8` (blue-700)

**Dark Mode:**
- Primary: `#60a5fa` (blue-400)
- Hover: `#3b82f6` (blue-500)
- Active: `#2563eb` (blue-600)

## Migration Notes

### From UTM Backend to Rex

All references have been updated:
- ✅ Site title
- ✅ Site description
- ✅ Logo and favicon
- ✅ Navigation
- ✅ Hero section
- ✅ Footer
- ✅ GitHub links
- ✅ Package name
- ✅ Documentation content (41 occurrences)

### Demo URL Configuration

The demo URL is now configurable instead of hardcoded:
- ✅ Navigation "Try Demo" link
- ✅ Home page demo button
- ✅ Environment variable based
- ✅ Docker build args supported
- ✅ Production deployment ready

## Next Steps

1. **Update Production URLs**: Set `VITE_DEMO_URL` in production deployment
2. **Update Branding**: If you want to use a different logo, replace `/docs/public/logo.svg`
3. **Customize Colors**: Edit `.vitepress/theme/custom.css` to change brand colors
4. **Add More Links**: Update navigation in `.vitepress/config.js`

## Support

For questions or issues:
- See [CONFIGURATION.md](./CONFIGURATION.md) for detailed setup instructions
- See [README.md](./README.md) for general documentation
- Open an issue on [GitHub](https://github.com/ysaakpr/rex/issues)

