// Package github implements the GitHub REST v3 adapter for the Code platform.
package github

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
const Name = "github"

const defaultBaseURL = "https://api.github.com"

// Descriptor describes what GitHub exposes through the platform.
var Descriptor = providers.Descriptor{
	Name:             Name,
	DisplayName:      "GitHub",
	WebsiteURL:       "https://github.com",
	DocumentationURL: "https://docs.github.com/rest",
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
		providers.CapabilityPackages:      providers.CapabilityUnavailable,
	},
}

// Factory builds a GitHub adapter for the registry.
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

// Adapter is the GitHub implementation of the provider capabilities.
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

// ── Repository ───────────────────────────────────────────────────────────────

type repositoryPayload struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	DefaultBranch string    `json:"default_branch"`
	Private       bool      `json:"private"`
	Visibility    string    `json:"visibility"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	Topics        []string  `json:"topics"`
	Stars         int       `json:"stargazers_count"`
	Forks         int       `json:"forks_count"`
	OpenIssues    int       `json:"open_issues_count"`
	Language      string    `json:"language"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
	Owner         struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"owner"`
}

// Repository fetches the repository metadata.
func (a *Adapter) Repository(ctx context.Context, ref providers.RepositoryRef) (*providers.Repository, error) {
	var payload repositoryPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s", url.PathEscape(ref.Owner), url.PathEscape(ref.Repo)),
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	return &providers.Repository{
		Provider:      Name,
		ID:            strconv.FormatInt(payload.ID, 10),
		Name:          payload.Name,
		FullName:      payload.FullName,
		Description:   payload.Description,
		URL:           payload.HTMLURL,
		CloneURL:      payload.CloneURL,
		SSHURL:        payload.SSHURL,
		DefaultBranch: payload.DefaultBranch,
		Visibility:    visibility(payload.Private, payload.Visibility),
		Archived:      payload.Archived,
		Fork:          payload.Fork,
		Topics:        orEmpty(payload.Topics),
		Stars:         payload.Stars,
		Forks:         payload.Forks,
		OpenIssues:    payload.OpenIssues,
		Language:      payload.Language,
		OwnerLogin:    payload.Owner.Login,
		OwnerType:     strings.ToLower(payload.Owner.Type),
		CreatedAt:     payload.CreatedAt,
		UpdatedAt:     payload.UpdatedAt,
		PushedAt:      payload.PushedAt,
	}, nil
}

// ── Branches ─────────────────────────────────────────────────────────────────

type branchPayload struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Commit    struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// Branches lists the repository branches.
func (a *Adapter) Branches(ctx context.Context, query providers.ListQuery) ([]providers.Branch, error) {
	var payload []branchPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s/branches", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query:  pagination(query.Limit, query.Page),
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	// GitHub does not flag the default branch on this endpoint: ask for the
	// repository metadata so the default flag is accurate.
	defaultBranch := ""
	if repository, repoErr := a.Repository(ctx, query.Ref); repoErr == nil {
		defaultBranch = repository.DefaultBranch
	}
	branches := make([]providers.Branch, 0, len(payload))
	for _, item := range payload {
		branches = append(branches, providers.Branch{
			Name:      item.Name,
			Default:   defaultBranch != "" && item.Name == defaultBranch,
			Protected: item.Protected,
			CommitSHA: item.Commit.SHA,
		})
	}
	return branches, nil
}

// ── Commits ──────────────────────────────────────────────────────────────────

type commitPayload struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
	HTMLURL string `json:"html_url"`
}

// Commits lists repository commits.
func (a *Adapter) Commits(ctx context.Context, query providers.CommitQuery) ([]providers.Commit, error) {
	listQuery := pagination(query.Limit, query.Page)
	if query.RefName != "" {
		listQuery.Set("sha", query.RefName)
	}
	if query.Path != "" {
		listQuery.Set("path", query.Path)
	}
	var payload []commitPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s/commits", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query:  listQuery,
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	commits := make([]providers.Commit, 0, len(payload))
	for _, item := range payload {
		commit := providers.Commit{
			SHA:         item.SHA,
			Message:     item.Commit.Message,
			AuthorName:  item.Commit.Author.Name,
			URL:         item.HTMLURL,
			CommittedAt: item.Commit.Author.Date,
		}
		if item.Author != nil {
			commit.AuthorLogin = item.Author.Login
		}
		for _, parent := range item.Parents {
			commit.Parents = append(commit.Parents, parent.SHA)
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// ── Files ────────────────────────────────────────────────────────────────────

type contentPayload struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// Tree lists a directory.
func (a *Adapter) Tree(ctx context.Context, query providers.TreeQuery) ([]providers.TreeEntry, error) {
	refName := query.RefName
	if refName == "" {
		refName = "HEAD"
	}
	path := strings.Trim(query.Path, "/")
	var payload []contentPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf(
			"/repos/%s/%s/contents/%s",
			url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo), escapePath(path),
		),
		Query:  url.Values{"ref": []string{refName}},
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	entries := make([]providers.TreeEntry, 0, len(payload))
	for _, item := range payload {
		kind := "file"
		if item.Type == "dir" {
			kind = "dir"
		}
		entries = append(entries, providers.TreeEntry{
			Path: item.Path,
			Name: item.Name,
			Type: kind,
			SHA:  item.SHA,
			Size: item.Size,
		})
	}
	return entries, nil
}

// Blob reads a file content.
func (a *Adapter) Blob(ctx context.Context, query providers.BlobQuery) (*providers.Blob, error) {
	refName := query.RefName
	if refName == "" {
		refName = "HEAD"
	}
	path := strings.Trim(query.Path, "/")
	var payload contentPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf(
			"/repos/%s/%s/contents/%s",
			url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo), escapePath(path),
		),
		Query:  url.Values{"ref": []string{refName}},
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	if payload.Type == "dir" {
		return nil, providers.NewError(providers.ErrorInvalid, Name, "blob", "the requested path is a directory", nil)
	}
	content := payload.Content
	binary := payload.Encoding == "" && payload.Content == ""
	if payload.Encoding == "base64" {
		content = strings.ReplaceAll(strings.ReplaceAll(payload.Content, "\n", ""), "\r", "")
	}
	return &providers.Blob{
		Path:     payload.Path,
		SHA:      payload.SHA,
		Encoding: payload.Encoding,
		Size:     payload.Size,
		Content:  content,
		Binary:   binary,
	}, nil
}

// ── Issues ───────────────────────────────────────────────────────────────────

type issuePayload struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	User   *struct {
		Login string `json:"login"`
	} `json:"user"`
	HTMLURL     string    `json:"html_url"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    time.Time `json:"closed_at"`
	PullRequest *struct{} `json:"pull_request"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// Issues lists repository issues (pull requests are excluded).
func (a *Adapter) Issues(ctx context.Context, query providers.ListQuery) ([]providers.Issue, error) {
	listQuery := pagination(query.Limit, query.Page)
	if query.State != "" {
		listQuery.Set("state", query.State)
	}
	var payload []issuePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s/issues", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query:  listQuery,
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	issues := make([]providers.Issue, 0, len(payload))
	for _, item := range payload {
		if item.PullRequest != nil {
			continue
		}
		issues = append(issues, providers.Issue{
			ID:          strconv.FormatInt(item.ID, 10),
			Number:      item.Number,
			Title:       item.Title,
			Body:        item.Body,
			State:       item.State,
			AuthorLogin: login(item.User),
			URL:         item.HTMLURL,
			Labels:      labelNames(item.Labels),
			Comments:    item.Comments,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			ClosedAt:    item.ClosedAt,
		})
	}
	return issues, nil
}

// ── Reviews (pull requests) ──────────────────────────────────────────────────

type pullPayload struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Draft  bool   `json:"draft"`
	User   *struct {
		Login string `json:"login"`
	} `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MergedAt  time.Time `json:"merged_at"`
	Merged    bool      `json:"merged"`
	Head      struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

