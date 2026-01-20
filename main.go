package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mini-fargate/infrastructure/config"
	"mini-fargate/infrastructure/docker"
	"mini-fargate/infrastructure/events"
	"mini-fargate/infrastructure/models"
	"mini-fargate/logger"
	"mini-fargate/transport"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// EurekaConfig holds Eureka registration configuration
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

// getEurekaConfig reads Eureka configuration from environment variables
func getEurekaConfig() *EurekaConfig {
	return &EurekaConfig{
		ServerURL:         getEnv("EUREKA_SERVER_URL", "http://localhost:8761/eureka"),
		AppName:           getEnv("EUREKA_APP_NAME", "FARGATE-SERVICE"),
		HostName:          getEnv("EUREKA_HOSTNAME", "localhost"),
		IPAddr:            getEnv("EUREKA_IP_ADDR", "127.0.0.1"),
		Port:              getEnvInt("SERVER_PORT", 8099),
		VipAddress:        getEnv("EUREKA_VIP_ADDRESS", "fargate-service"),
		InstanceID:        getEnv("EUREKA_INSTANCE_ID", "fargate-service:8099"),
		HeartbeatInterval: getEnvDuration("EUREKA_HEARTBEAT_INTERVAL", 30*time.Second),
	}
}

// registerWithEureka registers the service instance with Eureka server
func registerWithEureka(config *EurekaConfig) error {
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

	log.Printf("✅ Successfully registered with Eureka server at %s", url)
	return nil
}

// sendHeartbeat sends periodic heartbeats to Eureka server
func sendHeartbeat(config *EurekaConfig) {
	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	url := fmt.Sprintf("%s/apps/%s/%s", config.ServerURL, config.AppName, config.InstanceID)
	client := &http.Client{Timeout: 5 * time.Second}

	for range ticker.C {
		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			log.Printf("❌ Failed to create heartbeat request: %v", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ Failed to send heartbeat to Eureka: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("⚠️  Heartbeat failed with status %d: %s", resp.StatusCode, string(body))
		} else {
			log.Printf("💓 Heartbeat sent successfully to Eureka")
		}

		resp.Body.Close()
	}
}

// deregisterFromEureka removes the service instance from Eureka
func deregisterFromEureka(config *EurekaConfig) error {
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

	log.Printf("✅ Successfully deregistered from Eureka server")
	return nil
}

func main() {
	logger.Init()
	defer logger.Log.Sync()

	// 0. Initialize Configuration and Profiles
	config.InitProfiles()

	// Optionally load from Spring Cloud Config if specified
	configURL := getEnv("CONFIG_SERVER_URL", "")
	if configURL != "" {
		if err := config.LoadConfig(configURL); err != nil {
			logger.Log.Warn("Failed to load external configuration", zap.Error(err))
		}
	}

	// Eureka Configuration
	eurekaConfig := getEurekaConfig()
	if err := registerWithEureka(eurekaConfig); err != nil {
		logger.Log.Error("Failed to register with Eureka", zap.Error(err))
	} else {
		go sendHeartbeat(eurekaConfig)
	}

	// Initialize NATS Handler
	natsHandler, err := events.NewNATSHandler(getEnv("NATS_URL", nats.DefaultURL))
	if err != nil {
		logger.Log.Fatal("Failed to initialize NATS", zap.Error(err))
	}
	defer natsHandler.Close()

	// Subscribe to tasks
	err = natsHandler.SubscribeTasks("tasks.run", func(inv models.NATSInvocation, callback func(status, msg string, result *models.NATSResponse)) (string, string, int, error) {
		timeout := 300 * 1000
		logger.Log.Info("Starting Task", zap.String("task_id", inv.TaskID), zap.Int("timeout_ms", timeout))
		if inv.TimeoutMS > 0 {
			timeout = inv.TimeoutMS
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()

		return docker.RunLambdaContainer(ctx, inv, callback)
	})

	if err != nil {
		logger.Log.Fatal("Failed to subscribe to tasks", zap.Error(err))
	}

	// Initialize Gin
	r := gin.Default()

	// Routes
	r.POST("/api/v1/fargate/tasks", func(c *gin.Context) {
		transport.HandleTasksGin(c)
	})
	r.GET("/api/v1/fargate/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8099",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("listen: ", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	// Deregister from Eureka
	if err := deregisterFromEureka(eurekaConfig); err != nil {
		logger.Log.Error("Failed to deregister from Eureka", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
