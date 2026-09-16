package middleware

import (
	"log/slog"
	"time"

	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

// Context keys used by handlers to enrich the access log without ever logging
// credentials.
const (
	operationKey contextKey = "log_operation"
	providerKey  contextKey = "log_provider"
	projectKey   contextKey = "log_project_id"
)

// SetOperation records the business operation of the current request.
func SetOperation(c *gin.Context, operation string) {
	if operation != "" {
		c.Set(string(operationKey), operation)
	}
}

// SetProvider records the external provider involved in the request.
func SetProvider(c *gin.Context, provider string) {
	if provider != "" {
		c.Set(string(providerKey), provider)
	}
}

// SetProject records the project involved in the request.
func SetProject(c *gin.Context, projectID string) {
	if projectID != "" {
		c.Set(string(projectKey), projectID)
	}
}

func contextString(c *gin.Context, key contextKey) string {
	value, ok := c.Get(string(key))
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}

// Logging emits one structured access log line per request. It carries the
// correlation identifiers the platform needs (request, user, workspace,
// project, provider, operation) and never logs tokens, secrets or payloads.
func Logging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		attributes := []any{
			"request_id", utils.RequestIDFromGin(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if principal, ok := GetPrincipal(c); ok && principal.UserID != "" {
			attributes = append(attributes, "user_id", principal.UserID)
			if principal.WorkspaceID != "" {
				attributes = append(attributes, "workspace_id", principal.WorkspaceID)
			}
		}
		if projectID := contextString(c, projectKey); projectID != "" {
			attributes = append(attributes, "project_id", projectID)
		}
		if provider := contextString(c, providerKey); provider != "" {
			attributes = append(attributes, "provider", provider)
		}
		if operation := contextString(c, operationKey); operation != "" {
			attributes = append(attributes, "operation", operation)
		}
		if len(c.Errors) > 0 {
			attributes = append(attributes, "error", c.Errors.String())
		}

		switch {
		case c.Writer.Status() >= 500:
			logger.Error("http_request", attributes...)
		case c.Writer.Status() >= 400:
			logger.Warn("http_request", attributes...)
		default:
			logger.Info("http_request", attributes...)
		}
	}
}
