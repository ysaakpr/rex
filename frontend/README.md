# UTM Frontend - User & Tenant Management

A React-based frontend for the UTM Backend, built with Vite and SuperTokens for authentication.

## Features

- 🔐 **SuperTokens Authentication** - Email/password login with UI components
- 🏢 **Tenant Management** - Create and view tenants
- 📱 **Responsive Design** - Works on desktop and mobile
- ⚡ **Fast Development** - Powered by Vite
- 🎨 **Modern UI** - Clean, dark-mode interface

## Quick Start

### With Docker (Recommended)

```bash
# From project root
docker-compose up frontend
```

The frontend will be available at: http://localhost:3000

### Local Development

```bash
cd frontend
npm install
npm run dev
```

## How to Use

1. **Open the app**: Navigate to http://localhost:3000
2. **Sign Up**: Click "Sign Up" and create an account
3. **Sign In**: Log in with your credentials
4. **Create Tenant**: Fill in the form and create your first tenant
5. **View Tenants**: See all your tenants listed below

## Architecture

### Tech Stack

- **React 18** - UI library
- **Vite** - Build tool and dev server
- **SuperTokens** - Authentication
- **React Router** - Routing
- **Axios** - HTTP client

### File Structure

```
frontend/
├── src/
│   ├── components/
│   │   └── Dashboard.jsx     # Main dashboard component
│   ├── App.jsx                # Main app with routing
│   ├── App.css                # App styles
│   ├── main.jsx               # Entry point
│   └── index.css              # Global styles
├── index.html                 # HTML template
├── vite.config.js            # Vite configuration
├── package.json              # Dependencies
└── Dockerfile                # Container config

## SuperTokens Configuration

The app is configured to work with the backend's SuperTokens setup:

- **API Domain**: `http://localhost:3000` (proxied to backend)
- **Auth Base Path**: `/auth`
- **Session Mode**: Cookie-based (automatic)

The Vite dev server proxies `/auth` and `/api` requests to the backend at `http://localhost:8080`.

## API Integration

All API calls are made to the backend through the Vite proxy:

```javascript
// Fetch tenants
const response = await fetch('/api/v1/tenants', {
  method: 'GET',
  credentials: 'include' // Important for cookies
});
```

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build

## Troubleshooting

### Authentication Issues

1. Make sure backend is running: `docker-compose ps`
2. Check SuperTokens is healthy: `curl http://localhost:3567/hello`
3. Clear browser cookies and try again

### CORS Issues

The Vite proxy should handle CORS. If issues persist:
- Check `vite.config.js` proxy settings
- Ensure backend CORS middleware is configured correctly

### Port Already in Use

If port 3000 is taken, change it in `vite.config.js`:

```javascript
server: {
  port: 3001, // or any other port
}
```

## Production Build

To build for production:

```bash
npm run build
```

The built files will be in the `dist/` directory and can be served with any static file server or the production Docker image:

```bash
docker build -t utm-frontend:prod --target production ./frontend
docker run -p 80:80 utm-frontend:prod
```

## Customization

### Styling

Edit `src/App.css` and `src/index.css` to customize the look and feel.

### Adding Features

1. Create new components in `src/components/`
2. Add routes in `src/App.jsx`
3. Wrap protected routes with `<SessionAuth>`

### SuperTokens UI Customization

See the [SuperTokens documentation](https://supertokens.com/docs/emailpassword/common-customizations/styling/changing-style) for UI customization options.

## Learn More

- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://react.dev/)
- [SuperTokens Documentation](https://supertokens.com/docs/emailpassword/introduction)
- [React Router Documentation](https://reactrouter.com/)

