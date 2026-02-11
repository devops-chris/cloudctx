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

// Provider implements the cloud provider interface for AWS.
// When orgKey is set, it operates on that org (session cloudctx-<orgKey>, profile prefix orgKey/).
// When orgKey is empty, legacy single-org behavior (session cloudctx-cli, no prefix).
type Provider struct {
	orgKey        string
	ssoStartURL   string
	ssoRegion     string
	defaultRegion string
}

// NewProvider creates a new AWS provider for one org.
// orgKey: org name for multi-org (e.g. "work"); empty for legacy single-org.
func NewProvider(orgKey, ssoStartURL, ssoRegion, defaultRegion string) *Provider {
	return &Provider{
		orgKey:        orgKey,
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
	sessionName := p.ssoSessionName()
	cmd := exec.Command("aws", "sso", "login", "--sso-session", sessionName)
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

	sectionName := "sso-session " + p.ssoSessionName()
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

	// Get SSO access token from cache (for this org's start URL)
	accessToken, err := p.getAccessToken(p.ssoStartURL)
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

	// Remove only cloudctx-managed profiles for this org (preserve others)
	for _, section := range awsCfg.Sections() {
		name := section.Name()
		if !strings.HasPrefix(name, "profile ") || !section.HasKey("cloudctx_managed") {
			continue
		}
		profileOrg := section.Key("cloudctx_org").String()
		if p.orgKey == "" {
			// Legacy: remove only profiles without cloudctx_org (old single-org)
			if profileOrg == "" {
				awsCfg.DeleteSection(name)
			}
		} else if profileOrg == p.orgKey {
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
			if p.orgKey != "" {
				_, _ = section.NewKey("cloudctx_org", p.orgKey)
			}
			_, _ = section.NewKey("sso_session", p.ssoSessionName())
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
			// [sso] vs [manual] label: show [sso] when this profile uses SSO (has sso_session or sso_account_id).
			// cloudctx_managed is only used for sync/rename (which profiles we overwrite); it does not affect this label.
			// So: SSO profile added by hand (no cloudctx_managed) still shows [sso]; key-based profile shows [manual].
			usesSSO := section.HasKey("sso_session") || section.HasKey("sso_account_id")
			profileMap[profileName] = provider.Context{
				Name:      profileName,
				Cloud:     "aws",
				Org:       section.Key("cloudctx_org").String(),
				AccountID: section.Key("sso_account_id").String(),
				Role:      section.Key("sso_role_name").String(),
				Region:    section.Key("region").String(),
				Active:    profileName == currentProfile,
				Managed:   usesSSO,
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
			// Org left empty for credentials-file profiles
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
		// Copy all settings from config profile to default (sso_session already exists from login/sync)
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

func (p *Provider) ssoSessionName() string {
	if p.orgKey == "" {
		return "cloudctx-cli"
	}
	return "cloudctx-" + p.orgKey
}

func (p *Provider) buildProfileName(accountName, roleName string) string {
	// Lowercase and combine: "My Account:AdminRole" -> "my-account:adminrole"
	name := strings.ToLower(accountName)
	name = strings.ReplaceAll(name, " ", "-")
	role := strings.ToLower(roleName)
	base := fmt.Sprintf("%s:%s", name, role)
	// Only prefix with org when multi-org (not for "default" or legacy empty orgKey)
	if p.orgKey != "" && p.orgKey != "default" {
		return p.orgKey + "/" + base
	}
	return base
}

func (p *Provider) getAccessToken(ssoStartURL string) (string, error) {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".aws", "sso", "cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", fmt.Errorf("SSO cache not found. Run 'cloudctx aws login' first")
	}

	// When we have a specific start URL, find the cache file that matches it (AWS CLI stores startUrl in cache).
	// Otherwise (legacy) use the newest token file.
	var bestToken string
	var bestTime int64

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
		if !strings.Contains(content, `"accessToken"`) {
			continue
		}

		// If we have a start URL, only use this file if it matches (AWS cache has "startUrl" or "startUrl" in JSON)
		if ssoStartURL != "" {
			if !strings.Contains(content, ssoStartURL) {
				continue
			}
		}

		token := extractAccessToken(content)
		if token == "" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		t := info.ModTime().Unix()
		if t > bestTime {
			bestTime = t
			bestToken = token
		}
	}

	if bestToken == "" {
		if ssoStartURL != "" {
			return "", fmt.Errorf("no SSO token found for %s. Run 'cloudctx aws login --org <org>' first", ssoStartURL)
		}
		return "", fmt.Errorf("no valid SSO access token found. Run 'cloudctx aws login' first")
	}

	return bestToken, nil
}

func extractAccessToken(content string) string {
	marker := `"accessToken"`
	tokenStart := strings.Index(content, marker)
	if tokenStart == -1 {
		return ""
	}
	afterMarker := content[tokenStart+len(marker):]
	colonPos := strings.Index(afterMarker, ":")
	if colonPos == -1 {
		return ""
	}
	afterColon := strings.TrimLeft(afterMarker[colonPos+1:], " \t\n")
	if len(afterColon) == 0 || afterColon[0] != '"' {
		return ""
	}
	valueStart := 1
	valueEnd := strings.Index(afterColon[valueStart:], `"`)
	if valueEnd <= 0 {
		return ""
	}
	return afterColon[valueStart : valueStart+valueEnd]
}

// RenameOrg renames an org in ~/.aws/config and state: profile names old/... -> new/..., cloudctx_org, sso_session, and sso-session section.
func RenameOrg(awsConfigPath, stateDir, oldName, newName string) error {
	if oldName == newName {
		return nil
	}
	awsCfg, err := ini.Load(awsConfigPath)
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}

	oldSession := "cloudctx-" + oldName
	newSession := "cloudctx-" + newName
	if oldName == "default" {
		oldSession = "cloudctx-cli" // legacy single-org uses this
	}
	// Also handle cloudctx-default if they had already converted to organizations
	oldSessions := []string{oldSession}
	if oldName == "default" {
		oldSessions = append(oldSessions, "cloudctx-default")
	}

	// Rename sso-session section(s)
	newSessSection := "sso-session " + newSession
	for _, os := range oldSessions {
		oldSessSection := "sso-session " + os
		if sec := awsCfg.Section(oldSessSection); sec != nil && len(sec.Keys()) > 0 {
			newSec, _ := awsCfg.NewSection(newSessSection)
			for _, k := range sec.Keys() {
				_, _ = newSec.NewKey(k.Name(), k.Value())
			}
			awsCfg.DeleteSection(oldSessSection)
			break
		}
	}

	// Profiles to rename: cloudctx_org == oldName, or (legacy default) oldName=="default" and sso_session is cloudctx-cli with cloudctx_managed
	var toRename []string
	for _, sec := range awsCfg.Sections() {
		name := sec.Name()
		if !strings.HasPrefix(name, "profile ") {
			continue
		}
		profileOrg := sec.Key("cloudctx_org").String()
		if profileOrg == oldName {
			toRename = append(toRename, name)
			continue
		}
		// Legacy single-org: no cloudctx_org, session cloudctx-cli; treat as "default" when renaming default
		if oldName == "default" && profileOrg == "" && sec.Key("cloudctx_managed").String() != "" && sec.Key("sso_session").String() == "cloudctx-cli" {
			toRename = append(toRename, name)
		}
	}

	renamed := make(map[string]string) // old profile name -> new profile name
	for _, sectionName := range toRename {
		sec := awsCfg.Section(sectionName)
		profileName := strings.TrimPrefix(sectionName, "profile ")
		var newProfileName string
		if strings.HasPrefix(profileName, oldName+"/") {
			newProfileName = newName + "/" + strings.TrimPrefix(profileName, oldName+"/")
		} else {
			newProfileName = newName + "/" + profileName
		}
		renamed[profileName] = newProfileName
		newSectionName := "profile " + newProfileName
		newSec, err := awsCfg.NewSection(newSectionName)
		if err != nil {
			continue
		}
		for _, k := range sec.Keys() {
			v := k.Value()
			if k.Name() == "cloudctx_org" {
				v = newName
			}
			if k.Name() == "sso_session" && (v == oldSession || v == "cloudctx-default" || v == "cloudctx-cli") {
				v = newSession
			}
			_, _ = newSec.NewKey(k.Name(), v)
		}
		awsCfg.DeleteSection(sectionName)
	}

	// Update [default] # cloudctx_current if it pointed to a renamed profile
	defaultSec := awsCfg.Section("default")
	if key := defaultSec.Key("# cloudctx_current"); key != nil {
		cur := key.Value()
		if newCur, ok := renamed[cur]; ok {
			key.SetValue(newCur)
		} else if strings.HasPrefix(cur, oldName+"/") {
			key.SetValue(newName + "/" + strings.TrimPrefix(cur, oldName+"/"))
		}
	}

	if err := awsCfg.SaveTo(awsConfigPath); err != nil {
		return fmt.Errorf("save AWS config: %w", err)
	}

	// State file
	stateFile := filepath.Join(stateDir, "aws_current")
	if data, err := os.ReadFile(stateFile); err == nil {
		current := strings.TrimSpace(string(data))
		if newCur, ok := renamed[current]; ok {
			_ = os.WriteFile(stateFile, []byte(newCur), 0644)
		} else if strings.HasPrefix(current, oldName+"/") {
			newCurrent := newName + "/" + strings.TrimPrefix(current, oldName+"/")
			_ = os.WriteFile(stateFile, []byte(newCurrent), 0644)
		}
	}

	// Remove old profile names from ~/.aws/credentials (AWS CLI caches SSO creds there by profile name; leftover = [manual] ghost)
	awsCredsPath := filepath.Join(filepath.Dir(awsConfigPath), "credentials")
	if creds, err := ini.Load(awsCredsPath); err == nil {
		for oldProfile := range renamed {
			creds.DeleteSection(oldProfile)
		}
		_ = creds.SaveTo(awsCredsPath)
	}

	return nil
}

// CleanStaleCredentials removes from ~/.aws/credentials any section that looks like a
// cloudctx profile (e.g. org/account:role) but no longer exists in config (e.g. after rename).
// This fixes "ghost" [manual] entries left by the AWS CLI's credential cache.
func CleanStaleCredentials(awsConfigPath, awsCredsPath string) (removed int, err error) {
	configProfiles := make(map[string]bool)
	if awsCfg, errLoad := ini.Load(awsConfigPath); errLoad == nil {
		for _, sec := range awsCfg.Sections() {
			name := sec.Name()
			if strings.HasPrefix(name, "profile ") {
				configProfiles[strings.TrimPrefix(name, "profile ")] = true
			}
		}
	}
	creds, err := ini.Load(awsCredsPath)
	if err != nil {
		return 0, fmt.Errorf("load credentials: %w", err)
	}
	for _, sec := range creds.Sections() {
		name := sec.Name()
		if name == "DEFAULT" || name == "default" {
			continue
		}
		// Only remove sections that look like org/profile and are not in config
		if strings.Contains(name, "/") && !configProfiles[name] {
			creds.DeleteSection(name)
			removed++
		}
	}
	if removed > 0 {
		if err := creds.SaveTo(awsCredsPath); err != nil {
			return removed, fmt.Errorf("save credentials: %w", err)
		}
	}
	return removed, nil
}

// RemoveOrg removes an org from ~/.aws/config and state: deletes its sso-session and all profiles for that org.
func RemoveOrg(awsConfigPath, stateDir, orgName string) error {
	awsCfg, err := ini.Load(awsConfigPath)
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}

	sessionName := "cloudctx-" + orgName
	if orgName == "default" {
		sessionName = "cloudctx-cli"
	}
	oldSessions := []string{sessionName}
	if orgName == "default" {
		oldSessions = append(oldSessions, "cloudctx-default")
	}

	// Delete sso-session section(s)
	for _, sess := range oldSessions {
		awsCfg.DeleteSection("sso-session " + sess)
	}

	// Delete all profiles that belong to this org
	var removedProfiles []string
	for _, sec := range awsCfg.Sections() {
		name := sec.Name()
		if !strings.HasPrefix(name, "profile ") {
			continue
		}
		profileOrg := sec.Key("cloudctx_org").String()
		if profileOrg == orgName {
			removedProfiles = append(removedProfiles, strings.TrimPrefix(name, "profile "))
			awsCfg.DeleteSection(name)
			continue
		}
		if orgName == "default" && profileOrg == "" && sec.Key("cloudctx_managed").String() != "" && sec.Key("sso_session").String() == "cloudctx-cli" {
			removedProfiles = append(removedProfiles, strings.TrimPrefix(name, "profile "))
			awsCfg.DeleteSection(name)
		}
	}

	removedSet := make(map[string]bool)
	for _, p := range removedProfiles {
		removedSet[p] = true
	}

	// If [default] # cloudctx_current was one of the removed profiles, clear it
	defaultSec := awsCfg.Section("default")
	if key := defaultSec.Key("# cloudctx_current"); key != nil {
		cur := key.Value()
		if removedSet[cur] || (orgName != "" && strings.HasPrefix(cur, orgName+"/")) {
			key.SetValue("")
		}
	}

	if err := awsCfg.SaveTo(awsConfigPath); err != nil {
		return fmt.Errorf("save AWS config: %w", err)
	}

	stateFile := filepath.Join(stateDir, "aws_current")
	if data, err := os.ReadFile(stateFile); err == nil {
		current := strings.TrimSpace(string(data))
		if removedSet[current] || (orgName != "" && strings.HasPrefix(current, orgName+"/")) {
			_ = os.WriteFile(stateFile, []byte(""), 0644)
		}
	}

	// Remove those profile names from ~/.aws/credentials (AWS CLI caches SSO creds there; leftover = [manual] ghost)
	awsCredsPath := filepath.Join(filepath.Dir(awsConfigPath), "credentials")
	if creds, err := ini.Load(awsCredsPath); err == nil {
		for _, p := range removedProfiles {
			creds.DeleteSection(p)
		}
		_ = creds.SaveTo(awsCredsPath)
	}

	return nil
}
