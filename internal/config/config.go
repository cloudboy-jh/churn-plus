package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the merged global and project configuration
type Config struct {
	Global  *GlobalConfig  `json:"global"`
	Project *ProjectConfig `json:"project"`
}

// GlobalConfig is stored in ~/.churn/config.json
type GlobalConfig struct {
	APIKeys      APIKeys           `json:"api_keys"`
	DefaultModel ModelSelection    `json:"default_model"`
	Concurrency  ConcurrencyLimits `json:"concurrency"`
	Cache        CacheSettings     `json:"cache"`
	UI           UISettings        `json:"ui"`
}

// ProjectConfig is stored in .churn/config.json
type ProjectConfig struct {
	LastRun        time.Time       `json:"last_run,omitempty"`
	Model          ModelSelection  `json:"model,omitempty"`
	IgnorePatterns []string        `json:"ignore_patterns,omitempty"`
	CustomPasses   []string        `json:"custom_passes,omitempty"`
	Pipeline       *PipelineConfig `json:"pipeline,omitempty"`
}

// PipelineConfig defines the pipeline configuration
type PipelineConfig struct {
	Passes []PassConfig `json:"passes"`
}

// PassConfig defines configuration for a single pass
type PassConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Model       string `json:"model"`
	Provider    string `json:"provider"`
}

// APIKeys holds credentials for various LLM providers
type APIKeys struct {
	Anthropic string `json:"anthropic,omitempty"`
	OpenAI    string `json:"openai,omitempty"`
	Google    string `json:"google,omitempty"`
	// Ollama doesn't need API keys (local)
}

// ModelSelection specifies which model to use for each provider
type ModelSelection struct {
	Provider string `json:"provider"` // "anthropic", "openai", "google", "ollama"
	Model    string `json:"model"`    // e.g., "claude-3.5-sonnet", "gpt-4-turbo"
}

// ConcurrencyLimits controls parallel processing
type ConcurrencyLimits struct {
	Ollama    int `json:"ollama"`    // Default: 20
	OpenAI    int `json:"openai"`    // Default: 8
	Anthropic int `json:"anthropic"` // Default: 10
	Google    int `json:"google"`    // Default: 8
}

// CacheSettings controls caching behavior
type CacheSettings struct {
	Enabled bool `json:"enabled"`  // Default: true
	TTL     int  `json:"ttl"`      // Time-to-live in hours, default: 24
	MaxSize int  `json:"max_size"` // Max cache size in MB, default: 100
}

// UISettings controls UI behavior
type UISettings struct {
	ShowLineNumbers bool   `json:"show_line_numbers"` // Default: true
	SyntaxHighlight bool   `json:"syntax_highlight"`  // Default: true
	Theme           string `json:"theme"`             // Default: "default"
}

// Default configurations
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		APIKeys: APIKeys{},
		DefaultModel: ModelSelection{
			Provider: "anthropic",
			Model:    "claude-3.5-sonnet",
		},
		Concurrency: ConcurrencyLimits{
			Ollama:    20,
			OpenAI:    8,
			Anthropic: 10,
			Google:    8,
		},
		Cache: CacheSettings{
			Enabled: true,
			TTL:     24,
			MaxSize: 100,
		},
		UI: UISettings{
			ShowLineNumbers: true,
			SyntaxHighlight: true,
			Theme:           "default",
		},
	}
}

func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		IgnorePatterns: []string{
			"node_modules",
			".git",
			"dist",
			"build",
			"*.min.js",
			"*.map",
		},
		CustomPasses: []string{},
	}
}

// GetGlobalConfigPath returns ~/.churn/config.json
func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".churn", "config.json"), nil
}

// GetProjectConfigPath returns .churn/config.json in the given project root
func GetProjectConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".churn", "config.json")
}

// GetReportsDir returns .churn/reports/ directory path
func GetReportsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".churn", "reports")
}

