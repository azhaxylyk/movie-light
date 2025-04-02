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
	var isFavorite bool
	var favoriteCount int
	movieID := c.Param("id")
	if err == nil && sessionToken != "" {
		// Проверяем валидность токена
		userID, usernameFromDB, err := models.GetIDBySessionToken(sessionToken)
		if err == nil && userID != "" {
			loggedIn = true
			username = usernameFromDB
			if loggedIn {
				userHasReviewed = models.HasUserReviewedMovie(userID, movieID)
				isFavorite, err = models.IsMovieInFavorites(userID, movieID)
				if err != nil {
					log.Printf("Ошибка проверки избранного: %v", err)
				}
			}
		}
	}

	favoriteCount, err = models.GetFavoriteCount(movieID)
	if err != nil {
		log.Printf("Ошибка получения количества лайков: %v", err)
		favoriteCount = 0
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
		"IsFavorite":      isFavorite,
		"FavoriteCount":   favoriteCount,
	})
}

// Добавление/удаление из избранного
func HandleFavorites(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	userID, _, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверная сессия"})
		return
	}

	var request struct {
		MovieID string `json:"movie_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	// Для POST - добавление, для DELETE - удаление
	var success bool
	if c.Request.Method == "POST" {
		err = models.AddToFavorites(userID, request.MovieID)
		success = err == nil
	} else {
		err = models.RemoveFromFavorites(userID, request.MovieID)
		success = err == nil
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	// Получаем обновленное количество лайков
	count, err := models.GetFavoriteCount(request.MovieID)
	if err != nil {
		count = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"count":   count,
	})
}
