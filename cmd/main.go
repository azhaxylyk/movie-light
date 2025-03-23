package main

import (
	"log"
	"movie-light/internal/config"
	"movie-light/internal/handlers"
	"movie-light/internal/models"
	"movie-light/internal/sql"
	"net/http"

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
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")

	// Маршруты для аутентификации
	r.GET("/login", handlers.LoginHandler)
	r.POST("/login", handlers.LoginHandler)
	r.GET("/register", handlers.RegisterHandler)
	r.POST("/register", handlers.RegisterHandler)
	r.GET("/logout", handlers.LogoutHandler)

	// Маршруты для OAuth
	r.GET("/auth/google", handlers.GoogleAuthHandler)
	r.GET("/auth/google/callback", handlers.GoogleCallbackHandler)
	r.GET("/auth/github", handlers.GitHubAuthHandler)
	r.GET("/auth/github/callback", handlers.GitHubCallbackHandler)

	r.GET("/trending", func(c *gin.Context) {
		timeWindow := c.DefaultQuery("timeWindow", "day")
		movies, err := models.GetTrendingMovies(timeWindow)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, movies)
	})

	// Защищенные маршруты
	authGroup := r.Group("/")
	authGroup.Use(AuthMiddleware())
	{
		authGroup.GET("/", handlers.HomePage)
		authGroup.GET("/movie/:id", handlers.MovieDetailPage)
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
