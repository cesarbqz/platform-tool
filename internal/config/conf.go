package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Config struct {
	CurrentVersion              string `json:"version"`
	GitHubRepoOwner             string `json:"github_repo_owner"`
	GitHubRepoName              string `json:"github_repo_name"`
	DefaultAssumeRoleHCLVarName string `json:"hcl_assumerole_varname"`
	DefaultAWSRegion            string `json:"default_aws_region"`
	DefaultAWSProfile           string `json:"default_aws_profile"`
	TFCloudOrg                  string `json:"tf_cloud_org"`
}

//go:embed config.json
var embeddedConfig []byte

func LoadConfig() (*Config, error) {
	var config Config
	if err := json.Unmarshal(embeddedConfig, &config); err != nil {
		return nil, fmt.Errorf("error decoding embedded config file: %w", err)
	}

	// Apply sensible defaults for optional fields.
	// These can be overridden via flags at runtime.
	if config.DefaultAssumeRoleHCLVarName == "" {
		config.DefaultAssumeRoleHCLVarName = "workspace_iam_roles"
	}

	if config.DefaultAWSRegion == "" {
		config.DefaultAWSRegion = "us-east-1"
	}

	if config.DefaultAWSProfile == "" {
		config.DefaultAWSProfile = "default"
	}

	return &config, nil
}
