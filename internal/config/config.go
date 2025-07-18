package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIBaseURL string `json:"api_base_url"`
	AuthToken  string `json:"auth_token,omitempty"`
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
}

var globalConfig *Config

func InitConfig() {
	globalConfig = &Config{
		APIBaseURL: "https://modly.dashspace.dev",
	}
	loadConfig()
}

func GetConfig() *Config {
	return globalConfig
}

func SaveConfig() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func loadConfig() {
	configPath := filepath.Join(getConfigDir(), "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // Fichier n'existe pas, utiliser les valeurs par défaut
	}

	json.Unmarshal(data, globalConfig)
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dashspace")
}

func ClearAuth() {
	globalConfig.AuthToken = ""
	globalConfig.Username = ""
	globalConfig.Email = ""
	SaveConfig()
}
