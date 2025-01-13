package handlers

import (
	"fhonk/cmd/apple"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppleMusicLoginHandler handles the connection to Apple Music
func AppleMusicLoginHandler(c *gin.Context) {
	// Generate the developer token
	token, err := apple.GenerateDeveloperToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate developer token"})
		return
	}

	// Return the token to the client
	c.JSON(http.StatusOK, gin.H{"developer_token": token})
}
