# Documentation Index

## 📁 Quick Navigation

### Getting Started
- **[Main README](../README.md)** - Project overview and quick start
- **[QUICKSTART](QUICKSTART.md)** - Step-by-step getting started guide
- **[SSL Quick Start](SSL_QUICK_START.md)** - HTTPS setup for localhost and production ⭐ NEW
- **[Authentication Implementation](AUTHENTICATION_IMPLEMENTATION.md)** - Frontend/Backend auth setup + Stateless vs Stateful ⭐ NEW
- **[API Authentication Guide](API_AUTHENTICATION_GUIDE.md)** - Complete curl examples with authentication
- **[API Examples](API_EXAMPLES.md)** - API endpoint examples and usage

### Implementation Summaries
- **[RBAC Refactoring Summary](RBAC_REFACTORING_SUMMARY.md)** - Complete RBAC terminology refactoring (Relations→Roles, Roles→Policies)
- **[Phase 2 Improvements Complete](PHASE2_IMPROVEMENTS_COMPLETE.md)** - Tenant management experience improvements

### Change Documentation (Implementation Journey)
All implementation docs are in `changedoc/` with sequence numbers:

1. **[01-TESTING.md](changedoc/01-TESTING.md)** - Initial testing guide
2. **[02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md)** - Authentication testing comprehensive guide
3. **[03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md)** - 2-minute quick test guide
4. **[04-FRONTEND_DEMO.md](changedoc/04-FRONTEND_DEMO.md)** - Frontend features and demo
5. **[05-IMPLEMENTATION_COMPLETE.md](changedoc/05-IMPLEMENTATION_COMPLETE.md)** - Final implementation summary
6. **[06-PLATFORM_ADMIN_DESIGN.md](changedoc/06-PLATFORM_ADMIN_DESIGN.md)** - Platform admin architecture design
7. **[07-PLATFORM_ADMIN_COMPLETE.md](changedoc/07-PLATFORM_ADMIN_COMPLETE.md)** - Platform admin implementation guide
8. **[08-UI_REWORK_MODERN_DASHBOARD.md](changedoc/08-UI_REWORK_MODERN_DASHBOARD.md)** - Modern dashboard UI implementation
9. **[09-PHASE1_COMPLETE.md](changedoc/09-PHASE1_COMPLETE.md)** - Phase 1 milestone completion
10. **[10-PHASE2_TENANTS_COMPLETE.md](changedoc/10-PHASE2_TENANTS_COMPLETE.md)** - Phase 2 tenant features completion
11. **[11-AWS_DEPLOYMENT.md](changedoc/11-AWS_DEPLOYMENT.md)** - AWS deployment with Pulumi
12. **[19-GOOGLE_OAUTH.md](changedoc/19-GOOGLE_OAUTH.md)** - Optional Google OAuth integration
13. **[20-LETSENCRYPT_SSL_SETUP.md](changedoc/20-LETSENCRYPT_SSL_SETUP.md)** - SSL/HTTPS setup with Let's Encrypt ⭐ NEW

See [changedoc/README.md](changedoc/README.md) for detailed descriptions.

### Frontend Documentation
- **[Frontend README](../frontend/README.md)** - Frontend-specific documentation
- **[Frontend Package](../frontend/package.json)** - Dependencies and scripts

## 📚 Documentation by Topic

