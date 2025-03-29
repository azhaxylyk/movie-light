package handlers

import (
	"fmt"
	"log"
	"movie-light/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateRoom(c *gin.Context) {
	// Проверяем авторизацию
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	_, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Рендерим HTML-страницу с формой
	c.HTML(http.StatusOK, "room_create.html", gin.H{
		"Username": username,
	})
}

func CreateNewRoom(c *gin.Context) {
	// Проверяем авторизацию
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Авторизация талап етіледі"})
		return
	}

	userID, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Сессия мерзімі өтіп кетті"})
		return
	}

	// Читаем JSON из запроса
	var req struct {
		RoomName string `json:"roomName"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Қате сұраныс форматы"})
		return
	}

	// Создаем комнату
	room := models.NewRoom(userID, username)

	// Возвращаем JSON с ID комнаты
	c.JSON(http.StatusOK, gin.H{
		"roomId": room.ID,
	})
}

func HandleRoomWebSocket(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		return
	}

	userID, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		return
	}

	roomID := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add user to room
	err = models.AddUserToRoom(roomID, userID, username, conn)
	if err != nil {
		log.Printf("Add user to room error: %v", err)
		return
	}
	sendRoomStatus(roomID)
	room, _ := models.GetRoom(roomID)

	for {
		var msg models.WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch msg.Type {
		case "get_room_status":
			sendRoomStatus(roomID)
		case "swipe":
			models.ProcessSwipe(roomID, userID, msg.FilmID, msg.Action)

		case "select_genre":
			// Сохраняем выбранные жанры
			room.Settings[userID] = msg.Genres
			// Сохраняем жанры в профиль пользователя (пример для models.SetUserGenres)
			err := models.SetUserGenres(userID, msg.Genres)
			if err != nil {
				log.Printf("Ошибка сохранения жанров: %v", err)
			}
			// Проверяем, все ли выбрали жанры
			if len(room.Settings) == 2 {
				// Получаем жанры от обоих пользователей
				var genreIDs []int
				for _, g := range room.Settings {
					if userGenres, ok := g.([]int); ok {
						genreIDs = append(genreIDs, userGenres...)
					}
				}

				// Получаем ВСЕ фильмы по выбранным жанрам (без ограничения в 10)
				films, err := models.GetGenreFilms(genreIDs)
				if err != nil {
					log.Printf("Ошибка получения фильмов: %v", err)
					return
				}

				room.Films = films

				// Отправляем подтверждение и фильмы
				models.Broadcast(roomID, models.WSMessage{
					Type:   "both_selected",
					Genres: genreIDs,
					Data:   films,
				})
			}

		case "request_films":
			if genre, ok := room.Settings["genre"].(string); ok {
				films, err := models.GetRandomFilms(10, genre)
				if err == nil {
					room.Films = films
					conn.WriteJSON(models.WSMessage{
						Type: "films_list",
						Data: films,
					})
				}
			}
		}
	}

	models.RemoveUserFromRoom(roomID, userID)
	sendRoomStatus(roomID)
}

func sendRoomStatus(roomID string) {
	room, exists := models.GetRoom(roomID)
	if !exists {
		return
	}

	models.Broadcast(roomID, models.WSMessage{
		Type:       "room_status",
		UsersCount: len(room.Users),
	})
}

func RoomPage(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	userID, username, err := models.GetIDBySessionToken(sessionToken)
	fmt.Println(username)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	genres, _ := models.GetUserGenres(userID)
	roomID := c.Param("id")
	room, exists := models.GetRoom(roomID)
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error":   "404 Бөлме табылмады",
			"Message": "Мұндай ID-мен бөлме жоқ немесе ол жойылған",
		})
		return
	}

	isCreator := room.CreatorID == userID
	otherUser := getOtherUsername(room, userID)
	c.HTML(http.StatusOK, "room.html", gin.H{
		"RoomID":         roomID,
		"Username":       username,
		"IsCreator":      isCreator,
		"BaseURL":        getBaseURL(c),
		"Films":          room.Films,
		"LoggedIn":       true,
		"OtherUser":      otherUser,
		"SelectedGenres": genres,
	})
}

func getOtherUsername(room *models.Room, currentUserID string) string {
	for _, user := range room.Users {
		if user.UserID != currentUserID {
			return user.Username
		}
	}
	return ""
}

func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}
