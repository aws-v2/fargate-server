package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, taskHandler *TaskHandler, docsHandler *DocsHandler) {
	v1 := r.Group("/api/v1/fargate")
	{
		v1.POST("/tasks", taskHandler.HandleTask)
		v1.GET("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
	}

	docs := r.Group("/api/v1/fargate/docs")
	{
		docs.GET("", docsHandler.GetPublicManifest)
		docs.GET("/:slug", docsHandler.GetPublicDoc)
	}

	internal := r.Group("/api/v1/fargate/internal/docs")
	{
		internal.GET("", docsHandler.GetInternalManifest)
		internal.GET("/:slug", docsHandler.GetInternalDoc)
	}
}