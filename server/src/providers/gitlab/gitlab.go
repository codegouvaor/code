// Package gitlab implements the GitLab REST v4 adapter for the Code platform.
package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/providers"
)

// Name is the registry identifier of the adapter.
const Name = "gitlab"

const defaultBaseURL = "https://gitlab.com/api/v4"

// Descriptor describes what GitLab exposes through the platform.
var Descriptor = providers.Descriptor{
	Name:             Name,
	DisplayName:      "GitLab",
	WebsiteURL:       "https://gitlab.com",
	DocumentationURL: "https://docs.gitlab.com/ee/api/rest/",
	Capabilities: map[string]providers.Capability{
		providers.CapabilityRepositories:  providers.CapabilityIntegrated,
		providers.CapabilityBranches:      providers.CapabilityIntegrated,
		providers.CapabilityCommits:       providers.CapabilityIntegrated,
		providers.CapabilityFiles:         providers.CapabilityIntegrated,
		providers.CapabilityIssues:        providers.CapabilityIntegrated,
		providers.CapabilityReviews:       providers.CapabilityIntegrated,
		providers.CapabilityReleases:      providers.CapabilityIntegrated,
		providers.CapabilityOrganizations: providers.CapabilityIntegrated,
		providers.CapabilityWebhooks:      providers.CapabilityIntegrated,
		providers.CapabilityCICD:          providers.CapabilityIntegrated,
		providers.CapabilityDocumentation: providers.CapabilityUnavailable,
		providers.CapabilityAPIs:          providers.CapabilityNative,
		providers.CapabilitySDKs:          providers.CapabilityNative,
		providers.CapabilityServices:      providers.CapabilityNative,
		providers.CapabilityStandards:     providers.CapabilityNative,
		providers.CapabilityPackages:      providers.CapabilityIntegrated,
	},
}

// Factory builds a GitLab adapter for the registry.
func Factory(options providers.Options) (providers.Adapter, error) {
	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Adapter{
		client: providers.NewClient(providers.ClientConfig{
			Provider:    Name,
			BaseURL:     baseURL,
			Token:       options.Token,
			UserAgent:   options.UserAgent,
			Timeout:     options.Timeout,
			HTTPClient:  options.HTTPClient,
			MaxAttempts: options.MaxAttempts,
			BaseDelay:   options.BaseDelay,
		}),
	}, nil
}

// Adapter is the GitLab implementation of the provider capabilities.
type Adapter struct {
	client *providers.Client
}

var (
	_ providers.Adapter              = (*Adapter)(nil)
	_ providers.RepositoryProvider   = (*Adapter)(nil)
	_ providers.BranchProvider       = (*Adapter)(nil)
	_ providers.CommitProvider       = (*Adapter)(nil)
	_ providers.FileProvider         = (*Adapter)(nil)
	_ providers.IssueProvider        = (*Adapter)(nil)
	_ providers.ReviewProvider       = (*Adapter)(nil)
	_ providers.ReleaseProvider      = (*Adapter)(nil)
	_ providers.OrganizationProvider = (*Adapter)(nil)
)

// Name returns the registry identifier.
func (a *Adapter) Name() string { return Name }

// Descriptor returns the capability description of the adapter.
func (a *Adapter) Descriptor() providers.Descriptor { return Descriptor }

type projectPayload struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	PathWithNamespace string    `json:"path_with_namespace"`
	Description       string    `json:"description"`
	WebURL            string    `json:"web_url"`
	HTTPURLToRepo     string    `json:"http_url_to_repo"`
	SSHURLToRepo      string    `json:"ssh_url_to_repo"`
	DefaultBranch     string    `json:"default_branch"`
	Visibility        string    `json:"visibility"`
	Archived          bool      `json:"archived"`
	Topics            []string  `json:"topics"`
	StarCount         int       `json:"star_count"`
	ForksCount        int       `json:"forks_count"`
	OpenIssuesCount   int       `json:"open_issues_count"`
	CreatedAt         time.Time `json:"created_at"`
	LastActivityAt    time.Time `json:"last_activity_at"`
	Namespace         struct {
		FullPath string `json:"full_path"`
		Kind     string `json:"kind"`
	} `json:"namespace"`
}

// Repository fetches the project metadata of a GitLab repository.
func (a *Adapter) Repository(ctx context.Context, ref providers.RepositoryRef) (*providers.Repository, error) {
	var payload projectPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf("/projects/%s", url.PathEscape(ref.FullName())),
	}, &payload)
	if err != nil {
		return nil, err
	}
	ownerType := "organization"
	if payload.Namespace.Kind == "user" {
		ownerType = "user"
	}
	return &providers.Repository{
		Provider:      Name,
		ID:            strconv.FormatInt(payload.ID, 10),
		Name:          payload.Name,
		FullName:      payload.PathWithNamespace,
		Description:   payload.Description,
		URL:           payload.WebURL,
		CloneURL:      payload.HTTPURLToRepo,
		SSHURL:        payload.SSHURLToRepo,
		DefaultBranch: payload.DefaultBranch,
		Visibility:    strings.ToLower(payload.Visibility),
		Archived:      payload.Archived,
		Topics:        orEmpty(payload.Topics),
		Stars:         payload.StarCount,
		Forks:         payload.ForksCount,
		OpenIssues:    payload.OpenIssuesCount,
		OwnerLogin:    payload.Namespace.FullPath,
		OwnerType:     ownerType,
		CreatedAt:     payload.CreatedAt,
		UpdatedAt:     payload.LastActivityAt,
		PushedAt:      payload.LastActivityAt,
	}, nil
}

