package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatBotPage отображает страницу чат-бота
func ChatBotPage(c *gin.Context) {
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

	// Передаем данные в шаблон
	c.HTML(http.StatusOK, "ai.html", gin.H{
		"LoggedIn": loggedIn,
		"Username": username,
	})
}

func ChatBotHandler(c *gin.Context) {
	var request struct {
		Query string `json:"query"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("Ошибка парсинга запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
		return
	}

	query := request.Query
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пустой запрос"})
		return
	}

	genreIDs, hint := models.DetectGenres(query)
	if len(genreIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Жанры не найдены. Попробуйте использовать другие ключевые слова.",
			"hint":    hint,
		})
		return
	}

	movies, err := models.GetGenreFilms(genreIDs)
	if err != nil {
		log.Printf("Ошибка при получении фильмов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Қате орын алды. Кейінірек қайталап көріңіз.",
		})
		return
	}

	if len(movies) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Бұл жанрлар бойынша фильмдер табылмады.",
		})
		return
	}

	// Ограничиваем количество фильмов
	if len(movies) > 5 {
		movies = movies[:5]
	}

	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
	})
}
