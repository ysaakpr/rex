#!/bin/bash

# Script to create placeholder pages for all missing documentation

cd "$(dirname "$0")/docs"

# Create directory if it doesn't exist and add placeholder content
create_page() {
    local path="$1"
    local title="$2"
    local content="$3"
    
    mkdir -p "$(dirname "$path")"
    
    if [ ! -f "$path" ]; then
        cat > "$path" << EOF
# $title

$content

::: tip Coming Soon
This page is under development. Please check back later for detailed documentation.
:::

## Related Documentation

- [Overview](/introduction/overview)
- [Quick Start](/getting-started/quick-start)
- [API Reference](/api/overview)
EOF
        echo "Created: $path"
    else
        echo "Exists: $path"
    fi
}

# API Reference Pages
create_page "api/authentication.md" "Authentication API" "SuperTokens authentication endpoints for sign up, sign in, and session management."
create_page "api/tenants.md" "Tenants API" "API endpoints for managing tenants (create, read, update, delete)."
create_page "api/members.md" "Members API" "API endpoints for managing tenant members and role assignments."
create_page "api/invitations.md" "Invitations API" "API endpoints for creating and managing user invitations."
create_page "api/rbac.md" "RBAC API" "API endpoints for managing roles, policies, and permissions."
create_page "api/system-users.md" "System Users API" "API endpoints for managing system users (M2M authentication)."
create_page "api/platform-admin.md" "Platform Admin API" "API endpoints for platform administration."
create_page "api/users.md" "Users API" "API endpoints for user information and management."

# Frontend Integration
create_page "frontend/react-setup.md" "React Setup" "Setting up SuperTokens and React Router in your frontend."
create_page "frontend/api-calls.md" "Making API Calls" "How to make authenticated API calls from the frontend."
create_page "frontend/protected-routes.md" "Protected Routes" "Implementing protected routes with SuperTokens SessionAuth."
create_page "frontend/invitation-flow.md" "Invitation Flow" "Implementing the invitation acceptance flow in React."
create_page "frontend/component-examples.md" "Component Examples" "Example React components for common patterns."

# RBAC Guides
create_page "guides/rbac-overview.md" "RBAC Overview" "Understanding the 3-tier RBAC system (Roles → Policies → Permissions)."
create_page "guides/roles-policies.md" "Roles & Policies" "Managing roles and policies in your application."
create_page "guides/permissions.md" "Permissions" "Understanding and working with permissions."
create_page "guides/managing-rbac.md" "Managing RBAC" "Best practices for RBAC management."
create_page "guides/creating-tenants.md" "Creating Tenants" "Step-by-step guide to tenant creation."
create_page "guides/member-management.md" "Member Management" "Adding, removing, and managing tenant members."
create_page "guides/invitations.md" "Invitations" "User invitation system and workflows."
create_page "guides/session-management.md" "Session Management" "Advanced session management topics."

# Middleware
create_page "middleware/nodejs.md" "Node.js Middleware" "Authentication middleware for Node.js/Express."
create_page "middleware/python.md" "Python Middleware" "Authentication middleware for Python/Flask."
create_page "middleware/csharp.md" "C# Middleware" "Authentication middleware for ASP.NET Core."
create_page "middleware/go.md" "Go Middleware" "Authentication middleware for Go (reference implementation)."

# System Auth
create_page "system-auth/java.md" "Java Implementation" "System Auth Library for Java applications."
create_page "system-auth/custom-vaults.md" "Custom Vaults" "Implementing custom secret vaults for credentials."
create_page "system-auth/usage.md" "Usage Examples" "Real-world usage examples for the System Auth Library."

# Jobs
create_page "jobs/architecture.md" "Background Jobs Architecture" "How the background job system works with Asynq."
create_page "jobs/available-jobs.md" "Available Jobs" "List of built-in background jobs."
create_page "jobs/custom-jobs.md" "Creating Custom Jobs" "How to create your own background jobs."
create_page "jobs/monitoring.md" "Job Monitoring" "Monitoring and managing background jobs."

# Deployment
create_page "deployment/environment.md" "Environment Variables" "Complete environment variable reference."
create_page "deployment/migrations.md" "Database Migrations" "Managing database schema migrations."
create_page "deployment/first-admin.md" "First Admin Setup" "Creating the first platform administrator."

# Examples
create_page "examples/user-journey.md" "Complete User Journey" "End-to-end example of user workflows."
create_page "examples/m2m-integration.md" "M2M Integration" "Integrating external services with system users."
create_page "examples/custom-rbac.md" "Custom RBAC Setup" "Setting up custom roles and permissions."
create_page "examples/credential-rotation.md" "Credential Rotation" "Rotating system user credentials safely."

# Advanced
create_page "advanced/custom-middleware.md" "Custom Middleware" "Creating custom authentication middleware."
create_page "advanced/permission-hooks.md" "Permission Hooks" "Implementing dynamic permission checks."
create_page "advanced/webhooks.md" "Webhook System" "Setting up webhooks for events."

# Reference
create_page "reference/database-schema.md" "Database Schema" "Complete database schema reference."
create_page "reference/error-codes.md" "Error Codes" "API error codes and meanings."
create_page "reference/glossary.md" "Glossary" "Definitions of terms used in the documentation."

# Troubleshooting
create_page "troubleshooting/common-issues.md" "Common Issues" "Solutions to frequently encountered problems."
create_page "troubleshooting/debug-mode.md" "Debug Mode" "Enabling and using debug mode."

echo ""
echo "✅ All missing pages created!"
echo "Run 'npm run docs:dev' to view the documentation"
EOF

chmod +x /Users/vyshakhp/work/utm-backend/docs-website/create-missing-pages.sh

