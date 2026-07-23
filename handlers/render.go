package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// HTML renders a page body template inside the matching layout shell.
// Admin content (names starting with "admin/") uses admin/layout.html.
// Public content (names starting with "public/") uses public/layout.html.
// c is the Gin request context used for the response.
// code is the HTTP status code to write.
// name is the content template name (for example "admin/dashboard.html").
// data is the template data; ContentTemplate is set automatically.
func (h *Handler) HTML(c *gin.Context, code int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["ContentTemplate"] = name

	layout := "public/layout.html"
	if strings.HasPrefix(name, "admin/") {
		layout = "admin/layout.html"
	}

	c.HTML(code, layout, data)
}
