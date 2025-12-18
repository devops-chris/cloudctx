package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the cloudctx configuration
type Config struct {
	// DefaultCloud is the default cloud provider when none specified
	DefaultCloud string `mapstructure:"default_cloud"`

	// AWS configuration
	AWS AWSConfig `mapstructure:"aws"`

	// Azure configuration (future)
	// Azure AzureConfig `mapstructure:"azure"`
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	// SSOStartURL is the AWS SSO portal URL
	SSOStartURL string `mapstructure:"sso_start_url"`

	// SSORegion is the AWS SSO region
	SSORegion string `mapstructure:"sso_region"`

	// DefaultRegion is the default AWS region for profiles
	DefaultRegion string `mapstructure:"default_region"`
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

