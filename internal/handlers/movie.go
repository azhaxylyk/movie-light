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
	movieID := c.Param("id")
	fmt.Println(movieID)

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
	discussions, err := models.GetDiscussionsByMovieID(movieID)
	if err != nil {
		log.Printf("Ошибка получения обсуждений: %v", err)
		discussions = []models.Discussion{}
	}

	c.HTML(http.StatusOK, "movie.html", gin.H{
		"Movie":         movie,
		"SimilarMovies": similarMovies,
		"Reviews":       reviews,
		"Discussions":   discussions,
		"LoggedIn":      loggedIn, // Передаем информацию о том, авторизован ли пользователь
		"Username":      username,
	})
	fmt.Println(reviews)
}

func AddDiscussion(c *gin.Context) {
	var input struct {
		MovieID    int    `json:"movie_id"`
		Discussion string `json:"discussion"`
		ParentID   *int   `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
		return
	}

	userID := 1 // Заменить на `c.Get("userID")`, если есть аутентификация

	err := models.AddDiscussion(userID, input.MovieID, input.Discussion, input.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления обсуждения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Обсуждение добавлено"})
}

func AddReview(c *gin.Context) {
	var input struct {
		MovieID int    `json:"movie_id"`
		Rating  int    `json:"rating"`
		Review  string `json:"review"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
		return
	}

	// Заглушка: Получаем user_id (заменить на реальную аутентификацию)
	userID := 1 // Если у тебя есть аутентификация, используй c.Get("userID")

	if input.Rating < 1 || input.Rating > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Рейтинг должен быть от 1 до 10"})
		return
	}

	err := models.AddReview(userID, input.MovieID, input.Rating, input.Review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления отзыва"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Отзыв добавлен"})
}

func SearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	// Получаем параметры фильтров
	filters := map[string]string{
		"year":         c.Query("year"),
		"language":     c.Query("language"),
		"vote_average": c.Query("rating"),
		"with_genres":  c.Query("genre"),
	}

	movies, err := models.SearchMovies(query, filters)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Ошибка при поиске фильмов",
		})
		return
	}

	genres, err := models.GetGenres()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Ошибка при получении жанров",
		})
		return
	}

	// Проверяем аутентификацию
	loggedIn, username := checkAuth(c)

	c.HTML(http.StatusOK, "search.html", gin.H{
		"Title":          "Результаты поиска: " + query,
		"SearchQuery":    query,
		"Movies":         movies,
		"LoggedIn":       loggedIn,
		"Username":       username,
		"Genres":         genres,
		"CurrentFilters": filters,
	})
}

func checkAuth(c *gin.Context) (bool, string) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		return false, ""
	}

	_, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		return false, ""
	}

	return true, username
}
