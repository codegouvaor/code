package routes

import (
	"net/http"
	"strings"

	"github.com/codegouvaor/code/server/src/services"
	"github.com/codegouvaor/code/server/src/utils"
	"github.com/gin-gonic/gin"
)

// repositoryEnvelope is the metadata attached to every repository response so
// that the frontend knows which provider answered and which features exist.
func (h *apiHandler) repositoryEnvelope(resolved *services.ResolvedRepository, source string) gin.H {
	meta := gin.H{
		"provider":         resolved.Provider,
		"capabilities":     h.deps.RepositoryService.Capabilities(resolved),
		"reference":        resolved.Project.Reference,
		"projectId":        resolved.Project.ID,
		"viewerPermission": string(resolved.Access.Role),
	}
	if source != "" {
		meta["source"] = source
	}
	return meta
}

func (h *apiHandler) getRepository(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	repository, err := h.deps.RepositoryService.Repository(c.Request.Context(), resolved)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, gin.H{
		"project":    resolved.Project.Reference,
		"binding":    resolved.Binding,
		"repository": repository,
	}, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) listRepositoryBranches(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	branches, err := h.deps.RepositoryService.Branches(c.Request.Context(), resolved, parsePageSize(c), parsePage(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, branches, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) listRepositoryCommits(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	commits, err := h.deps.RepositoryService.Commits(
		c.Request.Context(), resolved,
		strings.TrimSpace(c.Query("ref")), strings.TrimSpace(c.Query("path")),
		parsePageSize(c), parsePage(c),
	)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, commits, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) getRepositoryTree(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	entries, err := h.deps.RepositoryService.Tree(
		c.Request.Context(), resolved,
		strings.TrimSpace(c.Query("path")), strings.TrimSpace(c.Query("ref")),
	)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, gin.H{
		"path":    strings.TrimSpace(c.Query("path")),
		"ref":     strings.TrimSpace(c.Query("ref")),
		"entries": entries,
	}, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) getRepositoryBlob(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	blob, err := h.deps.RepositoryService.Blob(c.Request.Context(), resolved, path, strings.TrimSpace(c.Query("ref")))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, blob, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) listRepositoryIssues(c *gin.Context) {
	h.repositoryIssues(c, false)
}

func (h *apiHandler) listRepositoryReviews(c *gin.Context) {
	h.repositoryIssues(c, true)
}

func (h *apiHandler) repositoryIssues(c *gin.Context, reviews bool) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	state := strings.TrimSpace(c.Query("state"))
	page := parsePage(c)
	pageSize := parsePageSize(c)
	if reviews {
		items, source, listErr := h.deps.RepositoryService.Reviews(c.Request.Context(), resolved, state, pageSize, page)
		if listErr != nil {
			utils.Error(c, listErr)
			return
		}
		utils.SuccessWithMeta(c, http.StatusOK, items, h.repositoryEnvelope(resolved, source))
		return
	}
	items, source, listErr := h.deps.RepositoryService.Issues(c.Request.Context(), resolved, state, pageSize, page)
	if listErr != nil {
		utils.Error(c, listErr)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, items, h.repositoryEnvelope(resolved, source))
}

func (h *apiHandler) listRepositoryReleases(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	items, source, err := h.deps.RepositoryService.Releases(c.Request.Context(), resolved, parsePageSize(c), parsePage(c))
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, items, h.repositoryEnvelope(resolved, source))
}

