package routes

import (
	"net/http"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

// ownerOptional resolves the caller when a token is present without failing
// anonymous requests: profiles must stay reachable without authentication.
func (h *apiHandler) ownerOptional(c *gin.Context) interfaces.Principal {
	header := c.GetHeader("Authorization")
	if header == "" || h.deps.IdentityProvider == nil {
		return interfaces.Principal{}
	}
	principal, err := h.deps.IdentityProvider.Authenticate(c.Request.Context(), bearerToken(header))
	if err != nil {
		return interfaces.Principal{}
	}
	return *principal
}

// getOwner resolves /[owner] into a user or an organization. It is the entry
// point of every profile page: the payload tells the shell which navigation
// entries (teams, people, insights, sponsoring) exist for the namespace.
func (h *apiHandler) getOwner(c *gin.Context) {
	principal := h.ownerOptional(c)
	owner, err := h.deps.OwnerService.Resolve(c.Request.Context(), principal.UserID, c.Param("owner"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, owner)
}

// getOwnerRepositories lists the repositories of an owner namespace, already
// enriched with the provider metadata the UI renders.
func (h *apiHandler) getOwnerRepositories(c *gin.Context) {
	principal := h.ownerOptional(c)
	owner := c.Param("owner")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	views, err := h.deps.ProjectService.ListByOwner(c.Request.Context(), principal, owner)
	if err != nil {
		utils.Error(c, err)
		return
	}
	items, hasMore := pageSlice(views, page, pageSize)
	utils.List(c, items, "", hasMore)
}

// getOwnerProjects lists the Code-native projects of an owner.
func (h *apiHandler) getOwnerProjects(c *gin.Context) {
	principal := h.ownerOptional(c)
	owner := c.Param("owner")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	views, err := h.deps.ProjectService.ListByOwner(c.Request.Context(), principal, owner)
	if err != nil {
		utils.Error(c, err)
		return
	}
	items, hasMore := pageSlice(views, page, pageSize)
	utils.List(c, items, "", hasMore)
}

// claimUsername assigns the Code handle of the authenticated account.
func (h *apiHandler) claimUsername(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	user, err := h.deps.OwnerService.ClaimUsername(c.Request.Context(), principal.UserID, req.Username)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"updatedAt": user.UpdatedAt,
	})
}

// getViewerProfile returns the profile of the authenticated account,
// including the handle used by the platform routes.
func (h *apiHandler) getViewerProfile(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	owner, err := h.deps.OwnerService.Resolve(c.Request.Context(), principal.UserID, principal.UserID)
	if err != nil {
		// The account may not have a username yet: report the identity anyway.
		user, userErr := h.deps.Repos.Users().GetByID(c.Request.Context(), principal.UserID)
		if userErr != nil {
			utils.Error(c, userErr)
			return
		}
		utils.Success(c, http.StatusOK, gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"avatarUrl":   user.AvatarURL,
			"username":    user.Username,
			"roles":       principal.Roles,
		})
		return
	}
	utils.Success(c, http.StatusOK, owner)
}

// listStarredProjects lists the projects starred by the caller.
func (h *apiHandler) listStarredProjects(c *gin.Context) {
	principal, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	page := parsePage(c)
	pageSize := parsePageSize(c)
	views, total, err := h.deps.ProjectService.StarredProjects(c.Request.Context(), principal, (page-1)*pageSize, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, views, "", int64(page*pageSize) < total)
}
