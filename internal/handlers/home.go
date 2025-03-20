package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
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
	})
}
