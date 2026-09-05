package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
)

func (h *apiHandler) listSeoConfigs(c *gin.Context) {
	items, err := h.deps.Repos.SeoConfigs().List(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) getSeoConfig(c *gin.Context) {
	pagePath := c.Query("pagePath")
	if strings.TrimSpace(pagePath) == "" {
		pagePath = c.Param("pagePath")
	}

	config, err := h.deps.Repos.SeoConfigs().GetByPagePath(c.Request.Context(), pagePath)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, config)
}

func (h *apiHandler) upsertSeoConfig(c *gin.Context) {
	var req struct {
		PagePath    string `json:"pagePath"`
		Title       string `json:"title"`
		Description string `json:"description"`
		OgImage     string `json:"ogImage"`
		Canonical   string `json:"canonical"`
		NoIndex     bool   `json:"noIndex"`
		Keywords    string `json:"keywords"`
		Locale      string `json:"locale"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.PagePath) == "" || strings.TrimSpace(req.Title) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "pagePath and title are required.", nil))
		return
	}

	now := time.Now().UTC()
	config := &models.SeoConfig{
		PagePath:    req.PagePath,
		Title:       req.Title,
		Description: req.Description,
		OgImage:     req.OgImage,
		Canonical:   req.Canonical,
		NoIndex:     req.NoIndex,
		Keywords:    req.Keywords,
		Locale:      req.Locale,
	}
	config.ID = utils.NewID()
	config.CreatedAt = now
	config.UpdatedAt = now

	if err := h.deps.Repos.SeoConfigs().Upsert(c.Request.Context(), config); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, config)
}
