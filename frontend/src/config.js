/**
 * Centralized Frontend Configuration
 * 
 * All environment-based configuration should be defined here.
 * This makes it easy to change deployment context (root, /demo, /admin, etc.)
 * by simply changing VITE_BASE_PATH in the .env file.
 */

// Get base path from environment variable
// Examples:
//   VITE_BASE_PATH=/demo  → App runs at /demo
//   VITE_BASE_PATH=/      → App runs at root
//   VITE_BASE_PATH=/admin → App runs at /admin
const BASE_PATH = import.meta.env.VITE_BASE_PATH || '/';

// Ensure base path starts with /
const normalizedBasePath = BASE_PATH.startsWith('/') ? BASE_PATH : `/${BASE_PATH}`;

// For React Router: basename should not have trailing slash
// '/' becomes ''
// '/demo/' becomes '/demo'
export const BASENAME = normalizedBasePath === '/' 
  ? '' 
  : normalizedBasePath.replace(/\/$/, '');

// For SuperTokens: websiteBasePath should not have trailing slash
// '/' + 'auth' → '/auth'
// '/demo' + '/auth' → '/demo/auth'
export const AUTH_PATH = normalizedBasePath === '/' 
  ? '/auth' 
  : `${normalizedBasePath.replace(/\/$/, '')}/auth`;

// API Configuration
// For API domain, prefer env variable, but allow dynamic fallback
const getApiDomain = () => {
  if (import.meta.env.VITE_API_DOMAIN) {
    return import.meta.env.VITE_API_DOMAIN;
  }
  // Fallback to window.location.origin if available
  if (typeof window !== 'undefined' && window.location) {
    return window.location.origin;
  }
  // During SSR or build, use empty string (will be set at runtime)
  return '';
};

export const API_DOMAIN = getApiDomain();
export const API_BASE_PATH = '/api/auth'; // SuperTokens API endpoints
export const WEBSITE_DOMAIN = typeof window !== 'undefined' && window.location ? window.location.origin : '';

// Full configuration object
export const config = {
  // App base path
  basePath: normalizedBasePath,
  basename: BASENAME,
  
  // Authentication
  authPath: AUTH_PATH,
  apiDomain: API_DOMAIN,
  websiteDomain: WEBSITE_DOMAIN,
  apiBasePath: API_BASE_PATH,
  
  // Application info
  appName: 'Rex',
};

// Log configuration in development
if (import.meta.env.DEV) {
  console.log('[Config] Frontend configuration:', {
    BASE_PATH: normalizedBasePath,
    BASENAME,
    AUTH_PATH,
    API_DOMAIN,
    WEBSITE_DOMAIN,
    API_BASE_PATH,
    env: import.meta.env.MODE,
    'VITE_API_DOMAIN (from env)': import.meta.env.VITE_API_DOMAIN,
    'VITE_BASE_PATH (from env)': import.meta.env.VITE_BASE_PATH,
  });
  
  // Warn if critical values are missing
  if (!API_DOMAIN) {
    console.warn('[Config] ⚠️ API_DOMAIN is empty! SuperTokens may fail to initialize.');
  }
}

export default config;

