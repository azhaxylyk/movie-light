package models

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type       string      `json:"type"`
	FilmID     int         `json:"filmId,omitempty"`
	Action     string      `json:"action,omitempty"`
	Genres     []int       `json:"genres,omitempty"` // Изменено с Genre на Genres
	Message    string      `json:"message,omitempty"`
	FilmTitle  string      `json:"filmTitle,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	UsersCount int         `json:"usersCount,omitempty"`
}

// Room представляет комнату для совместного просмотра
type Room struct {
	ID        string
	CreatorID string
	Settings  map[string]interface{}
	CreatedAt time.Time
	Users     map[string]*UserSession
	Films     []Movie
}

// UserSession представляет подключенного пользователя
type UserSession struct {
	UserID   string
	Username string
	Conn     *websocket.Conn
	Swiped   map[int]string // filmID: "like"/"dislike"
}

var Rooms = make(map[string]*Room)       // In-memory хранилище комнат
var userGenresCache = map[string][]int{} // userID -> genres

func GetUserGenres(userID string) ([]int, error) {
	return userGenresCache[userID], nil
}

func SetUserGenres(userID string, genres []int) error {
	userGenresCache[userID] = genres
	return nil
}

// NewRoom создает новую комнату
func NewRoom(creatorID, creatorName string) *Room {
	roomID := generateRoomID()
	room := &Room{
		ID:        roomID,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
		Users:     make(map[string]*UserSession),
		Settings:  make(map[string]interface{}),
	}
	Rooms[roomID] = room
	return room
}

// GetRandomFilms получает случайные фильмы по жанру
func GetRandomFilms(count int, genres ...string) ([]Movie, error) {
	// Получаем популярные фильмы
	movies, err := GetPopularMovies()
	if err != nil {
		return nil, err
	}

	// Если жанры не указаны, возвращаем все фильмы
	if len(genres) == 0 || (len(genres) == 1 && genres[0] == "") {
		if count > 0 && count < len(movies) {
			return movies[:count], nil
		}
		return movies, nil
	}

	// Получаем ID жанров
	genreIDs := make(map[int]bool)
	allGenres, err := GetGenres()
	if err != nil {
		return nil, err
	}

	log.Printf("Filtering movies for genres: %v", genres)
	for _, genreName := range genres {
		for _, g := range allGenres {
			if g.Name == genreName {
				genreIDs[g.ID] = true
				break
			}
		}
	}

	log.Printf("Total movies before filtering: %d", len(movies))
	// Фильтруем фильмы по жанрам
	var filtered []Movie
	for _, m := range movies {
		for _, gID := range m.GenreIDs {
			if genreIDs[gID] {
				filtered = append(filtered, m)
				break
			}
		}
	}

	// Если фильмов меньше, чем нужно, или count=0 (все фильмы), возвращаем все
	if count <= 0 || len(filtered) <= count {
		return filtered, nil
	}

	// Выбираем случайные фильмы
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	return filtered[:count], nil
}

// GetRoom возвращает комнату по ID
func GetRoom(roomID string) (*Room, bool) {
	room, exists := Rooms[roomID]
	return room, exists
}

// AddUserToRoom добавляет пользователя в комнату
func AddUserToRoom(roomID, userID, username string, conn *websocket.Conn) error {
	room, exists := Rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}

	// Максимум 2 пользователя в комнате
	if len(room.Users) >= 2 {
		return errors.New("room is full")
	}

	room.Users[userID] = &UserSession{
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Swiped:   make(map[int]string),
	}

	// Уведомляем других пользователей
	Broadcast(roomID, WSMessage{
		Type:    "user_joined",
		Message: fmt.Sprintf("%s joined the room", username),
	})

	return nil
}

// RemoveUserFromRoom удаляет пользователя из комнаты
func RemoveUserFromRoom(roomID, userID string) {
	if room, exists := Rooms[roomID]; exists {
		if user, ok := room.Users[userID]; ok {
			delete(room.Users, userID)
			Broadcast(roomID, WSMessage{
				Type:    "user_left",
				Message: fmt.Sprintf("%s left the room", user.Username),
			})
		}
	}
}

// ProcessSwipe обрабатывает свайп фильма
func ProcessSwipe(roomID, userID string, filmID int, action string) {
	room, exists := Rooms[roomID]
	if !exists {
		return
	}

	user, exists := room.Users[userID]
	if !exists {
		return
	}

	// Сохраняем выбор пользователя
	user.Swiped[filmID] = action

	// Проверяем, все ли проголосовали
	checkAllSwiped(room, filmID)
}

// checkAllSwiped проверяет, все ли пользователи проголосовали
func checkAllSwiped(room *Room, filmID int) {
	// Собираем все голоса
	likes := 0
	dislikes := 0
	totalUsers := len(room.Users)

	for _, user := range room.Users {
		if action, ok := user.Swiped[filmID]; ok {
			if action == "like" {
				likes++
			} else {
				dislikes++
			}
		}
	}

	// Если все проголосовали
	if likes+dislikes == totalUsers {
		var filmTitle string
		for _, f := range room.Films {
			if f.ID == filmID {
				filmTitle = f.Title
				break
			}
		}

		// Отправляем результаты
		Broadcast(room.ID, WSMessage{
			Type:      "swipe_result",
			FilmID:    filmID,
			FilmTitle: filmTitle,
			Data: map[string]interface{}{
				"likes":    likes,
				"dislikes": dislikes,
			},
		})

		// Если все лайкнули
		if likes == totalUsers {
			Broadcast(room.ID, WSMessage{
				Type:      "match",
				FilmID:    filmID,
				FilmTitle: filmTitle,
			})
		}
	}
}

// broadcast отправляет сообщение всем участникам комнаты
func Broadcast(roomID string, msg WSMessage) {
	if room, exists := Rooms[roomID]; exists {
		for _, user := range room.Users {
			if user.Conn != nil {
				user.Conn.WriteJSON(msg)
			}
		}
	}
}

// generateRoomID создает случайный ID комнаты
func generateRoomID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
