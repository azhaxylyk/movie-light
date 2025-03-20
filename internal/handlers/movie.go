package handlers

import (
	"fmt"
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MovieDetailPage отображает страницу с подробной информацией о фильме
func MovieDetailPage(c *gin.Context) {
	movieID := c.Param("id") // Получаем ID фильма из URL

	// Получаем подробную информацию о фильме
	movie, err := models.GetMovieDetails(movieID)
	fmt.Println(movie)
	if err != nil {
		log.Println("Ошибка получения данных о фильме:", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Ошибка получения данных о фильме"})
		return
	}

	// Получаем похожие фильмы
	similarMovies, err := models.GetSimilarMovies(movieID)
	if err != nil {
		log.Println("Ошибка получения похожих фильмов:", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Ошибка получения похожих фильмов"})
		return
	}

	// Передаем данные в шаблон
	c.HTML(http.StatusOK, "movie.html", gin.H{
		"Movie":         movie,
		"SimilarMovies": similarMovies,
	})
}
