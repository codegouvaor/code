package routes

import (
	"crypto/sha256"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/info/server/src/interfaces"
	"github.com/codegouvaor/info/server/src/models"
	"github.com/codegouvaor/info/server/src/services"
	"github.com/codegouvaor/info/server/src/utils"
)

func (h *apiHandler) register(c *gin.Context) {
	if h.deps.AuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	var req struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		DisplayName   string `json:"displayName"`
		WorkspaceName string `json:"workspaceName"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.AuthService.Register(c.Request.Context(), services.RegisterInput{
		Email:         req.Email,
		Password:      req.Password,
		DisplayName:   req.DisplayName,
		WorkspaceName: req.WorkspaceName,
	}, requestMetadata(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	h.deps.AuthService.SetRefreshCookie(c, result.RefreshToken)
	utils.Success(c, http.StatusCreated, result)
}

func (h *apiHandler) login(c *gin.Context) {
	if h.deps.AuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.AuthService.Login(c.Request.Context(), services.LoginInput{Email: req.Email, Password: req.Password}, requestMetadata(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	hasMfa := false
	if h.deps.MfaService != nil {
		hasMfa = h.deps.MfaService.HasMFA(c.Request.Context(), result.User.ID)
	}
	h.deps.AuthService.SetRefreshCookie(c, result.RefreshToken)
	utils.Success(c, http.StatusOK, gin.H{
		"user":         result.User,
		"accessToken":  result.AccessToken,
		"expiresIn":    result.ExpiresIn,
		"refreshToken": result.RefreshToken,
		"sessionId":    result.SessionID,
		"hasMfa":       hasMfa,
	})
}

func (h *apiHandler) refresh(c *gin.Context) {
	if !h.validateCSRFCookieOrigin(c) {
		return
	}
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = c.ShouldBindJSON(&body)
	refreshToken := strings.TrimSpace(body.RefreshToken)
	if refreshToken == "" {
		cookieToken, err := c.Cookie(h.deps.Config.Auth.RefreshCookieName)
		if err != nil {
			utils.Error(c, utils.ErrUnauthorized)
			return
		}
		refreshToken = cookieToken
	}
	result, err := h.deps.AuthService.Refresh(c.Request.Context(), refreshToken, requestMetadata(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	h.deps.AuthService.SetRefreshCookie(c, result.RefreshToken)
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) logout(c *gin.Context) {
	if !h.validateCSRFCookieOrigin(c) {
		return
	}
	if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
		if principal, err := h.deps.IdentityProvider.Authenticate(c.Request.Context(), strings.TrimPrefix(header, "Bearer ")); err == nil {
			_ = h.deps.AuthService.Logout(c.Request.Context(), principal.SessionID)
		}
	}
	h.deps.AuthService.ClearRefreshCookie(c)
	utils.Success(c, http.StatusOK, gin.H{"loggedOut": true})
}

func (h *apiHandler) logoutAll(c *gin.Context) {
	if !h.validateCSRFCookieOrigin(c) {
		return
	}
	principal, _ := h.principal(c)
	var req struct {
		ExceptCurrent bool `json:"exceptCurrent"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.deps.AuthService.LogoutAll(c.Request.Context(), principal.UserID, req.ExceptCurrent, principal.SessionID); err != nil {
		utils.Error(c, err)
		return
	}
	if !req.ExceptCurrent {
		h.deps.AuthService.ClearRefreshCookie(c)
	}
	utils.Success(c, http.StatusOK, gin.H{"loggedOut": true})
}

func (h *apiHandler) authMe(c *gin.Context) {
	principal, _ := h.principal(c)
	user, err := h.deps.AuthService.CurrentUser(c.Request.Context(), principal)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, user)
}

func (h *apiHandler) changePassword(c *gin.Context) {
	principal, _ := h.principal(c)
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	if err := h.deps.AuthService.ChangePassword(c.Request.Context(), principal, req.CurrentPassword, req.NewPassword, requestMetadata(c)); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"updated": true})
}

