package providers

import (
	"context"
	"errors"
	"testing"
)

type stubAdapter struct {
	descriptor Descriptor
}

func (s stubAdapter) Name() string           { return s.descriptor.Name }
func (s stubAdapter) Descriptor() Descriptor { return s.descriptor }

type stubRepository struct{ stubAdapter }

func (s stubRepository) Repository(context.Context, RepositoryRef) (*Repository, error) {
	return &Repository{Name: "repo"}, nil
}

func testRegistry() *Registry {
	registry := NewRegistry()
	registry.Register(Descriptor{
		Name:        "alpha",
		DisplayName: "Alpha",
		Capabilities: map[string]Capability{
			CapabilityRepositories: CapabilityIntegrated,
			CapabilityIssues:       CapabilityUnavailable,
		},
	}, func(Options) (Adapter, error) {
		return stubRepository{stubAdapter{descriptor: Descriptor{Name: "alpha"}}}, nil
	})
	registry.Register(Descriptor{
		Name:        "beta",
		DisplayName: "Beta",
		Capabilities: map[string]Capability{
			CapabilityRepositories: CapabilityNative,
		},
	}, func(Options) (Adapter, error) {
		return stubAdapter{descriptor: Descriptor{Name: "beta"}}, nil
	})
	return registry
}

func TestRegistryListsProviders(t *testing.T) {
	t.Parallel()

	registry := testRegistry()
	descriptors := registry.Descriptors()
	if len(descriptors) != 2 {
		t.Fatalf("expected 2 descriptors, got %d", len(descriptors))
	}
	if descriptors[0].Name != "alpha" || descriptors[1].Name != "beta" {
		t.Fatalf("descriptors are not sorted: %+v", descriptors)
	}
}

func TestRegistryRejectsUnknownProvider(t *testing.T) {
	t.Parallel()

	_, err := testRegistry().NewAdapter("gamma", Options{})
	var providerErr *Error
	if !errors.As(err, &providerErr) || providerErr.Kind != ErrorUnsupported {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestRegistryCapabilitiesDefaultsToUnavailable(t *testing.T) {
	t.Parallel()

	registry := testRegistry()
	capabilities := registry.Capabilities("alpha")
	if capabilities[CapabilityIssues] != CapabilityUnavailable {
		t.Fatalf("expected issues to be unavailable: %+v", capabilities)
	}
	unknown := registry.Capabilities("unknown")
	if len(unknown) != 0 {
		t.Fatalf("expected an empty capability map for an unknown provider: %+v", unknown)
	}
}

func TestCapabilityAccessors(t *testing.T) {
	t.Parallel()

	registry := testRegistry()
	adapter, err := registry.NewAdapter("alpha", Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := RepositoryOf(adapter); err != nil {
		t.Fatalf("expected the repository capability: %v", err)
	}
	_, err = IssuesOf(adapter)
	if !IsUnsupported(err) {
		t.Fatalf("expected an unsupported capability error, got %v", err)
	}
	// An adapter that implements nothing still reports a clear error instead
	// of panicking or returning a nil provider.
	narrow, err := registry.NewAdapter("beta", Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := RepositoryOf(narrow); !IsUnsupported(err) {
		t.Fatalf("expected an unsupported capability error, got %v", err)
	}
}

func TestDescriptorSupports(t *testing.T) {
	t.Parallel()

	descriptor := Descriptor{Capabilities: map[string]Capability{CapabilityIssues: CapabilityHybrid}}
	if !descriptor.Supports(CapabilityIssues) {
		t.Fatal("expected hybrid to be supported")
	}
	if descriptor.Supports(CapabilityPackages) {
		t.Fatal("expected an unknown capability to be unsupported")
	}
}
