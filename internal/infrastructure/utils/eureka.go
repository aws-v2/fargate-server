package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mini-fargate/logger"
	"net/http"
	"time"
)

type EurekaConfig struct {
	ServerURL         string
	AppName           string
	HostName          string
	IPAddr            string
	Port              int
	VipAddress        string
	InstanceID        string
	HeartbeatInterval time.Duration
}

func GetEurekaConfig() *EurekaConfig {
	return &EurekaConfig{
		ServerURL:         GetEnv("EUREKA_SERVER_URL", "http://localhost:8761/eureka"),
		AppName:           GetEnv("EUREKA_APP_NAME", "FARGATE-SERVICE"),
		HostName:          GetEnv("EUREKA_HOSTNAME", "localhost"),
		IPAddr:            GetEnv("EUREKA_IP_ADDR", "127.0.0.1"),
		Port:              GetEnvInt("SERVER_PORT", 8086),
		VipAddress:        GetEnv("EUREKA_VIP_ADDRESS", "fargate-service"),
		InstanceID:        GetEnv("EUREKA_INSTANCE_ID", "fargate-service:8086"),
		HeartbeatInterval: GetEnvDuration("EUREKA_HEARTBEAT_INTERVAL", 30*time.Second),
	}
}

func RegisterWithEureka(config *EurekaConfig) error {
	instance := map[string]interface{}{
		"instance": map[string]interface{}{
			"instanceId": config.InstanceID,
			"hostName":   config.HostName,
			"app":        config.AppName,
			"ipAddr":     config.IPAddr,
			"vipAddress": config.VipAddress,
			"status":     "UP",
			"port": map[string]interface{}{
				"$":        config.Port,
				"@enabled": "true",
			},
			"dataCenterInfo": map[string]interface{}{
				"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
				"name":   "MyOwn",
			},
			"healthCheckUrl": fmt.Sprintf("http://%s:%d/health", config.HostName, config.Port),
			"statusPageUrl":  fmt.Sprintf("http://%s:%d/health", config.HostName, config.Port),
			"homePageUrl":    fmt.Sprintf("http://%s:%d/", config.HostName, config.Port),
		},
	}

	jsonData, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal Eureka registration data: %w", err)
	}

	url := fmt.Sprintf("%s/apps/%s", config.ServerURL, config.AppName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register with Eureka: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("eureka registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Log.Info("Registered with Eureka", slog.String("url", url))
	return nil
}

func SendHeartbeat(config *EurekaConfig) {
	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	url := fmt.Sprintf("%s/apps/%s/%s", config.ServerURL, config.AppName, config.InstanceID)
	client := &http.Client{Timeout: 5 * time.Second}

	for range ticker.C {
		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			logger.Log.Error("Failed to create heartbeat request", slog.Any("error", err))
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("Failed to send heartbeat", slog.Any("error", err))
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			logger.Log.Warn("Heartbeat failed", slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
		} else {
			logger.Log.Debug("Heartbeat sent successfully")
		}
		resp.Body.Close()
	}
}

func DeregisterFromEureka(config *EurekaConfig) error {
	url := fmt.Sprintf("%s/apps/%s/%s", config.ServerURL, config.AppName, config.InstanceID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create deregistration request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deregister from Eureka: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deregistration failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Log.Info("Deregistered from Eureka")
	return nil
}