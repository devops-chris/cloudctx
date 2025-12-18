package aws

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/devops-chris/cloudctx/internal/provider"
	"gopkg.in/ini.v1"
)

// Provider implements the cloud provider interface for AWS
type Provider struct {
	ssoStartURL   string
	ssoRegion     string
	defaultRegion string
}

// NewProvider creates a new AWS provider
func NewProvider(ssoStartURL, ssoRegion, defaultRegion string) *Provider {
	return &Provider{
		ssoStartURL:   ssoStartURL,
		ssoRegion:     ssoRegion,
		defaultRegion: defaultRegion,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "aws"
}

// Login performs AWS SSO login
func (p *Provider) Login() error {
	// Check if AWS CLI is installed
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("AWS CLI not found. Please install AWS CLI v2: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
	}

	if p.ssoStartURL == "" {
		return fmt.Errorf("SSO start URL not configured. Run 'cloudctx aws init' first")
	}

	// Ensure we have an SSO session configured
	if err := p.ensureSSOSession(); err != nil {
		return fmt.Errorf("failed to configure SSO session: %w", err)
	}

	// Use AWS CLI for SSO login with our session
	cmd := exec.Command("aws", "sso", "login", "--sso-session", "cloudctx-cli")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ensureSSOSession creates an SSO session in ~/.aws/config
func (p *Provider) ensureSSOSession() error {
	awsConfigPath := p.awsConfigPath()
	
	// Ensure ~/.aws directory exists
	awsDir := filepath.Dir(awsConfigPath)
	if err := os.MkdirAll(awsDir, 0700); err != nil {
		return fmt.Errorf("failed to create AWS config directory: %w", err)
	}

	awsCfg, err := ini.Load(awsConfigPath)
	if err != nil {
		awsCfg = ini.Empty()
	}

	sectionName := "sso-session cloudctx-cli"
	section := awsCfg.Section(sectionName)

	// Clear and set SSO session settings
	for _, key := range section.Keys() {
		section.DeleteKey(key.Name())
	}

	_, _ = section.NewKey("sso_start_url", p.ssoStartURL)
	_, _ = section.NewKey("sso_region", p.ssoRegion)
	_, _ = section.NewKey("sso_registration_scopes", "sso:account:access")

	return awsCfg.SaveTo(awsConfigPath)
}

// Sync synchronizes profiles from AWS SSO
func (p *Provider) Sync() error {
	if p.ssoStartURL == "" {
		return fmt.Errorf("SSO start URL not configured. Run 'cloudctx aws init' first")
	}

	// Ensure SSO session exists (profiles will reference it)
	if err := p.ensureSSOSession(); err != nil {
		return fmt.Errorf("failed to configure SSO session: %w", err)
	}

	ctx := context.Background()

	// Get SSO access token from cache
	accessToken, err := p.getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get SSO access token (try 'cloudctx aws login' first): %w", err)
	}

	// Create SSO client
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(p.ssoRegion))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	ssoClient := sso.NewFromConfig(cfg)

	// List ALL accounts (with pagination)
	var allAccounts []ssotypes.AccountInfo
	var accountsNextToken *string
	for {
		accountsOutput, err := ssoClient.ListAccounts(ctx, &sso.ListAccountsInput{
			AccessToken: aws.String(accessToken),
			NextToken:   accountsNextToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list SSO accounts: %w", err)
		}
		allAccounts = append(allAccounts, accountsOutput.AccountList...)
		if accountsOutput.NextToken == nil {
			break
		}
		accountsNextToken = accountsOutput.NextToken
	}

	// Load existing AWS config
	awsConfigPath := p.awsConfigPath()
	awsCfg, err := ini.Load(awsConfigPath)
	if err != nil {
		// Create new if doesn't exist
		awsCfg = ini.Empty()
	}

	// Remove only cloudctx-managed profiles (preserve manually created ones)
	for _, section := range awsCfg.Sections() {
		name := section.Name()
		if strings.HasPrefix(name, "profile ") && section.HasKey("cloudctx_managed") {
			awsCfg.DeleteSection(name)
		}
	}

	// Generate profiles for each account/role (using sso_session reference)
	profileCount := 0
	for _, account := range allAccounts {
		// List ALL roles for this account (with pagination)
		var allRoles []ssotypes.RoleInfo
		var rolesNextToken *string
		for {
			rolesOutput, err := ssoClient.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
				AccessToken: aws.String(accessToken),
				AccountId:   account.AccountId,
				NextToken:   rolesNextToken,
			})
			if err != nil {
				break // Skip accounts we can't list roles for
			}
			allRoles = append(allRoles, rolesOutput.RoleList...)
			if rolesOutput.NextToken == nil {
				break
			}
			rolesNextToken = rolesOutput.NextToken
		}

		for _, role := range allRoles {
			profileName := p.buildProfileName(aws.ToString(account.AccountName), aws.ToString(role.RoleName))
			sectionName := fmt.Sprintf("profile %s", profileName)

			// Delete existing section first to avoid duplicates
			awsCfg.DeleteSection(sectionName)

			section, err := awsCfg.NewSection(sectionName)
			if err != nil {
				continue
			}

			_, _ = section.NewKey("cloudctx_managed", "true")
			_, _ = section.NewKey("sso_session", "cloudctx-cli")
			_, _ = section.NewKey("sso_account_id", aws.ToString(account.AccountId))
			_, _ = section.NewKey("sso_role_name", aws.ToString(role.RoleName))
			_, _ = section.NewKey("region", p.defaultRegion)
			_, _ = section.NewKey("output", "json")
			profileCount++
		}
	}

	// Save config
	return awsCfg.SaveTo(awsConfigPath)
}

