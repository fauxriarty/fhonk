package handlers

import (
	"bytes"
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

	// Log the raw request body before parsing
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading raw request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Restore the body for subsequent reads
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	log.Printf("Raw request body: %s", string(rawBody))

	// For POST requests, we need to read from the request body instead of URL query parameters
	var requestBody struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)

		// Try to debug the content type
		log.Printf("Content-Type header: %s", c.GetHeader("Content-Type"))

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
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
