// Package providers defines the forge abstraction used by the Code platform.
//
// The platform never talks to GitHub, GitLab or Giteria directly from the
// frontend: every external access goes through a small, capability-scoped
// adapter registered here. Adapters are intentionally independent so that a
// provider can expose a subset of the capabilities without forcing the others
// down to a lowest common denominator.
package providers

// Capability describes how a given feature is served for a provider.
type Capability string

const (
	// CapabilityNative means Code itself is the source of truth.
	CapabilityNative Capability = "native"
	// CapabilityIntegrated means Code reads (and sometimes writes) through the
	// provider API.
	CapabilityIntegrated Capability = "integrated"
	// CapabilityHybrid means part of the feature is Code-native and part is
	// delegated to the provider.
	CapabilityHybrid Capability = "hybrid"
	// CapabilityUnavailable means the provider does not expose the feature.
	CapabilityUnavailable Capability = "unavailable"
)

// Capability keys exposed through the API. They are stable identifiers that
// the frontend uses to decide which screens to render.
const (
	CapabilityRepositories  = "repositories"
	CapabilityBranches      = "branches"
	CapabilityCommits       = "commits"
	CapabilityFiles         = "files"
	CapabilityIssues        = "issues"
	CapabilityReviews       = "reviews"
	CapabilityReleases      = "releases"
	CapabilityOrganizations = "organizations"
	CapabilityDocumentation = "documentation"
	CapabilityAPIs          = "apis"
	CapabilitySDKs          = "sdks"
	CapabilityPackages      = "packages"
	CapabilityServices      = "services"
	CapabilityStandards     = "standards"
	CapabilityWebhooks      = "webhooks"
	CapabilityCICD          = "cicd"
)

// Descriptor is the public, safe description of a provider.
type Descriptor struct {
	Name             string                `json:"name"`
	DisplayName      string                `json:"displayName"`
	WebsiteURL       string                `json:"websiteUrl,omitempty"`
	DocumentationURL string                `json:"documentationUrl,omitempty"`
	Capabilities     map[string]Capability `json:"capabilities"`
}

// Capability returns the capability exposed for a feature key. Unknown keys
// are reported as unavailable instead of panicking.
func (d Descriptor) Capability(feature string) Capability {
	if d.Capabilities == nil {
		return CapabilityUnavailable
	}
	value, ok := d.Capabilities[feature]
	if !ok {
		return CapabilityUnavailable
	}
	return value
}

// Supports reports whether the feature is served in any way (native, hybrid or
// integrated) by the provider.
func (d Descriptor) Supports(feature string) bool {
	return d.Capability(feature) != CapabilityUnavailable
}
