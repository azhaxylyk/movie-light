package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
	// Получаем сессионный токен из cookies
	sessionToken, err := c.Cookie("session_token")
	loggedIn := false
	var username string

	if err == nil && sessionToken != "" {
		// Проверяем валидность токена
		userID, usernameFromDB, err := models.GetIDBySessionToken(sessionToken)
		if err == nil && userID != "" {
			loggedIn = true
			username = usernameFromDB
		}
	}

	// Получаем популярные фильмы
	movies, err := models.GetPopularMovies()
	if err != nil {
		log.Println("Ошибка загрузки популярных фильмов:", err)
		c.HTML(http.StatusInternalServerError, "web/templates/home.html", gin.H{"Movies": nil})
		return
	}

	// Получаем временной интервал для трендов (по умолчанию "day")
	timeWindow := c.DefaultQuery("timeWindow", "day")

	// Получаем трендовые фильмы
	trendingMovies, err := models.GetTrendingMovies(timeWindow)
	if err != nil {
		log.Println("Ошибка загрузки трендовых фильмов:", err)
		c.HTML(http.StatusInternalServerError, "web/templates/home.html", gin.H{"Movies": nil, "TrendingMovies": nil})
		return
	}

	// Передаем данные в шаблон
	c.HTML(http.StatusOK, "home.html", gin.H{
		"Movies":         movies,
		"TrendingMovies": trendingMovies,
		"LoggedIn":       loggedIn, // Передаем информацию о том, авторизован ли пользователь
		"Username":       username, // Передаем имя пользователя, если он авторизован
	})
}
