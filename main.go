package main

import (
	"context"
	"fhonk/cmd/db"
	"fhonk/cmd/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	apiVersion = "v1"
	appName    = "fhonk"
	version    = "0.1"
)

type Config struct {
	Port        string
	Environment string
	DatabaseURL string
}

func loadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in environment")
	}

	return &Config{
		Port:        port,
		Environment: env,
		DatabaseURL: dbURL,
	}, nil
}

func setupRouter() *gin.Engine {
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"app":     appName,
			"version": version,
		})
	})

	// API routes
	v1 := router.Group("/api/" + apiVersion)
	{
		auth := v1.Group("/auth")
		{
			auth.GET("/apple", handlers.AppleMusicLoginHandler)
			auth.GET("/spotify", handlers.SpotifyLoginHandler)
			auth.GET("/spotify/callback", handlers.SpotifyCallbackHandler)
		}
	}

	return router
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	db.ConnectDB(config.DatabaseURL)
	defer db.CloseDB()

	router := setupRouter()

	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
