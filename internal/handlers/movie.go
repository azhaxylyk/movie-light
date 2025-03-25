package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"
	"strconv"

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

func AddDiscussion(c *gin.Context) {
	// Получаем userID из сессии
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	userID, _, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительная сессия"})
		return
	}

	movieID, err := strconv.Atoi(c.PostForm("movie_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID фильма"})
		return
	}

	discussion := c.PostForm("discussion")
	if discussion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст обсуждения не может быть пустым"})
		return
	}

	var parentID *int
	if pid := c.PostForm("parent_id"); pid != "" {
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный parent_id"})
			return
		}
		parentID = &pidInt
	}

	err = models.AddDiscussion(userID, movieID, discussion, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления обсуждения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Обсуждение добавлено"})
}

func AddReview(c *gin.Context) {
	// Получаем userID из сессии
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	userID, _, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительная сессия"})
		return
	}

	movieID, err := strconv.Atoi(c.PostForm("movie_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID фильма"})
		return
	}

	rating, err := strconv.Atoi(c.PostForm("rating"))
	if err != nil || rating < 1 || rating > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Рейтинг должен быть числом от 1 до 10"})
		return
	}

	review := c.PostForm("review")
	if review == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст отзыва не может быть пустым"})
		return
	}

	err = models.AddReview(userID, movieID, rating, review)
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
