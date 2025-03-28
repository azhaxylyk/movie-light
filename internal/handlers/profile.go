package handlers

import (
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProfilePage(c *gin.Context) {
	// Получаем сессионный токен из cookies
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login") // Перенаправляем на страницу входа, если токен отсутствует
		return
	}

	// Получаем данные пользователя по токену
	_, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.Redirect(http.StatusFound, "/login") // Перенаправляем на страницу входа, если токен недействителен
		return
	}

	// Передаем данные в шаблон
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"Username": username,
	})
}
