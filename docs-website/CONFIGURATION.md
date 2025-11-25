# Docs Website Configuration Guide

This guide explains how to configure the Rex documentation website for different deployment scenarios.

## Environment Variables

### `VITE_DEMO_URL`

The URL to the live demo/frontend application. This is used for the "Try Demo" buttons in both the navigation and the home page hero section.

**Default**: `http://localhost/demo`

## How It Works

The demo URL is made configurable through:
1. **Environment Variable**: `VITE_DEMO_URL` is read from `.env` file or environment
2. **VitePress Config**: The URL is defined in `.vitepress/config.js` and made available globally
3. **Navigation**: The navigation "Try Demo" link uses the configured URL directly
4. **Home Page**: A custom Vue component (`DemoLink.vue`) dynamically renders the button with the configured URL

This allows the docs site to be deployed completely separately from the demo application.

## Configuration Scenarios

### 1. Local Development (Default)

For local development with Docker Compose:

```bash
# .env file
VITE_DEMO_URL=http://localhost/demo
```

Access:
- Docs: `http://localhost/docs`
- Demo: `http://localhost/demo`

### 2. Production - Same Domain

When docs and demo are on the same domain:

```bash
# .env file
VITE_DEMO_URL=https://yourdomain.com
```

Access:
- Docs: `https://yourdomain.com/docs`
- Demo: `https://yourdomain.com`

### 3. Production - Separate Subdomains

When docs and demo use different subdomains:

```bash
# .env file
VITE_DEMO_URL=https://demo.yourdomain.com
```

Access:
- Docs: `https://docs.yourdomain.com`
- Demo: `https://demo.yourdomain.com`

### 4. Production - Completely Separate Deployment

When docs site is deployed separately:

```bash
# .env file
VITE_DEMO_URL=https://your-production-app.com
```

Access:
- Docs: `https://your-docs-site.com`
- Demo: `https://your-production-app.com`

## Setup Instructions

### Using Environment File

1. Copy the example environment file:
   ```bash
   cd docs-website
   cp .env.example .env
   ```

2. Edit `.env` and set your demo URL:
   ```bash
   VITE_DEMO_URL=https://your-demo-url.com
   ```

3. Restart the docs service:
   ```bash
   docker-compose restart docs
   ```

### Using Docker Compose

Edit `docker-compose.yml` to set the environment variable:

```yaml
docs:
  build:
    context: ./docs-website
    dockerfile: Dockerfile
    target: development
    args:
      VITE_DEMO_URL: https://your-demo-url.com
  environment:
    - VITE_DEMO_URL=https://your-demo-url.com
```

### Using Docker Run

When running the docs container directly:

```bash
# Development
docker run -p 5173:5173 \
  -e VITE_DEMO_URL=https://demo.yourdomain.com \
  rex-docs

# Production (nginx)
docker build --target production \
  --build-arg VITE_DEMO_URL=https://demo.yourdomain.com \
  -t rex-docs .

docker run -p 80:80 rex-docs
```

### Using Docker Build Args

For production builds with custom demo URL:

```bash
docker build \
  --target production \
  --build-arg VITE_DEMO_URL=https://your-demo.com \
  -t rex-docs .
```

## Kubernetes/Cloud Deployments

### Environment Variable

Set the environment variable in your deployment manifest:

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
          value: "https://demo.yourdomain.com"
        ports:
        - containerPort: 5173
```

### ConfigMap

Or use a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rex-docs-config
data:
  VITE_DEMO_URL: "https://demo.yourdomain.com"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rex-docs
spec:
  template:
    spec:
      containers:
      - name: docs
        envFrom:
        - configMapRef:
            name: rex-docs-config
```

## Vercel/Netlify Deployment

### Vercel

1. Set environment variable in Vercel dashboard:
   - Key: `VITE_DEMO_URL`
   - Value: `https://your-demo.com`

2. Or use `vercel.json`:
   ```json
   {
     "build": {
       "env": {
         "VITE_DEMO_URL": "https://your-demo.com"
       }
     }
   }
   ```

### Netlify

1. Set environment variable in Netlify dashboard:
   - Key: `VITE_DEMO_URL`
   - Value: `https://your-demo.com`

2. Or use `netlify.toml`:
   ```toml
   [build.environment]
     VITE_DEMO_URL = "https://your-demo.com"
   ```

## Static Hosting (GitHub Pages, S3, etc.)

For static hosting, you need to build with the correct URL:

```bash
# Build with custom demo URL
VITE_DEMO_URL=https://your-demo.com npm run docs:build

# Deploy the dist folder
# Output is in: docs/.vitepress/dist/
```

## Verifying Configuration

After deployment, verify the demo link:

1. Open the docs website
2. Check the "Try Demo" button in the top navigation
3. Click it to ensure it opens the correct demo URL

## Troubleshooting

### Demo Link Not Working

If the "Try Demo" link doesn't work:

1. Check the environment variable is set:
   ```bash
   echo $VITE_DEMO_URL
   ```

2. Verify the build used the correct value:
   - View page source
   - Check the navigation bar links

3. Restart the container:
   ```bash
   docker-compose restart docs
   ```

### Demo Link Shows localhost

If production shows localhost:

1. The environment variable wasn't set during build
2. Rebuild with correct environment variable:
   ```bash
   docker-compose build --no-cache docs
   ```

## Best Practices

1. **Use Environment Variables**: Never hardcode URLs in config files
2. **Document Your URLs**: Keep a record of all demo URLs for different environments
3. **Test After Deployment**: Always verify the demo link works
4. **Use HTTPS in Production**: Always use HTTPS for production demo URLs
5. **Keep URLs Updated**: Update the demo URL if your demo site moves

## Support

For issues or questions:
- Check the [main documentation](/)
- Open an issue on [GitHub](https://github.com/ysaakpr/rex/issues)

