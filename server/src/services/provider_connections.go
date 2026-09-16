package services

import (
	"context"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/config"
	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
)

// ProviderConnectionService manages the credentials the platform uses to talk
// to external forges. It reuses the OAuth machinery for the authorization
// code flow but keeps the resulting grant strictly separate from the Code
// identity: connecting GitHub to read repositories never affects how the user
// signs in.
type ProviderConnectionService struct {
	connections interfaces.ProviderConnectionRepository
	registry    *providers.Registry
	cfg         config.ProvidersConfig
	box         *secretBox
	events      interfaces.EventBus
}

// ConnectionInput is what the OAuth callback hands over.
type ConnectionInput struct {
	Provider          string
	ProviderAccountID string
	AccountLogin      string
	AvatarURL         string
	Scopes            string
	AccessToken       string
	RefreshToken      string
	ExpiresAt         *time.Time
}

// ProviderConnectionView is the safe API representation of a connection.
type ProviderConnectionView struct {
	ID                string            `json:"id"`
	Provider          string            `json:"provider"`
	ProviderName      string            `json:"providerName"`
	ProviderAccountID string            `json:"providerAccountId"`
	AccountLogin      string            `json:"accountLogin"`
	AvatarURL         *string           `json:"avatarUrl,omitempty"`
	Scopes            []string          `json:"scopes"`
	Status            string            `json:"status"`
	TokenExpiresAt    *time.Time        `json:"tokenExpiresAt,omitempty"`
	LastValidatedAt   *time.Time        `json:"lastValidatedAt,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
	Capabilities      map[string]string `json:"capabilities"`
	AuthorizeURL      string            `json:"authorizeUrl,omitempty"`
}

// NewProviderConnectionService builds the connection service.
func NewProviderConnectionService(
	connections interfaces.ProviderConnectionRepository,
	registry *providers.Registry,
	cfg config.ProvidersConfig,
	events interfaces.EventBus,
) (*ProviderConnectionService, error) {
	box, err := newSecretBox(cfg.EncryptionKey)
	if err != nil {
		// The service stays usable without encryption support (it will refuse
		// to persist credentials) so that a misconfigured environment does not
		// prevent the platform from starting.
		box = nil
	}
	return &ProviderConnectionService{
		connections: connections,
		registry:    registry,
		cfg:         cfg,
		box:         box,
		events:      events,
	}, nil
}

// Upsert stores (or refreshes) the grant obtained through OAuth. Existing
// tokens are replaced, never duplicated.
func (s *ProviderConnectionService) Upsert(ctx context.Context, userID string, input ConnectionInput) (*models.ProviderConnection, error) {
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if !s.registry.Has(provider) {
		return nil, utils.ErrProviderNotSupported
	}
	if s.box == nil {
		return nil, utils.ErrEncryptionUnavailable
	}
	accessToken, err := s.box.Seal(input.AccessToken)
	if err != nil {
		return nil, utils.ErrEncryptionUnavailable
	}
	refreshToken, err := s.box.Seal(input.RefreshToken)
	if err != nil {
		return nil, utils.ErrEncryptionUnavailable
	}
	now := time.Now().UTC()
	existing, lookupErr := s.connections.GetByProviderAccount(ctx, provider, input.ProviderAccountID)

	connection := &models.ProviderConnection{}
	if lookupErr == nil && existing != nil && existing.ID != "" {
		connection = existing
		if connection.UserID != userID {
			// The same forge account cannot be attached to two Code accounts:
			// it would blur the ownership of the repositories.
			return nil, utils.ErrProviderConnectionExists
		}
	} else {
		connection.Common = models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now}
	}
	connection.UserID = userID
	connection.Provider = provider
	connection.ProviderAccountID = input.ProviderAccountID
	connection.AccountLogin = input.AccountLogin
	connection.Scopes = input.Scopes
	connection.AccessTokenEnc = accessToken
	connection.RefreshTokenEnc = refreshToken
	connection.TokenExpiresAt = input.ExpiresAt
	connection.Status = models.ConnectionActive
	connection.RevokedAt = nil
	connection.LastValidatedAt = &now
	connection.UpdatedAt = now
	if avatar := strings.TrimSpace(input.AvatarURL); avatar != "" {
		connection.AvatarURL = &avatar
	}
	if connection.CreatedAt.IsZero() {
		connection.CreatedAt = now
	}

	if lookupErr == nil && existing != nil && existing.ID != "" {
		if err := s.connections.Update(ctx, connection); err != nil {
			return nil, err
		}
	} else {
		if err := s.connections.Create(ctx, connection); err != nil {
			return nil, err
		}
	}
	s.publish(ctx, "provider.connected", userID, map[string]any{
		"provider": provider,
		"login":    connection.AccountLogin,
	})
	return connection, nil
}

// List returns the connections of a user with their safe representation.
func (s *ProviderConnectionService) List(ctx context.Context, userID string) ([]ProviderConnectionView, error) {
	items, err := s.connections.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderConnectionView, 0, len(items))
	for index := range items {
		out = append(out, s.ToView(&items[index]))
	}
	return out, nil
}

// ToView converts a stored connection into its API representation. Tokens are
// never exposed.
func (s *ProviderConnectionService) ToView(connection *models.ProviderConnection) ProviderConnectionView {
	view := ProviderConnectionView{
		ID:                connection.ID,
		Provider:          connection.Provider,
		ProviderName:      connection.Provider,
		ProviderAccountID: connection.ProviderAccountID,
		AccountLogin:      connection.AccountLogin,
		AvatarURL:         connection.AvatarURL,
		Scopes:            strings.Fields(connection.Scopes),
		Status:            connection.Status,
		TokenExpiresAt:    connection.TokenExpiresAt,
		LastValidatedAt:   connection.LastValidatedAt,
		CreatedAt:         connection.CreatedAt,
		Capabilities:      map[string]string{},
	}
	if descriptor, ok := s.registry.Descriptor(connection.Provider); ok {
		view.ProviderName = descriptor.DisplayName
		view.Capabilities = stringifyCapabilities(descriptor.Capabilities)
		if descriptor.DocumentationURL != "" {
			view.AuthorizeURL = ""
		}
	}
	if view.Scopes == nil {
		view.Scopes = []string{}
	}
	return view
}

// Get returns a single connection of a user.
func (s *ProviderConnectionService) Get(ctx context.Context, userID, provider string) (*models.ProviderConnection, error) {
	connection, err := s.connections.GetByUserAndProvider(ctx, userID, strings.ToLower(strings.TrimSpace(provider)))
	if err != nil {
		return nil, utils.ErrProviderConnectionNotFound
	}
	return connection, nil
}

// Disconnect revokes a connection. Tokens are wiped, the row is kept so that
// the audit trail of what was connected (and when) survives.
func (s *ProviderConnectionService) Disconnect(ctx context.Context, userID, provider string) error {
	connection, err := s.connections.GetByUserAndProvider(ctx, userID, strings.ToLower(strings.TrimSpace(provider)))
	if err != nil {
		return utils.ErrProviderConnectionNotFound
	}
	now := time.Now().UTC()
	connection.Status = models.ConnectionRevoked
	connection.AccessTokenEnc = ""
	connection.RefreshTokenEnc = ""
	connection.RevokedAt = &now
	connection.UpdatedAt = now
	if err := s.connections.Update(ctx, connection); err != nil {
		return err
	}
	s.publish(ctx, "provider.disconnected", userID, map[string]any{"provider": connection.Provider})
	return nil
}

// AccessToken decrypts the credential of a connection. The clear-text value is
// only ever held in memory for the duration of a provider call.
func (s *ProviderConnectionService) AccessToken(connection *models.ProviderConnection) (string, error) {
	if connection == nil {
		return "", utils.ErrProviderConnectionNotFound
	}
	if connection.Status != models.ConnectionActive {
		return "", utils.ErrProviderUnauthorized
	}
	if connection.AccessTokenEnc == "" {
		return "", utils.ErrProviderUnauthorized
	}
	if s.box == nil {
		return "", utils.ErrEncryptionUnavailable
	}
	token, err := s.box.Open(connection.AccessTokenEnc)
	if err != nil {
		return "", utils.ErrProviderUnauthorized
	}
	return token, nil
}

// Adapter returns a provider adapter bound to the credential of a user. When
// the user has no connection, the platform service token (if configured) is
// used so that public repositories stay readable.
func (s *ProviderConnectionService) Adapter(ctx context.Context, userID, provider string) (providers.Adapter, error) {
	name := strings.ToLower(strings.TrimSpace(provider))
	if !s.registry.Has(name) {
		return nil, utils.ErrProviderNotSupported
	}
	token := s.serviceToken(name)
	if userID != "" {
		connection, err := s.connections.GetByUserAndProvider(ctx, userID, name)
		if err == nil && connection != nil && connection.ID != "" {
			connected, tokenErr := s.AccessToken(connection)
			if tokenErr != nil {
				return nil, tokenErr
			}
			token = connected
		} else if err != nil && utils.AsAppError(err).Code != "PROVIDER_CONNECTION_NOT_FOUND" {
			return nil, err
		}
	}
	return s.registry.NewAdapter(name, providers.Options{
		BaseURL:   s.baseURL(name),
		Token:     token,
		Timeout:   s.cfg.Timeout,
		BaseDelay: 100 * time.Millisecond,
	})
}

// AdapterForConnection builds an adapter from an explicit connection.
func (s *ProviderConnectionService) AdapterForConnection(connection *models.ProviderConnection) (providers.Adapter, error) {
	if connection == nil {
		return nil, utils.ErrProviderConnectionNotFound
	}
	token, err := s.AccessToken(connection)
	if err != nil {
		return nil, err
	}
	return s.registry.NewAdapter(connection.Provider, providers.Options{
		BaseURL: s.baseURL(connection.Provider),
		Token:   token,
		Timeout: s.cfg.Timeout,
	})
}

// ConnectionFor resolves the connection that should serve a binding: the
// explicit one when set, otherwise the owner's connection, otherwise the
// platform service credential.
func (s *ProviderConnectionService) ConnectionFor(ctx context.Context, binding *models.RepositoryBinding, ownerUserID string) (*models.ProviderConnection, error) {
	if binding != nil && binding.ConnectionID != nil && *binding.ConnectionID != "" {
		connection, err := s.connections.GetByID(ctx, *binding.ConnectionID)
		if err != nil {
			return nil, utils.ErrProviderConnectionNotFound
		}
		return connection, nil
	}
	if ownerUserID == "" {
		return nil, utils.ErrProviderConnectionNotFound
	}
	connection, err := s.connections.GetByUserAndProvider(ctx, ownerUserID, binding.Provider)
	if err != nil {
		return nil, utils.ErrProviderConnectionNotFound
	}
	return connection, nil
}

// Registry exposes the provider registry for capability lookups.
func (s *ProviderConnectionService) Registry() *providers.Registry { return s.registry }

func (s *ProviderConnectionService) baseURL(provider string) string {
	switch provider {
	case models.ProviderGitHub:
		return s.cfg.GitHub.BaseURL
	case models.ProviderGitLab:
		return s.cfg.GitLab.BaseURL
	case models.ProviderGiteria:
		return s.cfg.Giteria.BaseURL
	default:
		return ""
	}
}

func (s *ProviderConnectionService) serviceToken(provider string) string {
	switch provider {
	case models.ProviderGitHub:
		return s.cfg.GitHub.Token
	case models.ProviderGitLab:
		return s.cfg.GitLab.Token
	case models.ProviderGiteria:
		return s.cfg.Giteria.Token
	default:
		return ""
	}
}

func (s *ProviderConnectionService) publish(ctx context.Context, eventType, actorID string, payload map[string]any) {
	if s.events == nil {
		return
	}
	_ = s.events.Publish(ctx, interfaces.Event{
		ID:        utils.NewID(),
		Topic:     eventType,
		Type:      eventType,
		ActorID:   actorID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:   payload,
	})
}
