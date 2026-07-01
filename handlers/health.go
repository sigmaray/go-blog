package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