// Reviews lists pull requests (the review workflow of the forge).
func (a *Adapter) Reviews(ctx context.Context, query providers.ListQuery) ([]providers.Review, error) {
	listQuery := pagination(query.Limit, query.Page)
	if query.State != "" {
		listQuery.Set("state", query.State)
	}
	var payload []pullPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s/pulls", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query:  listQuery,
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	reviews := make([]providers.Review, 0, len(payload))
	for _, item := range payload {
		reviews = append(reviews, providers.Review{
			ID:           strconv.FormatInt(item.ID, 10),
			Number:       item.Number,
			Title:        item.Title,
			Body:         item.Body,
			State:        item.State,
			AuthorLogin:  login(item.User),
			URL:          item.HTMLURL,
			SourceBranch: item.Head.Ref,
			TargetBranch: item.Base.Ref,
			Draft:        item.Draft,
			Merged:       item.Merged,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
			MergedAt:     item.MergedAt,
		})
	}
	return reviews, nil
}

// ── Releases ─────────────────────────────────────────────────────────────────

type releasePayload struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Author      *struct {
		Login string `json:"login"`
	} `json:"author"`
}

// Releases lists repository releases.
func (a *Adapter) Releases(ctx context.Context, query providers.ListQuery) ([]providers.Release, error) {
	var payload []releasePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   fmt.Sprintf("/repos/%s/%s/releases", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query:  pagination(query.Limit, query.Page),
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	releases := make([]providers.Release, 0, len(payload))
	for _, item := range payload {
		createdAt := item.CreatedAt
		if createdAt.IsZero() {
			createdAt = item.PublishedAt
		}
		releases = append(releases, providers.Release{
			ID:          strconv.FormatInt(item.ID, 10),
			TagName:     item.TagName,
			Name:        item.Name,
			Body:        item.Body,
			URL:         item.HTMLURL,
			AuthorLogin: login(item.Author),
			Draft:       item.Draft,
			Prerelease:  item.Prerelease,
			CreatedAt:   createdAt,
			PublishedAt: item.PublishedAt,
		})
	}
	return releases, nil
}

// ── Organizations ────────────────────────────────────────────────────────────

type organizationPayload struct {
	ID          int64  `json:"id"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	AvatarURL   string `json:"avatar_url"`
}

// Organizations lists the organizations visible to the credential.
func (a *Adapter) Organizations(ctx context.Context, query providers.OrganizationQuery) ([]providers.ExternalOrganization, error) {
	path := "/user/orgs"
	values := pagination(query.Limit, 0)
	if query.Login != "" {
		path = fmt.Sprintf("/orgs/%s", url.PathEscape(query.Login))
		values = nil
	}
	var payload []organizationPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:   path,
		Query:  values,
		Accept: "application/vnd.github+json",
	}, &payload)
	if err != nil {
		return nil, err
	}
	organizations := make([]providers.ExternalOrganization, 0, len(payload))
	for _, item := range payload {
		organizations = append(organizations, providers.ExternalOrganization{
			Provider:    Name,
			ID:          strconv.FormatInt(item.ID, 10),
			Login:       item.Login,
			Name:        item.Name,
			Description: item.Description,
			URL:         item.HTMLURL,
			AvatarURL:   item.AvatarURL,
			Type:        "organization",
		})
	}
	return organizations, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

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

func escapePath(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func login(user *struct {
	Login string `json:"login"`
}) string {
	if user == nil {
		return ""
	}
	return user.Login
}

func labelNames(labels []struct {
	Name string `json:"name"`
}) []string {
	if len(labels) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		out = append(out, label.Name)
	}
	return out
}

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func visibility(private bool, declared string) string {
	if declared != "" {
		return strings.ToLower(declared)
	}
	if private {
		return "private"
	}
	return "public"
}