func (h *apiHandler) listAuthSessions(c *gin.Context) {
	principal, _ := h.principal(c)
	items, err := h.deps.AuthService.ListSessions(c.Request.Context(), principal)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.List(c, items, "", false)
}

func (h *apiHandler) deleteAuthSession(c *gin.Context) {
	principal, _ := h.principal(c)
	if err := h.deps.AuthService.RevokeSession(c.Request.Context(), principal, c.Param("sessionId")); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *apiHandler) forgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	if err := h.deps.AuthService.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"accepted": true})
}

func (h *apiHandler) resetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Token == "" || len(req.Password) < 12 {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	tokenHash := hashToken(req.Token)
	token, err := h.deps.Repos.PasswordResetTokens().GetByHash(c.Request.Context(), tokenHash)
	if err != nil {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}
	now := time.Now().UTC()
	if token.ConsumedAt != nil || token.ExpiresAt.Before(now) {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	hasher := services.NewPasswordHasher(h.deps.Config.Auth)
	newHash, err := hasher.Hash(req.Password)
	if err != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	credential, err := h.deps.Repos.LocalCredentials().GetByUserID(c.Request.Context(), token.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	credential.PasswordHash = newHash
	credential.UpdatedAt = now
	_ = h.deps.Repos.LocalCredentials().Update(c.Request.Context(), credential)

	user, _ := h.deps.Repos.Users().GetByID(c.Request.Context(), token.UserID)
	user.PasswordChangedAt = &now
	user.UpdatedAt = now
	_ = h.deps.Repos.Users().Update(c.Request.Context(), user)

	token.ConsumedAt = &now
	_ = h.deps.Repos.PasswordResetTokens().Update(c.Request.Context(), token)

	_ = h.deps.AuthService.LogoutAll(c.Request.Context(), token.UserID, false, "")

	utils.Success(c, http.StatusOK, gin.H{"reset": true})
}

func (h *apiHandler) verifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Token == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	tokenHash := hashToken(req.Token)
	token, err := h.deps.Repos.EmailVerificationTokens().GetByHash(c.Request.Context(), tokenHash)
	if err != nil {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}
	now := time.Now().UTC()
	if token.ConsumedAt != nil || token.ExpiresAt.Before(now) {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	user, err := h.deps.Repos.Users().GetByID(c.Request.Context(), token.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now
	_ = h.deps.Repos.Users().Update(c.Request.Context(), user)

	token.ConsumedAt = &now
	_ = h.deps.Repos.EmailVerificationTokens().Update(c.Request.Context(), token)

	utils.Success(c, http.StatusOK, gin.H{"verified": true})
}

func (h *apiHandler) resendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Email == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	_, emailNorm := normalizeEmailAuth(req.Email)
	user, err := h.deps.Repos.Users().GetByEmail(c.Request.Context(), emailNorm)
	if err != nil {
		utils.Success(c, http.StatusAccepted, gin.H{"accepted": true})
		return
	}
	if user.EmailVerifiedAt != nil {
		utils.Success(c, http.StatusAccepted, gin.H{"accepted": true, "alreadyVerified": true})
		return
	}
	now := time.Now().UTC()
	token, hash, _ := issueOpaqueTokenAuth(32)
	_ = token
	model := &models.EmailVerificationToken{
		Common:       models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		UserID:       user.ID,
		TokenHash:    hash,
		ExpiresAt:    now.Add(24 * time.Hour),
		LastSentAt:   &now,
		RequestCount: 1,
	}
	_ = h.deps.Repos.EmailVerificationTokens().Create(c.Request.Context(), model)
	utils.Success(c, http.StatusAccepted, gin.H{"accepted": true})
}

func (h *apiHandler) oauthLogin(c *gin.Context) {
	provider := c.Param("provider")
	if h.deps.OAuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	action := c.DefaultQuery("action", "login")
	var userID string
	if action == "link" {
		p, ok := h.principal(c)
		if !ok {
			utils.Error(c, utils.ErrUnauthorized)
			return
		}
		userID = p.UserID
	}
	authURL, err := h.deps.OAuthService.GetAuthorizationURL(c.Request.Context(), provider, action, userID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"url": authURL})
}

func (h *apiHandler) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if h.deps.OAuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.OAuthService.HandleCallback(c.Request.Context(), provider, code, state, requestMetadata(c))
	if err != nil {
		frontendURL := h.deps.Config.App.FrontendURL
		c.Redirect(http.StatusFound, frontendURL+"/auth/callback?error="+err.Error())
		return
	}
	h.deps.OAuthService.SetOAuthCookies(c, result)
	frontendURL := h.deps.Config.App.FrontendURL
	c.Redirect(http.StatusFound, frontendURL+"/auth/callback")
}

