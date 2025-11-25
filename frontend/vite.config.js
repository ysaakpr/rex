import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd(), '')
  
  // Get base path from environment or default to '/'
  const basePath = env.VITE_BASE_PATH || '/'
  
  // Ensure base path has trailing slash for Vite
  const base = basePath.endsWith('/') ? basePath : `${basePath}/`
  
  return {
  plugins: [react()],
    base,
  server: {
    host: '0.0.0.0',
    port: 3000,
    // Allow requests from any host (useful for custom domains, staging environments, etc.)
    allowedHosts: 'all',
    watch: {
      usePolling: true,
    },
    hmr: {
      // Enable HMR to work through nginx proxy
      clientPort: 80,
        path: `${base}@vite/client`,
    },
    proxy: {
      // Proxy all /api requests (including /api/auth for SuperTokens) to backend
      '/api': {
        target: 'http://api:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
        cookieDomainRewrite: 'localhost',
        cookiePathRewrite: '/',
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('proxy error', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            // Manually forward cookies
            if (req.headers.cookie) {
              proxyReq.setHeader('Cookie', req.headers.cookie);
            }
            console.log('Sending Request:', req.method, req.url, 'Cookies:', req.headers.cookie ? 'YES' : 'NO');
          });
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('Received Response:', proxyRes.statusCode, req.url);
          });
        },
        }
      }
    }
  }
})
