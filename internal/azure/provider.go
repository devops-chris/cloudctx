package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/devops-chris/cloudctx/internal/provider"
)

// Provider implements the cloud provider interface for Azure
type Provider struct {
	defaultLocation string
}

// Subscription represents an Azure subscription from az cli
type Subscription struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	State           string `json:"state"`
	IsDefault       bool   `json:"isDefault"`
	TenantID        string `json:"tenantId"`
	HomeTenantID    string `json:"homeTenantId"`
	ManagedByTenants []struct {
		TenantID string `json:"tenantId"`
	} `json:"managedByTenants"`
}

// Account represents the current Azure account
type Account struct {
	EnvironmentName string `json:"environmentName"`
	ID              string `json:"id"`
	IsDefault       bool   `json:"isDefault"`
	Name            string `json:"name"`
	State           string `json:"state"`
	TenantID        string `json:"tenantId"`
	User            struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"user"`
}

// NewProvider creates a new Azure provider
func NewProvider(defaultLocation string) *Provider {
	if defaultLocation == "" {
		defaultLocation = "eastus"
	}
	return &Provider{
		defaultLocation: defaultLocation,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "azure"
}

// Login performs Azure authentication
func (p *Provider) Login() error {
	// Check if az cli is installed
	if err := p.verifyAzureCLI(); err != nil {
		return err
	}

	// Disable Azure CLI's v2 login experience (built-in subscription picker)
	// We have our own prettier picker via 'ctx azure'
	_ = exec.Command("az", "config", "set", "core.login_experience_v2=off").Run()

	// Run az login with --output none to suppress JSON output
	// Only show stderr for browser instructions/errors
	cmd := exec.Command("az", "login", "--output", "none")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Sync is a no-op for Azure (subscriptions are always fetched live)
func (p *Provider) Sync() error {
	// Azure doesn't need sync - subscriptions are fetched live
	// Just verify we're logged in
	_, err := p.ListContexts()
	if err != nil {
		return fmt.Errorf("failed to list subscriptions (are you logged in?): %w", err)
	}
	return nil
}

// ListContexts returns all Azure subscriptions
func (p *Provider) ListContexts() ([]provider.Context, error) {
	// Run az account list
	cmd := exec.Command("az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	var contexts []provider.Context
	for _, sub := range subscriptions {
		if sub.State != "Enabled" {
			continue // Skip disabled subscriptions
		}
		contexts = append(contexts, provider.Context{
			Name:      sub.Name,
			Cloud:     "azure",
			AccountID: sub.ID,
			Region:    p.defaultLocation,
			Active:    sub.IsDefault,
			Managed:   true, // All Azure subscriptions are "managed" by Azure
		})
	}

	// Sort by name
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts, nil
}

// SetContext sets the active Azure subscription
func (p *Provider) SetContext(name string) error {
	// Find subscription by name or ID
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	var subscriptionID string
	for _, ctx := range contexts {
		if ctx.Name == name || ctx.AccountID == name {
			subscriptionID = ctx.AccountID
			break
		}
	}

	if subscriptionID == "" {
		return fmt.Errorf("subscription '%s' not found", name)
	}

	// Set the subscription
	cmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription: %w", err)
	}

	// Save to state file for quick lookup
	stateDir := p.stateDir()
	if err := os.MkdirAll(stateDir, 0755); err == nil {
		_ = os.WriteFile(filepath.Join(stateDir, "azure_current"), []byte(name), 0644)
	}

	return nil
}

// CurrentContext returns the currently active Azure subscription
func (p *Provider) CurrentContext() (*provider.Context, error) {
	cmd := exec.Command("az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil // Not logged in or no subscription set
	}

	var account Account
	if err := json.Unmarshal(output, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account: %w", err)
	}

	return &provider.Context{
		Name:      account.Name,
		Cloud:     "azure",
		AccountID: account.ID,
		Active:    true,
		Managed:   true,
	}, nil
}

// WhoAmI returns the current Azure identity
func (p *Provider) WhoAmI() (*provider.Identity, error) {
	cmd := exec.Command("az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("not logged in to Azure")
	}

	var account Account
	if err := json.Unmarshal(output, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account: %w", err)
	}

	return &provider.Identity{
		Cloud:       "azure",
		AccountID:   account.ID,
		AccountName: account.Name,
		UserID:      account.User.Name,
		ARN:         fmt.Sprintf("/subscriptions/%s", account.ID), // Azure resource path
		Region:      p.defaultLocation,
	}, nil
}

// Helper functions

func (p *Provider) stateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cloudctx")
}

func (p *Provider) verifyAzureCLI() error {
	cmd := exec.Command("az", "version", "--output", "none")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Azure CLI not found. Install it with: brew install azure-cli")
	}
	return nil
}

