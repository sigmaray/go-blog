package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		respondHealth(c, http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		respondHealth(c, http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	respondHealth(c, http.StatusOK, gin.H{"status": "ok"})
}

func respondHealth(c *gin.Context, status int, body gin.H) {
	if c.Request.Method == http.MethodHead {
		c.Status(status)
		return
	}
	c.JSON(status, body)
}
