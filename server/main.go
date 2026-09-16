package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	redisclient "github.com/codegouvaor/code/server/internal/redis"
	"github.com/codegouvaor/code/server/src/config"
	"github.com/codegouvaor/code/server/src/middleware"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/providers/giteria"
	"github.com/codegouvaor/code/server/src/providers/github"
	"github.com/codegouvaor/code/server/src/providers/gitlab"
	"github.com/codegouvaor/code/server/src/routes"
	"github.com/codegouvaor/code/server/src/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type runtimeMode string

const (
	modeAPI    runtimeMode = "api"
	modeServer runtimeMode = "server"
)

func connectDatabase(ctx context.Context, logger *slog.Logger, cfg config.Config) (*services.DatabaseService, error) {
	db, err := services.NewDatabaseService(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info(
		"database connected",
		"service", "database",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"name", cfg.Database.Name,
	)

	return db, nil
}

func parseRuntimeMode(args []string) (runtimeMode, error) {
	if len(args) == 0 {
		return modeAPI, nil
	}
	switch runtimeMode(args[0]) {
	case modeAPI, modeServer:
		return modeAPI, nil
	default:
		return "", fmt.Errorf("unknown mode %q", args[0])
	}
}

// buildProviderRegistry registers the forge adapters the platform knows about.
// A provider that is not configured is simply not registered: capabilities and
// descriptors are always derived from the registry, never hard-coded in a
// handler.
func buildProviderRegistry(cfg config.ProvidersConfig) *providers.Registry {
	registry := providers.NewRegistry()
	if cfg.GitHub.Enabled {
		registry.Register(github.Descriptor, github.Factory)
	}
	if cfg.GitLab.Enabled {
		registry.Register(gitlab.Descriptor, gitlab.Factory)
	}
	if cfg.Giteria.Enabled {
		registry.Register(giteria.Descriptor, giteria.Factory)
	}
	return registry
}

// oauthScopesFor mirrors the scopes requested by the OAuth service for the
// providers that support repository access.
func oauthScopesFor(provider string) []string {
	switch provider {
	case models.ProviderGitHub:
		return []string{"read:user", "user:email"}
	case models.ProviderGitLab:
		return []string{"read_api", "read_user"}
	case models.ProviderGiteria:
		return []string{"read:repository", "read:user", "read:organization"}
	default:
		return nil
	}
}

// startReconciliation schedules the periodic reconciliation job that refreshes
// bindings which have not been synchronised recently.
func startReconciliation(ctx context.Context, logger *slog.Logger, jobs *services.JobService, interval time.Duration) {
	if jobs == nil || interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bucket := time.Now().UTC().Format("20060102T15")
				if _, err := jobs.Submit(ctx, services.EnqueueInput{
					Kind:           models.SyncJobReconciliation,
					IdempotencyKey: "reconciliation:" + bucket,
				}); err != nil {
					logger.Warn("reconciliation not scheduled", "error", err)
				}
			}
		}
	}()
}

