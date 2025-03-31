package handlers

import (
	"log"
	"movie-light/internal/models"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var watchUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWatchWebSocket(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		log.Println("Ошибка получения session_token:", err)
		return
	}

	userID, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		log.Println("Ошибка получения userID по session_token:", err)
		return
	}

	roomID := c.Param("id")
	conn, err := watchUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer func() {
		models.RemoveUserFromWatchRoom(roomID, userID)
		conn.Close()
		log.Printf("Клиент %s отключился от комнаты %s", userID, roomID)
	}()

	err = models.AddUserToWatchRoom(roomID, userID, username, conn)
	if err != nil {
		log.Printf("Ошибка добавления пользователя в комнату: %v", err)
		return
	}
	log.Printf("Клиент %s подключился к комнате %s", userID, roomID)

	// Отправляем текущий статус комнаты новому пользователю
	models.SendWatchRoomStatus(roomID)

	// Чтение сообщений от клиента
	for {
		var msg struct {
			Type    string      `json:"type"`
			MovieID interface{} `json:"movieID"`
			Text    string      `json:"text"`
			Data    interface{} `json:"data"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "chat_message":
			models.BroadcastWatchRoom(roomID, models.WSMessage{
				Type:     "chat_message",
				Username: username,
				Text:     msg.Text,
				Data:     msg.Data,
			})

		case "request_trailer":
			var movieID int
			switch v := msg.MovieID.(type) {
			case float64:
				movieID = int(v)
			case int:
				movieID = v
			default:
				log.Printf("Неподдерживаемый тип movieID: %T", msg.MovieID)
				continue
			}

			trailerURL, err := models.GetMovieTrailer(movieID)
			if err != nil {
				log.Printf("Ошибка получения трейлера: %v", err)
				continue
			}

			videoID := extractYouTubeID(trailerURL)
			if videoID == "" {
				log.Printf("Не удалось извлечь YouTube ID из URL: %s", trailerURL)
				continue
			}

			models.BroadcastWatchRoom(roomID, models.WSMessage{
				Type:    "movie_trailer",
				VideoId: videoID,
				Data:    msg.Data,
			})
		}
	}
}

// Функция для извлечения YouTube ID из URL
func extractYouTubeID(url string) string {
	patterns := []string{
		`youtube\.com\/watch\?v=([^"&?\/\s]{11})`,
		`youtube\.com\/embed\/([^"&?\/\s]{11})`,
		`youtu\.be\/([^"&?\/\s]{11})`,
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Printf("Ошибка компиляции регулярного выражения: %v", err)
			continue
		}
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
func WatchRoomPage(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	c.Header("Content-Security-Policy", "frame-src https://www.youtube.com;")
	c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
	_, username, err := models.GetIDBySessionToken(sessionToken)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	roomID := c.Param("roomID")
	movieID := c.Query("movie_id") // Получаем ID фильма из query параметра

	// Получаем информацию о фильме для заголовка
	var movieTitle string
	if movieID != "" {
		movie, err := models.GetMovieDetails(movieID)
		if err == nil {
			movieTitle = movie.Title
		}
	}

	c.HTML(http.StatusOK, "watch_room.html", gin.H{
		"RoomID":     roomID,
		"Username":   username,
		"MovieID":    movieID,
		"MovieTitle": movieTitle,
	})
}
