package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	spotifyAuthURL      = "https://accounts.spotify.com/authorize"
	spotifyTokenURL     = "https://accounts.spotify.com/api/token"
	spotifyClientID     = os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRedirectURI  = os.Getenv("SPOTIFY_REDIRECT_URI")
	state               = "randomStateString"
)

func SpotifyLoginHandler(c *gin.Context) {
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&state=%s&scope=user-read-private",
		spotifyAuthURL, spotifyClientID, spotifyRedirectURI, state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func SpotifyCallbackHandler(c *gin.Context) {
	log.Println("Received Spotify callback request")

	// Try to bind JSON data first
	var requestBody struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil || requestBody.Code == "" || requestBody.State == "" {
		// Fallback to query parameters if JSON binding fails or data is missing
		requestBody.Code = c.Query("code")
		requestBody.State = c.Query("state")
	}

	log.Printf("Parsed request body - Code: %s, State: %s", requestBody.Code, requestBody.State)

	if requestBody.Code == "" {
		log.Println("Error: Empty authorization code received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty authorization code"})
		return
	}

	if requestBody.State != state {
		log.Printf("State mismatch: got %s, expected %s", requestBody.State, state)
		c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
		return
	}

	code := requestBody.Code
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", spotifyRedirectURI)
	data.Set("client_id", spotifyClientID)
	data.Set("client_secret", spotifyClientSecret)

	log.Printf("Spotify token request data: %v", data)
	log.Printf("Redirect URI: %s", spotifyRedirectURI)

	req, err := http.NewRequest("POST", spotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Println("Sending request to Spotify token endpoint")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token"})
		return
	}
	defer resp.Body.Close()

	log.Printf("Spotify token response status: %d %s", resp.StatusCode, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	log.Printf("Spotify token response body: %s", string(body))

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Printf("Failed to parse response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Check if there was an error in the token response
	if tokenResponse["error"] != nil {
		log.Printf("Spotify API error: %v", tokenResponse["error"])
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Spotify API error",
			"details": tokenResponse["error"],
		})
		return
	}

	log.Println("Successfully retrieved Spotify tokens")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenResponse["access_token"],
		"refresh_token": tokenResponse["refresh_token"],
	})
}
