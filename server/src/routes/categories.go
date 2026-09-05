package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

func (h *apiHandler) listCategories(c *gin.Context) {
	categories, err := h.deps.Repos.Categories().List(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, categories, "", false)
}

func (h *apiHandler) createCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		ParentID    *string `json:"parentId"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Name is required.", nil))
		return
	}

	slug := slugify(req.Name)
	now := time.Now().UTC()
	category := &models.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Color:       req.Color,
		ParentID:    req.ParentID,
	}
	category.ID = utils.NewID()
	category.CreatedAt = now
	category.UpdatedAt = now

	if err := h.deps.Repos.Categories().Create(c.Request.Context(), category); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, category)
}

func (h *apiHandler) getCategory(c *gin.Context) {
	id := c.Param("id")
	category, err := h.deps.Repos.Categories().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, category)
}

func (h *apiHandler) updateCategory(c *gin.Context) {
	id := c.Param("id")
	category, err := h.deps.Repos.Categories().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		SortOrder   *int    `json:"sortOrder"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	category.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.Categories().Update(c.Request.Context(), category); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, category)
}

func (h *apiHandler) deleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.deps.Repos.Categories().Delete(c.Request.Context(), id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}
