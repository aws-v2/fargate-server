package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"mini-fargate/logger"

	"go.uber.org/zap"
)

type ConfigResponse struct {
	Name            string           `json:"name"`
	Profiles        []string         `json:"profiles"`
	PropertySources []PropertySource `json:"propertySources"`
}

type PropertySource struct {
	Name   string                 `json:"name"`
	Source map[string]interface{} `json:"source"`
}

// LoadConfig fetches configuration from Spring Cloud Config Server and sets them as environment variables
func LoadConfig(url string) error {
	logger.Log.Info("Fetching configuration from external server", zap.String("url", url))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config server returned status: %d", resp.StatusCode)
	}

	var configResp ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return fmt.Errorf("failed to decode config response: %w", err)
	}

	// Iterate through property sources and set environment variables
	// Higher priority sources come first in the slice
	for i := len(configResp.PropertySources) - 1; i >= 0; i-- {
		ps := configResp.PropertySources[i]
		for k, v := range ps.Source {
			val := fmt.Sprintf("%v", v)
			os.Setenv(k, val)
		}
	}

	logger.Log.Info("Configuration loaded successfully from external server")
	return nil
}
