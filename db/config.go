package db

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// Config represents persistent database configuration for ATS-TUI.
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
	SeedDemo bool   `json:"seed_demo"`
}

// DefaultConfig provides standard local PostgreSQL defaults.
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Database: "ats",
		SSLMode:  "disable",
		SeedDemo: true,
	}
}

// ConnectionURI constructs a standard PostgreSQL connection URI.
func (c Config) ConnectionURI() string {
	portStr := strconv.Itoa(c.Port)
	if c.Port <= 0 {
		portStr = "5432"
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", c.Host, portStr),
		Path:   "/" + c.Database,
	}

	if c.User != "" {
		if c.Password != "" {
			u.User = url.UserPassword(c.User, c.Password)
		} else {
			u.User = url.User(c.User)
		}
	}

	q := u.Query()
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	} else {
		q.Set("sslmode", "disable")
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// GetConfigFilePath returns the platform-specific path to config.json.
func GetConfigFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ats-tui", "config.json"), nil
}

// LoadConfig reads the configuration file from disk if it exists.
func LoadConfig() (*Config, error) {
	path, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig persists the configuration to disk.
func SaveConfig(cfg Config) error {
	path, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}
