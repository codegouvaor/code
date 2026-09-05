package routes

import (
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

func (h *apiHandler) listNewsletterSubscribers(c *gin.Context) {
	status := c.Query("status")
	page := parsePage(c)
	pageSize := parsePageSize(c)
	offset := (page - 1) * pageSize

	items, total, err := h.deps.Repos.NewsletterSubscribers().List(c.Request.Context(), status, offset, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", int64(page*pageSize) < total)
	_ = total
}

func (h *apiHandler) subscribeNewsletter(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "A valid email is required.", nil))
		return
	}

	normalized := strings.ToLower(email)
	now := time.Now().UTC()
	subscriber := &models.NewsletterSubscriber{
		Email:            email,
		EmailNormalized:  normalized,
		Status:           "active",
		SubscribedAt:     now,
		UnsubscribeToken: generateUnsubscribeToken(),
	}
	subscriber.ID = utils.NewID()
	subscriber.CreatedAt = now
	subscriber.UpdatedAt = now

	if err := h.deps.Repos.NewsletterSubscribers().Create(c.Request.Context(), subscriber); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, gin.H{"subscribed": true})
}

func (h *apiHandler) unsubscribeNewsletter(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		utils.Error(c, utils.NewError(http.StatusBadRequest, "VALIDATION_ERROR", "Email is required.", nil))
		return
	}

	subscriber, err := h.deps.Repos.NewsletterSubscribers().GetByEmail(c.Request.Context(), strings.ToLower(email))
	if err != nil {
		utils.Error(c, err)
		return
	}

	now := time.Now().UTC()
	subscriber.Status = "unsubscribed"
	subscriber.UnsubscribedAt = &now
	subscriber.UpdatedAt = now

	if err := h.deps.Repos.NewsletterSubscribers().Update(c.Request.Context(), subscriber); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"unsubscribed": true})
}

func generateUnsubscribeToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var result []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == ' ' {
			if r == ' ' {
				result = append(result, '-')
			} else {
				result = append(result, r)
			}
		}
	}
	return strings.Trim(string(result), "-")
}
