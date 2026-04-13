package config

import (
	_ "embed"
	"encoding/json"
	"errors"
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

	if config.CurrentVersion == "" {
		return nil, errors.New("currentVersion is missing in config file")
	}

	if config.GitHubRepoOwner == "" {
		return nil, errors.New("gitHub repo owner is missing in config file")
	}

	if config.GitHubRepoName == "" {
		return nil, errors.New("gitHub repo name is missing in config file")
	}

	if config.DefaultAssumeRoleHCLVarName == "" {
		return nil, errors.New("default assume role var name is missing in config file")
	}

	if config.DefaultAWSRegion == "" {
		return nil, errors.New("default AWS region is missing in config file")
	}

	if config.DefaultAWSProfile == "" {
		return nil, errors.New("default AWS profile is missing in config file")
	}

	if config.TFCloudOrg == "" {
		return nil, errors.New("tfcloud org is missing in config file")
	}
	return &config, nil
}
