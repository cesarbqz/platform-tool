package tfcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"platform-tool/internal/log"
)

func GetTFCloudToken() (string, error) {
	if token, err := getTFCloudTokenFromCredentials(); err == nil {
		return token, nil
	}

	if token := os.Getenv("TFC_TOKEN"); token != "" {
		return token, nil
	}

	return "", errors.New("terraform Cloud token not found in credentials.tfrc.json or TFC_TOKEN environment variable")
}

func getTFCloudTokenFromCredentials() (string, error) {
	var credentialsPath string

	if os.PathSeparator == '\\' { // Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		credentialsPath = filepath.Join(appData, "terraform.d", "credentials.tfrc.json")
	} else { // Linux/macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		credentialsPath = filepath.Join(homeDir, ".terraform.d", "credentials.tfrc.json")
	}

	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("credentials file not found at %s", credentialsPath)
	}

	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	var credentials struct {
		Credentials map[string]struct {
			Token string `json:"token"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal(data, &credentials); err != nil {
		return "", fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if tfCloud, exists := credentials.Credentials["app.terraform.io"]; exists && tfCloud.Token != "" {
		log.PrintfColored("Terraform Cloud token found")
		return tfCloud.Token, nil
	}

	return "", errors.New("no token found in credentials file for app.terraform.io")
}
