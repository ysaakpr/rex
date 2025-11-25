import { defineConfig } from 'vitepress'

// Load demo URL from environment variable with fallback
const demoUrl = process.env.VITE_DEMO_URL || 'http://localhost/demo'

export default defineConfig({
  title: "Rex",
  description: "Open-Source User Management Platform with Multi-Tenancy and RBAC",
  base: '/',
  
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }]
  ],

  // Make demo URL available globally
  vite: {
    define: {
      __DEMO_URL__: JSON.stringify(demoUrl)
    }
  },

  themeConfig: {
    logo: '/logo.svg',
    
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started/quick-start' },
      { text: 'Guides', link: '/guides/authentication' },
      { text: 'API Reference', link: '/x-api/overview' },
      { text: 'Examples', link: '/examples/user-journey' },
      { text: 'Try Demo', link: demoUrl }
    ],

    sidebar: [
      {
        text: 'Introduction',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/introduction/overview' },
          { text: 'Architecture', link: '/introduction/architecture' },
          { text: 'Core Concepts', link: '/introduction/core-concepts' }
        ]
      },
      {
        text: 'Getting Started',
        collapsed: false,
        items: [
          { text: 'Quick Start', link: '/getting-started/quick-start' },
          { text: 'Installation', link: '/getting-started/installation' },
          { text: 'Project Structure', link: '/getting-started/project-structure' },
          { text: 'Configuration', link: '/getting-started/configuration' }
        ]
      },
      {
        text: 'Authentication',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/guides/authentication' },
          { text: 'User Authentication', link: '/guides/user-authentication' },
          { text: 'System Users (M2M)', link: '/guides/system-users' },
          { text: 'Session Management', link: '/guides/session-management' }
        ]
      },
      {
        text: 'Authorization (RBAC)',
        collapsed: false,
        items: [
          { text: 'RBAC Overview', link: '/guides/rbac-overview' },
          { text: 'Roles & Policies', link: '/guides/roles-policies' },
          { text: 'Permissions', link: '/guides/permissions' },
          { text: 'Managing RBAC', link: '/guides/managing-rbac' }
        ]
      },
      {
        text: 'Tenant Management',
        collapsed: false,
        items: [
          { text: 'Multi-Tenancy', link: '/guides/multi-tenancy' },
          { text: 'Creating Tenants', link: '/guides/creating-tenants' },
          { text: 'Member Management', link: '/guides/member-management' },
          { text: 'Invitations', link: '/guides/invitations' }
        ]
      },
      {
        text: 'API Reference',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/x-api/overview' },
          { text: 'Authentication', link: '/x-api/authentication' },
          { text: 'Tenants', link: '/x-api/tenants' },
          { text: 'Members', link: '/x-api/members' },
          { text: 'Invitations', link: '/x-api/invitations' },
          { text: 'RBAC', link: '/x-api/rbac' },
          { text: 'System Users', link: '/x-api/system-users' },
          { text: 'Platform Admin', link: '/x-api/platform-admin' },
          { text: 'Users', link: '/x-api/users' }
        ]
      },
      {
        text: 'Frontend Integration',
        collapsed: false,
        items: [
          { text: 'React Setup', link: '/frontend/react-setup' },
          { text: 'Making API Calls', link: '/frontend/api-calls' },
          { text: 'Protected Routes', link: '/frontend/protected-routes' },
          { text: 'Invitation Flow', link: '/frontend/invitation-flow' },
          { text: 'Component Examples', link: '/frontend/component-examples' }
        ]
      },
      {
        text: 'Middleware Examples',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/middleware/overview' },
          { text: 'Go (Reference)', link: '/middleware/go' },
          { text: 'Node.js/Express', link: '/middleware/nodejs' },
          { text: 'Python/Flask', link: '/middleware/python' },
          { text: 'Java (Manual)', link: '/middleware/java' },
          { text: 'C# (.NET)', link: '/middleware/csharp' }
        ]
      },
      {
        text: 'System Auth Library',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/system-auth/overview' },
          { text: 'Go Implementation', link: '/system-auth/go' },
          { text: 'Java Implementation', link: '/system-auth/java' },
          { text: 'Custom Vaults', link: '/system-auth/custom-vaults' },
          { text: 'Usage Examples', link: '/system-auth/usage' }
        ]
      },
      {
        text: 'Background Jobs',
        collapsed: false,
        items: [
          { text: 'Architecture', link: '/jobs/architecture' },
          { text: 'Available Jobs', link: '/jobs/available-jobs' },
          { text: 'Creating Custom Jobs', link: '/jobs/custom-jobs' },
          { text: 'Monitoring', link: '/jobs/monitoring' }
        ]
      },
      {
        text: 'Deployment',
        collapsed: false,
        items: [
          { text: 'Environment Variables', link: '/deployment/environment' },
          { text: 'Docker', link: '/deployment/docker' },
          { text: 'Database Migrations', link: '/deployment/migrations' },
          { text: 'First Admin Setup', link: '/deployment/first-admin' }
        ]
      },
      {
        text: 'Examples & Recipes',
        collapsed: false,
        items: [
          { text: 'Complete User Journey', link: '/examples/user-journey' },
          { text: 'M2M Integration', link: '/examples/m2m-integration' },
          { text: 'Custom RBAC Setup', link: '/examples/custom-rbac' },
          { text: 'Credential Rotation', link: '/examples/credential-rotation' }
        ]
      },
      {
        text: 'Advanced Topics',
        collapsed: false,
        items: [
          { text: 'Custom Middleware', link: '/advanced/custom-middleware' },
          { text: 'Permission Hooks', link: '/advanced/permission-hooks' },
          { text: 'Webhook System', link: '/advanced/webhooks' }
        ]
      },
      {
        text: 'Troubleshooting',
        collapsed: false,
        items: [
          { text: 'Common Issues', link: '/troubleshooting/common-issues' },
          { text: 'Debug Mode', link: '/troubleshooting/debug-mode' }
        ]
      },
      {
        text: 'Reference',
        collapsed: false,
        items: [
          { text: 'Database Schema', link: '/reference/database-schema' },
          { text: 'Error Codes', link: '/reference/error-codes' },
          { text: 'Glossary', link: '/reference/glossary' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/ysaakpr/rex' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present'
    },

    search: {
      provider: 'local'
    }
  }
})

