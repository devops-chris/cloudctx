package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds the cloudctx configuration
type Config struct {
	// DefaultCloud is the default cloud provider when none specified
	DefaultCloud string `mapstructure:"default_cloud" yaml:"default_cloud"`

	// AWS configuration
	AWS AWSConfig `mapstructure:"aws" yaml:"aws"`

	// Azure configuration
	Azure AzureConfig `mapstructure:"azure" yaml:"azure"`
}

// AWSOrgConfig holds settings for one AWS organization (one SSO portal).
type AWSOrgConfig struct {
	SSOStartURL   string `mapstructure:"sso_start_url" yaml:"sso_start_url"`
	SSORegion     string `mapstructure:"sso_region" yaml:"sso_region"`
	DefaultRegion string `mapstructure:"default_region" yaml:"default_region"`
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	// Legacy single-org (used when organizations is not set)
	SSOStartURL   string `mapstructure:"sso_start_url" yaml:"sso_start_url,omitempty"`
	SSORegion     string `mapstructure:"sso_region" yaml:"sso_region,omitempty"`
	DefaultRegion string `mapstructure:"default_region" yaml:"default_region,omitempty"`

	// Multi-org: named organizations (key = org name, e.g. "work", "personal")
	Organizations map[string]AWSOrgConfig `mapstructure:"organizations" yaml:"organizations,omitempty"`

	// DefaultOrganization is the org used for login/sync when --org is not set
	DefaultOrganization string `mapstructure:"default_organization" yaml:"default_organization,omitempty"`
}

// AzureConfig holds Azure-specific configuration
type AzureConfig struct {
	// DefaultLocation is the default Azure location/region
	DefaultLocation string `mapstructure:"default_location" yaml:"default_location"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultCloud: "aws",
		AWS: AWSConfig{
			SSOStartURL:   "",
			SSORegion:     "us-east-1",
			DefaultRegion: "us-east-1",
		},
		Azure: AzureConfig{
			DefaultLocation: "eastus",
		},
	}
}

// Load loads configuration from file and environment
func Load(configFile string) *Config {
	cfg := DefaultConfig()

	v := viper.New()

	// Set defaults
	v.SetDefault("default_cloud", cfg.DefaultCloud)
	v.SetDefault("aws.sso_start_url", cfg.AWS.SSOStartURL)
	v.SetDefault("aws.sso_region", cfg.AWS.SSORegion)
	v.SetDefault("aws.default_region", cfg.AWS.DefaultRegion)
	v.SetDefault("azure.default_location", cfg.Azure.DefaultLocation)

	// Environment variables
	v.SetEnvPrefix("CLOUDCTX")
	v.AutomaticEnv()

	// Config file
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "cloudctx"))
			v.AddConfigPath(filepath.Join(home, ".cloudctx"))
		}
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read config file (ignore if not found)
	_ = v.ReadInConfig()

	// Unmarshal into struct
	_ = v.Unmarshal(cfg)

	return cfg
}

// ConfigDir returns the cloudctx config directory
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cloudctx"
	}
	return filepath.Join(home, ".config", "cloudctx")
}

// ConfigPath returns the default config file path (when no --config flag is set).
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// WriteConfig writes the configuration to path as YAML.
func WriteConfig(path string, c *Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AWSOrgs returns the effective map of org key -> org config.
// If organizations is set, returns it. Otherwise if legacy sso_start_url is set,
// returns a single org "default" with those settings.
func (c *Config) AWSOrgs() map[string]AWSOrgConfig {
	if len(c.AWS.Organizations) > 0 {
		return c.AWS.Organizations
	}
	if c.AWS.SSOStartURL != "" {
		return map[string]AWSOrgConfig{
			"default": {
				SSOStartURL:   c.AWS.SSOStartURL,
				SSORegion:     c.AWS.SSORegion,
				DefaultRegion: c.AWS.DefaultRegion,
			},
		}
	}
	return nil
}

// AWSDefaultOrg returns the default org key for login/sync when --org is not set.
// Returns "" if no orgs configured.
func (c *Config) AWSDefaultOrg() string {
	orgs := c.AWSOrgs()
	if len(orgs) == 0 {
		return ""
	}
	if c.AWS.DefaultOrganization != "" {
		if _, ok := orgs[c.AWS.DefaultOrganization]; ok {
			return c.AWS.DefaultOrganization
		}
	}
	// Single org: use its key; multi: prefer "default" if present, else first key
	if _, ok := orgs["default"]; ok {
		return "default"
	}
	for k := range orgs {
		return k
	}
	return ""
}

