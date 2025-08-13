package infrastructure

import (
	"errors"
	"os"
	"sync"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwtSecret     []byte
	loadSecretOnce sync.Once
)

func loadJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set in environment")
	}
	jwtSecret = []byte(secret)
}

type JWTService struct{}

func NewJWTService() *JWTService {
	loadSecretOnce.Do(loadJWTSecret)
	return &JWTService{}
}

func (j *JWTService) GenerateToken(userID, username, role string) (userpkg.TokenResult, error) {
	accessExp := time.Now().Add(15 * time.Minute)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"_id":      userID,
		"username": username,
		"role":     role,
		"exp":      accessExp.Unix(),
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	refreshExp := time.Now().Add(7 * 24 * time.Hour)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"_id": userID,
		"exp": refreshExp.Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	return userpkg.TokenResult{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}


func (j *JWTService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}