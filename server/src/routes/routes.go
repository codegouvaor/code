package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	redisclient "github.com/codegouvaor/code/server/internal/redis"
	"github.com/codegouvaor/code/server/src/config"
	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/middleware"
	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
)

type Dependencies struct {
	Config           config.Config
	Logger           *slog.Logger
	RuntimeRole      string
	Database         interfaces.Database
	Redis            *redisclient.Client
	EventBus         interfaces.EventBus
	IdentityProvider interfaces.IdentityProvider
	AuthService      *services.AuthService
	OAuthService     *services.OAuthService
	UserService      *services.UserService
	WorkspaceService *services.WorkspaceService
	Repos            *services.Repositories
	MfaService       *services.MfaService
}

func SetupRoutes(router *gin.Engine, deps Dependencies) {
	handler := &apiHandler{deps: deps}

	router.GET("/health/live", handler.live)
	router.GET("/health/ready", handler.ready)
	router.GET("/metrics", handler.metrics)

	api := router.Group("/api/v1")
	api.GET("/health", handler.health)
	api.GET("/ready", handler.ready)
	api.POST("/integrations/webhooks/:provider/:integrationId", handler.webhook)

	auth := api.Group("/auth")
	{
		auth.POST("/register", handler.register)
		auth.POST("/login", handler.login)
		auth.POST("/refresh", handler.refresh)
		auth.POST("/logout", handler.logout)
		auth.POST("/forgot-password", handler.forgotPassword)
		auth.POST("/reset-password", handler.resetPassword)
		auth.POST("/verify-email", handler.verifyEmail)
		auth.POST("/resend-verification", handler.resendVerification)
		auth.GET("/oauth/:provider", handler.oauthLogin)
		auth.GET("/oauth/:provider/callback", handler.oauthCallback)
		authProtected := auth.Group("")
		authProtected.Use(middleware.Auth(deps.IdentityProvider))
		{
			authProtected.POST("/logout-all", handler.logoutAll)
			authProtected.POST("/change-password", handler.changePassword)
			authProtected.GET("/me", handler.authMe)
			authProtected.GET("/sessions", handler.listAuthSessions)
			authProtected.DELETE("/sessions/:sessionId", handler.deleteAuthSession)
			authProtected.GET("/oauth/accounts", handler.listOAuthAccounts)
			authProtected.DELETE("/oauth/:provider", handler.unlinkOAuthAccount)
			authProtected.POST("/ensure-first-user-admin", handler.ensureFirstUserIsAdmin)
			authProtected.GET("/first-user", handler.getFirstUserInfo)
			authProtected.POST("/ensure-user-owner", handler.ensureUserIsOwner)
		}

		mfaProtected := auth.Group("/mfa")
		mfaProtected.Use(middleware.Auth(deps.IdentityProvider))
		{
			mfaProtected.POST("/setup", handler.mfaSetup)
			mfaProtected.POST("/verify", handler.mfaVerify)
			mfaProtected.POST("/disable", handler.mfaDisable)
			mfaProtected.POST("/validate-login", handler.mfaValidateLogin)
			mfaProtected.POST("/recovery-codes", handler.mfaRecoveryCodes)
			mfaProtected.POST("/regenerate-recovery-codes", handler.mfaRegenerateRecoveryCodes)
		}
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(deps.IdentityProvider))
	protected.Use(middleware.WorkspaceContext())
	{
		protected.GET("/me", handler.me)
		protected.PATCH("/me", handler.updateMe)

		protected.GET("/workspaces", handler.listWorkspaces)
		protected.POST("/workspaces", handler.createWorkspace)
		protected.GET("/workspaces/:workspaceId", handler.getWorkspace)
		protected.PATCH("/workspaces/:workspaceId", handler.updateWorkspace)
		protected.DELETE("/workspaces/:workspaceId", handler.deleteWorkspace)
		protected.GET("/workspaces/:workspaceId/members", handler.listWorkspaceMembers)
		protected.POST("/workspaces/:workspaceId/members", handler.createWorkspaceMember)
		protected.POST("/workspaces/:workspaceId/members/provision", handler.provisionWorkspaceUser)
		protected.PATCH("/workspaces/:workspaceId/members/:userId", handler.updateWorkspaceMember)
		protected.DELETE("/workspaces/:workspaceId/members/:userId", handler.deleteWorkspaceMember)

		// Articles (slug route must come before :id to avoid conflict)
		protected.GET("/articles", handler.listArticles)
		protected.POST("/articles", handler.createArticle)
		protected.GET("/articles/slug/:slug", handler.getArticleBySlug)
		protected.GET("/articles/:id", handler.getArticle)
		protected.PATCH("/articles/:id", handler.updateArticle)
		protected.DELETE("/articles/:id", handler.deleteArticle)
		protected.POST("/articles/:id/publish", handler.publishArticle)
		protected.POST("/articles/:id/schedule", handler.scheduleArticle)
		protected.POST("/articles/:id/archive", handler.archiveArticle)

		// Categories
		protected.GET("/categories", handler.listCategories)
		protected.POST("/categories", handler.createCategory)
		protected.GET("/categories/:id", handler.getCategory)
		protected.PATCH("/categories/:id", handler.updateCategory)
		protected.DELETE("/categories/:id", handler.deleteCategory)

		// Media
		protected.GET("/media", handler.listMedia)
		protected.POST("/media", handler.createMedia)
		protected.GET("/media/:id", handler.getMedia)
		protected.PATCH("/media/:id", handler.updateMedia)
		protected.DELETE("/media/:id", handler.deleteMedia)

		// Webhooks
		protected.GET("/webhooks", handler.listWebhooks)
		protected.POST("/webhooks", handler.createWebhook)
		protected.GET("/webhooks/:id", handler.getWebhook)
		protected.PATCH("/webhooks/:id", handler.updateWebhook)
		protected.DELETE("/webhooks/:id", handler.deleteWebhook)
		protected.POST("/webhooks/:id/test", handler.testWebhook)

		// SEO
		protected.GET("/seo", handler.listSeoConfigs)
		protected.GET("/seo/config", handler.getSeoConfig)
		protected.POST("/seo/config", handler.upsertSeoConfig)

		// Newsletter
		protected.GET("/newsletter/subscribers", handler.listNewsletterSubscribers)
		protected.POST("/newsletter/subscribe", handler.subscribeNewsletter)
		protected.DELETE("/newsletter/unsubscribe", handler.unsubscribeNewsletter)

		// Schedule
		protected.GET("/schedule", handler.listSchedules)
		protected.POST("/schedule", handler.createSchedule)
		protected.GET("/schedule/:id", handler.getSchedule)
		protected.POST("/schedule/:id/cancel", handler.cancelSchedule)
		protected.DELETE("/schedule/:id", handler.deleteSchedule)

		// Admin Docker
		protected.GET("/admin/docker/containers", handler.listDockerContainers)
		protected.GET("/admin/docker/logs", handler.getDockerLogs)
		protected.GET("/admin/docker/logs/stream", handler.streamDockerLogs)
		protected.POST("/admin/docker/exec", handler.execDockerCommand)
		protected.GET("/admin/docker/updates", handler.checkDockerUpdates)
		protected.POST("/admin/docker/update", handler.updateDockerContainer)

		// Admin Settings
		protected.GET("/admin/settings", handler.getAdminSettings)
		protected.PATCH("/admin/settings", handler.updateAdminSettings)
	}
}

