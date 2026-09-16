package routes

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func parsePage(c *gin.Context) int {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	return page
}

func parsePageSize(c *gin.Context) int {
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageSize
}

// bearerToken extracts the token of an Authorization header.
func bearerToken(header string) string {
	if len(header) <= len("Bearer ") || !strings.EqualFold(header[:len("Bearer ")], "Bearer ") {
		return ""
	}
	return strings.TrimSpace(header[len("Bearer "):])
}

// pageSlice applies in-memory pagination to a slice that was fully computed by
// a service (provider-backed listings are already paginated upstream).
func pageSlice[T any](items []T, page, pageSize int) ([]T, bool) {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []T{}, false
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], end < len(items)
}

// queryBool reads an optional boolean query parameter.
func queryBool(c *gin.Context, name string) bool {
	value := strings.ToLower(strings.TrimSpace(c.Query(name)))
	return value == "1" || value == "true" || value == "yes"
}

// marshalJSON encodes a value as JSONB, falling back to an empty array.
func marshalJSON(v any) datatypes.JSON {
	data, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(data)
}
