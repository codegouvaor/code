// Package giteria implements the adapter for Giteria, the forge operated by
// the State. Its HTTP API is Gitea-compatible (GET /api/v1/repos/{owner}/{repo}
// and friends), so the adapter also works against any Gitea/Forgejo instance.
package giteria

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/providers"
)

// Name is the registry identifier of the adapter.
const Name = "giteria"

// DefaultBaseURL points at the public Giteria instance. It is always
// overridden by configuration in practice (self-hosted instances).
const DefaultBaseURL = "https://giteria.gouv.aor/api/v1"

// Descriptor describes what Giteria exposes through the platform.
var Descriptor = providers.Descriptor{
	Name:             Name,
	DisplayName:      "Giteria",
	WebsiteURL:       "https://giteria.gouv.aor",
	DocumentationURL: "https://giteria.gouv.aor/api/swagger",
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
		providers.CapabilityPackages:      providers.CapabilityIntegrated,
		// Documentation, APIs, SDKs, services and standards are Code-native:
		// Giteria only hosts the source code.
		providers.CapabilityDocumentation: providers.CapabilityHybrid,
		providers.CapabilityAPIs:          providers.CapabilityNative,
		providers.CapabilitySDKs:          providers.CapabilityNative,
		providers.CapabilityServices:      providers.CapabilityNative,
		providers.CapabilityStandards:     providers.CapabilityNative,
	},
}

// Factory builds a Giteria adapter for the registry.
func Factory(options providers.Options) (providers.Adapter, error) {
	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
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

// Adapter is the Giteria implementation of the provider capabilities.
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

type repositoryPayload struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	HTMLURL     string    `json:"html_url"`
	CloneURL    string    `json:"clone_url"`
	SSHURL      string    `json:"ssh_url"`
	Empty       bool      `json:"empty"`
	Private     bool      `json:"private"`
	Archived    bool      `json:"archived"`
	Fork        bool      `json:"fork"`
	Website     string    `json:"website"`
	Stars       int       `json:"stars_count"`
	Forks       int       `json:"forks_count"`
	OpenIssues  int       `json:"open_issues_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Owner       struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"owner"`
	// Gitea exposes the default branch as "default_branch" and, on the search
	// endpoints, as "default_branch"; topics live in a separate field.
	DefaultBranch string   `json:"default_branch"`
	Topics        []string `json:"topics"`
	Language      string   `json:"language"`
}

// Repository fetches the repository metadata.
func (a *Adapter) Repository(ctx context.Context, ref providers.RepositoryRef) (*providers.Repository, error) {
	var payload repositoryPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf("/repos/%s/%s", url.PathEscape(ref.Owner), url.PathEscape(ref.Repo)),
	}, &payload)
	if err != nil {
		return nil, err
	}
	visibility := "public"
	if payload.Private {
		visibility = "private"
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
		Visibility:    visibility,
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
		PushedAt:      payload.UpdatedAt,
	}, nil
}

type branchPayload struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Commit    struct {
		ID         string    `json:"id"`
		Timestamp  time.Time `json:"timestamp"`
		Committer  time.Time `json:"-"`
		AuthorName string    `json:"-"`
	} `json:"commit"`
}

// Branches lists the repository branches.
func (a *Adapter) Branches(ctx context.Context, query providers.ListQuery) ([]providers.Branch, error) {
	var repositoryDefault string
	if repository, err := a.Repository(ctx, query.Ref); err == nil {
		repositoryDefault = repository.DefaultBranch
	}
	values := pagination(query.Limit, query.Page)
	var payload []branchPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/repos/%s/%s/branches", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	branches := make([]providers.Branch, 0, len(payload))
	for _, item := range payload {
		branches = append(branches, providers.Branch{
			Name:      item.Name,
			Default:   repositoryDefault != "" && item.Name == repositoryDefault,
			Protected: item.Protected,
			CommitSHA: item.Commit.ID,
			UpdatedAt: item.Commit.Timestamp,
		})
	}
	return branches, nil
}

