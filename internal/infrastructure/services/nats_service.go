package services

import (
	"context"
	"log/slog"
	"mini-fargate/internal/infrastructure/docker"
	"mini-fargate/internal/infrastructure/events"
	"mini-fargate/internal/infrastructure/models"
	"mini-fargate/logger"
	"time"
)

func StartNATSSubscription(natsHandler *events.NATSHandler) error {
	return natsHandler.SubscribeTasks(func(inv models.NATSInvocation, callback func(status, msg string, result *models.NATSResponse)) (string, string, int, error) {
		timeout := 300 * 1000
		logger.Log.Info("Starting Task", slog.String("task_id", inv.TaskID), slog.Int("timeout_ms", timeout))
		if inv.TimeoutMS > 0 {
			timeout = inv.TimeoutMS
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()

		return docker.RunLambdaContainer(ctx, inv, callback)
	})
}