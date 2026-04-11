package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"mini-fargate/infrastructure/models"
	"mini-fargate/logger"

	"github.com/nats-io/nats.go"
)

type NATSHandler struct {
	nc      *nats.Conn
	profile string
}

func NewNATSHandler(url, username, password, profile string) (*NATSHandler, error) {
	opts := []nats.Option{
		nats.Name("Fargate Server"),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Log.Warn("NATS disconnected", slog.Any("error", err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Log.Info("NATS reconnected", slog.String("url", nc.ConnectedUrl()))
		}),
	}
	if username != "" && password != "" {
		opts = append(opts, nats.UserInfo(username, password))
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	return &NATSHandler{nc: nc, profile: profile}, nil
}

// BuildSubject constructs a dynamic NATS subject based on profile
func (h *NATSHandler) BuildSubject(service, version, domain, action string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", h.profile, service, version, domain, action)
}

func (h *NATSHandler) SubscribeTasks(runner func(inv models.NATSInvocation, callback func(status, msg string, result *models.NATSResponse)) (string, string, int, error)) error {
	// Pattern: <profile>.<service>.<version>.<domain>.<action>
	subject := h.BuildSubject("fargate", "v1", "tasks", "run")

	logger.Log.Info("Subscribing to NATS subject", slog.String("subject", subject))

	_, err := h.nc.Subscribe(subject, func(m *nats.Msg) {
		var inv models.NATSInvocation
		if err := json.Unmarshal(m.Data, &inv); err != nil {
			logger.Log.Error("Error unmarshaling NATS message", slog.Any("error", err))
			return
		}

		logger.Log.Info("NATS Received Task", slog.String("task_id", inv.TaskID))

		// Define callback for status updates
		statusCallback := func(status, msg string, result *models.NATSResponse) {
			update := models.NATSStatusUpdate{
				TaskID:  inv.TaskID,
				Status:  status,
				Message: msg,
				Result:  result,
			}
			data, _ := json.Marshal(update)

			// Publish status update to a dynamic status subject
			// e.g., staging.fargate.v1.tasks.status
			statusSubject := h.BuildSubject("fargate", "v1", "tasks", "status")
			h.nc.Publish(statusSubject, data)
			logger.Log.Debug("NATS Published Status",
				slog.String("status", status),
				slog.String("task_id", inv.TaskID),
				slog.String("subject", statusSubject),
			)
		}

		stdout, stderr, exitCode, err := runner(inv, statusCallback)

		respStatus := "success"
		executionResult := "Execution completed"
		if err != nil {
			respStatus = "error"
			executionResult = err.Error()
		} else if exitCode != 0 {
			respStatus = "error"
			executionResult = fmt.Sprintf("Process exited with code %d", exitCode)
		}

		resp := models.NATSResponse{
			TaskID:          inv.TaskID,
			Status:          respStatus,
			ExecutionResult: executionResult,
			Stdout:          stdout,
			Stderr:          stderr,
			ExitCode:        exitCode,
		}

		// Send final status update with the full result
		statusCallback("task.status.complete", "Task finished", &resp)

		respData, _ := json.Marshal(resp)
		m.Respond(respData)
		logger.Log.Info("NATS Sent Final Response",
			slog.String("task_id", inv.TaskID),
			slog.String("status", respStatus),
		)
	})
	return err
}

func (h *NATSHandler) Close() {
	h.nc.Close()
}