type commitPayload struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			When time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author  *userPayload `json:"author"`
	HTMLURL string       `json:"html_url"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

type userPayload struct {
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// Commits lists repository commits.
func (a *Adapter) Commits(ctx context.Context, query providers.CommitQuery) ([]providers.Commit, error) {
	values := pagination(query.Limit, query.Page)
	if query.RefName != "" {
		values.Set("sha", query.RefName)
	}
	if query.Path != "" {
		values.Set("path", query.Path)
	}
	var payload []commitPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/repos/%s/%s/commits", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query: values,
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
			CommittedAt: item.Commit.Author.When,
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

type contentPayload struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Type     string `json:"type"`
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
	var payload []contentPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf(
			"/repos/%s/%s/contents/%s",
			url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo), escapePath(strings.Trim(query.Path, "/")),
		),
		Query: url.Values{"ref": []string{refName}},
	}, &payload)
	if err != nil {
		return nil, err
	}
	entries := make([]providers.TreeEntry, 0, len(payload))
	for _, item := range payload {
		kind := "file"
		if item.Type == "dir" || item.Type == "submodule" {
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
	var payload contentPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path: fmt.Sprintf(
			"/repos/%s/%s/contents/%s",
			url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo), escapePath(strings.Trim(query.Path, "/")),
		),
		Query: url.Values{"ref": []string{refName}},
	}, &payload)
	if err != nil {
		return nil, err
	}
	if payload.Type == "dir" {
		return nil, providers.NewError(providers.ErrorInvalid, Name, "blob", "the requested path is a directory", nil)
	}
	content := payload.Content
	if payload.Encoding == "base64" {
		content = strings.ReplaceAll(strings.ReplaceAll(content, "\n", ""), "\r", "")
	}
	return &providers.Blob{
		Path:     payload.Path,
		SHA:      payload.SHA,
		Encoding: payload.Encoding,
		Size:     payload.Size,
		Content:  content,
	}, nil
}

// DecodeContent decodes a base64 payload returned by the Gitea-compatible API.
func DecodeContent(content, encoding string) (string, error) {
	if encoding != "base64" {
		return content, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.ReplaceAll(content, "\n", ""), "\r", ""))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

type issuePayload struct {
	ID          int64        `json:"id"`
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	State       string       `json:"state"`
	User        *userPayload `json:"user"`
	HTMLURL     string       `json:"html_url"`
	Comments    int          `json:"comments"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	ClosedAt    *time.Time   `json:"closed_at"`
	PullRequest *struct {
		Merged bool `json:"merged"`
	} `json:"pull_request"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// Issues lists repository issues (pull requests are excluded).
func (a *Adapter) Issues(ctx context.Context, query providers.ListQuery) ([]providers.Issue, error) {
	values := pagination(query.Limit, query.Page)
	values.Set("type", "issues")
	if query.State != "" {
		values.Set("state", query.State)
	}
	var payload []issuePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/repos/%s/%s/issues", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	issues := make([]providers.Issue, 0, len(payload))
	for _, item := range payload {
		if item.PullRequest != nil {
			continue
		}
		issue := providers.Issue{
			ID:          strconv.FormatInt(item.ID, 10),
			Number:      item.Number,
			Title:       item.Title,
			Body:        item.Body,
			State:       item.State,
			AuthorLogin: loginOf(item.User),
			URL:         item.HTMLURL,
			Labels:      labelNames(item.Labels),
			Comments:    item.Comments,
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

type pullPayload struct {
	ID        int64        `json:"id"`
	Number    int          `json:"number"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	State     string       `json:"state"`
	Draft     bool         `json:"draft"`
	User      *userPayload `json:"user"`
	HTMLURL   string       `json:"html_url"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	MergedAt  *time.Time   `json:"merged_at"`
	Merged    bool         `json:"merged"`
	Head      struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Comments int `json:"comments"`
}

// Reviews lists pull requests (the review workflow of the forge).
func (a *Adapter) Reviews(ctx context.Context, query providers.ListQuery) ([]providers.Review, error) {
	values := pagination(query.Limit, query.Page)
	if query.State != "" {
		values.Set("state", query.State)
	}
	var payload []pullPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/repos/%s/%s/pulls", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query: values,
	}, &payload)
	if err != nil {
		return nil, err
	}
	reviews := make([]providers.Review, 0, len(payload))
	for _, item := range payload {
		review := providers.Review{
			ID:           strconv.FormatInt(item.ID, 10),
			Number:       item.Number,
			Title:        item.Title,
			Body:         item.Body,
			State:        item.State,
			AuthorLogin:  loginOf(item.User),
			URL:          item.HTMLURL,
			SourceBranch: item.Head.Ref,
			TargetBranch: item.Base.Ref,
			Draft:        item.Draft,
			Merged:       item.Merged,
			Comments:     item.Comments,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
		if item.MergedAt != nil {
			review.MergedAt = *item.MergedAt
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

type releasePayload struct {
	ID          int64        `json:"id"`
	TagName     string       `json:"tag_name"`
	Name        string       `json:"name"`
	Body        string       `json:"body"`
	URL         string       `json:"url"`
	Draft       bool         `json:"draft"`
	Prerelease  bool         `json:"prerelease"`
	CreatedAt   time.Time    `json:"created_at"`
	PublishedAt time.Time    `json:"published_at"`
	Author      *userPayload `json:"author"`
	HTMLURL     string       `json:"html_url"`
}

// Releases lists repository releases.
func (a *Adapter) Releases(ctx context.Context, query providers.ListQuery) ([]providers.Release, error) {
	var payload []releasePayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  fmt.Sprintf("/repos/%s/%s/releases", url.PathEscape(query.Ref.Owner), url.PathEscape(query.Ref.Repo)),
		Query: pagination(query.Limit, query.Page),
	}, &payload)
	if err != nil {
		return nil, err
	}
	releases := make([]providers.Release, 0, len(payload))
	for _, item := range payload {
		releaseURL := item.HTMLURL
		if releaseURL == "" {
			releaseURL = item.URL
		}
		releases = append(releases, providers.Release{
			ID:          strconv.FormatInt(item.ID, 10),
			TagName:     item.TagName,
			Name:        item.Name,
			Body:        item.Body,
			URL:         releaseURL,
			AuthorLogin: loginOf(item.Author),
			Draft:       item.Draft,
			Prerelease:  item.Prerelease,
			CreatedAt:   item.CreatedAt,
			PublishedAt: item.PublishedAt,
		})
	}
	return releases, nil
}

type organizationPayload struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	AvatarURL   string `json:"avatar_url"`
	Location    string `json:"location"`
	Visibility  string `json:"visibility"`
}

// Organizations lists the organizations of the instance.
func (a *Adapter) Organizations(ctx context.Context, query providers.OrganizationQuery) ([]providers.ExternalOrganization, error) {
	if query.Login != "" {
		var payload organizationPayload
		_, err := a.client.Do(ctx, providers.Request{
			Path: fmt.Sprintf("/orgs/%s", url.PathEscape(query.Login)),
		}, &payload)
		if err != nil {
			return nil, err
		}
		return []providers.ExternalOrganization{organizationToExternal(payload, a.client.BaseURL())}, nil
	}
	var payload []organizationPayload
	_, err := a.client.Do(ctx, providers.Request{
		Path:  "/orgs",
		Query: pagination(query.Limit, 0),
	}, &payload)
	if err != nil {
		return nil, err
	}
	organizations := make([]providers.ExternalOrganization, 0, len(payload))
	for _, item := range payload {
		organizations = append(organizations, organizationToExternal(item, a.client.BaseURL()))
	}
	return organizations, nil
}

func organizationToExternal(organization organizationPayload, baseURL string) providers.ExternalOrganization {
	name := organization.Name
	if name == "" {
		name = organization.FullName
	}
	return providers.ExternalOrganization{
		Provider:    Name,
		ID:          strconv.FormatInt(organization.ID, 10),
		Login:       organization.Name,
		Name:        name,
		Description: organization.Description,
		URL:         strings.TrimSuffix(baseURL, "/api/v1") + "/" + organization.Name,
		AvatarURL:   organization.AvatarURL,
		Type:        "organization",
	}
}

func pagination(limit, page int) url.Values {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(clamp(limit, 1, 50)))
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

func loginOf(user *userPayload) string {
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
