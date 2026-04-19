package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mini-fargate/logger"
	"net/http"
	"os"
	"strings"
)

// GetProfile returns the current application profile (default: dev)
func GetProfile() string {
	profile := os.Getenv("APP_PROFILE")
	if profile == "" {
		profile = "dev"
	}
	return strings.ToLower(profile)
}

// SpringCloudConfig represents the structure of the response from Spring Cloud Config server
type SpringCloudConfig struct {
	Name            string           `json:"name"`
	Profiles        []string         `json:"profiles"`
	Label           string           `json:"label"`
	Version         string           `json:"version"`
	State           interface{}      `json:"state"`
	PropertySources []PropertySource `json:"propertySources"`
}

type PropertySource struct {
	Name   string                 `json:"name"`
	Source map[string]interface{} `json:"source"`
}

// LoadConfig fetches configuration from Spring Cloud Config server and sets environment variables
func LoadConfig(url string) error {
	logger.Log.Info("Fetching configuration from server", slog.String("url", url))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config server returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read config response: %w", err)
	}

	var scConfig SpringCloudConfig
	if err := json.Unmarshal(body, &scConfig); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	for _, ps := range scConfig.PropertySources {
		for k, v := range ps.Source {
			// Convert Spring Cloud Config dot notation to environment variable notation
			envKey := strings.ToUpper(strings.ReplaceAll(k, ".", "_"))
			envValue := fmt.Sprintf("%v", v)

			// Only set if not already set manually to allow local overrides
			if os.Getenv(envKey) == "" {
				os.Setenv(envKey, envValue)
				logger.Log.Debug("Set config from server", slog.String("key", envKey))
			}
		}
	}

	logger.Log.Info("Successfully loaded configuration from external server")
	return nil
}
