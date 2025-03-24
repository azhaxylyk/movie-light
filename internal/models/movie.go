package models

import (
	"database/sql"
	"time"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

type Discussion struct {
	ID         int
	UserID     int
	MovieID    int
	Discussion string
	ParentID   *int // Может быть NULL
	CreatedAt  time.Time
	Replies    []Discussion // Список ответов
}

type Review struct {
	ID        int
	UserID    string
	MovieID   int
	Rating    int
	Review    string
	CreatedAt time.Time
}

func GetReviewsByMovieID(movieID string) ([]Review, error) {
	rows, err := db.Query(`
		SELECT id, user_id, movie_id, rating, review, created_at
		FROM reviews 
		WHERE movie_id = ? 
		ORDER BY created_at DESC
	`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		err := rows.Scan(&r.ID, &r.UserID, &r.MovieID, &r.Rating, &r.Review, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func GetDiscussionsByMovieID(movieID string) ([]Discussion, error) {
	// Получаем все обсуждения
	rows, err := db.Query(`
		SELECT id, user_id, movie_id, discussion, parent_id, created_at
		FROM discussions 
		WHERE movie_id = ? 
		ORDER BY created_at ASC
	`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discussions []Discussion
	var repliesMap = make(map[int][]Discussion)

	// Разбиваем данные на обсуждения и ответы
	for rows.Next() {
		var d Discussion
		err := rows.Scan(&d.ID, &d.UserID, &d.MovieID, &d.Discussion, &d.ParentID, &d.CreatedAt)
		if err != nil {
			return nil, err
		}

		if d.ParentID == nil {
			discussions = append(discussions, d) // Основные обсуждения
		} else {
			repliesMap[*d.ParentID] = append(repliesMap[*d.ParentID], d) // Ответы
		}
	}

	// Добавляем ответы к обсуждениям
	for i := range discussions {
		if replies, ok := repliesMap[discussions[i].ID]; ok {
			discussions[i].Replies = replies
		}
	}

	return discussions, nil
}

// AddDiscussion добавляет новое обсуждение или ответ
func AddDiscussion(userID, movieID int, discussion string, parentID *int) error {
	_, err := db.Exec(`
		INSERT INTO discussions (user_id, movie_id, discussion, parent_id) 
		VALUES (?, ?, ?, ?)
	`, userID, movieID, discussion, parentID)
	return err
}

func AddReview(userID, movieID, rating int, reviewText string) error {
	_, err := db.Exec(`
		INSERT INTO reviews (user_id, movie_id, rating, review, created_at) 
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, movieID, rating, reviewText)
	return err
}
