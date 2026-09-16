package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/codegouvaor/code/server/src/utils"
)

// RepositoryRef identifies a repository on a provider without carrying any
// credential.
type RepositoryRef struct {
	Owner string
	Repo  string
}

// FullName renders the canonical "owner/repo" form.
func (r RepositoryRef) FullName() string { return r.Owner + "/" + r.Repo }

// ListQuery is the common pagination/state input of list operations.
type ListQuery struct {
	Ref   RepositoryRef
	State string
	Limit int
	Page  int
}

// TreeQuery describes a directory listing request.
type TreeQuery struct {
	Ref        RepositoryRef
	Path       string
	RefName    string
	Recursive  bool
	MaxEntries int
}

// BlobQuery describes a single file request.
type BlobQuery struct {
	Ref     RepositoryRef
	Path    string
	RefName string
}

// CommitQuery describes a commit listing request.
type CommitQuery struct {
	Ref     RepositoryRef
	RefName string
	Path    string
	Limit   int
	Page    int
}

// OrganizationQuery describes a provider organization listing request.
type OrganizationQuery struct {
	Login string
	Limit int
}

// Repository is the provider-side view of a repository.
type Repository struct {
	Provider      string    `json:"provider"`
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"fullName"`
	Description   string    `json:"description"`
	URL           string    `json:"url"`
	CloneURL      string    `json:"cloneUrl"`
	SSHURL        string    `json:"sshUrl"`
	DefaultBranch string    `json:"defaultBranch"`
	Visibility    string    `json:"visibility"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	Topics        []string  `json:"topics"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	OpenIssues    int       `json:"openIssues"`
	Language      string    `json:"language,omitempty"`
	OwnerLogin    string    `json:"ownerLogin"`
	OwnerType     string    `json:"ownerType"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	PushedAt      time.Time `json:"pushedAt"`
}

// Branch is a provider branch.
type Branch struct {
	Name      string    `json:"name"`
	Default   bool      `json:"default"`
	Protected bool      `json:"protected"`
	CommitSHA string    `json:"commitSha"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Commit is a provider commit.
type Commit struct {
	SHA         string    `json:"sha"`
	Message     string    `json:"message"`
	AuthorLogin string    `json:"authorLogin,omitempty"`
	AuthorName  string    `json:"authorName,omitempty"`
	URL         string    `json:"url,omitempty"`
	Parents     []string  `json:"parents,omitempty"`
	CommittedAt time.Time `json:"committedAt"`
}

// TreeEntry is one entry of a directory listing.
type TreeEntry struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"` // "file" or "dir"
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
}

// Blob is a file content at a given revision.
type Blob struct {
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Encoding string `json:"encoding"`
	Size     int64  `json:"size"`
	Content  string `json:"content"`
	TooLarge bool   `json:"tooLarge"`
	Binary   bool   `json:"binary"`
}

// Issue is a provider issue.
type Issue struct {
	ID          string    `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body,omitempty"`
	State       string    `json:"state"`
	AuthorLogin string    `json:"authorLogin,omitempty"`
	URL         string    `json:"url"`
	Labels      []string  `json:"labels"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ClosedAt    time.Time `json:"closedAt,omitzero"`
}

// Review is a provider change proposal (pull request or merge request).
type Review struct {
	ID           string    `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body,omitempty"`
	State        string    `json:"state"`
	AuthorLogin  string    `json:"authorLogin,omitempty"`
	URL          string    `json:"url"`
	SourceBranch string    `json:"sourceBranch,omitempty"`
	TargetBranch string    `json:"targetBranch,omitempty"`
	Draft        bool      `json:"draft"`
	Merged       bool      `json:"merged"`
	Comments     int       `json:"comments"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MergedAt     time.Time `json:"mergedAt,omitzero"`
}

// Release is a provider release.
type Release struct {
	ID          string    `json:"id"`
	TagName     string    `json:"tagName"`
	Name        string    `json:"name"`
	Body        string    `json:"body,omitempty"`
	URL         string    `json:"url"`
	AuthorLogin string    `json:"authorLogin,omitempty"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"createdAt"`
	PublishedAt time.Time `json:"publishedAt,omitzero"`
}

// ExternalOrganization is a provider organization.
type ExternalOrganization struct {
	Provider    string `json:"provider"`
	ID          string `json:"id"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	Type        string `json:"type"`
}

