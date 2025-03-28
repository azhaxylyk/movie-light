package handlers

import (
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
