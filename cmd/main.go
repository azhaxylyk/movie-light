package main

import (
	"encoding/json"
	"html/template"
	"log"
	"movie-light/internal/config"
	"movie-light/internal/handlers"
	"movie-light/internal/models"
	"movie-light/internal/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := sql.InitDB()
	if err != nil {
		log.Fatal(err, "server shutdown")
	}
	models.SetDB(db)
	config.LoadEnv()
	handlers.InitOAuthConfigs()
	models.InitAPI()

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"until": func(n int) []int {
			var result []int
			for i := 1; i <= n; i++ {
				result = append(result, i)
			}
			return result
		},
		"now": func() time.Time {
			return time.Now()
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"toString": func(i int) string {
			return strconv.Itoa(i)
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"formatYear": func(date string) string {
			if len(date) >= 4 {
				return date[:4]
			}
			return date
		},
		"json": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	})
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")

	// Маршруты для аутентификации
	r.GET("/login", handlers.LoginHandler)
	r.POST("/login", handlers.LoginHandler)
	r.GET("/register", handlers.RegisterHandler)
	r.POST("/register", handlers.RegisterHandler)
	r.POST("/logout", handlers.LogoutHandler)

	// Маршруты для OAuth
	r.GET("/auth/google", handlers.GoogleAuthHandler)
	r.GET("/auth/google/callback", handlers.GoogleCallbackHandler)

	// Главная страница (доступна без авторизации)
	r.GET("/", handlers.HomePage)
	r.GET("/movie/:id", handlers.MovieDetailPage)

	// Избранное
	r.POST("/favorites", handlers.HandleFavorites)
	r.DELETE("/favorites", handlers.HandleFavorites)

	r.POST("/add-discussion", handlers.AddDiscussion)
	r.POST("/add-review", handlers.AddReview)

	r.GET("/ai", handlers.ChatBotPage)
	r.POST("/ai/query", handlers.ChatBotHandler)

	r.GET("/trending", func(c *gin.Context) {
		timeWindow := c.DefaultQuery("timeWindow", "day") // По умолчанию "day"
		movies, err := models.GetTrendingMovies(timeWindow)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, movies)
	})

	// Добавьте после других маршрутов
	r.GET("/room/create", handlers.CreateRoom) // Показывает страницу с формой
	r.POST("/room/new", handlers.CreateNewRoom)
	r.GET("/room/:id", handlers.RoomPage)
	r.GET("/ws/room/:id", handlers.HandleRoomWebSocket)

	r.GET("/watch/:roomID", handlers.WatchRoomPage)
	r.GET("/ws/watch/:id", handlers.HandleWatchWebSocket)

	// Добавьте этот маршрут перед authGroup
	r.GET("/search", handlers.SearchHandler)
	// Защищенные маршруты (требуют авторизации)
	authGroup := r.Group("/")
	authGroup.Use(AuthMiddleware())
	{
		authGroup.GET("/profile", handlers.ProfilePage)
	}

	r.Run(":8080")
}

// AuthMiddleware проверяет наличие valid session token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("session_token")
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		userID, username, err := models.GetIDBySessionToken(sessionToken)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()
	}
}