// Adapter is anything the registry can hand out for a provider.
type Adapter interface {
	Name() string
	Descriptor() Descriptor
}

// The capability interfaces below are deliberately small and independent: an
// adapter implements only what the provider actually exposes.

type RepositoryProvider interface {
	Repository(ctx context.Context, ref RepositoryRef) (*Repository, error)
}

type BranchProvider interface {
	Branches(ctx context.Context, query ListQuery) ([]Branch, error)
}

type CommitProvider interface {
	Commits(ctx context.Context, query CommitQuery) ([]Commit, error)
}

type FileProvider interface {
	Tree(ctx context.Context, query TreeQuery) ([]TreeEntry, error)
	Blob(ctx context.Context, query BlobQuery) (*Blob, error)
}

type IssueProvider interface {
	Issues(ctx context.Context, query ListQuery) ([]Issue, error)
}

type ReviewProvider interface {
	Reviews(ctx context.Context, query ListQuery) ([]Review, error)
}

type ReleaseProvider interface {
	Releases(ctx context.Context, query ListQuery) ([]Release, error)
}

type OrganizationProvider interface {
	Organizations(ctx context.Context, query OrganizationQuery) ([]ExternalOrganization, error)
}

// ── Errors ───────────────────────────────────────────────────────────────────

// ErrorKind classifies every provider failure so that the platform can react
// consistently (retry, back off, surface a capability error, …).
type ErrorKind string

const (
	ErrorUnavailable  ErrorKind = "provider_unavailable"
	ErrorRateLimited  ErrorKind = "provider_rate_limited"
	ErrorNotFound     ErrorKind = "provider_not_found"
	ErrorUnauthorized ErrorKind = "provider_unauthorized"
	ErrorInvalid      ErrorKind = "provider_invalid_request"
	ErrorConflict     ErrorKind = "provider_conflict"
	ErrorUnsupported  ErrorKind = "provider_unsupported"
	ErrorUnknown      ErrorKind = "provider_error"
)

// Error is the typed error returned by every adapter.
type Error struct {
	Kind       ErrorKind
	Provider   string
	Operation  string
	Message    string
	Status     int
	RetryAfter time.Duration
	Cause      error
}

func (e *Error) Error() string {
	if e.Provider == "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.Provider, e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// NewError builds a typed provider error.
func NewError(kind ErrorKind, provider, operation, message string, cause error) *Error {
	return &Error{Kind: kind, Provider: provider, Operation: operation, Message: message, Cause: cause}
}

// Retryable reports whether the operation may be retried as-is.
func (e *Error) Retryable() bool {
	switch e.Kind {
	case ErrorRateLimited, ErrorUnavailable, ErrorUnknown:
		return true
	default:
		return false
	}
}

// AsAppError converts any provider error into the API error contract used by
// the rest of the server, keeping the response envelope consistent.
func AsAppError(err error) *utils.AppError {
	var providerErr *Error
	if !errors.As(err, &providerErr) {
		return utils.AsAppError(err)
	}
	switch providerErr.Kind {
	case ErrorNotFound:
		return utils.ErrProviderResourceNotFound
	case ErrorUnauthorized:
		return utils.ErrProviderUnauthorized
	case ErrorRateLimited:
		return utils.ErrProviderRateLimited
	case ErrorUnavailable:
		return utils.ErrProviderUnavailable
	case ErrorUnsupported:
		return utils.ErrProviderCapabilityMissing
	default:
		return utils.NewError(http.StatusBadGateway, "PROVIDER_ERROR", "The provider returned an unexpected response.", nil)
	}
}
