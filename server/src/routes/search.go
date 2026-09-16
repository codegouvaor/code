package routes

import (
	"net/http"
	"strings"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

// search runs the global platform search across the requested kinds. The
// engine is an implementation detail: today PostgreSQL, later a dedicated
// index, without any change for the client.
func (h *apiHandler) search(c *gin.Context) {
	principal := h.ownerOptional(c)
	page := parsePage(c)
	pageSize := parsePageSize(c)

	kinds := []string{}
	for _, raw := range strings.Split(c.Query("type"), ",") {
		if value := strings.TrimSpace(raw); value != "" {
			kinds = append(kinds, value)
		}
	}
	normalized, err := services.NormalizeSearchKinds(kinds)
	if err != nil {
		utils.Error(c, err)
		return
	}

	response, err := h.deps.SearchService.Search(c.Request.Context(), principal.UserID, interfaces.SearchFilter{
		Query:  strings.TrimSpace(c.Query("q")),
		Kinds:  normalized,
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, response.Hits, gin.H{
		"total":   response.Total,
		"hasMore": response.HasMore,
		"query":   response.Query,
		"kinds":   response.Kinds,
		"engines": h.deps.SearchService.Engines(),
	})
}
