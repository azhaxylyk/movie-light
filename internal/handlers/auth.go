package handlers

import (
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		email := c.PostForm("email")
		username := c.PostForm("username")
		password := c.PostForm("password")
		moderatorRequest := c.PostForm("moderator_request") == "on"

		sessionToken, err := models.RegisterUser(email, username, password, moderatorRequest)
		if err != nil {
			c.HTML(http.StatusOK, "register.html", gin.H{"Error": "Registration failed"})
			return
		}

		c.SetCookie("session_token", sessionToken, 86400, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "register.html", nil)
}

func LoginHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		email := c.PostForm("email")
		password := c.PostForm("password")

		sessionToken, err := models.AuthenticateUser(email, password)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid email or password"})
			return
		}

		c.SetCookie("session_token", sessionToken, 86400, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func LogoutHandler(c *gin.Context) {
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}