// ListContexts returns all AWS profiles from both ~/.aws/config and ~/.aws/credentials
func (p *Provider) ListContexts() ([]provider.Context, error) {
	currentProfile := os.Getenv("AWS_PROFILE")
	profileMap := make(map[string]provider.Context) // Use map to dedupe

	// Read from ~/.aws/config (profiles use [profile name] format)
	awsConfigPath := p.awsConfigPath()
	if awsCfg, err := ini.Load(awsConfigPath); err == nil {
		for _, section := range awsCfg.Sections() {
			name := section.Name()
			if !strings.HasPrefix(name, "profile ") {
				continue
			}

			profileName := strings.TrimPrefix(name, "profile ")
			profileMap[profileName] = provider.Context{
				Name:      profileName,
				Cloud:     "aws",
				AccountID: section.Key("sso_account_id").String(),
				Role:      section.Key("sso_role_name").String(),
				Region:    section.Key("region").String(),
				Active:    profileName == currentProfile,
				Managed:   section.HasKey("cloudctx_managed"),
			}
		}
	}

	// Read from ~/.aws/credentials (profiles use [name] format, no "profile " prefix)
	awsCredsPath := p.awsCredentialsPath()
	if awsCreds, err := ini.Load(awsCredsPath); err == nil {
		for _, section := range awsCreds.Sections() {
			name := section.Name()
			// Skip DEFAULT section and any already in config
			if name == "DEFAULT" || name == "default" {
				continue
			}
			// Only add if not already in config (config takes precedence)
			if _, exists := profileMap[name]; !exists {
				profileMap[name] = provider.Context{
					Name:    name,
					Cloud:   "aws",
					Region:  section.Key("region").String(),
					Active:  name == currentProfile,
					Managed: false, // Credentials file profiles are always manual
				}
			}
		}
	}

	// Convert map to slice
	var contexts []provider.Context
	for _, ctx := range profileMap {
		contexts = append(contexts, ctx)
	}

	// Sort by name
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts, nil
}

// SetContext sets the active AWS profile by updating [default] in ~/.aws/config
// For credentials-file profiles, also updates [default] in ~/.aws/credentials
func (p *Provider) SetContext(name string) error {
	awsConfigPath := p.awsConfigPath()
	awsCfg, err := ini.Load(awsConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Check if profile exists in config file
	sourceSectionName := fmt.Sprintf("profile %s", name)
	sourceSection := awsCfg.Section(sourceSectionName)
	foundInConfig := sourceSection != nil && len(sourceSection.Keys()) > 0

	// Check if profile exists in credentials file
	var foundInCreds bool
	var credsSection *ini.Section
	awsCredsPath := p.awsCredentialsPath()
	awsCreds, credsErr := ini.Load(awsCredsPath)
	if credsErr == nil {
		credsSection = awsCreds.Section(name)
		foundInCreds = credsSection != nil && len(credsSection.Keys()) > 0
	}

	if !foundInConfig && !foundInCreds {
		return fmt.Errorf("profile '%s' not found in config or credentials", name)
	}

	// Delete and recreate default section in config to avoid stale keys
	awsCfg.DeleteSection("default")
	defaultConfigSection, err := awsCfg.NewSection("default")
	if err != nil {
		return fmt.Errorf("failed to create default section: %w", err)
	}

	if foundInConfig {
		// Ensure SSO session exists (only needed for SSO profiles)
		if err := p.ensureSSOSession(); err != nil {
			return fmt.Errorf("failed to configure SSO session: %w", err)
		}

		// Copy all settings from config profile to default
		for _, key := range sourceSection.Keys() {
			// Skip our internal marker
			if key.Name() == "cloudctx_managed" {
				continue
			}
			_, _ = defaultConfigSection.NewKey(key.Name(), key.Value())
		}

		// Clear any credentials from credentials file default (avoid conflict)
		if awsCreds != nil {
			awsCreds.DeleteSection("default")
			if defaultCredSection, err := awsCreds.NewSection("default"); err == nil {
				_, _ = defaultCredSection.NewKey("# cloudctx_managed", "true")
				_ = awsCreds.SaveTo(awsCredsPath)
			}
		}
	} else {
		// For credentials-file profiles, copy credentials to [default] in credentials file
		if awsCreds == nil {
			return fmt.Errorf("cannot load credentials file")
		}

		// Update credentials file [default] section
		awsCreds.DeleteSection("default")
		defaultCredSection, err := awsCreds.NewSection("default")
		if err != nil {
			return fmt.Errorf("failed to create default credentials section: %w", err)
		}

		// Copy credentials from source profile to default
		for _, key := range credsSection.Keys() {
			_, _ = defaultCredSection.NewKey(key.Name(), key.Value())
		}
		_, _ = defaultCredSection.NewKey("# cloudctx_source", name)

		if err := awsCreds.SaveTo(awsCredsPath); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		// Set region in config file default
		_, _ = defaultConfigSection.NewKey("region", p.defaultRegion)
	}

	// Mark which profile is current in config
	_, _ = defaultConfigSection.NewKey("# cloudctx_current", name)

	// Save config
	if err := awsCfg.SaveTo(awsConfigPath); err != nil {
		return fmt.Errorf("failed to save AWS config: %w", err)
	}

	// Also save to our state file for quick lookup
	stateDir := p.stateDir()
	if err := os.MkdirAll(stateDir, 0755); err == nil {
		_ = os.WriteFile(filepath.Join(stateDir, "aws_current"), []byte(name), 0644)
	}

	return nil
}

func (p *Provider) stateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cloudctx")
}

