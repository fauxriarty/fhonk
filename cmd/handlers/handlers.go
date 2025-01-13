package handlers

import (
	"fhonk/cmd/spotify"
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
	frontendCallback := r.URL.Query().Get("callback_url")
	if frontendCallback == "" {
		http.Error(w, "Missing callback URL", http.StatusBadRequest)
		return
	}

	authURL := spotify.GenerateAuthURL(frontendCallback)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func SpotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	frontendCallback := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	tokens, err := spotify.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL := frontendCallback + "?access_token=" + tokens.AccessToken + "&refresh_token=" + tokens.RefreshToken
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
