package handlers

import (
	"net/http"

	"go-blog/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LoginPage renders the login form, or redirects authenticated users to admin.
// c is the Gin request context used to read the session and write the response.
func (h *Handler) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user") != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}
	c.HTML(http.StatusOK, "admin/login.html", gin.H{})
}

// Login authenticates username/password, stores the user in the session, and
// redirects to the admin dashboard on success.
// c is the Gin request context carrying the login form POST body.
func (h *Handler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := models.FindUserByUsername(h.DB, username)
	if err != nil {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	if user == nil || !models.CheckPassword(user.PasswordHash, password) {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user", user.Username)
	if err := session.Save(); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create session")
		return
	}
	c.Redirect(http.StatusFound, "/admin/")
}

// Logout clears the session cookie and redirects to the public homepage.
// c is the Gin request context used to clear the session and redirect.
func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.String(http.StatusInternalServerError, "Failed to clear session")
		return
	}
	c.Redirect(http.StatusFound, "/")
}