type apiHandler struct {
	deps Dependencies
}

func (h *apiHandler) principal(c *gin.Context) (interfaces.Principal, bool) {
	return middleware.PrincipalFromGin(c)
}

func (h *apiHandler) live(c *gin.Context) {
	utils.Success(c, http.StatusOK, gin.H{
		"status":  "alive",
		"role":    h.runtimeRole(),
		"version": h.deps.Config.App.Version,
	})
}

func (h *apiHandler) runtimeRole() string {
	if h.deps.RuntimeRole == "" {
		return "api"
	}
	return h.deps.RuntimeRole
}

func (h *apiHandler) health(c *gin.Context) {
	status := "healthy"
	redisStatus := "disabled"
	if err := h.deps.Database.Ping(c.Request.Context()); err != nil {
		status = "degraded"
	}
	if h.deps.Redis != nil {
		redisHealth := h.deps.Redis.Health(c.Request.Context())
		redisStatus = string(redisHealth.Status)
		if redisHealth.Status != redisclient.StatusHealthy {
			status = "degraded"
		}
	}
	utils.Success(c, http.StatusOK, gin.H{
		"status":   status,
		"database": "healthy",
		"redis":    redisStatus,
		"role":     h.runtimeRole(),
		"version":  h.deps.Config.App.Version,
	})
}

