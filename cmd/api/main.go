package main

import (
	"context"
	"fmt"
	"log/slog"
	"mini-fargate/internal/application"
	"mini-fargate/internal/infrastructure/config"
	"mini-fargate/internal/infrastructure/events"
	"mini-fargate/internal/infrastructure/services"
	"mini-fargate/internal/infrastructure/utils"
	transportHTTP "mini-fargate/internal/transport/http"
	"mini-fargate/logger"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load .env
	_ = godotenv.Load()

	// 1. Profile + Logger
	profile := config.GetProfile()
	logger.Init(profile)
	logger.Log.Info("Starting Fargate Server", slog.String("profile", profile))

	// 2. External config
	if configURL := utils.GetEnv("CONFIG_SERVER_URL", ""); configURL != "" {
		if err := config.LoadConfig(configURL); err != nil {
			logger.Log.Warn("Failed to load external config", slog.Any("error", err))
		}
	}

	// 3. NATS reachability check
	natsURLStr := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	natsPrefix :=utils.GetEnv("NATS_PREFIX","dev.v1")
	parsedURL, err := url.Parse(natsURLStr)
	if err != nil {
		logger.Log.Error("Failed to parse NATS_URL", slog.Any("error", err))
		os.Exit(1)
	}
	natsHost := parsedURL.Host
	if !strings.Contains(natsHost, ":") {
		natsHost += ":4222"
	}
	if err := utils.CheckReachability(natsHost, 5, 2*time.Second); err != nil {
		logger.Log.Error("NATS unreachable", slog.Any("error", err))
		os.Exit(1)
	}

	// 4. NATS handler
	natsHandler, err := events.NewNATSHandler(
		natsURLStr,
		utils.GetEnv("NATS_USERNAME", "auth-server"),
		utils.GetEnv("NATS_PASSWORD", "auth-secret"),
		profile,
		natsPrefix,
	)
	if err != nil {
		logger.Log.Error("Failed to initialize NATS", slog.Any("error", err))
		os.Exit(1)
	}
	defer natsHandler.Close()

	// 5. NATS subscription
	if err := services.StartNATSSubscription(natsHandler); err != nil {
		logger.Log.Error("Failed to subscribe to tasks", slog.Any("error", err))
		os.Exit(1)
	}

	// 6. Eureka
	eurekaConfig := utils.GetEurekaConfig()
	if err := utils.RegisterWithEureka(eurekaConfig); err != nil {
		logger.Log.Error("Failed to register with Eureka", slog.Any("error", err))
	} else {
		go utils.SendHeartbeat(eurekaConfig)
	}

	// 7. Handlers + Router
	gin.SetMode(gin.ReleaseMode)
	if profile == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	docsService := application.NewDocsService(utils.GetEnv("DOCS_PATH", "./docs"))
	docsHandler := transportHTTP.NewDocsHandler(docsService)
	taskHandler := transportHTTP.NewTaskHandler()

	r := gin.New()
	r.Use(gin.Recovery())
	transportHTTP.SetupRoutes(r, taskHandler, docsHandler)

	// 8. Start server
	port := utils.GetEnvInt("SERVER_PORT", 8086)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Log.Info("Listening", slog.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down...")

	if err := utils.DeregisterFromEureka(eurekaConfig); err != nil {
		logger.Log.Error("Failed to deregister from Eureka", slog.Any("error", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Forced shutdown", slog.Any("error", err))
	}
	logger.Log.Info("Server exited")
}

 