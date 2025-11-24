import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
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
})

