package models

import (
	"database/sql"
	"log"
	"time"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

type Discussion struct {
	ID         int
	UserID     string
	Username   string `json:"username"`
	MovieID    int
	Discussion string
	ParentID   *int // Может быть NULL
	CreatedAt  time.Time
	Replies    []Discussion // Список ответов
}

type Review struct {
	ID        int
	UserID    string
	Username  string `json:"username"`
	MovieID   int
	Rating    int
	Review    string
	CreatedAt time.Time
}

func GetReviewsByMovieID(movieID string) ([]Review, error) {
	rows, err := db.Query(`
        SELECT r.id, r.user_id, u.username, r.movie_id, r.rating, r.review, r.created_at
        FROM reviews r
        JOIN users u ON r.user_id = u.id
        WHERE r.movie_id = $1
        ORDER BY r.created_at DESC
    `, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		// Обратите внимание на порядок сканирования - он должен соответствовать порядку в SELECT
		err := rows.Scan(&r.ID, &r.UserID, &r.Username, &r.MovieID, &r.Rating, &r.Review, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func GetDiscussionsByMovieID(movieID string) ([]Discussion, error) {
	rows, err := db.Query(`
        SELECT d.id, d.user_id, u.username, d.movie_id, d.discussion, d.parent_id, d.created_at
        FROM discussions d
        JOIN users u ON d.user_id = u.id
        WHERE d.movie_id = $1
        ORDER BY d.created_at DESC
    `, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discussions []Discussion
	var repliesMap = make(map[int][]Discussion)

	for rows.Next() {
		var d Discussion
		// Сканируем все поля, включая username
		err := rows.Scan(&d.ID, &d.UserID, &d.Username, &d.MovieID, &d.Discussion, &d.ParentID, &d.CreatedAt)
		if err != nil {
			return nil, err
		}

		if d.ParentID == nil {
			discussions = append(discussions, d)
		} else {
			repliesMap[*d.ParentID] = append(repliesMap[*d.ParentID], d)
		}
	}

	for i := range discussions {
		if replies, ok := repliesMap[discussions[i].ID]; ok {
			discussions[i].Replies = replies
		}
	}

	return discussions, nil
}

// AddDiscussion добавляет новое обсуждение или ответ
func AddDiscussion(userID string, movieID int, discussion string, parentID *int) error {
	_, err := db.Exec(`
		INSERT INTO discussions (user_id, movie_id, discussion, parent_id) 
		VALUES (?, ?, ?, ?)
	`, userID, movieID, discussion, parentID)
	return err
}

func AddReview(userID string, movieID, rating int, reviewText string) error {
	_, err := db.Exec(`
		INSERT INTO reviews (user_id, movie_id, rating, review, created_at) 
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, movieID, rating, reviewText)
	return err
}

func HasUserReviewedMovie(userID string, movieID string) bool {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM reviews 
        WHERE user_id = ? AND movie_id = ?
    `, userID, movieID).Scan(&count)

	if err != nil {
		log.Printf("Ошибка проверки отзыва: %v", err)
		return false
	}

	return count > 0
}
