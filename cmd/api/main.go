package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/recipe/thirdparty"
	"github.com/supertokens/supertokens-golang/recipe/thirdparty/tpmodels"
	"github.com/supertokens/supertokens-golang/recipe/usermetadata"
	"github.com/supertokens/supertokens-golang/supertokens"
	"github.com/ysaakpr/rex/internal/api/handlers"
	"github.com/ysaakpr/rex/internal/api/router"
	"github.com/ysaakpr/rex/internal/config"
	"github.com/ysaakpr/rex/internal/database"
	"github.com/ysaakpr/rex/internal/jobs"
	"github.com/ysaakpr/rex/internal/repository"
	"github.com/ysaakpr/rex/internal/services"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting UTM Backend API Server",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Database connection established")

	// Initialize SuperTokens
	if err := initSuperTokens(cfg); err != nil {
		logger.Fatal("Failed to initialize SuperTokens", zap.Error(err))
	}
	logger.Info("SuperTokens initialized")

	// Initialize job client
	jobClient, err := jobs.NewClient(cfg.GetRedisAddr(), cfg.Redis.Password)
	if err != nil {
		logger.Fatal("Failed to initialize job client", zap.Error(err))
	}
	defer jobClient.Close()
	logger.Info("Job client initialized")

	// Initialize repositories
	tenantRepo := repository.NewTenantRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	rbacRepo := repository.NewRBACRepository(db)
	platformAdminRepo := repository.NewPlatformAdminRepository(db)
	systemUserRepo := repository.NewSystemUserRepository(db)

	// Initialize services
	rbacService := services.NewRBACService(rbacRepo)
	tenantService := services.NewTenantService(tenantRepo, memberRepo, invitationRepo, rbacRepo, jobClient)
	memberService := services.NewMemberService(memberRepo, tenantRepo, rbacRepo)
	invitationService := services.NewInvitationService(invitationRepo, memberRepo, tenantRepo, rbacRepo, jobClient, cfg)
	platformAdminService := services.NewPlatformAdminService(platformAdminRepo)
	systemUserService := services.NewSystemUserService(systemUserRepo)

	// Initialize handlers
	tenantHandler := handlers.NewTenantHandler(tenantService, db)
	memberHandler := handlers.NewMemberHandler(memberService)
	invitationHandler := handlers.NewInvitationHandler(invitationService, cfg)
	rbacHandler := handlers.NewRBACHandler(rbacService)
	platformAdminHandler := handlers.NewPlatformAdminHandler(platformAdminService)
	userHandler := handlers.NewUserHandler(logger, db)
	systemUserHandler := handlers.NewSystemUserHandler(systemUserService, logger)
	authConfigHandler := handlers.NewAuthConfigHandler(cfg)

	// Setup router
	routerDeps := &router.RouterDeps{
		TenantHandler:        tenantHandler,
		MemberHandler:        memberHandler,
		InvitationHandler:    invitationHandler,
		RBACHandler:          rbacHandler,
		PlatformAdminHandler: platformAdminHandler,
		UserHandler:          userHandler,
		SystemUserHandler:    systemUserHandler,
		AuthConfigHandler:    authConfigHandler,
		MemberRepo:           memberRepo,
		RBACService:          rbacService,
		Logger:               logger,
		DB:                   db,
	}

	r := router.SetupRouter(routerDeps)

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.App.Port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("API server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.App.Env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrString(s string) *string {
	return &s
}

func initSuperTokens(cfg *config.Config) error {
	apiBasePath := cfg.SuperTokens.APIBasePath
	websiteBasePath := "/auth"

	// Build recipe list dynamically based on configuration
	recipeList := []supertokens.Recipe{
		emailpassword.Init(nil),
		usermetadata.Init(nil), // For storing system user metadata
	}

	// Add Google OAuth if enabled
	if cfg.IsGoogleOAuthEnabled() {
		log.Printf("Google OAuth enabled with client ID: %s", cfg.SuperTokens.GoogleClientID)
		recipeList = append(recipeList, thirdparty.Init(&tpmodels.TypeInput{
			SignInAndUpFeature: tpmodels.TypeInputSignInAndUp{
				Providers: []tpmodels.ProviderInput{
					{
						Config: tpmodels.ProviderConfig{
							ThirdPartyId: "google",
							Clients: []tpmodels.ProviderClientConfig{
								{
									ClientID:     cfg.SuperTokens.GoogleClientID,
									ClientSecret: cfg.SuperTokens.GoogleClientSecret,
								},
							},
						},
					},
				},
			},
		}))
	} else {
		log.Println("Google OAuth disabled - no credentials provided")
	}

	// Add session recipe with automatic secure cookie detection
	recipeList = append(recipeList, session.Init(&sessmodels.TypeInput{
		// Don't set CookieSecure - let SuperTokens auto-detect from the request protocol
		// This allows multiple frontends (HTTP and HTTPS) to work with the same backend
		// SuperTokens will set secure=true for HTTPS requests and secure=false for HTTP requests
		CookieSameSite: ptrString("lax"), // Allow cross-origin requests with lax policy
		// Don't set CookieDomain - let it default to the request domain
		// This allows cookies to work properly with the frontend proxy

		// Support both cookie and header-based authentication
		GetTokenTransferMethod: func(req *http.Request, forCreateNewSession bool, userContext supertokens.UserContext) sessmodels.TokenTransferMethod {
			// Check if client requested header-based auth
			authMode := req.Header.Get("st-auth-mode")
			if authMode == "header" {
				return sessmodels.HeaderTransferMethod
			}
			// Default to cookie-based auth
			return sessmodels.CookieTransferMethod
		},

		// Override session creation to add metadata and customize expiry
		Override: &sessmodels.OverrideStruct{
			Functions: func(originalImplementation sessmodels.RecipeInterface) sessmodels.RecipeInterface {
				originalCreateNewSession := *originalImplementation.CreateNewSession

				*originalImplementation.CreateNewSession = func(
					userID string,
					accessTokenPayload map[string]interface{},
					sessionDataInDatabase map[string]interface{},
					disableAntiCsrf *bool,
					tenantId string,
					userContext supertokens.UserContext,
				) (sessmodels.SessionContainer, error) {
					// Fetch user metadata from SuperTokens
					metadata, err := usermetadata.GetUserMetadata(userID)
					if err == nil && metadata != nil {
						// Check if this is a system user
						if isSystemUser, ok := metadata["is_system_user"].(bool); ok && isSystemUser {
							// Add system user flags to token payload
							accessTokenPayload["is_system_user"] = true
							if serviceName, ok := metadata["service_name"].(string); ok {
								accessTokenPayload["service_name"] = serviceName
							}
							if serviceType, ok := metadata["service_type"].(string); ok {
								accessTokenPayload["service_type"] = serviceType
							}
						}
					}

					return originalCreateNewSession(userID, accessTokenPayload, sessionDataInDatabase, disableAntiCsrf, tenantId, userContext)
				}

				return originalImplementation
			},
		},
	}))

	err := supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.SuperTokens.ConnectionURI,
			APIKey:        cfg.SuperTokens.APIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         "UTM Backend",
			APIDomain:       cfg.SuperTokens.APIDomain,
			WebsiteDomain:   cfg.SuperTokens.WebsiteDomain,
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: recipeList,
	})

	return err
}
