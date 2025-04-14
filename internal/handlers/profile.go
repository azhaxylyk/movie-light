package handlers

import (
	"fmt"
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProfilePage(c *gin.Context) {
	// Получаем сессионный токен из cookies
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Получаем данные пользователя по токену
	userID, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var favorites []models.MovieDetail
	favoritesFromDB, err := models.GetUserFavorites(userID)
	if err != nil {
		log.Print("Ошибка получения избранных фильмов:", err)
		favorites = []models.MovieDetail{} // Чтобы избежать nil в шаблоне
	} else {
		// Преобразуем []models.Movie в []models.MovieDetail
		for _, movie := range favoritesFromDB {
			details, err := models.GetMovieDetails(fmt.Sprintf("%d", movie.ID))
			if err == nil {
				favorites = append(favorites, *details)
			}
		}
	}
	log.Printf("Favorites data: %+v", favorites)
	// Получаем отзывы пользователя
	reviews, err := models.GetUserReviews(userID)
	if err != nil {
		log.Printf("Ошибка получения отзывов пользователя: %v", err)
		reviews = []models.UserReview{}
	}

	// Передаем данные в шаблон
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"Username":  username,
		"Favorites": favorites,
		"Reviews":   reviews,
	})
}
