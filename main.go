package main

import (
	"fhonk/cmd/db"
	"fhonk/cmd/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello from da fhonk",
	})
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	router := gin.Default()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set in environment")
	}
	db.ConnectDB(dsn)
	defer db.CloseDB()

	// drop existing table
	// if err := db.DB.Migrator().DropTable(&models.User{}, &models.UserData{}); err != nil {
	// 	log.Fatalf("Failed to drop existing UserData table: %v", err)
	// }

	//migration
	// if err := db.DB.AutoMigrate(&models.User{}, &models.UserData{}); err != nil {
	// 	log.Fatalf("Failed to migrate UserData table: %v", err)
	// }

	// log.Println("Database migrated successfully!")

	router.GET("/", Status)

	router.GET("/login/spotify", func(c *gin.Context) {
		handlers.SpotifyLoginHandler(c.Writer, c.Request)
	})

	router.GET("/spotify-callback", func(c *gin.Context) {
		handlers.SpotifyCallbackHandler(c.Writer, c.Request)
	})

	port := "8080"
	log.Printf("listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
