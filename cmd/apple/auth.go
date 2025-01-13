package apple

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GenerateDeveloperToken() (string, error) {
	// Load environment variables
	teamID := os.Getenv("TEAM_ID")
	keyID := os.Getenv("KEY_ID")
	privateKey := os.Getenv("PRIVATE_KEY")

	// Parse the private key
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": teamID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(6 * time.Hour).Unix(), // Token validity: 6 hours
		"kid": keyID,
	})

	// Sign the token
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}
