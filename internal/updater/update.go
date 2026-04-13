package updater

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"platform-tool/internal/config"
	"platform-tool/internal/log"

	"github.com/creativeprojects/go-selfupdate"
)

func Update(version string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.FatalfColored("Failed to load config: %v", err)
	}

	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create new updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.NewRepositorySlug(cfg.GitHubRepoOwner, cfg.GitHubRepoName))
	if err != nil {
		return "", fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return "", fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		log.Printf("Current version (%s) is the latest", version)
		return version, nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return "", errors.New("could not locate executable path")
	}

	_, err = updater.UpdateCommand(context.Background(), exe, version, selfupdate.NewRepositorySlug(cfg.GitHubRepoOwner, cfg.GitHubRepoName))
	if err != nil {
		return "", fmt.Errorf("error occurred while updating binary: %w", err)
	}

	log.PrintfColored("Successfully updated to version %s", latest.Version())
	return latest.Version(), nil
}
