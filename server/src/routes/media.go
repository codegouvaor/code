package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

func (h *apiHandler) listMedia(c *gin.Context) {
	mimeType := c.Query("mimeType")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	offset := (page - 1) * pageSize

	items, total, err := h.deps.Repos.Media().List(c.Request.Context(), "", mimeType, offset, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", int64(page*pageSize) < total)
	_ = total
}

func (h *apiHandler) createMedia(c *gin.Context) {
	principal, _ := h.principal(c)
	var req struct {
		Name     string `json:"name"`
		FileName string `json:"fileName"`
		URL      string `json:"url"`
		MimeType string `json:"mimeType"`
		Size     int64  `json:"size"`
		Alt      string `json:"alt"`
		Caption  string `json:"caption"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Name and URL are required.", nil))
		return
	}

	now := time.Now().UTC()
	media := &models.Media{
		Name:       req.Name,
		FileName:   req.FileName,
		URL:        req.URL,
		MimeType:   req.MimeType,
		Size:       req.Size,
		Alt:        req.Alt,
		Caption:    req.Caption,
		UploadedBy: principal.UserID,
	}
	media.ID = utils.NewID()
	media.CreatedAt = now
	media.UpdatedAt = now

	if err := h.deps.Repos.Media().Create(c.Request.Context(), media); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, media)
}

func (h *apiHandler) getMedia(c *gin.Context) {
	id := c.Param("id")
	media, err := h.deps.Repos.Media().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, media)
}

func (h *apiHandler) updateMedia(c *gin.Context) {
	id := c.Param("id")
	media, err := h.deps.Repos.Media().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	var req struct {
		Name    *string `json:"name"`
		Alt     *string `json:"alt"`
		Caption *string `json:"caption"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if req.Name != nil {
		media.Name = *req.Name
	}
	if req.Alt != nil {
		media.Alt = *req.Alt
	}
	if req.Caption != nil {
		media.Caption = *req.Caption
	}
	media.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.Media().Update(c.Request.Context(), media); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, media)
}

func (h *apiHandler) deleteMedia(c *gin.Context) {
	id := c.Param("id")
	if err := h.deps.Repos.Media().Delete(c.Request.Context(), id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}
