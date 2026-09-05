package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/utils"
)

func (h *apiHandler) listArticles(c *gin.Context) {
	status := c.Query("status")
	categoryID := c.Query("categoryId")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	offset := (page - 1) * pageSize

	articles, total, err := h.deps.Repos.Articles().List(c.Request.Context(), "", status, categoryID, offset, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, articles, "", int64(page*pageSize) < total)
	_ = total
}

func (h *apiHandler) createArticle(c *gin.Context) {
	principal, _ := h.principal(c)
	var req struct {
		Title          string   `json:"title"`
		Excerpt        string   `json:"excerpt"`
		Content        string   `json:"content"`
		Type           string   `json:"type"`
		CategoryID     *string  `json:"categoryId"`
		Team           string   `json:"team"`
		Tags           []string `json:"tags"`
		SeoTitle       string   `json:"seoTitle"`
		SeoDescription string   `json:"seoDescription"`
		SeoOgImage     string   `json:"seoOgImage"`
		Priority       string   `json:"priority"`
		Channel        string   `json:"channel"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Title is required.", nil))
		return
	}

	slug := slugify(req.Title)
	now := time.Now().UTC()
	article := &models.Article{
		Title:          req.Title,
		Slug:           slug,
		Excerpt:        req.Excerpt,
		Content:        req.Content,
		Type:           req.Type,
		Status:         "draft",
		CategoryID:     req.CategoryID,
		Team:           req.Team,
		AuthorID:       principal.UserID,
		Tags:           marshalJSON(req.Tags),
		SeoTitle:       req.SeoTitle,
		SeoDescription: req.SeoDescription,
		SeoOgImage:     req.SeoOgImage,
		Priority:       req.Priority,
		Channel:        req.Channel,
	}
	article.ID = utils.NewID()
	article.CreatedAt = now
	article.UpdatedAt = now

	if err := h.deps.Repos.Articles().Create(c.Request.Context(), article); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, article)
}

func (h *apiHandler) getArticle(c *gin.Context) {
	id := c.Param("id")
	article, err := h.deps.Repos.Articles().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}

func (h *apiHandler) getArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")
	article, err := h.deps.Repos.Articles().GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}

func (h *apiHandler) updateArticle(c *gin.Context) {
	id := c.Param("id")
	article, err := h.deps.Repos.Articles().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	var req struct {
		Title          *string  `json:"title"`
		Excerpt        *string  `json:"excerpt"`
		Content        *string  `json:"content"`
		Type           *string  `json:"type"`
		CategoryID     *string  `json:"categoryId"`
		Team           *string  `json:"team"`
		Tags           []string `json:"tags"`
		SeoTitle       *string  `json:"seoTitle"`
		SeoDescription *string  `json:"seoDescription"`
		SeoOgImage     *string  `json:"seoOgImage"`
		Priority       *string  `json:"priority"`
		Channel        *string  `json:"channel"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if req.Title != nil {
		article.Title = *req.Title
	}
	if req.Excerpt != nil {
		article.Excerpt = *req.Excerpt
	}
	if req.Content != nil {
		article.Content = *req.Content
	}
	if req.Type != nil {
		article.Type = *req.Type
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Team != nil {
		article.Team = *req.Team
	}
	if req.Tags != nil {
		article.Tags = marshalJSON(req.Tags)
	}
	if req.SeoTitle != nil {
		article.SeoTitle = *req.SeoTitle
	}
	if req.SeoDescription != nil {
		article.SeoDescription = *req.SeoDescription
	}
	if req.SeoOgImage != nil {
		article.SeoOgImage = *req.SeoOgImage
	}
	if req.Priority != nil {
		article.Priority = *req.Priority
	}
	if req.Channel != nil {
		article.Channel = *req.Channel
	}
	article.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.Articles().Update(c.Request.Context(), article); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}

func (h *apiHandler) deleteArticle(c *gin.Context) {
	id := c.Param("id")
	if err := h.deps.Repos.Articles().Delete(c.Request.Context(), id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) publishArticle(c *gin.Context) {
	id := c.Param("id")
	article, err := h.deps.Repos.Articles().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	now := time.Now().UTC()
	article.Status = "published"
	article.PublishedAt = &now
	article.UpdatedAt = now

	if err := h.deps.Repos.Articles().Update(c.Request.Context(), article); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}

func (h *apiHandler) scheduleArticle(c *gin.Context) {
	id := c.Param("id")
	article, err := h.deps.Repos.Articles().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	var req struct {
		ScheduledAt time.Time `json:"scheduledAt"`
	}
	if c.ShouldBindJSON(&req) != nil || req.ScheduledAt.IsZero() {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	article.Status = "scheduled"
	article.ScheduledAt = &req.ScheduledAt
	article.UpdatedAt = time.Now().UTC()

	if err := h.deps.Repos.Articles().Update(c.Request.Context(), article); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}

func (h *apiHandler) archiveArticle(c *gin.Context) {
	id := c.Param("id")
	article, err := h.deps.Repos.Articles().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	now := time.Now().UTC()
	article.Status = "archived"
	article.ArchivedAt = &now
	article.UpdatedAt = now

	if err := h.deps.Repos.Articles().Update(c.Request.Context(), article); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, article)
}