type branchPayload struct {
	Name      string `json:"name"`
	Default   bool   `json:"default"`
	Protected bool   `json:"protected"`
	Commit    struct {
		ID          string    `json:"id"`
		CommittedAt time.Time `json:"committed_date"`
	} `json:"commit"`
}

// Branches lists the repository branches.
func (a *Adapter) Branches(ctx context.Context, query providers.ListQuery) ([]providers.Branch, error) {
	var payload []branchPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/repository/branches", url.PathEscape(query.Ref.FullName())),
		Query: pagination(query.Limit, query.Page),
	}, &payload)
	if err != nil {
		return nil, err
	}
	branches := make([]providers.Branch, 0, len(payload))
	for _, item := range payload {
		branches = append(branches, providers.Branch{
			Name:      item.Name,
			Default:   item.Default,
			Protected: item.Protected,
			CommitSHA: item.Commit.ID,
			UpdatedAt: item.Commit.CommittedAt,
		})
	}
	return branches, nil
}

type commitPayload struct {
	ID             string    `json:"id"`
	Message        string    `json:"message"`
	Title          string    `json:"title"`
	AuthorName     string    `json:"author_name"`
	CommittedDate  time.Time `json:"committed_date"`
	WebURL         string    `json:"web_url"`
	ParentIDs      []string  `json:"parent_ids"`
	AuthorEmail    string    `json:"author_email"`
	CommitterEmail string    `json:"committer_email"`
}

// Commits lists repository commits.
func (a *Adapter) Commits(ctx context.Context, query providers.CommitQuery) ([]providers.Commit, error) {
	values := pagination(query.Limit, query.Page)
	if query.RefName != "" {
		values.Set("ref_name", query.RefName)
	}
	if query.Path != "" {
		values.Set("path", query.Path)
	}
	var payload []commitPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/repository/commits", url.PathEscape(query.Ref.FullName())),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	commits := make([]providers.Commit, 0, len(payload))
	for _, item := range payload {
		message := item.Message
		if message == "" {
			message = item.Title
		}
		commits = append(commits, providers.Commit{
			SHA:         item.ID,
			Message:     message,
			AuthorName:  item.AuthorName,
			URL:         item.WebURL,
			Parents:     item.ParentIDs,
			CommittedAt: item.CommittedDate,
		})
	}
	return commits, nil
}

type treePayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// Tree lists a directory.
func (a *Adapter) Tree(ctx context.Context, query providers.TreeQuery) ([]providers.TreeEntry, error) {
	values := pagination(query.MaxEntries, 0)
	values.Set("path", strings.Trim(query.Path, "/"))
	if query.RefName != "" {
		values.Set("ref", query.RefName)
	}
	if query.Recursive {
		values.Set("recursive", "true")
	}
	var payload []treePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/repository/tree", url.PathEscape(query.Ref.FullName())),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	entries := make([]providers.TreeEntry, 0, len(payload))
	for _, item := range payload {
		kind := "file"
		if item.Type == "tree" {
			kind = "dir"
		}
		entries = append(entries, providers.TreeEntry{
			Path: item.Path,
			Name: item.Name,
			Type: kind,
			SHA:  item.ID,
		})
	}
	return entries, nil
}

type filePayload struct {
	FileName     string `json:"file_name"`
	FilePath     string `json:"file_path"`
	Size         int64  `json:"size"`
	Encoding     string `json:"encoding"`
	Content      string `json:"content"`
	ContentSHA   string `json:"content_sha256"`
	BlobID       string `json:"blob_id"`
	Ref          string `json:"ref"`
	LastCommitID string `json:"last_commit_id"`
}

// Blob reads a file content.
func (a *Adapter) Blob(ctx context.Context, query providers.BlobQuery) (*providers.Blob, error) {
	values := url.Values{}
	values.Set("file_path", strings.Trim(query.Path, "/"))
	if query.RefName != "" {
		values.Set("ref", query.RefName)
	}
	var payload filePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/repository/files/%s", url.PathEscape(query.Ref.FullName()), url.PathEscape(strings.Trim(query.Path, "/"))),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	return &providers.Blob{
		Path:     payload.FilePath,
		SHA:      payload.BlobID,
		Encoding: payload.Encoding,
		Size:     payload.Size,
		Content:  strings.ReplaceAll(payload.Content, "\n", ""),
	}, nil
}

