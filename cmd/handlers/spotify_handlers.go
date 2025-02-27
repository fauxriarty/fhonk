package handlers

import (
	"encoding/json"
	"fhonk/cmd/db"
	"fhonk/cmd/db/models"
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

	code := c.Query("code")
	receivedState := c.Query("state")

	log.Printf("Received from Spotify - Code: %s, State: %s", code, receivedState)

	if code == "" {
		log.Println("Error: Empty authorization code received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty authorization code"})
		return
	}

	if receivedState != state {
		log.Printf("State mismatch: got %s, expected %s", receivedState, state)
		c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
		return
	}

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

	accessToken := tokenResponse["access_token"].(string)
	refreshToken := tokenResponse["refresh_token"].(string)

	// Retrieve user profile from Spotify
	userProfile, err := getSpotifyUserProfile(accessToken)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	// Store user data in the database
	userData := models.UserData{
		UserID:       userProfile.ID,
		SpotifyID:    userProfile.ID,
		DisplayName:  userProfile.DisplayName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if err := db.DB.Create(&userData).Error; err != nil {
		log.Printf("Failed to save user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	log.Println("Successfully retrieved and stored Spotify user data")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type SpotifyUserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func getSpotifyUserProfile(accessToken string) (*SpotifyUserProfile, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user profile: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var userProfile SpotifyUserProfile
	if err := json.Unmarshal(body, &userProfile); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &userProfile, nil
}
