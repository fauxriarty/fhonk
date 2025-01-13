package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	ClientID     = getEnv("SPOTIFY_CLIENT_ID", "")
	ClientSecret = getEnv("SPOTIFY_CLIENT_SECRET", "")
	RedirectURI  = getEnv("SPOTIFY_REDIRECT_URI", "http://localhost:8080/callback")
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// creates the Spotify authorization URL
func GenerateAuthURL(state string) string {
	authURL := "https://accounts.spotify.com/authorize"
	params := url.Values{}
	params.Add("client_id", ClientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", RedirectURI)
	params.Add("scope", "user-read-recently-played user-top-read") // add required scopes
	params.Add("state", state)                                     // state should contain the callback URL
	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// exchanges the authorization code for an access token
func ExchangeCodeForToken(code string) (*TokenResponse, error) {
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", RedirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+basicAuth(ClientID, ClientSecret))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func basicAuth(clientID, clientSecret string) string {
	return base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
}
