package handlers

import (
	"movie-light/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