func (h *apiHandler) ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	result := gin.H{"database": "healthy", "redis": "disabled"}
	if err := h.deps.Database.Ping(ctx); err != nil {
		result["database"] = "unhealthy"
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	if h.deps.Config.Redis.Enabled {
		if h.deps.Redis == nil || !h.deps.Redis.IsAvailable() {
			if h.deps.Config.Redis.Required {
				result["redis"] = "unhealthy"
				utils.Error(c, utils.ErrDependencyUnavailable)
				return
			}
			result["redis"] = "unavailable"
		} else {
			result["redis"] = "healthy"
		}
	}
	utils.Success(c, http.StatusOK, gin.H{
		"status":   "ready",
		"database": result["database"],
		"redis":    result["redis"],
		"role":     h.runtimeRole(),
		"version":  h.deps.Config.App.Version,
	})
}

func (h *apiHandler) me(c *gin.Context) {
	principal, _ := h.principal(c)
	user, err := h.deps.UserService.GetMe(c.Request.Context(), principal)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"displayName":       user.DisplayName,
		"avatarUrl":         user.AvatarURL,
		"status":            user.Status,
		"presenceStatus":    "offline",
		"roles":             principal.Roles,
		"permissions":       principal.Permissions,
		"workspaceId":       principal.WorkspaceID,
		"createdAt":         user.CreatedAt,
		"updatedAt":         user.UpdatedAt,
		"lastSeenAt":        nil,
		"disabledAt":        user.DisabledAt,
		"emailVerifiedAt":   user.EmailVerifiedAt,
		"passwordChangedAt": user.PasswordChangedAt,
	})
}

