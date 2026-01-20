package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"mini-fargate/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

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
	// First, load local profiles/env if not already done
	InitProfiles()

	logger.Log.Info("Fetching configuration from", zap.String("url", url))

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
			// Convert Spring Cloud Config dot notation to environment variable notation (optional but common)
			// e.g., nats.url -> NATS_URL
			envKey := strings.ToUpper(strings.ReplaceAll(k, ".", "_"))
			envValue := fmt.Sprintf("%v", v)

			// Only set if not already set manually to allow local overrides
			if os.Getenv(envKey) == "" {
				os.Setenv(envKey, envValue)
				logger.Log.Debug("Set config from server", zap.String("key", envKey))
			}
		}
	}

	logger.Log.Info("Successfully loaded configuration from external server")
	return nil
}

// InitProfiles loads .env if present and maps environment variables based on APP_PROFILE.
// For example, if APP_PROFILE=DEV, then DEV_DATASOURCE_URL will be mapped to DATASOURCE_URL.
func InitProfiles() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Log.Debug("No .env file found, relying on system environment variables")
	}

	profile := os.Getenv("APP_PROFILE")
	if profile == "" {
		profile = "DEV"
		logger.Log.Info("No APP_PROFILE set, defaulting to DEV")
	}

	prefix := strings.ToUpper(profile) + "_"
	logger.Log.Info("Initializing environment profiles", zap.String("profile", profile))

	count := 0
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := pair[0]
		value := pair[1]

		if strings.HasPrefix(key, prefix) {
			newKey := strings.TrimPrefix(key, prefix)
			os.Setenv(newKey, value)
			count++
		}
	}

	if count > 0 {
		logger.Log.Info("Successfully mapped profile variables",
			zap.String("profile", profile),
			zap.Int("mapped_count", count))
	}
}