func (h *apiHandler) listOAuthAccounts(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if h.deps.OAuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	accounts, err := h.deps.OAuthService.ListLinkedAccounts(c.Request.Context(), p.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	type accountDTO struct {
		Provider          string  `json:"provider"`
		ProviderAccountID string  `json:"providerAccountId"`
		Scopes            *string `json:"scopes,omitempty"`
		CreatedAt         string  `json:"createdAt"`
	}
	dtos := make([]accountDTO, 0, len(accounts))
	for _, a := range accounts {
		dtos = append(dtos, accountDTO{
			Provider:          a.Provider,
			ProviderAccountID: a.ProviderAccountID,
			Scopes:            a.Scopes,
			CreatedAt:         a.CreatedAt.Format(time.RFC3339),
		})
	}
	utils.Success(c, http.StatusOK, gin.H{"accounts": dtos})
}

func (h *apiHandler) unlinkOAuthAccount(c *gin.Context) {
	provider := c.Param("provider")
	p, ok := h.principal(c)
	if !ok {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	if h.deps.OAuthService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	if err := h.deps.OAuthService.UnlinkAccount(c.Request.Context(), p.UserID, provider); err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"unlinked": true})
}

func requestMetadata(c *gin.Context) services.RequestMetadata {
	return services.RequestMetadata{
		UserAgent: strings.TrimSpace(c.GetHeader("User-Agent")),
		IPAddress: strings.TrimSpace(c.ClientIP()),
	}
}

