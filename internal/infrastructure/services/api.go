package services

import (
	"mini-fargate/internal/infrastructure/docker"
	"mini-fargate/internal/infrastructure/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleTasksGin handles task creation requests using Gin context
func HandleTasksGin(c *gin.Context) {
	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := docker.RunContainer(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.TaskResponse{
		ID:     id,
		Status: "running",
	})
}

func HandleTasks(w http.ResponseWriter, r *http.Request) {
	// Keep for legacy or internal use if needed
}
