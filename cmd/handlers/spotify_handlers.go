package handlers

import (
	"encoding/json"
	"fmt"
	"io"
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
	// For POST requests, we need to read from the request body instead of URL query parameters
	var requestBody struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if requestBody.State != state {
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

	req, err := http.NewRequest("POST", spotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenResponse["access_token"],
		"refresh_token": tokenResponse["refresh_token"],
	})
}
