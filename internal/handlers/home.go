package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
	movies, err := models.GetPopularMovies()
	if err != nil {
		log.Println("Ошибка загрузки фильмов:", err)
		c.HTML(http.StatusInternalServerError, "web/templates/home.html", gin.H{"Movies": nil})
		return
	}

	c.HTML(http.StatusOK, "home.html", gin.H{"Movies": movies})
}
