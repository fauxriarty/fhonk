package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", Status)

	port := "8080"
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

}

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello from da fhonk",
	})
}
