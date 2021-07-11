package utils

import (
	"fmt"
	"goventy/config"

	jwt "github.com/dgrijalva/jwt-go"
)

func GetJwtToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(config.ENV()["JWT_SECRET_KEY"]), nil
	})

	if err != nil {
		return nil, err
	}
	return token, nil
}
