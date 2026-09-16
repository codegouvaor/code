package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

func (h *apiHandler) listOrganizations(c *gin.Context) {
	principal := h.ownerOptional(c)
	items, err := h.deps.OrganizationService.ListByViewer(c.Request.Context(), principal)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) createOrganization(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		WebsiteURL  string `json:"websiteUrl"`
		Location    string `json:"location"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	organization, err := h.deps.OrganizationService.Create(c.Request.Context(), principal, services.CreateOrganizationInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Visibility:  req.Visibility,
		WebsiteURL:  req.WebsiteURL,
		Location:    req.Location,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, organization)
}

func (h *apiHandler) getOrganization(c *gin.Context) {
	principal := h.ownerOptional(c)
	organization, err := h.deps.OrganizationService.Get(c.Request.Context(), principal, c.Param("organization"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, organization)
}

func (h *apiHandler) updateOrganization(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		WebsiteURL  string `json:"websiteUrl"`
		Location    string `json:"location"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	organization, err := h.deps.OrganizationService.Update(c.Request.Context(), principal, c.Param("organization"), services.CreateOrganizationInput{
		Name:        req.Name,
		Description: req.Description,
		Visibility:  req.Visibility,
		WebsiteURL:  req.WebsiteURL,
		Location:    req.Location,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, organization)
}

func (h *apiHandler) deleteOrganization(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.OrganizationService.Archive(c.Request.Context(), principal, c.Param("organization")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) listOrganizationMembers(c *gin.Context) {
	principal := h.ownerOptional(c)
	items, err := h.deps.OrganizationService.ListMembers(c.Request.Context(), principal, c.Param("organization"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) addOrganizationMember(c *gin.Context) {
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
	member, err := h.deps.OrganizationService.AddMember(c.Request.Context(), principal, c.Param("organization"), req.User, req.Role)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, member)
}

func (h *apiHandler) updateOrganizationMember(c *gin.Context) {
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
	member, err := h.deps.OrganizationService.UpdateMember(c.Request.Context(), principal, c.Param("organization"), c.Param("userId"), req.Role)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, member)
}

func (h *apiHandler) removeOrganizationMember(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.OrganizationService.RemoveMember(c.Request.Context(), principal, c.Param("organization"), c.Param("userId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) listOrganizationTeams(c *gin.Context) {
	principal := h.ownerOptional(c)
	items, err := h.deps.OrganizationService.ListTeams(c.Request.Context(), principal, c.Param("organization"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) createOrganizationTeam(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	team, err := h.deps.OrganizationService.CreateTeam(c.Request.Context(), principal, c.Param("organization"), req.Name, req.Description)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, team)
}

func (h *apiHandler) deleteOrganizationTeam(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if err := h.deps.OrganizationService.DeleteTeam(c.Request.Context(), principal, c.Param("organization"), c.Param("team")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

// checkOwnerName tells the "new organization" form whether a namespace is
// still free (usernames and organization slugs share one namespace).
func (h *apiHandler) checkOwnerName(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		name = strings.TrimSpace(c.Param("name"))
	}
	if err := services.ValidateName(name); err != nil {
		utils.Error(c, err)
		return
	}
	available, err := h.deps.OwnerService.IsNameAvailable(c.Request.Context(), name)
	if err != nil {
		utils.Error(c, err)
		return
	}
	status := http.StatusOK
	utils.Success(c, status, gin.H{"name": strings.ToLower(name), "available": available})
}

// parsePositiveInt is shared by the query helpers of the platform routes.
func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, err
	}
	return parsed, nil
}
