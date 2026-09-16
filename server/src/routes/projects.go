package routes

import (
	"net/http"
	"strings"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

func (h *apiHandler) listProjects(c *gin.Context) {
	principal := h.ownerOptional(c)
	page := parsePage(c)
	pageSize := parsePageSize(c)
	filter := interfaces.ProjectFilter{
		Query:          strings.TrimSpace(c.Query("q")),
		Topic:          strings.TrimSpace(c.Query("topic")),
		Offset:         (page - 1) * pageSize,
		Limit:          pageSize,
		OrganizationID: strings.TrimSpace(c.Query("organizationId")),
		OwnerID:        strings.TrimSpace(c.Query("ownerId")),
	}
	if visibility := strings.TrimSpace(c.Query("visibility")); visibility != "" {
		filter.Visibility = visibility
	}
	views, total, err := h.deps.ProjectService.List(c.Request.Context(), principal, filter)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, views, "", int64(page*pageSize) < total)
}

func (h *apiHandler) createProject(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Owner        string   `json:"owner"`
		Organization string   `json:"organization"`
		Name         string   `json:"name"`
		Slug         string   `json:"slug"`
		Description  string   `json:"description"`
		Visibility   string   `json:"visibility"`
		Topics       []string `json:"topics"`
		HomepageURL  string   `json:"homepageUrl"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	project, err := h.deps.ProjectService.Create(c.Request.Context(), principal, services.CreateProjectInput{
		Owner:        req.Owner,
		Organization: req.Organization,
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		Visibility:   req.Visibility,
		Topics:       req.Topics,
		HomepageURL:  req.HomepageURL,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	view, err := h.deps.ProjectService.View(c.Request.Context(), principal, project.ID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, view)
}

func (h *apiHandler) getProject(c *gin.Context) {
	principal := h.ownerOptional(c)
	view, err := h.deps.ProjectService.View(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, view)
}

func (h *apiHandler) updateProject(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Visibility  string   `json:"visibility"`
		Topics      []string `json:"topics"`
		HomepageURL string   `json:"homepageUrl"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	view, err := h.deps.ProjectService.Update(c.Request.Context(), principal, c.Param("projectId"), services.CreateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		Visibility:  req.Visibility,
		Topics:      req.Topics,
		HomepageURL: req.HomepageURL,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, view)
}

func (h *apiHandler) deleteProject(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.ProjectService.Archive(c.Request.Context(), principal, c.Param("projectId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) listProjectMembers(c *gin.Context) {
	principal := h.ownerOptional(c)
	items, err := h.deps.ProjectService.ListMembers(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) addProjectMember(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		User string `json:"user"`
		Role string `json:"role"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	member, err := h.deps.ProjectService.AddMember(c.Request.Context(), principal, c.Param("projectId"), req.User, req.Role)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, member)
}

func (h *apiHandler) updateProjectMember(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	member, err := h.deps.ProjectService.UpdateMember(c.Request.Context(), principal, c.Param("projectId"), c.Param("userId"), req.Role)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, member)
}

func (h *apiHandler) removeProjectMember(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.ProjectService.RemoveMember(c.Request.Context(), principal, c.Param("projectId"), c.Param("userId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

// ── Code-native assets (documentation, APIs, SDKs, services, standards) ──────

func (h *apiHandler) listProjectAssets(c *gin.Context) {
	principal := h.ownerOptional(c)
	kind := strings.TrimSpace(c.Query("kind"))
	assets, err := h.deps.ProjectService.ListAssets(c.Request.Context(), principal, c.Param("projectId"), kind)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, assets, "", false)
}

func (h *apiHandler) createProjectAsset(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Kind        string         `json:"kind"`
		Slug        string         `json:"slug"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		URL         string         `json:"url"`
		Visibility  string         `json:"visibility"`
		Position    int            `json:"position"`
		Metadata    map[string]any `json:"metadata"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	asset, err := h.deps.ProjectService.CreateAsset(c.Request.Context(), principal, c.Param("projectId"), services.CreateAssetInput{
		Kind:        req.Kind,
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Visibility:  req.Visibility,
		Position:    req.Position,
		Metadata:    req.Metadata,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, asset)
}

func (h *apiHandler) updateProjectAsset(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		URL         string         `json:"url"`
		Visibility  string         `json:"visibility"`
		Metadata    map[string]any `json:"metadata"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	asset, err := h.deps.ProjectService.UpdateAsset(c.Request.Context(), principal, c.Param("projectId"), c.Param("assetId"), services.CreateAssetInput{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Visibility:  req.Visibility,
		Metadata:    req.Metadata,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, asset)
}

func (h *apiHandler) deleteProjectAsset(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.ProjectService.DeleteAsset(c.Request.Context(), principal, c.Param("projectId"), c.Param("assetId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

// getProjectCapabilities exposes the capability matrix of a project so that the
// UI never has to know which forge is used.
func (h *apiHandler) getProjectCapabilities(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.ResolveProject(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"projectId":    resolved.Project.ID,
		"reference":    resolved.Project.Reference,
		"provider":     resolved.Provider,
		"descriptor":   h.deps.RepositoryService.Descriptor(resolved),
		"capabilities": h.deps.RepositoryService.Capabilities(resolved),
	})
}

// ── Stars and watches ────────────────────────────────────────────────────────

func (h *apiHandler) starProject(c *gin.Context)   { h.projectSubscription(c, "star", true) }
func (h *apiHandler) unstarProject(c *gin.Context) { h.projectSubscription(c, "star", false) }
func (h *apiHandler) watchProject(c *gin.Context)  { h.projectSubscription(c, "watch", true) }
func (h *apiHandler) unwatchProject(c *gin.Context) {
	h.projectSubscription(c, "watch", false)
}

func (h *apiHandler) projectSubscription(c *gin.Context, kind string, enabled bool) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var (
		view *services.ProjectView
		err  error
	)
	switch {
	case kind == "star" && enabled:
		view, err = h.deps.ProjectService.Star(c.Request.Context(), principal, c.Param("projectId"))
	case kind == "star" && !enabled:
		view, err = h.deps.ProjectService.Unstar(c.Request.Context(), principal, c.Param("projectId"))
	case kind == "watch" && enabled:
		view, err = h.deps.ProjectService.Watch(c.Request.Context(), principal, c.Param("projectId"), c.Query("level"))
	default:
		view, err = h.deps.ProjectService.Unwatch(c.Request.Context(), principal, c.Param("projectId"))
	}
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, view)
}
