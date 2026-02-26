package events

import (
	"encoding/json"
	"fmt"
	"log"
	"mini-fargate/infrastructure/models"

	"github.com/nats-io/nats.go"
)

type NATSHandler struct {
	nc *nats.Conn
}

func NewNATSHandler(url, username, password string) (*NATSHandler, error) {
	opts := []nats.Option{}
	if username != "" && password != "" {
		opts = append(opts, nats.UserInfo(username, password))
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	return &NATSHandler{nc: nc}, nil
}

func (h *NATSHandler) SubscribeTasks(subject string, runner func(inv models.NATSInvocation, callback func(status, msg string, result *models.NATSResponse)) (string, string, int, error)) error {
	_, err := h.nc.Subscribe(subject, func(m *nats.Msg) {
		var inv models.NATSInvocation
		if err := json.Unmarshal(m.Data, &inv); err != nil {
			log.Printf("Error unmarshaling NATS message: %v", err)
			return
		}

		log.Printf("NATS Received Task: %s", inv.TaskID)

		// Define callback for status updates
		statusCallback := func(status, msg string, result *models.NATSResponse) {
			update := models.NATSStatusUpdate{
				TaskID:  inv.TaskID,
				Status:  status,
				Message: msg,
				Result:  result,
			}
			data, _ := json.Marshal(update)
			h.nc.Publish(status, data)
			log.Printf("NATS Published Status: %s for Task: %s", status, inv.TaskID)
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
		log.Printf("NATS Sent Final Response for Task: %s [Status: %s]", inv.TaskID, respStatus)
	})
	return err
}

func (h *NATSHandler) Close() {
	h.nc.Close()
}
