// Package config provides configuration loading via koanf.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DBProvider   string `yaml:"db_provider"`
	DBConnection string `yaml:"db_connection"`
}

func defaultConfig(homeDir string) Config {
	return Config{
		DBProvider:   "sqlite",
		DBConnection: filepath.Join(homeDir, ".taskerr.db"),
	}
}

func LoadConfig() (*Config, error) {
	k := koanf.New(".")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error resolving home directory: %w", err)
	}
	configFilePath := filepath.Join(homeDir, ".taskerr")

	// Load from file if it exists
	if _, err := os.Stat(configFilePath); err == nil {
		if err := k.Load(file.Provider(configFilePath), kyaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading config file: %w", err)
		}
	} else if os.IsNotExist(err) {
		// Write default config if file doesn't exist
		if err := writeDefaultConfig(configFilePath, homeDir); err != nil {
			return nil, fmt.Errorf("error writing default config: %w", err)
		}
		if err := k.Load(file.Provider(configFilePath), kyaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading default config file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("error checking config file: %w", err)
	}

	// Override with environment variables
	if err := k.Load(env.Provider("TASKERR_", ".", func(s string) string {
		// Convert TASKERR_DB_PROVIDER -> db_provider
		return strings.ToLower(strings.TrimPrefix(s, "TASKERR_"))
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func writeDefaultConfig(path, homeDir string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	encoder.SetIndent(2)
	return encoder.Encode(defaultConfig(homeDir))
}
