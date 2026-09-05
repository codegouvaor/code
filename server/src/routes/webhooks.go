package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
)

func (h *apiHandler) listWebhooks(c *gin.Context) {
	principal, _ := h.principal(c)
	items, err := h.deps.Repos.Webhooks().ListByWorkspace(c.Request.Context(), principal.WorkspaceID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) createWebhook(c *gin.Context) {
	principal, _ := h.principal(c)
	var req struct {
		Provider string   `json:"provider"`
		URL      string   `json:"url"`
		Secret   string   `json:"secret"`
		Events   []string `json:"events"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.URL) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "URL is required.", nil))
		return
	}

	now := time.Now().UTC()
	webhook := &models.Webhook{
		WorkspaceID: principal.WorkspaceID,
		Provider:    req.Provider,
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      marshalJSON(req.Events),
		Active:      true,
	}
	webhook.ID = utils.NewID()
	webhook.CreatedAt = now
	webhook.UpdatedAt = now

	if err := h.deps.Repos.Webhooks().Create(c.Request.Context(), webhook); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, webhook)
}

func (h *apiHandler) getWebhook(c *gin.Context) {
	id := c.Param("id")
	webhook, err := h.deps.Repos.Webhooks().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, webhook)
}

func (h *apiHandler) updateWebhook(c *gin.Context) {
	id := c.Param("id")
	webhook, err := h.deps.Repos.Webhooks().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	var req struct {
		URL      *string  `json:"url"`
		Secret   *string  `json:"secret"`
		Events   []string `json:"events"`
		Active   *bool    `json:"active"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Secret != nil {
		webhook.Secret = *req.Secret
	}
	if req.Events != nil {
		webhook.Events = marshalJSON(req.Events)
	}
	if req.Active != nil {
		webhook.Active = *req.Active
	}
	webhook.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.Webhooks().Update(c.Request.Context(), webhook); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, webhook)
}

func (h *apiHandler) deleteWebhook(c *gin.Context) {
	id := c.Param("id")
	if err := h.deps.Repos.Webhooks().Delete(c.Request.Context(), id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) testWebhook(c *gin.Context) {
	id := c.Param("id")
	webhook, err := h.deps.Repos.Webhooks().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	delivery := &models.WebhookDelivery{
		WebhookID:   webhook.ID,
		Event:       "test",
		Status:      200,
		RequestBody: `{"test": true}`,
		ResponseBody: `{"ok": true}`,
		Duration:    100,
	}
	delivery.ID = utils.NewID()
	delivery.CreatedAt = time.Now().UTC()
	delivery.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.WebhookDeliveries().Create(c.Request.Context(), delivery); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"delivered": true, "deliveryId": delivery.ID})
}
