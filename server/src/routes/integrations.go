package routes

import (
	"io"
	"net/http"
	"strings"

	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

// ── Provider catalogue ───────────────────────────────────────────────────────

// listProviders exposes the registered forges and their capabilities. It is a
// public endpoint: the UI needs it before any authentication to render the
// "connect a provider" screens.
func (h *apiHandler) listProviders(c *gin.Context) {
	descriptors := h.deps.ConnectionService.Registry().Descriptors()
	utils.Success(c, http.StatusOK, descriptors)
}

// ── Connections ──────────────────────────────────────────────────────────────

func (h *apiHandler) listConnections(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	items, err := h.deps.ConnectionService.List(c.Request.Context(), principal.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

// authorizeConnection starts the OAuth flow used to grant repository access.
// The Code identity is untouched: the callback stores a ProviderConnection.
func (h *apiHandler) authorizeConnection(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if !h.deps.ConnectionService.Registry().Has(provider) {
		utils.Error(c, utils.ErrProviderNotSupported)
		return
	}
	url, err := h.deps.OAuthService.GetAuthorizationURL(c.Request.Context(), provider, "connect", principal.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"provider":     provider,
		"authorizeUrl": url,
		"state":        "oauth",
	})
}

func (h *apiHandler) deleteConnection(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.ConnectionService.Disconnect(c.Request.Context(), principal.UserID, c.Param("provider")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"disconnected": true})
}

// ── Bindings and synchronisation ─────────────────────────────────────────────

func (h *apiHandler) listProjectRepositories(c *gin.Context) {
	principal := h.ownerOptional(c)
	bindings, err := h.deps.SyncService.ListBindings(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, bindings, "", false)
}

func (h *apiHandler) connectProjectRepository(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Provider     string `json:"provider"`
		Owner        string `json:"owner"`
		Repo         string `json:"repo"`
		ConnectionID string `json:"connectionId"`
		IsPrimary    bool   `json:"isPrimary"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	binding, err := h.deps.SyncService.ConnectBinding(c.Request.Context(), principal, c.Param("projectId"), services.ConnectBindingInput{
		Provider:     req.Provider,
		Owner:        req.Owner,
		Repo:         req.Repo,
		ConnectionID: req.ConnectionID,
		IsPrimary:    req.IsPrimary,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, binding)
}

func (h *apiHandler) disconnectProjectRepository(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.SyncService.DisconnectBinding(c.Request.Context(), principal, c.Param("projectId"), c.Param("bindingId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"detached": true})
}

// syncProjectRepository refreshes a binding immediately instead of waiting for
// the background reconciliation.
func (h *apiHandler) syncProjectRepository(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	project, err := h.deps.ProjectService.Get(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	binding, err := h.deps.Repos.RepositoryBindings().GetByID(c.Request.Context(), c.Param("bindingId"))
	if err != nil || binding.ProjectID != project.ID {
		utils.Error(c, utils.ErrRepositoryBindingNotFound)
		return
	}
	synced, err := h.deps.SyncService.SyncBinding(c.Request.Context(), binding.ID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, synced)
}

// listJobs exposes the background queue: observability for operators and for
// the settings screens.
func (h *apiHandler) listJobs(c *gin.Context) {
	if _, ok := h.principal(c); !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	status := strings.TrimSpace(c.Query("status"))
	limit := 50
	if value := strings.TrimSpace(c.Query("limit")); value != "" {
		if parsed, err := parsePositiveInt(value); err == nil {
			limit = parsed
		}
	}
	items, err := h.deps.JobService.List(c.Request.Context(), status, limit)
	if err != nil {
		utils.Error(c, err)
		return
	}
	stats, err := h.deps.JobService.Stats(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, items, gin.H{"stats": stats, "worker": "in-process"})
}

// ── Webhooks ─────────────────────────────────────────────────────────────────

// providerWebhook receives forge deliveries. The route is public by design (the
// provider cannot authenticate as a Code user) and its authenticity is
// established by the signature/token check performed by the sync service.
func (h *apiHandler) providerWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 5<<20))
	if err != nil {
		utils.Error(c, utils.ErrWebhookPayloadInvalid)
		return
	}
	if len(body) == 0 {
		utils.Error(c, utils.ErrWebhookPayloadInvalid)
		return
	}
	outcome, err := h.deps.SyncService.HandleWebhook(
		c.Request.Context(),
		c.Param("provider"),
		c.Param("bindingId"),
		c.Request.Header,
		body,
	)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusAccepted, outcome)
}
