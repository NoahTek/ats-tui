package db

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", cfg.Port)
	}
	if cfg.Database != "ats" {
		t.Errorf("Expected database 'ats', got '%s'", cfg.Database)
	}

	uri := cfg.ConnectionURI()
	expected := "postgres://postgres@localhost:5432/ats?sslmode=disable"
	if uri != expected {
		t.Errorf("Expected URI %q, got %q", expected, uri)
	}
}

func TestConfigWithPassword(t *testing.T) {
	cfg := Config{
		Host:     "db.example.com",
		Port:     5433,
		User:     "myuser",
		Password: "secret-pass@123",
		Database: "myats",
		SSLMode:  "require",
	}

	uri := cfg.ConnectionURI()
	expected := "postgres://myuser:secret-pass%40123@db.example.com:5433/myats?sslmode=require"
	if uri != expected {
		t.Errorf("Expected URI %q, got %q", expected, uri)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ats-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override config dir temporarily for test
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	testCfg := Config{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "admin",
		Password: "password",
		Database: "ats_custom",
		SSLMode:  "disable",
		SeedDemo: true,
	}

	if err := SaveConfig(testCfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.Host != testCfg.Host || loaded.Port != testCfg.Port || loaded.Database != testCfg.Database {
		t.Errorf("Loaded config %+v does not match saved %+v", loaded, testCfg)
	}
}