type issuePayload struct {
	ID          int64  `json:"id"`
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	WebURL    string     `json:"web_url"`
	Labels    []string   `json:"labels"`
	UserNotes int        `json:"user_notes_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// Issues lists repository issues.
func (a *Adapter) Issues(ctx context.Context, query providers.ListQuery) ([]providers.Issue, error) {
	values := pagination(query.Limit, query.Page)
	if query.State != "" {
		values.Set("state", query.State)
	}
	var payload []issuePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/issues", url.PathEscape(query.Ref.FullName())),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	issues := make([]providers.Issue, 0, len(payload))
	for _, item := range payload {
		issue := providers.Issue{
			ID:          strconv.FormatInt(item.ID, 10),
			Number:      item.IID,
			Title:       item.Title,
			Body:        item.Description,
			State:       item.State,
			AuthorLogin: item.Author.Username,
			URL:         item.WebURL,
			Labels:      orEmpty(item.Labels),
			Comments:    item.UserNotes,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
		if item.ClosedAt != nil {
			issue.ClosedAt = *item.ClosedAt
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

type mergeRequestPayload struct {
	ID           int64  `json:"id"`
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	Draft        bool   `json:"draft"`
	WebURL       string `json:"web_url"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author       struct {
		Username string `json:"username"`
	} `json:"author"`
	UserNotesCount int        `json:"user_notes_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	MergedAt       *time.Time `json:"merged_at"`
}

// Reviews lists merge requests (the review workflow of the forge).
func (a *Adapter) Reviews(ctx context.Context, query providers.ListQuery) ([]providers.Review, error) {
	values := pagination(query.Limit, query.Page)
	if query.State != "" {
		values.Set("state", query.State)
	}
	var payload []mergeRequestPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/merge_requests", url.PathEscape(query.Ref.FullName())),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	reviews := make([]providers.Review, 0, len(payload))
	for _, item := range payload {
		review := providers.Review{
			ID:           strconv.FormatInt(item.ID, 10),
			Number:       item.IID,
			Title:        item.Title,
			Body:         item.Description,
			State:        item.State,
			AuthorLogin:  item.Author.Username,
			URL:          item.WebURL,
			SourceBranch: item.SourceBranch,
			TargetBranch: item.TargetBranch,
			Draft:        item.Draft,
			Comments:     item.UserNotesCount,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
		if item.MergedAt != nil {
			review.Merged = true
			review.MergedAt = *item.MergedAt
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

type releasePayload struct {
	TagName         string    `json:"tag_name"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	ReleasedAt      time.Time `json:"released_at"`
	UpcomingRelease bool      `json:"upcoming_release"`
	Links           struct {
		SelfURL string `json:"self_url"`
	} `json:"_links"`
}

// Releases lists repository releases.
func (a *Adapter) Releases(ctx context.Context, query providers.ListQuery) ([]providers.Release, error) {
	var payload []releasePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/projects/%s/releases", url.PathEscape(query.Ref.FullName())),
		Query: pagination(query.Limit, query.Page),
	}, &payload)
	if err != nil {
		return nil, err
	}
	releases := make([]providers.Release, 0, len(payload))
	for _, item := range payload {
		published := item.ReleasedAt
		if published.IsZero() {
			published = item.CreatedAt
		}
		releases = append(releases, providers.Release{
			ID:          item.TagName,
			TagName:     item.TagName,
			Name:        item.Name,
			Body:        item.Description,
			URL:         item.Links.SelfURL,
			Draft:       item.UpcomingRelease,
			CreatedAt:   item.CreatedAt,
			PublishedAt: published,
		})
	}
	return releases, nil
}

type groupPayload struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	FullPath    string `json:"full_path"`
	Description string `json:"description"`
	WebURL      string `json:"web_url"`
	AvatarURL   string `json:"avatar_url"`
}

// Organizations lists the GitLab groups visible to the credential.
func (a *Adapter) Organizations(ctx context.Context, query providers.OrganizationQuery) ([]providers.ExternalOrganization, error) {
	if query.Login != "" {
		var payload groupPayload
		_, err := a.client.Do(ctx, providers.Request{
			Path: fmt.Sprintf("/groups/%s", url.PathEscape(query.Login)),
		}, &payload)
		if err != nil {
			return nil, err
		}
		return []providers.ExternalOrganization{groupToExternal(payload)}, nil
	}
	var payload []groupPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  "/groups",
		Query: pagination(query.Limit, 0),
	}, &payload)
	if err != nil {
		return nil, err
	}
	organizations := make([]providers.ExternalOrganization, 0, len(payload))
	for _, item := range payload {
		organizations = append(organizations, groupToExternal(item))
	}
	return organizations, nil
}

func groupToExternal(group groupPayload) providers.ExternalOrganization {
	return providers.ExternalOrganization{
		Provider:    Name,
		ID:          strconv.FormatInt(group.ID, 10),
		Login:       group.FullPath,
		Name:        group.Name,
		Description: group.Description,
		URL:         group.WebURL,
		AvatarURL:   group.AvatarURL,
		Type:        "organization",
	}
}

func pagination(limit, page int) url.Values {
	values := url.Values{}
	if limit > 0 {
		values.Set("per_page", strconv.Itoa(clamp(limit, 1, 100)))
	}
	if page > 0 {
		values.Set("page", strconv.Itoa(page))
	}
	return values
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
