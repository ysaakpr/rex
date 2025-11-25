# Rex Documentation Website

Comprehensive documentation for the Rex project, built with VitePress.

## Setup

```bash
npm install
```

## Configuration

Copy the example environment file and configure:

```bash
cp .env.example .env
```

### Environment Variables

- `VITE_DEMO_URL`: URL to the live demo/frontend application
  - Local development: `http://localhost/demo`
  - Production: `https://yourdomain.com` or `https://demo.yourdomain.com`

## Development

Run the documentation site locally:

```bash
npm run docs:dev
```

The site will be available at `http://localhost:5173`

## Build

Build for production:

```bash
npm run docs:build
```

Preview the production build:

```bash
npm run docs:preview
```

## Docker Deployment

The documentation site can be deployed independently using Docker:

```bash
# Build the image
docker build -t rex-docs .

# Run the container
docker run -p 5173:5173 -e VITE_DEMO_URL=https://demo.yourdomain.com rex-docs
```

## Deployment Configurations

### Local Development (Default)
```env
VITE_DEMO_URL=http://localhost/demo
```

### Production with Same Domain
```env
VITE_DEMO_URL=https://yourdomain.com
```

### Production with Separate Demo Domain
```env
VITE_DEMO_URL=https://demo.yourdomain.com
```

### Deployed Separately
When deploying the docs website separately from the main application:
```env
VITE_DEMO_URL=https://your-production-demo.com
```

## Structure

```
docs/
├── .vitepress/
│   ├── config.js              # VitePress configuration
│   └── theme/
│       ├── index.js           # Custom theme setup
│       ├── custom.css         # Custom styling
│       └── components/
│           └── DemoLink.vue   # Dynamic demo link component
├── introduction/              # Overview and concepts
├── getting-started/           # Quick start guides
├── guides/                    # Feature guides
├── api/                       # API documentation
├── frontend/                  # Frontend integration
├── middleware/                # Language-specific middleware
├── system-auth/               # System authentication
├── jobs/                      # Background jobs
├── deployment/                # Deployment guides
├── examples/                  # Code examples
├── advanced/                  # Advanced topics
├── troubleshooting/           # Common issues
├── reference/                 # Technical reference
└── public/                    # Static assets (logo, favicon)
```

## Custom Components

### DemoLink Component

The `DemoLink.vue` component provides a dynamic "Try Demo" button that uses the configured `VITE_DEMO_URL`. This ensures the demo link works correctly in all deployment scenarios.

**Usage in Markdown:**
```markdown
<DemoLink />
```

The component is automatically registered globally and can be used in any markdown file.

## Contributing

When adding new documentation:

1. Create the `.md` file in the appropriate directory
2. Update the sidebar in `.vitepress/config.js`
3. Test locally with `npm run docs:dev`
4. Ensure all links work correctly
5. If using custom components, test them with different `VITE_DEMO_URL` values

## Tech Stack

- **VitePress**: Static site generator
- **Vue 3**: Component framework
- **Markdown**: Content format
- **Vite**: Build tool
