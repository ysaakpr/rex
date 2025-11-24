package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ysaakpr/rex/internal/api/handlers"
	"github.com/ysaakpr/rex/internal/api/middleware"
	"github.com/ysaakpr/rex/internal/repository"
	"github.com/ysaakpr/rex/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RouterDeps struct {
	TenantHandler        *handlers.TenantHandler
	MemberHandler        *handlers.MemberHandler
	InvitationHandler    *handlers.InvitationHandler
	RBACHandler          *handlers.RBACHandler
	PlatformAdminHandler *handlers.PlatformAdminHandler
	UserHandler          *handlers.UserHandler
	SystemUserHandler    *handlers.SystemUserHandler
	AuthConfigHandler    *handlers.AuthConfigHandler
	MemberRepo           repository.MemberRepository
	RBACService          services.RBACService
	Logger               *zap.Logger
	DB                   *gorm.DB
}

func SetupRouter(deps *RouterDeps) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(deps.Logger))
	router.Use(middleware.SuperTokensMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes - invitation details can be viewed before authentication
		v1.GET("/invitations/:token", deps.InvitationHandler.GetInvitationByToken)

		// Auth configuration endpoint - returns which OAuth providers are enabled
		v1.GET("/auth/config", deps.AuthConfigHandler.GetAuthConfig)

		// Protected routes (require authentication)
		auth := v1.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// Tenant routes
			tenants := auth.Group("/tenants")
			{
				tenants.POST("", deps.TenantHandler.CreateTenant)
				tenants.POST("/managed", deps.TenantHandler.CreateManagedTenant)
				tenants.GET("", deps.TenantHandler.ListTenants)

				// Tenant-scoped routes (require tenant membership or platform admin) - using :id consistently
				tenantScoped := tenants.Group("/:id")
				tenantScoped.Use(middleware.TenantAccessMiddleware(deps.MemberRepo, deps.DB))
				{
					// Tenant info routes
					tenantScoped.GET("", deps.TenantHandler.GetTenant)
					tenantScoped.PATCH("", deps.TenantHandler.UpdateTenant)
					tenantScoped.DELETE("", deps.TenantHandler.DeleteTenant)
					tenantScoped.GET("/status", deps.TenantHandler.GetTenantStatus)

					// Member routes
					tenantScoped.POST("/members", deps.MemberHandler.AddMember)
					tenantScoped.GET("/members", deps.MemberHandler.ListMembers)
					tenantScoped.GET("/members/:user_id", deps.MemberHandler.GetMember)
					tenantScoped.PATCH("/members/:user_id", deps.MemberHandler.UpdateMember)
					tenantScoped.DELETE("/members/:user_id", deps.MemberHandler.RemoveMember)
					tenantScoped.POST("/members/:user_id/roles", deps.MemberHandler.AssignRoles)
					tenantScoped.DELETE("/members/:user_id/roles/:role_id", deps.MemberHandler.RemoveRole)

					// Invitation routes
					tenantScoped.POST("/invitations", deps.InvitationHandler.CreateInvitation)
					tenantScoped.GET("/invitations", deps.InvitationHandler.ListInvitations)
				}
			}

			// Invitation management
			invitations := auth.Group("/invitations")
			{
				invitations.POST("/:token/accept", deps.InvitationHandler.AcceptInvitation)
				invitations.POST("/check-pending", deps.InvitationHandler.CheckPendingInvitations)
				invitations.DELETE("/:id", deps.InvitationHandler.CancelInvitation)
			}

			// User routes (for fetching user details)
			users := auth.Group("/users")
			{
				users.GET("/me", deps.UserHandler.GetCurrentUser)
				users.GET("", deps.UserHandler.ListUsers)
				users.GET("/search", deps.UserHandler.SearchUsers)
				users.GET("/:user_id", deps.UserHandler.GetUserDetails)
				users.GET("/:user_id/tenants", deps.UserHandler.GetUserTenants)
				users.POST("/batch", deps.UserHandler.GetBatchUserDetails)
			}

			// Platform Admin routes (require platform admin access)
			platform := auth.Group("/platform")
			platform.Use(middleware.PlatformAdminMiddleware(deps.DB))
			{
				// Platform admin management
				admins := platform.Group("/admins")
				{
					admins.POST("", deps.PlatformAdminHandler.CreateAdmin)
					admins.GET("", deps.PlatformAdminHandler.ListAdmins)
					admins.GET("/:user_id", deps.PlatformAdminHandler.GetAdmin)
					admins.DELETE("/:user_id", deps.PlatformAdminHandler.DeleteAdmin)
				}

				// Tenants management (all tenants)
				platform.GET("/tenants", deps.TenantHandler.ListAllTenants)
				platform.GET("/tenants/:id", deps.TenantHandler.GetTenantForPlatformAdmin)

				// System users (M2M authentication)
				systemUsers := platform.Group("/system-users")
				{
					systemUsers.POST("", deps.SystemUserHandler.CreateSystemUser)
					systemUsers.GET("", deps.SystemUserHandler.ListSystemUsers)
					systemUsers.GET("/:id", deps.SystemUserHandler.GetSystemUser)
					systemUsers.PATCH("/:id", deps.SystemUserHandler.UpdateSystemUser)
					systemUsers.POST("/:id/regenerate-password", deps.SystemUserHandler.RegeneratePassword)
					systemUsers.POST("/:id/rotate", deps.SystemUserHandler.RotateWithGracePeriod)
					systemUsers.DELETE("/:id", deps.SystemUserHandler.DeactivateSystemUser)
				}

				// Application-level credential management
				applications := platform.Group("/applications")
				{
					applications.GET("/:application_name/credentials", deps.SystemUserHandler.GetApplicationCredentials)
					applications.POST("/:application_name/revoke-old", deps.SystemUserHandler.RevokeOldCredentials)
				}

				// Roles (platform-level - user's role in tenant: Admin, Writer, etc.)
				roles := platform.Group("/roles")
				{
					roles.POST("", deps.RBACHandler.CreateRole)
					roles.GET("", deps.RBACHandler.ListRoles)
					roles.GET("/:id", deps.RBACHandler.GetRole)
					roles.PATCH("/:id", deps.RBACHandler.UpdateRole)
					roles.DELETE("/:id", deps.RBACHandler.DeleteRole)
					// Role-to-policy mapping
					roles.POST("/:id/policies", deps.RBACHandler.AssignPoliciesToRole)
					roles.GET("/:id/policies", deps.RBACHandler.GetRolePolicies)
					roles.DELETE("/:id/policies/:policy_id", deps.RBACHandler.RevokePolicyFromRole)
				}

				// Policies (platform-level - group of permissions)
				policies := platform.Group("/policies")
				{
					policies.POST("", deps.RBACHandler.CreatePolicy)
					policies.GET("", deps.RBACHandler.ListPolicies)
					policies.GET("/:id", deps.RBACHandler.GetPolicy)
					policies.PATCH("/:id", deps.RBACHandler.UpdatePolicy)
					policies.DELETE("/:id", deps.RBACHandler.DeletePolicy)
					policies.POST("/:id/permissions", deps.RBACHandler.AssignPermissionsToPolicy)
					policies.DELETE("/:id/permissions/:permission_id", deps.RBACHandler.RevokePermissionFromPolicy)
				}

				// Permissions (platform-level)
				permissions := platform.Group("/permissions")
				{
					permissions.POST("", deps.RBACHandler.CreatePermission)
					permissions.GET("", deps.RBACHandler.ListPermissions)
					permissions.GET("/:id", deps.RBACHandler.GetPermission)
					permissions.DELETE("/:id", deps.RBACHandler.DeletePermission)
				}
			}

			// Platform admin check endpoint (accessible to all authenticated users)
			auth.GET("/platform/admins/check", deps.PlatformAdminHandler.CheckPlatformAdmin)

			// Keep legacy routes for backward compatibility (deprecated)
			roles := auth.Group("/roles")
			{
				roles.GET("", deps.RBACHandler.ListRoles)
				roles.GET("/:id", deps.RBACHandler.GetRole)
			}

			policies := auth.Group("/policies")
			{
				policies.GET("", deps.RBACHandler.ListPolicies)
				policies.GET("/:id", deps.RBACHandler.GetPolicy)
			}

			permissions := auth.Group("/permissions")
			{
				permissions.GET("", deps.RBACHandler.ListPermissions)
				permissions.GET("/:id", deps.RBACHandler.GetPermission)
				permissions.GET("/user", deps.RBACHandler.GetUserPermissions)
			}

			// Authorization check endpoint
			auth.POST("/authorize", deps.RBACHandler.Authorize)
		}
	}

	return router
}
