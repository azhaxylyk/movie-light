package models

import (
	"log"

	"github.com/gorilla/websocket"
)

// Структура для комнаты просмотра
type WatchRoom struct {
	ID    string
	Users map[string][]*UserSession
}

// Глобальная карта всех комнат
var watchRooms = make(map[string]*WatchRoom)

// Добавление пользователя в комнату
func AddUserToWatchRoom(roomID, userID, username string, conn *websocket.Conn) error {
	room, exists := watchRooms[roomID]
	if !exists {
		room = &WatchRoom{
			ID:    roomID,
			Users: make(map[string][]*UserSession),
		}
		watchRooms[roomID] = room
	}

	room.Users[userID] = append(room.Users[userID], &UserSession{UserID: userID, Username: username, Conn: conn})

	log.Printf("Комната %s, текущее количество пользователей: %d", roomID, len(room.Users))
	SendWatchRoomStatus(roomID) // Отправляем обновленный статус

	return nil
}

// Удаление пользователя из комнаты
func RemoveUserFromWatchRoom(roomID, userID string) {
	if room, exists := watchRooms[roomID]; exists {
		delete(room.Users, userID)
		if len(room.Users) == 0 {
			delete(watchRooms, roomID)
		}
	}
	SendWatchRoomStatus(roomID) // Обновляем статус после удаления
}

// Отправка данных о количестве пользователей
func SendWatchRoomStatus(roomID string) {
	if room, exists := watchRooms[roomID]; exists {
		usersCount := len(room.Users)
		message := WSMessage{
			Type:       "room_status",
			UsersCount: usersCount,
		}
		BroadcastWatchRoom(roomID, message)
		log.Printf("Отправлен статус комнаты %s: %d пользователей", roomID, usersCount)
	}
}

// Трансляция сообщений в комнату
func BroadcastWatchRoom(roomID string, msg WSMessage) {
	if room, exists := watchRooms[roomID]; exists {
		for _, sessions := range room.Users {
			for _, user := range sessions {
				if user.Conn != nil {
					// Убедимся, что сообщение содержит все необходимые поля
					fullMsg := WSMessage{
						Type:       msg.Type,
						Username:   msg.Username,
						Text:       msg.Text,
						UsersCount: msg.UsersCount,
						MovieID:    msg.MovieID,
						VideoId:    msg.VideoId,
						Data:       msg.Data,
					}

					if err := user.Conn.WriteJSON(fullMsg); err != nil {
						log.Printf("Ошибка отправки сообщения пользователю %s: %v", user.UserID, err)
						// Можно добавить обработку ошибки, например, удаление отключенного пользователя
					}
				}
			}
		}
	}
}

// Отправка трейлера в комнату
func SendMovieTrailerToRoom(roomID string, movieID int) {
	trailerURL, err := GetMovieTrailer(movieID)
	if err != nil {
		log.Printf("Ошибка при получении трейлера: %v", err)
		return
	}

	// Отправляем трейлер всем участникам комнаты
	BroadcastWatchRoom(roomID, WSMessage{
		Type:    "movie_trailer",
		FilmID:  movieID,
		Message: trailerURL,
	})
}
