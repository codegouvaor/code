package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	redisclient "github.com/codegouvaor/info/server/internal/redis"
	"github.com/codegouvaor/info/server/src/interfaces"
)

type PresenceService struct {
	logger     *slog.Logger
	redis      *redisclient.Client
	eventBus   interfaces.EventBus
	users      interfaces.UserRepository
	ttl        time.Duration
	mu         sync.RWMutex
	localState map[string]string
	closeCh    chan struct{}
}

func NewPresenceService(
	logger *slog.Logger,
	redis *redisclient.Client,
	eventBus interfaces.EventBus,
	users interfaces.UserRepository,
	ttl time.Duration,
) *PresenceService {
	s := &PresenceService{
		logger:     logger,
		redis:      redis,
		eventBus:   eventBus,
		users:      users,
		ttl:        ttl,
		localState: make(map[string]string),
		closeCh:    make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

func (s *PresenceService) RefreshUserState(ctx context.Context, userID string) error {
	if s.redis != nil && s.redis.Raw != nil {
		key := "presence:" + userID
		return s.redis.Raw.Set(ctx, key, "online", s.ttl).Err()
	}
	s.mu.Lock()
	s.localState[userID] = "online"
	s.mu.Unlock()
	return nil
}

func (s *PresenceService) SetStatus(ctx context.Context, userID, status string) error {
	if s.redis != nil && s.redis.Raw != nil {
		key := "presence:" + userID
		return s.redis.Raw.Set(ctx, key, status, s.ttl).Err()
	}
	s.mu.Lock()
	s.localState[userID] = status
	s.mu.Unlock()
	return nil
}

func (s *PresenceService) GetStatus(ctx context.Context, userID string) string {
	if s.redis != nil && s.redis.Raw != nil {
		key := "presence:" + userID
		status, err := s.redis.Raw.Get(ctx, key).Result()
		if err != nil {
			return "offline"
		}
		return status
	}
	s.mu.RLock()
	status, ok := s.localState[userID]
	s.mu.RUnlock()
	if !ok {
		return "offline"
	}
	return status
}

func (s *PresenceService) cleanupLoop() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

func (s *PresenceService) cleanup() {
	if s.redis != nil && s.redis.Raw != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for userID, status := range s.localState {
		if status == "offline" {
			delete(s.localState, userID)
		}
	}
}

func (s *PresenceService) Close() {
	close(s.closeCh)
}
