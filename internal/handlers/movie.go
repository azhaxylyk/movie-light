package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MovieDetailPage отображает страницу с подробной информацией о фильме
func MovieDetailPage(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	loggedIn := false
	var username string
	userHasReviewed := false
	movieID := c.Param("id")
	if err == nil && sessionToken != "" {
		// Проверяем валидность токена
		userID, usernameFromDB, err := models.GetIDBySessionToken(sessionToken)
		if err == nil && userID != "" {
			loggedIn = true
			username = usernameFromDB
			if loggedIn {
				userHasReviewed = models.HasUserReviewedMovie(userID, movieID)
			}
		}
	}

	// Получаем данные о фильме с актерами
	movie, err := models.GetMovieDetails(movieID)
	if err != nil {
		log.Printf("Ошибка получения данных о фильме: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Не удалось загрузить информацию о фильме",
		})
		return
	}

	// Получаем похожие фильмы
	similarMovies, err := models.GetSimilarMovies(movieID)
	if err != nil {
		log.Printf("Ошибка получения похожих фильмов: %v", err)
		// Не прерываем выполнение, просто покажем без похожих фильмов
		similarMovies = []models.Movie{}
	}

	// Получаем отзывы
	reviews, err := models.GetReviewsByMovieID(movieID)
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		reviews = []models.Review{}
	}

	// Получаем обсуждения
	// Получаем обсуждения
	discussions, err := models.GetDiscussionsByMovieID(movieID) // <- Опечатка в имени переменной
	if err != nil {
		log.Printf("Ошибка получения обсуждений: %v", err)
		discussions = []models.Discussion{}
	}

	c.HTML(http.StatusOK, "movie.html", gin.H{
		"Movie":           movie,
		"SimilarMovies":   similarMovies,
		"Reviews":         reviews,
		"Discussions":     discussions,
		"LoggedIn":        loggedIn,
		"UserHasReviewed": userHasReviewed, // Передаем информацию о том, авторизован ли пользователь
		"Username":        username,
	})
}
