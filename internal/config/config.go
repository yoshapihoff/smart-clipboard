package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MaxHistorySize int           `yaml:"max_history_size"`
	CheckInterval  time.Duration `yaml:"check_interval_ms"`
	StoragePath    string        `yaml:"storage_path"`
	StartMinimized bool          `yaml:"start_minimized"`
}

func DefaultConfig() *Config {
	return &Config{
		MaxHistorySize: 100,
		CheckInterval:  1000 * time.Millisecond,
		StoragePath:    getDefaultStoragePath(),
		StartMinimized: false,
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
	return filepath.Join(configDir, "clipboard-history", "config.yaml")
}

func getDefaultStoragePath() string {
	dataDir, _ := os.UserDataDir()
	return filepath.Join(dataDir, "clipboard-history", "history.json")
}
