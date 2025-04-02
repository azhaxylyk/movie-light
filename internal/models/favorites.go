package models

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