// LoadGlobalConfig loads configuration from ~/.churn/config.json
func LoadGlobalConfig() (*GlobalConfig, error) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, create default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultGlobalConfig()
		if err := SaveGlobalConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to create default global config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	// Merge with defaults for any missing fields
	return mergeGlobalWithDefaults(&cfg), nil
}

// LoadProjectConfig loads configuration from .churn/config.json
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	path := GetProjectConfigPath(projectRoot)

	// If config doesn't exist, create default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultProjectConfig()
		if err := SaveProjectConfig(projectRoot, cfg); err != nil {
			return nil, fmt.Errorf("failed to create default project config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return mergeProjectWithDefaults(&cfg), nil
}

// SaveGlobalConfig writes global configuration to ~/.churn/config.json
func SaveGlobalConfig(cfg *GlobalConfig) error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveProjectConfig writes project configuration to .churn/config.json
func SaveProjectConfig(projectRoot string, cfg *ProjectConfig) error {
	path := GetProjectConfigPath(projectRoot)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Load loads and merges global and project configurations
func Load(projectRoot string) (*Config, error) {
	global, err := LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	project, err := LoadProjectConfig(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	// Override API keys from environment variables if present
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		global.APIKeys.Anthropic = key
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		global.APIKeys.OpenAI = key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		global.APIKeys.Google = key
	}

	return &Config{
		Global:  global,
		Project: project,
	}, nil
}

// GetAPIKey returns the API key for a given provider
func (c *Config) GetAPIKey(provider string) string {
	switch provider {
	case "anthropic":
		return c.Global.APIKeys.Anthropic
	case "openai":
		return c.Global.APIKeys.OpenAI
	case "google":
		return c.Global.APIKeys.Google
	default:
		return ""
	}
}

// GetConcurrencyLimit returns the concurrency limit for a provider
func (c *Config) GetConcurrencyLimit(provider string) int {
	switch provider {
	case "ollama":
		return c.Global.Concurrency.Ollama
	case "openai":
		return c.Global.Concurrency.OpenAI
	case "anthropic":
		return c.Global.Concurrency.Anthropic
	case "google":
		return c.Global.Concurrency.Google
	default:
		return 5
	}
}

// GetModelSelection returns the active model selection
// Project config overrides global config
func (c *Config) GetModelSelection() ModelSelection {
	if c.Project.Model.Provider != "" && c.Project.Model.Model != "" {
		return c.Project.Model
	}
	return c.Global.DefaultModel
}

// mergeGlobalWithDefaults fills in missing fields from defaults
func mergeGlobalWithDefaults(cfg *GlobalConfig) *GlobalConfig {
	defaults := DefaultGlobalConfig()

	if cfg.Concurrency.Ollama == 0 {
		cfg.Concurrency.Ollama = defaults.Concurrency.Ollama
	}
	if cfg.Concurrency.OpenAI == 0 {
		cfg.Concurrency.OpenAI = defaults.Concurrency.OpenAI
	}
	if cfg.Concurrency.Anthropic == 0 {
		cfg.Concurrency.Anthropic = defaults.Concurrency.Anthropic
	}
	if cfg.Concurrency.Google == 0 {
		cfg.Concurrency.Google = defaults.Concurrency.Google
	}

	if cfg.Cache.TTL == 0 {
		cfg.Cache.TTL = defaults.Cache.TTL
	}
	if cfg.Cache.MaxSize == 0 {
		cfg.Cache.MaxSize = defaults.Cache.MaxSize
	}

	if cfg.UI.Theme == "" {
		cfg.UI.Theme = defaults.UI.Theme
	}

	return cfg
}

// mergeProjectWithDefaults fills in missing fields from defaults
func mergeProjectWithDefaults(cfg *ProjectConfig) *ProjectConfig {
	defaults := DefaultProjectConfig()

	if len(cfg.IgnorePatterns) == 0 {
		cfg.IgnorePatterns = defaults.IgnorePatterns
	}

	if cfg.CustomPasses == nil {
		cfg.CustomPasses = defaults.CustomPasses
	}

	return cfg
}
