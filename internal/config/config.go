package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	CurrentUserID   string `json:"current_user_id"`
}

const configFileName string = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := fmt.Sprintf("%s/%s", homeDir, configFileName)
	return configPath, nil
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	var jsonData Config
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return Config{}, err
	}
	return jsonData, nil
}

func (c Config) SetUser(userName string, userID string) error {
	c.CurrentUserName = userName
	c.CurrentUserID = userID
	err := write(c)
	return err
}

func write(cfg Config) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, jsonData, os.FileMode(os.O_CREATE))
	if err != nil {
		return err
	}
	return nil
}