func (h *apiHandler) updateMe(c *gin.Context) {
	var req struct {
		DisplayName string `json:"displayName"`
		AvatarURL   string `json:"avatarUrl"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	principal, _ := h.principal(c)
	user, err := h.deps.UserService.UpdateMe(c.Request.Context(), principal, req.DisplayName, req.AvatarURL, req.Status)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"displayName":       user.DisplayName,
		"avatarUrl":         user.AvatarURL,
		"status":            user.Status,
		"presenceStatus":    "offline",
		"roles":             principal.Roles,
		"permissions":       principal.Permissions,
		"workspaceId":       principal.WorkspaceID,
		"createdAt":         user.CreatedAt,
		"updatedAt":         user.UpdatedAt,
		"lastSeenAt":        nil,
		"disabledAt":        user.DisabledAt,
		"emailVerifiedAt":   user.EmailVerifiedAt,
		"passwordChangedAt": user.PasswordChangedAt,
	})
}

func (h *apiHandler) listWorkspaces(c *gin.Context) {
	principal, _ := h.principal(c)
	items, err := h.deps.WorkspaceService.List(c.Request.Context(), principal)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) createWorkspace(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	principal, _ := h.principal(c)
	item, err := h.deps.WorkspaceService.Create(c.Request.Context(), principal, req.Name, req.Slug, req.Description)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, item)
}

func (h *apiHandler) getWorkspace(c *gin.Context)    { h.workspaceResource(c, "get") }
func (h *apiHandler) updateWorkspace(c *gin.Context) { h.workspaceResource(c, "update") }
func (h *apiHandler) deleteWorkspace(c *gin.Context) { h.workspaceResource(c, "delete") }

func (h *apiHandler) workspaceResource(c *gin.Context, action string) {
	principal, _ := h.principal(c)
	id := c.Param("workspaceId")
	switch action {
	case "get":
		item, err := h.deps.WorkspaceService.Get(c.Request.Context(), principal, id)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusOK, item)
	case "update":
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if c.ShouldBindJSON(&req) != nil {
			utils.Error(c, utils.ErrValidationFailed)
			return
		}
		item, err := h.deps.WorkspaceService.Update(c.Request.Context(), principal, id, req.Name, req.Description)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusOK, item)
	case "delete":
		if err := h.deps.WorkspaceService.Archive(c.Request.Context(), principal, id); err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusOK, gin.H{"deleted": true})
	}
}

func (h *apiHandler) listWorkspaceMembers(c *gin.Context)   { h.membersResource(c, "list") }
func (h *apiHandler) createWorkspaceMember(c *gin.Context)  { h.membersResource(c, "create") }
func (h *apiHandler) provisionWorkspaceUser(c *gin.Context) { h.membersResource(c, "provision") }
func (h *apiHandler) updateWorkspaceMember(c *gin.Context)  { h.membersResource(c, "update") }
func (h *apiHandler) deleteWorkspaceMember(c *gin.Context)  { h.membersResource(c, "delete") }

func (h *apiHandler) membersResource(c *gin.Context, action string) {
	principal, _ := h.principal(c)
	workspaceID := c.Param("workspaceId")
	switch action {
	case "list":
		items, err := h.deps.WorkspaceService.ListMembers(c.Request.Context(), principal, workspaceID)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.List(c, items, "", false)
	case "create":
		var req struct {
			UserID string `json:"userId"`
			Email  string `json:"email"`
			Role   string `json:"role"`
		}
		if c.ShouldBindJSON(&req) != nil {
			utils.Error(c, utils.ErrValidationFailed)
			return
		}
		item, err := h.deps.WorkspaceService.AddMember(c.Request.Context(), principal, workspaceID, req.UserID, req.Email, req.Role)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusCreated, item)
	case "provision":
		var req struct {
			Email             string `json:"email"`
			DisplayName       string `json:"displayName"`
			Role              string `json:"role"`
			TemporaryPassword string `json:"temporaryPassword"`
		}
		if c.ShouldBindJSON(&req) != nil {
			utils.Error(c, utils.ErrValidationFailed)
			return
		}
		item, err := h.deps.WorkspaceService.ProvisionWorkspaceUser(
			c.Request.Context(),
			principal,
			workspaceID,
			services.ProvisionWorkspaceUserInput{
				Email:             req.Email,
				DisplayName:       req.DisplayName,
				Role:              req.Role,
				TemporaryPassword: req.TemporaryPassword,
			},
		)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusCreated, item)
	case "update":
		var req struct{ Role string }
		if c.ShouldBindJSON(&req) != nil {
			utils.Error(c, utils.ErrValidationFailed)
			return
		}
		item, err := h.deps.WorkspaceService.UpdateMember(c.Request.Context(), principal, workspaceID, c.Param("userId"), req.Role)
		if err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusOK, item)
	case "delete":
		if err := h.deps.WorkspaceService.RemoveMember(c.Request.Context(), principal, workspaceID, c.Param("userId")); err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, http.StatusOK, gin.H{"deleted": true})
	}
}

func (h *apiHandler) metrics(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func (h *apiHandler) webhook(c *gin.Context) {
	payload, _ := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	_ = payload
	utils.Success(c, http.StatusAccepted, gin.H{"accepted": true})
}