### Authentication & Security
- [AUTHENTICATION_IMPLEMENTATION.md](AUTHENTICATION_IMPLEMENTATION.md) - Complete implementation guide: Frontend setup + Backend stateless vs stateful ⭐ NEW
- **[SECURITY_ARCHITECTURE.md](SECURITY_ARCHITECTURE.md)** - SuperTokens security architecture and best practices ⭐ NEW
- **[19-GOOGLE_OAUTH.md](changedoc/19-GOOGLE_OAUTH.md)** - Optional Google OAuth login integration ⭐ NEW
- **[RBAC_AUTHORIZATION_GUIDE.md](RBAC_AUTHORIZATION_GUIDE.md)** - Complete RBAC authorization guide with backend/frontend examples ⭐ NEW
- [API_AUTHENTICATION_GUIDE.md](API_AUTHENTICATION_GUIDE.md) - Complete auth guide + Token reference (Access, Refresh, Front)
- [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) - Cookie and header-based auth
- [Main README](../README.md#authentication) - Authentication overview

### API Testing
- [API_AUTHENTICATION_GUIDE.md](API_AUTHENTICATION_GUIDE.md) - Complete authentication workflow with curl + Token deep dive ⭐ UPDATED
- [API_EXAMPLES.md](API_EXAMPLES.md) - Complete API reference with cURL examples
- [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) - Authentication in API calls
- [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md) - Quick API verification

### Frontend Development
- [04-FRONTEND_DEMO.md](changedoc/04-FRONTEND_DEMO.md) - UI features and customization
- [Frontend README](../frontend/README.md) - Technical documentation
- [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md#test-1-frontend) - Frontend testing

### Database
- [Main README](../README.md#database) - Database schema overview
- [QUICKSTART.md](QUICKSTART.md) - Migration commands

### Deployment
- **[AWS Deployment Guide](AWS_DEPLOYMENT_GUIDE.md)** - Complete AWS deployment guide ⭐ NEW
- **[20-LETSENCRYPT_SSL_SETUP.md](changedoc/20-LETSENCRYPT_SSL_SETUP.md)** - SSL/HTTPS certificates with Let's Encrypt ⭐ NEW
- **[Platform Compatibility](PLATFORM_COMPATIBILITY.md)** - ARM64/AMD64 multi-arch support ⭐ NEW
- **[Nginx Proxy Guide](NGINX_PROXY_GUIDE.md)** - Reverse proxy setup and configuration ⭐ NEW
- **[Custom Domains Guide](CUSTOM_DOMAINS.md)** - Custom domains and host configuration ⭐ NEW
- **[11-AWS_DEPLOYMENT.md](changedoc/11-AWS_DEPLOYMENT.md)** - Pulumi infrastructure documentation ⭐ NEW
- [Infrastructure Code](../infra/) - Pulumi Go infrastructure
- [Docker Compose](../docker-compose.yml) - Local development orchestration
- [Dockerfile](../Dockerfile) - Development container
- [Dockerfile.prod](../Dockerfile.prod) - Production container (API & Worker)
- [Frontend Dockerfile.prod](../frontend/Dockerfile.prod) - Production frontend container

### Troubleshooting
- [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md#troubleshooting) - Auth issues
- [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md#troubleshooting) - Common problems
- [01-TESTING.md](changedoc/01-TESTING.md#troubleshooting) - Testing issues

## 🎯 Documentation for Different Roles

### New Developers
1. Start: [Main README](../README.md)
2. Setup: [QUICKSTART.md](QUICKSTART.md)
3. Test: [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md)
4. Understand: [05-IMPLEMENTATION_COMPLETE.md](changedoc/05-IMPLEMENTATION_COMPLETE.md)
5. Build: [04-FRONTEND_DEMO.md](changedoc/04-FRONTEND_DEMO.md)

### Frontend Developers
1. [04-FRONTEND_DEMO.md](changedoc/04-FRONTEND_DEMO.md) - UI overview
2. [Frontend README](../frontend/README.md) - Setup and dev workflow
3. [RBAC_AUTHORIZATION_GUIDE.md](RBAC_AUTHORIZATION_GUIDE.md) - Permission checks in React ⭐ NEW
4. [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) - Auth integration
5. [API_EXAMPLES.md](API_EXAMPLES.md) - Backend API reference

### Backend Developers
1. [Main README](../README.md) - Architecture
2. [API_EXAMPLES.md](API_EXAMPLES.md) - API patterns
3. [RBAC_AUTHORIZATION_GUIDE.md](RBAC_AUTHORIZATION_GUIDE.md) - Authorization implementation ⭐ NEW
4. [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) - Auth implementation
5. [QUICKSTART.md](QUICKSTART.md) - Database migrations

### QA/Testers
1. [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md) - Quick verification
2. [01-TESTING.md](changedoc/01-TESTING.md) - Comprehensive testing
3. [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) - Auth testing
4. [API_EXAMPLES.md](API_EXAMPLES.md) - API endpoints

### DevOps/SRE
1. **[AWS Deployment Guide](AWS_DEPLOYMENT_GUIDE.md)** - Production deployment ⭐ NEW
2. **[20-LETSENCRYPT_SSL_SETUP.md](changedoc/20-LETSENCRYPT_SSL_SETUP.md)** - SSL/HTTPS setup ⭐ NEW
3. **[Nginx Proxy Guide](NGINX_PROXY_GUIDE.md)** - Reverse proxy configuration ⭐ NEW
4. **[11-AWS_DEPLOYMENT.md](changedoc/11-AWS_DEPLOYMENT.md)** - Infrastructure as code ⭐ NEW
5. [Infrastructure README](../infra/README.md) - Pulumi infrastructure docs
6. [Docker Compose](../docker-compose.yml) - Local service definitions
7. [Makefile](../Makefile) - Common commands
8. [QUICKSTART.md](QUICKSTART.md) - Setup procedures
9. [05-IMPLEMENTATION_COMPLETE.md](changedoc/05-IMPLEMENTATION_COMPLETE.md) - System overview

## 📊 Documentation Metrics

- **Total Docs**: 22+ files (UPDATED)
- **Change Docs**: 13 sequenced files (UPDATED)
- **Code Examples**: 250+ snippets (UPDATED)
- **API Endpoints**: 40+ documented
- **Testing Guides**: 4 comprehensive guides
- **Implementation Guides**: 5 (Frontend/Backend auth, API usage, RBAC authorization, AWS deployment, SSL/HTTPS) (UPDATED)
- **Infrastructure Code**: 15+ Pulumi modules in Go
- **Deployment Scripts**: 8 automated scripts (UPDATED)

## 🔄 Documentation Maintenance

### Updating Documentation

When making significant changes:

1. Create new numbered doc in `changedoc/` if it's a major milestone
2. Update relevant existing docs
3. Update this INDEX.md
4. Update main README.md if needed

### Naming Convention

- Change docs: `##-DESCRIPTIVE_NAME.md` (sequence number + description)
- Feature docs: `FEATURE_NAME.md` (uppercase with underscores)
- Technical docs: `lowercase-with-dashes.md`

## 🔗 External Resources

- **SuperTokens**: https://supertokens.com/docs
- **Gin Framework**: https://gin-gonic.com/docs/
- **Vite**: https://vitejs.dev
- **React**: https://react.dev
- **PostgreSQL**: https://www.postgresql.org/docs/

## 📞 Quick Links

| What I Need | Where to Look |
|-------------|---------------|
| Quick start project | [QUICKSTART.md](QUICKSTART.md) |
| **Deploy to AWS** | **[AWS_DEPLOYMENT_GUIDE.md](AWS_DEPLOYMENT_GUIDE.md)** ⭐ NEW |
| **Setup HTTPS/SSL** | **[20-LETSENCRYPT_SSL_SETUP.md](changedoc/20-LETSENCRYPT_SSL_SETUP.md)** ⭐ NEW |
| **Setup nginx proxy** | **[NGINX_PROXY_GUIDE.md](NGINX_PROXY_GUIDE.md)** ⭐ NEW |
| **Infrastructure code** | **[infra/README.md](../infra/README.md)** ⭐ NEW |
| Setup frontend auth | [AUTHENTICATION_IMPLEMENTATION.md#part-1-frontend-authentication-setup](AUTHENTICATION_IMPLEMENTATION.md#part-1-frontend-authentication-setup) |
| **Add Google login** | **[19-GOOGLE_OAUTH.md](changedoc/19-GOOGLE_OAUTH.md)** ⭐ NEW |
| Stateless vs Stateful | [AUTHENTICATION_IMPLEMENTATION.md#part-2-backend-token-verification](AUTHENTICATION_IMPLEMENTATION.md#part-2-backend-token-verification) |
| Test with curl | [API_AUTHENTICATION_GUIDE.md](API_AUTHENTICATION_GUIDE.md) |
| Understand tokens | [API_AUTHENTICATION_GUIDE.md#understanding-supertokens-tokens](API_AUTHENTICATION_GUIDE.md#understanding-supertokens-tokens) |
| **Check permissions** | **[RBAC_AUTHORIZATION_GUIDE.md](RBAC_AUTHORIZATION_GUIDE.md)** ⭐ NEW |
| Test the system | [03-QUICK_TEST.md](changedoc/03-QUICK_TEST.md) |
| Use the API | [API_EXAMPLES.md](API_EXAMPLES.md) |
| Understand auth | [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md) |
| Build frontend | [04-FRONTEND_DEMO.md](changedoc/04-FRONTEND_DEMO.md) |
| See what's done | [05-IMPLEMENTATION_COMPLETE.md](changedoc/05-IMPLEMENTATION_COMPLETE.md) |
| Troubleshoot | [02-AUTH_TESTING.md](changedoc/02-AUTH_TESTING.md#troubleshooting) |

---

**Last Updated**: November 24, 2025  
**Maintained By**: Development Team  
**Status**: AWS Deployment Ready with HTTPS 🚀 ✅ 🔒  
**Authentication**: Email/Password + Optional Google OAuth  
**Security**: TLS 1.2/1.3 + Let's Encrypt + Auto-renewal

