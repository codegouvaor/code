package utils

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Status  int            `json:"-"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(status int, code, message string, details map[string]any) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Details: details}
}

func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred.", nil)
}

var (
	ErrUnauthorized               = NewError(http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required.", nil)
	ErrForbidden                  = NewError(http.StatusForbidden, "FORBIDDEN", "You are not allowed to access this resource.", nil)
	ErrValidationFailed           = NewError(http.StatusBadRequest, "VALIDATION_ERROR", "The request payload is invalid.", nil)
	ErrEmailAlreadyExists         = NewError(http.StatusConflict, "EMAIL_ALREADY_EXISTS", "An account with this email already exists.", nil)
	ErrInvalidCredentials         = NewError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password.", nil)
	ErrAccountSuspended           = NewError(http.StatusForbidden, "ACCOUNT_SUSPENDED", "This account is suspended.", nil)
	ErrAccountPending             = NewError(http.StatusForbidden, "ACCOUNT_PENDING", "This account is pending verification.", nil)
	ErrSessionExpired             = NewError(http.StatusUnauthorized, "SESSION_EXPIRED", "The session has expired.", nil)
	ErrSessionRevoked             = NewError(http.StatusUnauthorized, "SESSION_REVOKED", "The session has been revoked.", nil)
	ErrTokenReuseDetected         = NewError(http.StatusUnauthorized, "TOKEN_REUSE_DETECTED", "Refresh token reuse was detected.", nil)
	ErrWorkspaceNotFound          = NewError(http.StatusNotFound, "WORKSPACE_NOT_FOUND", "The requested workspace was not found.", nil)
	ErrMembershipRequired         = NewError(http.StatusForbidden, "MEMBERSHIP_REQUIRED", "Workspace membership is required.", nil)
	ErrDependencyUnavailable      = NewError(http.StatusServiceUnavailable, "DEPENDENCY_UNAVAILABLE", "A required dependency is unavailable.", nil)
	ErrOAuthProviderNotConfigured = NewError(http.StatusServiceUnavailable, "OAUTH_PROVIDER_NOT_CONFIGURED", "The OAuth provider is not configured.", nil)
	ErrOAuthProviderMismatch      = NewError(http.StatusBadRequest, "OAUTH_PROVIDER_MISMATCH", "The OAuth provider does not match.", nil)
	ErrOAuthAccountAlreadyLinked  = NewError(http.StatusConflict, "OAUTH_ACCOUNT_ALREADY_LINKED", "This OAuth account is already linked to a user.", nil)
	ErrOAuthAccountNotFound       = NewError(http.StatusNotFound, "OAUTH_ACCOUNT_NOT_FOUND", "The OAuth account was not found.", nil)
	ErrOAuthStateInvalid          = NewError(http.StatusBadRequest, "OAUTH_STATE_INVALID", "The OAuth state parameter is invalid or expired.", nil)

	ErrNotFound             = NewError(http.StatusNotFound, "NOT_FOUND", "The requested resource was not found.", nil)
	ErrSlugAlreadyExists    = NewError(http.StatusConflict, "SLUG_ALREADY_EXISTS", "A resource with this slug already exists.", nil)
	ErrAnimeNotFound        = NewError(http.StatusNotFound, "ANIME_NOT_FOUND", "The requested anime was not found.", nil)
	ErrEpisodeNotFound      = NewError(http.StatusNotFound, "EPISODE_NOT_FOUND", "The requested episode was not found.", nil)
	ErrGenreNotFound        = NewError(http.StatusNotFound, "GENRE_NOT_FOUND", "The requested genre was not found.", nil)
	ErrStudioNotFound       = NewError(http.StatusNotFound, "STUDIO_NOT_FOUND", "The requested studio was not found.", nil)
	ErrCharacterNotFound    = NewError(http.StatusNotFound, "CHARACTER_NOT_FOUND", "The requested character was not found.", nil)
	ErrMediaNotFound        = NewError(http.StatusNotFound, "MEDIA_NOT_FOUND", "The requested media asset was not found.", nil)
	ErrReviewNotFound       = NewError(http.StatusNotFound, "REVIEW_NOT_FOUND", "The requested review was not found.", nil)
	ErrCommentNotFound      = NewError(http.StatusNotFound, "COMMENT_NOT_FOUND", "The requested comment was not found.", nil)
	ErrNotificationNotFound = NewError(http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "The requested notification was not found.", nil)
	ErrReportNotFound       = NewError(http.StatusNotFound, "REPORT_NOT_FOUND", "The requested report was not found.", nil)
	ErrWatchlistNotFound    = NewError(http.StatusNotFound, "WATCHLIST_NOT_FOUND", "The requested watchlist was not found.", nil)
	ErrSimulcastNotFound    = NewError(http.StatusNotFound, "SIMULCAST_NOT_FOUND", "The requested simulcast was not found.", nil)
	ErrDuplicateReview      = NewError(http.StatusConflict, "DUPLICATE_REVIEW", "You have already reviewed this anime.", nil)
	ErrInvalidToken         = NewError(http.StatusBadRequest, "INVALID_TOKEN", "The provided token is invalid or expired.", nil)
	ErrTooManyRequests      = NewError(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Too many requests. Please try again later.", nil)
	ErrEncodingJobNotFound  = NewError(http.StatusNotFound, "ENCODING_JOB_NOT_FOUND", "The requested encoding job was not found.", nil)
	ErrAuditLogNotFound     = NewError(http.StatusNotFound, "AUDIT_LOG_NOT_FOUND", "The requested audit log was not found.", nil)
	ErrContactNotFound      = NewError(http.StatusNotFound, "CONTACT_NOT_FOUND", "The requested contact was not found.", nil)
	ErrContactGroupNotFound = NewError(http.StatusNotFound, "CONTACT_GROUP_NOT_FOUND", "The requested contact group was not found.", nil)
	ErrSystemSettingNotFound = NewError(http.StatusNotFound, "SYSTEM_SETTING_NOT_FOUND", "The requested system setting was not found.", nil)

	// Content errors
	ErrArticleNotFound    = NewError(http.StatusNotFound, "ARTICLE_NOT_FOUND", "The requested article was not found.", nil)
	ErrCategoryNotFound   = NewError(http.StatusNotFound, "CATEGORY_NOT_FOUND", "The requested category was not found.", nil)
	ErrTagNotFound        = NewError(http.StatusNotFound, "TAG_NOT_FOUND", "The requested tag was not found.", nil)
	ErrWebhookNotFound    = NewError(http.StatusNotFound, "WEBHOOK_NOT_FOUND", "The requested webhook was not found.", nil)
	ErrSeoConfigNotFound  = NewError(http.StatusNotFound, "SEO_CONFIG_NOT_FOUND", "The requested SEO config was not found.", nil)
	ErrSubscriberNotFound = NewError(http.StatusNotFound, "SUBSCRIBER_NOT_FOUND", "The requested subscriber was not found.", nil)
	ErrScheduleNotFound   = NewError(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "The requested schedule was not found.", nil)
)
