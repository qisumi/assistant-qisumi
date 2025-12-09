package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Secret []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{Secret: []byte(secret)}
}

func (j *JWTManager) GenerateToken(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j *JWTManager) ParseToken(tokenStr string) (uint64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(float64); ok {
			return uint64(sub), nil
		}
	}
	return 0, jwt.ErrTokenMalformed
}
