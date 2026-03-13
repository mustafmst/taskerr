package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigWritesDefaults(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.DBProvider != "sqlite" {
		t.Fatalf("DBProvider = %q, want sqlite", cfg.DBProvider)
	}

	wantDBPath := filepath.Join(tempHome, ".taskerr.db")
	if cfg.DBConnection != wantDBPath {
		t.Fatalf("DBConnection = %q, want %q", cfg.DBConnection, wantDBPath)
	}

	if _, err := os.Stat(filepath.Join(tempHome, ".taskerr")); err != nil {
		t.Fatalf("expected default config file to be created: %v", err)
	}
}

func TestLoadConfigAppliesEnvironmentOverrides(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("TASKERR_DB_PROVIDER", "postgres")
	t.Setenv("TASKERR_DB_CONNECTION", "postgres://taskerr:test@localhost/taskerr")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.DBProvider != "postgres" {
		t.Fatalf("DBProvider = %q, want postgres", cfg.DBProvider)
	}

	if cfg.DBConnection != "postgres://taskerr:test@localhost/taskerr" {
		t.Fatalf("DBConnection = %q, want env override", cfg.DBConnection)
	}
}
