package utils

import (
	"mcpdocs/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey string = config.GetJWTSecretKey()

func GenerateJWT(userID string, expiryDuration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiryDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
