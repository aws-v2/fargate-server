package utils

import (
	"fmt"
	"log/slog"
	"mini-fargate/logger"
	"net"
	"os"
	"strconv"
	"time"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func CheckReachability(target string, attempts int, delay time.Duration) error {
	for i := 1; i <= attempts; i++ {
		logger.Log.Info("Checking reachability",
			slog.String("target", target),
			slog.Int("attempt", i),
			slog.Int("max", attempts),
		)

		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err == nil {
			conn.Close()
			logger.Log.Info("Target reachable", slog.String("target", target))
			return nil
		}

		if i < attempts {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("target %s unreachable after %d attempts", target, attempts)
}