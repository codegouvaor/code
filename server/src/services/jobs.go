package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
)

const httpStatusTooManyRequests = http.StatusTooManyRequests

// JobHandler processes one job kind.
type JobHandler func(ctx context.Context, job *models.SyncJob) error

// JobService is the lightweight durable worker queue of the platform. Long
// operations (repository synchronisation, webhook processing, search
// indexing, reconciliation) never run inside the HTTP request: they are
// persisted as rows and claimed by workers, which makes them retryable,
// observable and safe to replay.
type JobService struct {
	jobs         interfaces.SyncJobRepository
	logger       *slog.Logger
	events       interfaces.EventBus
	handlers     map[string]JobHandler
	workerID     string
	pollInterval time.Duration
	lockFor      time.Duration
	backoffBase  time.Duration

	mu      sync.Mutex
	started bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewJobService builds the job service.
func NewJobService(jobs interfaces.SyncJobRepository, events interfaces.EventBus, logger *slog.Logger) *JobService {
	if logger == nil {
		logger = slog.Default()
	}
	return &JobService{
		jobs:         jobs,
		logger:       logger,
		events:       events,
		handlers:     map[string]JobHandler{},
		workerID:     "worker-" + utils.NewID()[:8],
		pollInterval: 2 * time.Second,
		lockFor:      2 * time.Minute,
		backoffBase:  30 * time.Second,
		done:         make(chan struct{}),
	}
}

// Register binds a handler to a job kind.
func (s *JobService) Register(kind string, handler JobHandler) {
	s.handlers[kind] = handler
}

// Enqueue persists a job. The idempotency key makes the call safe to repeat:
// enqueueing twice returns the existing job instead of duplicating the work.
func (s *JobService) Enqueue(ctx context.Context, job *models.SyncJob) (*models.SyncJob, error) {
	if job.Kind == "" || job.IdempotencyKey == "" {
		return nil, utils.ErrValidationFailed
	}
	if existing, err := s.jobs.GetByIdempotencyKey(ctx, job.IdempotencyKey); err == nil && existing != nil && existing.ID != "" {
		return existing, nil
	}
	now := time.Now().UTC()
	if job.ID == "" {
		job.ID = utils.NewID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now
	if job.Status == "" {
		job.Status = models.SyncJobQueued
	}
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 5
	}
	if job.RunAfter.IsZero() {
		job.RunAfter = now
	}
	if err := s.jobs.Create(ctx, job); err != nil {
		// A concurrent enqueue with the same key wins: read it back instead of
		// failing the caller.
		if existing, getErr := s.jobs.GetByIdempotencyKey(ctx, job.IdempotencyKey); getErr == nil && existing != nil && existing.ID != "" {
			return existing, nil
		}
		return nil, err
	}
	return job, nil
}

// EnqueueInput is the convenient payload used by services.
type EnqueueInput struct {
	Kind           string
	IdempotencyKey string
	ProjectID      *string
	BindingID      *string
	Payload        map[string]any
	RunAfter       time.Time
}

// Submit enqueues a job from a plain input.
func (s *JobService) Submit(ctx context.Context, input EnqueueInput) (*models.SyncJob, error) {
	job := &models.SyncJob{
		Common:         models.Common{ID: utils.NewID(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		Kind:           input.Kind,
		IdempotencyKey: input.IdempotencyKey,
		ProjectID:      input.ProjectID,
		BindingID:      input.BindingID,
		RunAfter:       input.RunAfter,
	}
	if input.Payload != nil {
		job.Payload = encodeMap(input.Payload)
	}
	return s.Enqueue(ctx, job)
}

// Start launches the worker loop in the background.
func (s *JobService) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return
	}
	workerCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.started = true
	go s.loop(workerCtx)
}

// Stop stops the worker loop.
func (s *JobService) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	started := s.started
	s.started = false
	s.cancel = nil
	s.mu.Unlock()
	if started && cancel != nil {
		cancel()
	}
}

// ProcessOnce claims and runs at most one job. It is exported so that tests and
// operators can drain the queue deterministically.
func (s *JobService) ProcessOnce(ctx context.Context) (bool, error) {
	job, err := s.jobs.ClaimNext(ctx, s.workerID, time.Now().UTC(), s.lockFor)
	if err != nil {
		if utils.AsAppError(err).Code == "SYNC_JOB_NOT_FOUND" {
			return false, nil
		}
		return false, err
	}
	s.run(ctx, job)
	return true, nil
}

