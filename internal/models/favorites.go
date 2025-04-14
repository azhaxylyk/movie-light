package models

import (
	"fmt"
	"log"
	"time"
)

// UserReview представляет отзыв пользователя с информацией о фильме
type UserReview struct {
	ID         int
	MovieID    int
	MovieTitle string
	PosterPath string
	Rating     int
	Review     string
	CreatedAt  time.Time
}

// Добавляет фильм в избранное
func AddToFavorites(userID, movieID string) error {
	_, err := db.Exec("INSERT INTO favorites (user_id, movie_id) VALUES (?, ?)", userID, movieID)
	return err
}

// Удаляет фильм из избранного
func RemoveFromFavorites(userID, movieID string) error {
	_, err := db.Exec("DELETE FROM favorites WHERE user_id = ? AND movie_id = ?", userID, movieID)
	return err
}

// Проверяет, есть ли фильм в избранном у пользователя
func IsMovieInFavorites(userID, movieID string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM favorites WHERE user_id = ? AND movie_id = ?", userID, movieID).Scan(&count)
	return count > 0, err
}

// Возвращает количество добавлений в избранное для фильма
func GetFavoriteCount(movieID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM favorites WHERE movie_id = ?", movieID).Scan(&count)
	return count, err
}

func GetUserFavorites(userID string) ([]MovieDetail, error) {
	var movieIDs []int
	rows, err := db.Query("SELECT movie_id FROM favorites WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		movieIDs = append(movieIDs, id)
	}

	// Получаем данные о фильмах через API
	var favorites []MovieDetail
	for _, id := range movieIDs {
		movie, err := GetMovieDetails(fmt.Sprintf("%d", id)) // Запрашиваем у API
		if err != nil {
			continue // Можно логировать ошибку
		}
		favorites = append(favorites, *movie)
	}

	return favorites, nil
}

func GetUserReviews(userID string) ([]UserReview, error) {
	rows, err := db.Query(`
        SELECT r.id, r.movie_id, r.rating, r.review, r.created_at
        FROM reviews r
        WHERE r.user_id = ?
        ORDER BY r.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []UserReview
	for rows.Next() {
		var r UserReview
		err := rows.Scan(&r.ID, &r.MovieID, &r.Rating, &r.Review, &r.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Получаем информацию о фильме через API
		movieDetails, err := GetMovieDetails(fmt.Sprintf("%d", r.MovieID))
		if err != nil {
			log.Printf("Ошибка получения данных о фильме %d: %v", r.MovieID, err)
			r.MovieTitle = "Неизвестный фильм"
			r.PosterPath = "" // Если API недоступно
		} else {
			r.MovieTitle = movieDetails.Title
			r.PosterPath = movieDetails.PosterPath
		}

		reviews = append(reviews, r)
	}

	return reviews, nil
}