func (h *apiHandler) getRepositoryLanguages(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	languages, err := h.deps.RepositoryService.Languages(c.Request.Context(), resolved)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, languages, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) getRepositoryContributions(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.Resolve(c.Request.Context(), principal, c.Param("owner"), c.Param("repo"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	days := 30
	if value := strings.TrimSpace(c.Query("days")); value != "" {
		if parsed, convErr := parsePositiveInt(value); convErr == nil {
			days = parsed
		}
	}
	contributions, err := h.deps.RepositoryService.Contributions(c.Request.Context(), resolved, days)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, contributions, h.repositoryEnvelope(resolved, "provider"))
}

// ── Project-scoped variants ──────────────────────────────────────────────────

func (h *apiHandler) getProjectRepository(c *gin.Context) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.ResolveProject(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	if resolved.Binding == nil {
		utils.Error(c, utils.ErrRepositoryBindingNotFound)
		return
	}
	repository, err := h.deps.RepositoryService.Repository(c.Request.Context(), resolved)
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.SuccessWithMeta(c, http.StatusOK, gin.H{
		"project":    resolved.Project.Reference,
		"binding":    resolved.Binding,
		"repository": repository,
	}, h.repositoryEnvelope(resolved, "provider"))
}

func (h *apiHandler) listProjectRepositoryTree(c *gin.Context) {
	h.projectRepositoryResource(c, "tree")
}

func (h *apiHandler) getProjectRepositoryBlob(c *gin.Context) {
	h.projectRepositoryResource(c, "blob")
}

func (h *apiHandler) listProjectRepositoryBranches(c *gin.Context) {
	h.projectRepositoryResource(c, "branches")
}

func (h *apiHandler) listProjectRepositoryCommits(c *gin.Context) {
	h.projectRepositoryResource(c, "commits")
}

func (h *apiHandler) listProjectRepositoryIssues(c *gin.Context) {
	h.projectRepositoryResource(c, "issues")
}

func (h *apiHandler) listProjectRepositoryReviews(c *gin.Context) {
	h.projectRepositoryResource(c, "reviews")
}

func (h *apiHandler) listProjectRepositoryReleases(c *gin.Context) {
	h.projectRepositoryResource(c, "releases")
}

// projectRepositoryResource implements every project-scoped repository read so
// that /projects/:id/repository/* mirrors the /repositories/:owner/:repo/* API.
func (h *apiHandler) projectRepositoryResource(c *gin.Context, resource string) {
	principal := h.ownerOptional(c)
	resolved, err := h.deps.RepositoryService.ResolveProject(c.Request.Context(), principal, c.Param("projectId"))
	if err != nil {
		utils.Error(c, err)
		return
	}
	if resolved.Binding == nil {
		utils.Error(c, utils.ErrRepositoryBindingNotFound)
		return
	}
	ctx := c.Request.Context()
	refName := strings.TrimSpace(c.Query("ref"))
	page := parsePage(c)
	pageSize := parsePageSize(c)
	var (
		payload any
		source  string
	)
	switch resource {
	case "tree":
		entries, treeErr := h.deps.RepositoryService.Tree(ctx, resolved, strings.TrimSpace(c.Query("path")), refName)
		if treeErr != nil {
			utils.Error(c, treeErr)
			return
		}
		payload = gin.H{"path": c.Query("path"), "ref": refName, "entries": entries}
	case "blob":
		path := strings.TrimSpace(c.Query("path"))
		if path == "" {
			utils.Error(c, utils.ErrValidationFailed)
			return
		}
		blob, blobErr := h.deps.RepositoryService.Blob(ctx, resolved, path, refName)
		if blobErr != nil {
			utils.Error(c, blobErr)
			return
		}
		payload = blob
	case "branches":
		branches, branchErr := h.deps.RepositoryService.Branches(ctx, resolved, pageSize, page)
		if branchErr != nil {
			utils.Error(c, branchErr)
			return
		}
		payload = branches
	case "commits":
		commits, commitErr := h.deps.RepositoryService.Commits(ctx, resolved, refName, strings.TrimSpace(c.Query("path")), pageSize, page)
		if commitErr != nil {
			utils.Error(c, commitErr)
			return
		}
		payload = commits
	case "issues":
		items, itemSource, listErr := h.deps.RepositoryService.Issues(ctx, resolved, strings.TrimSpace(c.Query("state")), pageSize, page)
		if listErr != nil {
			utils.Error(c, listErr)
			return
		}
		payload, source = items, itemSource
	case "reviews":
		items, itemSource, listErr := h.deps.RepositoryService.Reviews(ctx, resolved, strings.TrimSpace(c.Query("state")), pageSize, page)
		if listErr != nil {
			utils.Error(c, listErr)
			return
		}
		payload, source = items, itemSource
	case "releases":
		items, itemSource, listErr := h.deps.RepositoryService.Releases(ctx, resolved, pageSize, page)
		if listErr != nil {
			utils.Error(c, listErr)
			return
		}
		payload, source = items, itemSource
	}
	utils.SuccessWithMeta(c, http.StatusOK, payload, h.repositoryEnvelope(resolved, source))
}