func (s *JobService) loop(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				processed, err := s.ProcessOnce(ctx)
				if err != nil {
					s.logger.Error("job claim failed", "error", err)
					break
				}
				if !processed {
					break
				}
				if ctx.Err() != nil {
					return
				}
			}
		}
	}
}

// run executes a claimed job and records the outcome.
func (s *JobService) run(ctx context.Context, job *models.SyncJob) {
	handler, ok := s.handlers[job.Kind]
	if !ok {
		s.fail(ctx, job, errors.New("no handler registered for job kind "+job.Kind), false)
		return
	}
	s.logger.Info("job started",
		"job_id", job.ID, "kind", job.Kind, "attempt", job.Attempts,
		"project_id", derefStringPointer(job.ProjectID), "binding_id", derefStringPointer(job.BindingID))
	started := time.Now()
	err := handler(ctx, job)
	duration := time.Since(started)
	if err == nil {
		now := time.Now().UTC()
		job.Status = models.SyncJobSucceeded
		job.FinishedAt = &now
		job.LastError = ""
		job.LockedBy = ""
		job.LockedUntil = nil
		job.UpdatedAt = now
		if updateErr := s.jobs.Update(ctx, job); updateErr != nil {
			s.logger.Error("job completion not persisted", "job_id", job.ID, "error", updateErr)
		}
		s.logger.Info("job finished",
			"job_id", job.ID, "kind", job.Kind, "status", job.Status,
			"duration_ms", duration.Milliseconds())
		return
	}
	s.fail(ctx, job, err, isRetryableError(err))
}

func (s *JobService) fail(ctx context.Context, job *models.SyncJob, err error, retryable bool) {
	now := time.Now().UTC()
	job.LastError = err.Error()
	job.FinishedAt = &now
	job.LockedBy = ""
	job.LockedUntil = nil
	job.UpdatedAt = now
	if retryable && job.Attempts < job.MaxAttempts {
		job.Status = models.SyncJobFailed
		delay := s.backoff(job.Attempts)
		// Honour the provider retry hint when it is explicit (Rate-Limit
		// headers), so that the platform stops hammering a forge.
		if hinted, ok := providers.RetryDelay(err); ok && hinted > delay {
			delay = hinted
		}
		job.RunAfter = now.Add(delay)
	} else {
		job.Status = models.SyncJobDead
	}
	if updateErr := s.jobs.Update(ctx, job); updateErr != nil {
		s.logger.Error("job failure not persisted", "job_id", job.ID, "error", updateErr)
	}
	s.logger.Warn("job failed",
		"job_id", job.ID, "kind", job.Kind, "status", job.Status,
		"attempt", job.Attempts, "error", err.Error())
	if s.events != nil {
		_ = s.events.Publish(ctx, interfaces.Event{
			ID:        utils.NewID(),
			Topic:     "job.failed",
			Type:      "job.failed",
			Timestamp: now.Format(time.RFC3339Nano),
			Payload: map[string]any{
				"jobId":  job.ID,
				"kind":   job.Kind,
				"status": job.Status,
			},
		})
	}
}

func (s *JobService) backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := s.backoffBase * time.Duration(attempt*attempt)
	if delay > time.Hour {
		delay = time.Hour
	}
	return delay
}

// Stats returns the queue depth by status, for observability endpoints.
func (s *JobService) Stats(ctx context.Context) (map[string]int64, error) {
	out := map[string]int64{}
	for _, status := range []string{models.SyncJobQueued, models.SyncJobRunning, models.SyncJobFailed, models.SyncJobDead, models.SyncJobSucceeded} {
		count, err := s.jobs.CountByStatus(ctx, status)
		if err != nil {
			return nil, err
		}
		out[status] = count
	}
	return out, nil
}

// List returns the recent jobs, optionally filtered by status.
func (s *JobService) List(ctx context.Context, status string, limit int) ([]models.SyncJob, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if status == "" {
		return s.jobs.ListByStatus(ctx, models.SyncJobQueued, limit)
	}
	return s.jobs.ListByStatus(ctx, status, limit)
}

// isRetryableError decides whether a failed job deserves another attempt.
// Provider rate limits and unavailable forges are retried; authorization and
// validation failures are not.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if providers.IsRetryable(err) {
		return true
	}
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return appErr.Status >= 500 || appErr.Status == httpStatusTooManyRequests
	}
	return false
}

func derefStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