// CurrentContext returns the current AWS profile
func (p *Provider) CurrentContext() (*provider.Context, error) {
	// First check AWS_PROFILE env var (takes precedence)
	profile := os.Getenv("AWS_PROFILE")

	// If not set, check our state file
	if profile == "" {
		stateFile := filepath.Join(p.stateDir(), "aws_current")
		if data, err := os.ReadFile(stateFile); err == nil {
			profile = strings.TrimSpace(string(data))
		}
	}

	// If still not set, check the marker in [default] section
	if profile == "" {
		awsConfigPath := p.awsConfigPath()
		if awsCfg, err := ini.Load(awsConfigPath); err == nil {
			defaultSection := awsCfg.Section("default")
			if key := defaultSection.Key("# cloudctx_current"); key != nil {
				profile = key.Value()
			}
		}
	}

	if profile == "" {
		return nil, nil
	}

	contexts, err := p.ListContexts()
	if err != nil {
		return nil, err
	}

	for _, ctx := range contexts {
		if ctx.Name == profile {
			ctx.Active = true
			return &ctx, nil
		}
	}

	// Profile exists but not in our list
	return &provider.Context{
		Name:   profile,
		Cloud:  "aws",
		Active: true,
	}, nil
}

// WhoAmI returns the current AWS identity
func (p *Provider) WhoAmI() (*provider.Identity, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	return &provider.Identity{
		Cloud:     "aws",
		AccountID: aws.ToString(output.Account),
		UserID:    aws.ToString(output.UserId),
		ARN:       aws.ToString(output.Arn),
		Region:    cfg.Region,
	}, nil
}

// Helper functions

func (p *Provider) awsConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

func (p *Provider) awsCredentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}

func (p *Provider) buildProfileName(accountName, roleName string) string {
	// Lowercase and combine: "My Account:AdminRole"
	name := strings.ToLower(accountName)
	name = strings.ReplaceAll(name, " ", "-")
	role := strings.ToLower(roleName)
	return fmt.Sprintf("%s:%s", name, role)
}

func (p *Provider) getAccessToken() (string, error) {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".aws", "sso", "cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", fmt.Errorf("SSO cache not found. Run 'cloudctx aws login' first")
	}

	// Find the most recent cache file with accessToken
	var newestToken string
	var newestTime int64

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(cacheDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		content := string(data)

		// Look for accessToken in the JSON
		if !strings.Contains(content, `"accessToken"`) {
			continue
		}

		// Extract token using simple string search
		// Format is: "accessToken": "value" or "accessToken":"value"
		marker := `"accessToken"`
		tokenStart := strings.Index(content, marker)
		if tokenStart == -1 {
			continue
		}

		// Move past the marker
		afterMarker := content[tokenStart+len(marker):]

		// Find the colon and skip any whitespace
		colonPos := strings.Index(afterMarker, ":")
		if colonPos == -1 {
			continue
		}

		// Move past colon and whitespace to find the opening quote
		afterColon := strings.TrimLeft(afterMarker[colonPos+1:], " \t\n")
		if len(afterColon) == 0 || afterColon[0] != '"' {
			continue
		}

		// Find the closing quote
		valueStart := 1 // skip opening quote
		valueEnd := strings.Index(afterColon[valueStart:], `"`)
		if valueEnd <= 0 {
			continue
		}

		token := afterColon[valueStart : valueStart+valueEnd]

		// Check file modification time - use newest
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Unix() > newestTime {
			newestTime = info.ModTime().Unix()
			newestToken = token
		}
	}

	if newestToken == "" {
		return "", fmt.Errorf("no valid SSO access token found. Run 'cloudctx aws login' first")
	}

	return newestToken, nil
}

