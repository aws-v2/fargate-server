package http

import (
	"mini-fargate/internal/application"

	"github.com/gin-gonic/gin"
)

type DocsHandler struct {
	service *application.DocsService
}

func NewDocsHandler(service *application.DocsService) *DocsHandler {
	return &DocsHandler{service: service}
}

func (h *DocsHandler) GetPublicManifest(c *gin.Context) {
	data, err := h.service.GetManifest(false)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *DocsHandler) GetInternalManifest(c *gin.Context) {
	data, err := h.service.GetManifest(true)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *DocsHandler) GetPublicDoc(c *gin.Context) {
	slug := c.Param("slug")
	doc, err := h.service.GetDoc(slug, false)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, gin.H{"data": doc})
}

func (h *DocsHandler) GetInternalDoc(c *gin.Context) {
	slug := c.Param("slug")
	doc, err := h.service.GetDoc(slug, true)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, gin.H{"data": doc})
}