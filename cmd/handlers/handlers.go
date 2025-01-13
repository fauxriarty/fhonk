package handlers

import (
	"fhonk/cmd/spotify"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SpotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	// a random state parameter
	state := "random_generated_state" // replace with actual random state generation
	authURL := spotify.GenerateAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func SpotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// exchange code for token
	tokenData, err := spotify.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	accessToken, ok := tokenData["access_token"].(string)
	if !ok {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	refreshToken, ok := tokenData["refresh_token"].(string)
	if !ok {
		http.Error(w, "Failed to get refresh token", http.StatusInternalServerError)
		return
	}

	// return the access token to the frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"access_token": "%s" , "refresh_token": "%s"}`, accessToken, refreshToken)))
}
