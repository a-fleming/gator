package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	CurrentUserID   string `json:"current_user_id"`
}

const configFileName string = ".gatorconfig.json"

func configFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: get user home dir: %w", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

func Read() (Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("config: read file %s: %w", configFilePath, err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("config: unmarshal json: %w", err)
	}
	return cfg, nil
}

func (c *Config) SetUser(userName string, userID string) error {
	c.CurrentUserName = userName
	c.CurrentUserID = userID
	err := write(*c)
	return err
}

func write(cfg Config) error {
	configPath, err := configFilePath()
	if err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal json: %w", err)
	}
	// 0600: user read.write only, typical for config with secrets
	err = os.WriteFile(configPath, jsonData, 0600)
	if err != nil {
		return err
	}
	return nil
}
