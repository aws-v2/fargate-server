package http

import (
	"net/http"

	"mini-fargate/internal/infrastructure/models"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct{}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

func (h *TaskHandler) HandleTask(c *gin.Context) {
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// delegate to docker directly or inject a service if needed
	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}