package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatBotPage отображает страницу чат-бота
func ChatBotPage(c *gin.Context) {
	c.HTML(http.StatusOK, "ai.html", nil)
}

// ChatBotHandler обрабатывает POST-запросы от чат-бота
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

	if len(genreIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Жанры не найдены. Попробуйте использовать другие ключевые слова.",
		})
		return
	}

	var responses []string

	for _, genreID := range genreIDs {
		resp, err := models.ChatResponseByGenre(genreID)
		if err != nil {
			log.Printf("Ошибка получения ответа по жанру %d: %v", genreID, err)
			continue
		}
		responses = append(responses, resp)
	}

	if len(responses) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Фильмы не найдены по указанным жанрам.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": responses,
	})
}
