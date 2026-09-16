package providers

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Options carries everything an adapter needs to talk to a provider on behalf
// of a single connection. Credentials never leak into the domain models.
type Options struct {
	BaseURL   string
	Token     string
	UserAgent string
	Timeout   time.Duration
	// HTTPClient is injected in tests (and for shared connection pooling).
	HTTPClient *http.Client
	// MaxAttempts and BaseDelay tune the retry policy of the adapter.
	MaxAttempts int
	BaseDelay   time.Duration
}

// Factory builds an adapter for one provider.
type Factory func(options Options) (Adapter, error)

// Registry holds the known providers. It is the only place where the platform
// decides which adapter serves a binding, so adding a provider never requires
// touching handlers or services.
type Registry struct {
	mu          sync.RWMutex
	descriptors map[string]Descriptor
	factories   map[string]Factory
	order       []string
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		descriptors: map[string]Descriptor{},
		factories:   map[string]Factory{},
	}
}

// Register adds a provider. Registering the same name twice replaces the
// previous definition, which keeps tests and future overrides simple.
func (r *Registry) Register(descriptor Descriptor, factory Factory) {
	if descriptor.Name == "" {
		panic("providers: descriptor name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.descriptors[descriptor.Name]; !exists {
		r.order = append(r.order, descriptor.Name)
	}
	r.descriptors[descriptor.Name] = descriptor
	r.factories[descriptor.Name] = factory
}

// Names returns the registered provider names in registration order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// Descriptors returns every registered descriptor, sorted by name for stable
// API responses.
func (r *Registry) Descriptors() []Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Descriptor, 0, len(r.descriptors))
	for _, name := range r.order {
		out = append(out, r.descriptors[name])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Has reports whether the provider is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.descriptors[name]
	return ok
}

// Descriptor returns the descriptor of a registered provider.
func (r *Registry) Descriptor(name string) (Descriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	descriptor, ok := r.descriptors[name]
	return descriptor, ok
}

// Capabilities returns the capability map of a registered provider. Unknown
// providers report every feature as unavailable rather than failing, so the
// frontend can always render a degraded view.
func (r *Registry) Capabilities(name string) map[string]Capability {
	descriptor, ok := r.Descriptor(name)
	if !ok {
		return map[string]Capability{}
	}
	out := make(map[string]Capability, len(descriptor.Capabilities))
	for key, value := range descriptor.Capabilities {
		out[key] = value
	}
	return out
}

// NewAdapter instantiates the adapter of a provider.
func (r *Registry) NewAdapter(name string, options Options) (Adapter, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()
	if !ok || factory == nil {
		return nil, &Error{
			Kind:      ErrorUnsupported,
			Provider:  name,
			Operation: "registry.new_adapter",
			Message:   fmt.Sprintf("provider %q is not registered", name),
		}
	}
	adapter, err := factory(options)
	if err != nil {
		var providerErr *Error
		if asProviderError(err, &providerErr) {
			return nil, err
		}
		return nil, NewError(ErrorUnavailable, name, "registry.new_adapter", "the provider adapter could not be created", err)
	}
	return adapter, nil
}

func asProviderError(err error, target **Error) bool {
	providerErr, ok := err.(*Error)
	if ok {
		*target = providerErr
	}
	return ok
}

// ── Capability-scoped accessors ──────────────────────────────────────────────
//
// A missing capability is a normal situation: adapters implement only what the
// forge exposes. The accessors turn that into a typed ErrorUnsupported so that
// services can translate it into a 501 API error instead of a nil dereference.

func capabilityError(adapter Adapter, capability string) error {
	return &Error{
		Kind:      ErrorUnsupported,
		Provider:  adapter.Name(),
		Operation: capability,
		Message:   fmt.Sprintf("provider %q does not implement the %s capability", adapter.Name(), capability),
	}
}

// RepositoryOf returns the repository reader of an adapter.
func RepositoryOf(adapter Adapter) (RepositoryProvider, error) {
	provider, ok := adapter.(RepositoryProvider)
	if !ok {
		return nil, capabilityError(adapter, "repository")
	}
	return provider, nil
}

// BranchesOf returns the branch reader of an adapter.
func BranchesOf(adapter Adapter) (BranchProvider, error) {
	provider, ok := adapter.(BranchProvider)
	if !ok {
		return nil, capabilityError(adapter, "branches")
	}
	return provider, nil
}

// CommitsOf returns the commit reader of an adapter.
func CommitsOf(adapter Adapter) (CommitProvider, error) {
	provider, ok := adapter.(CommitProvider)
	if !ok {
		return nil, capabilityError(adapter, "commits")
	}
	return provider, nil
}

// FilesOf returns the file reader of an adapter.
func FilesOf(adapter Adapter) (FileProvider, error) {
	provider, ok := adapter.(FileProvider)
	if !ok {
		return nil, capabilityError(adapter, "files")
	}
	return provider, nil
}

// IssuesOf returns the issue reader of an adapter.
func IssuesOf(adapter Adapter) (IssueProvider, error) {
	provider, ok := adapter.(IssueProvider)
	if !ok {
		return nil, capabilityError(adapter, "issues")
	}
	return provider, nil
}

// ReviewsOf returns the review reader of an adapter.
func ReviewsOf(adapter Adapter) (ReviewProvider, error) {
	provider, ok := adapter.(ReviewProvider)
	if !ok {
		return nil, capabilityError(adapter, "reviews")
	}
	return provider, nil
}

// ReleasesOf returns the release reader of an adapter.
func ReleasesOf(adapter Adapter) (ReleaseProvider, error) {
	provider, ok := adapter.(ReleaseProvider)
	if !ok {
		return nil, capabilityError(adapter, "releases")
	}
	return provider, nil
}

// OrganizationsOf returns the organization reader of an adapter.
func OrganizationsOf(adapter Adapter) (OrganizationProvider, error) {
	provider, ok := adapter.(OrganizationProvider)
	if !ok {
		return nil, capabilityError(adapter, "organizations")
	}
	return provider, nil
}
