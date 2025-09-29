package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MaxItems      int           `yaml:"max_items"`
	CheckInterval time.Duration `yaml:"check_interval_ms"`
	StoragePath   string        `yaml:"storage_path"`
	DebugMode     bool          `yaml:"debug_mode"`
}

func DefaultConfig() *Config {
	return &Config{
		MaxItems:      40,
		CheckInterval: 1000 * time.Millisecond,
		StoragePath:   getDefaultStoragePath(),
		DebugMode:     false,
	}
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	configPath := getConfigPath()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Создаем директорию если не существует
	os.MkdirAll(filepath.Dir(configPath), 0755)

	return os.WriteFile(configPath, data, 0644)
}

func getConfigPath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "smart-clipboard", "config.yaml")
}

func getDefaultStoragePath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "smart-clipboard", "history.json")
}
