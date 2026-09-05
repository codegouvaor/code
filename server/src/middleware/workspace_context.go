package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const workspaceIDKey contextKey = "workspace_id"

func WorkspaceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := strings.TrimSpace(c.GetHeader("X-Workspace-ID"))
		if workspaceID == "" {
			workspaceID = strings.TrimSpace(c.Query("workspaceId"))
		}
		if workspaceID != "" {
			c.Set(string(workspaceIDKey), workspaceID)
		}
		c.Next()
	}
}

func WorkspaceIDFromGin(c *gin.Context) string {
	value, _ := c.Get(string(workspaceIDKey))
	if id, ok := value.(string); ok {
		return id
	}
	return ""
}
