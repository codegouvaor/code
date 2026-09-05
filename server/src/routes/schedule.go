package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

func (h *apiHandler) listSchedules(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	offset := (page - 1) * pageSize

	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}

	items, total, err := h.deps.Repos.Schedules().List(c.Request.Context(), from, to, offset, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", int64(page*pageSize) < total)
	_ = total
}

func (h *apiHandler) createSchedule(c *gin.Context) {
	principal, _ := h.principal(c)
	var req struct {
		EntityType  string `json:"entityType"`
		EntityID    string `json:"entityId"`
		Title       string `json:"title"`
		ScheduledAt string `json:"scheduledAt"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	if strings.TrimSpace(req.EntityType) == "" || strings.TrimSpace(req.Title) == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "entityType and title are required.", nil))
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid scheduledAt format.", nil))
		return
	}

	now := time.Now().UTC()
	schedule := &models.Schedule{
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Title:       req.Title,
		ScheduledAt: scheduledAt,
		Status:      "pending",
		CreatedBy:   principal.UserID,
	}
	schedule.ID = utils.NewID()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	if err := h.deps.Repos.Schedules().Create(c.Request.Context(), schedule); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, schedule)
}

func (h *apiHandler) getSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.deps.Repos.Schedules().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, schedule)
}

func (h *apiHandler) cancelSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.deps.Repos.Schedules().GetByID(c.Request.Context(), id)
	if err != nil {
		utils.Error(c, err)
		return
	}

	now := time.Now().UTC()
	schedule.Status = "cancelled"
	schedule.CancelledAt = &now
	schedule.UpdatedAt = now

	if err := h.deps.Repos.Schedules().Update(c.Request.Context(), schedule); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, schedule)
}

func (h *apiHandler) deleteSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := h.deps.Repos.Schedules().Delete(c.Request.Context(), id); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}
