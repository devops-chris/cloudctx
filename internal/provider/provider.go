// Package provider defines the interface for cloud providers
package provider

// Context represents a cloud context (AWS profile, Azure subscription, etc.)
type Context struct {
	Name        string
	Cloud       string // "aws", "azure", "gcp"
	AccountID   string
	AccountName string
	Role        string // AWS role, Azure role, etc.
	Region      string
	Active      bool
	Managed     bool   // true if created/managed by cloudctx (SSO sync)
}

// Identity represents the current authenticated identity
type Identity struct {
	Cloud       string
	AccountID   string
	AccountName string
	UserID      string
	ARN         string // AWS ARN or equivalent
	Region      string
}

// Provider defines the interface that all cloud providers must implement
type Provider interface {
	// Name returns the provider name (e.g., "aws", "azure")
	Name() string

	// Login performs authentication (e.g., SSO login)
	Login() error

	// Sync synchronizes available contexts from the cloud (e.g., fetch SSO accounts/roles)
	Sync() error

	// ListContexts returns all available contexts
	ListContexts() ([]Context, error)

	// SetContext sets the active context
	SetContext(name string) error

	// CurrentContext returns the currently active context
	CurrentContext() (*Context, error)

	// WhoAmI returns the current authenticated identity
	WhoAmI() (*Identity, error)
}

