package main

import (
	"log"
	"movie-light/internal/config"
	"movie-light/internal/handlers"
	"movie-light/internal/models"
	"movie-light/internal/sql"

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
	r.GET("/", handlers.HomePage)
	r.GET("/movie/:id", handlers.MovieDetailPage)
	r.Run(":8080")
}