func (h *apiHandler) validateCSRFCookieOrigin(c *gin.Context) bool {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		utils.Error(c, utils.ErrForbidden)
		return false
	}
	for _, allowed := range h.deps.Config.CORS.AllowedOrigins {
		if allowed == "" {
			continue
		}
		if allowed == origin {
			return true
		}
		allowedURL, parseErr := url.Parse(allowed)
		if parseErr == nil && allowedURL.Scheme == parsed.Scheme && allowedURL.Host == parsed.Host {
			return true
		}
	}
	utils.Error(c, utils.ErrForbidden)
	return false
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func issueOpaqueTokenAuth(length int) (string, string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func normalizeEmailAuth(email string) (string, string) {
	trimmed := strings.TrimSpace(email)
	return trimmed, strings.ToLower(trimmed)
}

// ensureFirstUserIsAdmin s'assure que le premier utilisateur a les rôles admin
// Cette route peut être appelée manuellement si nécessaire
func (h *apiHandler) ensureFirstUserIsAdmin(c *gin.Context) {
	// Vérifier que l'utilisateur actuel est admin (pour éviter les abus)
	principal, exists := c.Get("principal")
	if !exists {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	
	p := principal.(interfaces.Principal)
	isAdmin := false
	for _, role := range p.Roles {
		if role == "admin" || role == "superadmin" {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin {
		utils.Error(c, utils.ErrForbidden)
		return
	}
	
	if err := h.deps.AuthService.EnsureFirstUserHasAdminRoles(c.Request.Context()); err != nil {
		utils.Error(c, err)
		return
	}
	
	utils.Success(c, http.StatusOK, gin.H{"message": "First user has been ensured admin roles"})
}

// getFirstUserInfo retourne les informations du premier utilisateur
func (h *apiHandler) getFirstUserInfo(c *gin.Context) {
	// Trouver le premier utilisateur
	var firstUser models.User
	if err := h.deps.Database.Gorm().
		Model(&models.User{}).
		Order("created_at ASC").
		First(&firstUser).Error; err != nil {
		utils.Error(c, utils.NewError(http.StatusNotFound, "NOT_FOUND", "No users found", nil))
		return
	}
	
	userRoles, _ := h.deps.Repos.UserRoles().ListByUser(c.Request.Context(), firstUser.ID)
	var roles []string
	for _, ur := range userRoles {
		role, err := h.deps.Repos.Roles().GetByID(c.Request.Context(), ur.RoleID)
		if err == nil {
			roles = append(roles, role.Slug)
		}
	}

	utils.Success(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":          firstUser.ID,
			"email":       firstUser.Email,
			"displayName": firstUser.DisplayName,
			"roles":       roles,
			"createdAt":   firstUser.CreatedAt,
		},
	})
}

// ensureUserIsOwner force un utilisateur spécifique à être owner/superadmin
// Endpoint temporaire pour corriger manuellement un utilisateur
func (h *apiHandler) ensureUserIsOwner(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Email == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	
	// Vérifier que l'utilisateur actuel est admin
	principal, exists := c.Get("principal")
	if !exists {
		utils.Error(c, utils.ErrUnauthorized)
		return
	}
	
	p := principal.(interfaces.Principal)
	isAdmin := false
	for _, role := range p.Roles {
		if role == "admin" || role == "superadmin" {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin {
		utils.Error(c, utils.ErrForbidden)
		return
	}
	
	// Trouver l'utilisateur par email
	var user models.User
	if err := h.deps.Database.Gorm().
		Model(&models.User{}).
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {
		utils.Error(c, utils.NewError(http.StatusNotFound, "NOT_FOUND", "User not found", nil))
		return
	}
	
	// Mettre à jour les rôles via UserRoles repository
	adminSlugs := []string{"superadmin", "admin", "owner"}
	for _, slug := range adminSlugs {
		role, err := h.deps.Repos.Roles().GetBySlug(c.Request.Context(), slug)
		if err != nil {
			continue
		}
		existing, _ := h.deps.Repos.UserRoles().GetByUserAndRole(c.Request.Context(), user.ID, role.ID)
		if existing == nil {
			_ = h.deps.Repos.UserRoles().Assign(c.Request.Context(), &models.UserRole{
				UserID:     user.ID,
				RoleID:     role.ID,
				AssignedAt: time.Now().UTC(),
			})
		}
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "User roles updated to superadmin",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"roles": adminSlugs,
		},
	})
}

func (h *apiHandler) mfaSetup(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	result, err := h.deps.MfaService.Setup(c.Request.Context(), principal.UserID)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) mfaVerify(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	var req struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Code) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.MfaService.VerifyAndEnable(c.Request.Context(), principal.UserID, strings.TrimSpace(req.Code))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) mfaDisable(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	var req struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Code) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.MfaService.Disable(c.Request.Context(), principal.UserID, strings.TrimSpace(req.Code))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) mfaValidateLogin(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	var req struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Code) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.MfaService.ValidateLogin(c.Request.Context(), principal.UserID, strings.TrimSpace(req.Code))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) mfaRecoveryCodes(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	result, err := h.deps.MfaService.ValidateRecoveryCode(c.Request.Context(), principal.UserID, c.Query("code"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}

func (h *apiHandler) mfaRegenerateRecoveryCodes(c *gin.Context) {
	if h.deps.MfaService == nil {
		utils.Error(c, utils.ErrDependencyUnavailable)
		return
	}
	principal, _ := h.principal(c)
	var req struct {
		Code string `json:"code"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Code) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	result, err := h.deps.MfaService.RegenerateRecoveryCodes(c.Request.Context(), principal.UserID, strings.TrimSpace(req.Code))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, http.StatusOK, result)
}