func runHTTPServer(ctx context.Context, logger *slog.Logger, cfg config.Config, handler http.Handler, serviceRole runtimeMode) error {
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "service", "http", "port", cfg.Server.Port, "mode", string(serviceRole))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := connectDatabase(context.Background(), logger, cfg)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	redis, err := redisclient.New(redisclient.Config{
		Enabled:        cfg.Redis.Enabled,
		Required:       cfg.Redis.Required,
		URL:            cfg.Redis.URL,
		Host:           cfg.Redis.Host,
		Port:           cfg.Redis.Port,
		Password:       cfg.Redis.Password,
		DB:             cfg.Redis.DB,
		KeyPrefix:      cfg.Redis.KeyPrefix,
		DefaultTTL:     cfg.Redis.DefaultTTL,
		ConnectTimeout: cfg.Redis.ConnectTimeout,
		ReadTimeout:    cfg.Redis.ReadTimeout,
		WriteTimeout:   cfg.Redis.WriteTimeout,
		MaxRetries:     cfg.Redis.MaxRetries,
	})
	if err != nil {
		logger.Error("redis initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if redis != nil {
			_ = redis.Close()
		}
	}()

	repos := services.NewRepositories(db.Gorm())
	identityProvider := services.NewIdentityProvider(cfg.Auth, repos)
	authLimiter := services.NewAuthRateLimiter(redis)
	eventBus := services.NewEventBus(cfg, redis)
	defer eventBus.Close()
	presence := services.NewPresenceService(logger, redis, eventBus, repos.Users(), 75*time.Second)
	defer presence.Close()

	userService := services.NewUserService(repos.Users(), repos.UserSettings(), repos.NotificationPreferences(), presence)
	workspaceService := services.NewWorkspaceService(db, cfg.Auth, repos.Users(), repos)
	authService := services.NewAuthService(cfg.Auth, db, repos, identityProvider, authLimiter, workspaceService)

	// S'assurer que le premier utilisateur a les rôles superadmin
	if err := authService.EnsureFirstUserHasAdminRoles(context.Background()); err != nil {
		logger.Warn("failed to ensure first user has admin roles", "error", err)
	}

	oauthService := services.NewOAuthService(cfg.OAuth, repos, authService, identityProvider, workspaceService, nil)
	mfaService := services.NewMfaService(cfg.Auth, db, repos)

	// ── Platform layer ───────────────────────────────────────────────────────
	// The Go API is the business facade of Code: projects, organizations and
	// provider access all live here, never in the frontend.
	registry := buildProviderRegistry(cfg.Providers)
	providerCache := redisclient.NewCache(redis, cfg.Providers.CacheTTL)
	authorizer := services.NewAuthorizer(
		repos.Projects(),
		repos.ProjectMembers(),
		repos.Organizations(),
		repos.OrganizationMembers(),
		repos.OrganizationTeamMembers(),
	)
	connectionService, err := services.NewProviderConnectionService(
		repos.ProviderConnections(), registry, cfg.Providers, eventBus,
	)
	if err != nil {
		logger.Error("provider connection service unavailable", "error", err)
		os.Exit(1)
	}
	jobService := services.NewJobService(repos.SyncJobs(), eventBus, logger)
	ownerService := services.NewOwnerService(
		repos.Users(), repos.Organizations(), repos.OrganizationMembers(), repos.OrganizationTeams(),
		repos.Projects(), repos.RepositoryBindings(), repos.ProjectStars(), authorizer,
	)
	organizationService := services.NewOrganizationService(
		repos.Organizations(), repos.OrganizationMembers(), repos.OrganizationTeams(),
		repos.OrganizationTeamMembers(), repos.Users(), eventBus, authorizer, ownerService,
	)
	projectService := services.NewProjectService(
		repos.Projects(), repos.ProjectMembers(), repos.ProjectAssets(), repos.ProjectStars(),
		repos.ProjectWatches(), repos.RepositoryBindings(), repos.Organizations(),
		repos.OrganizationMembers(), repos.Users(), registry, eventBus, authorizer,
	)
	syncService := services.NewSyncService(
		repos.Projects(), repos.RepositoryBindings(), repos.ExternalResources(), repos.SyncCursors(),
		repos.WebhookSubscriptions(), repos.Users(), connectionService, eventBus, jobService,
		authorizer, logger,
	)
	syncService.RegisterHandlers()
	repositoryService := services.NewRepositoryService(
		repos.Projects(), repos.RepositoryBindings(), repos.Organizations(), repos.Users(),
		repos.ExternalResources(), connectionService, authorizer, providerCache,
		cfg.Redis.KeyPrefix, cfg.Providers.CacheTTL, eventBus, logger,
	)
	searchService := services.NewSearchService(
		services.NewPostgresSearchEngine(repos.Projects(), repos.Organizations(), repos.ProjectAssets(), repos.Users()),
	)

	// The OAuth flow is reused for repository access: the callback stores a
	// ProviderConnection, never a second identity system.
	oauthService.SetConnectionHandler(func(
		ctx context.Context, provider, userID string, info *services.OAuthUserInfo, token *oauth2.Token,
	) error {
		var expiresAt *time.Time
		if !token.Expiry.IsZero() {
			expiry := token.Expiry.UTC()
			expiresAt = &expiry
		}
		_, upsertErr := connectionService.Upsert(ctx, userID, services.ConnectionInput{
			Provider:          provider,
			ProviderAccountID: info.ID,
			AccountLogin:      info.Login,
			AvatarURL:         info.AvatarURL,
			Scopes:            strings.Join(oauthScopesFor(provider), " "),
			AccessToken:       token.AccessToken,
			RefreshToken:      token.RefreshToken,
			ExpiresAt:         expiresAt,
		})
		return upsertErr
	})

	jobService.Start(ctx)
	defer jobService.Stop()
	startReconciliation(ctx, logger, jobService, 30*time.Minute)

	mode, err := parseRuntimeMode(os.Args[1:])
	if err != nil {
		logger.Error("unknown mode", "error", err)
		os.Exit(1)
	}

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.Recovery(logger), middleware.CORS(cfg.CORS.AllowedOrigins))
	if cfg.App.AccessLogs {
		router.Use(middleware.Logging(logger))
	}
	if len(cfg.App.TrustedProxies) > 0 {
		_ = router.SetTrustedProxies(cfg.App.TrustedProxies)
	}
	routes.SetupRoutes(router, routes.Dependencies{
		Config:              cfg,
		Logger:              logger,
		Database:            db,
		Redis:               redis,
		EventBus:            eventBus,
		IdentityProvider:    identityProvider,
		AuthService:         authService,
		OAuthService:        oauthService,
		UserService:         userService,
		WorkspaceService:    workspaceService,
		Repos:               repos,
		MfaService:          mfaService,
		RuntimeRole:         string(mode),
		OwnerService:        ownerService,
		OrganizationService: organizationService,
		ProjectService:      projectService,
		RepositoryService:   repositoryService,
		ConnectionService:   connectionService,
		SyncService:         syncService,
		JobService:          jobService,
		SearchService:       searchService,
	})

	if err := runHTTPServer(ctx, logger, cfg, router, mode); err != nil {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
