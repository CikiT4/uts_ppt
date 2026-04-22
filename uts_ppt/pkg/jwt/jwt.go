package jwt

import (
	"errors"
	"time"

	"legal-consultation-api/internal/config"
	"legal-consultation-api/internal/models"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID        `json:"user_id"`
	Email  string           `json:"email"`
	Role   models.UserRole  `json:"role"`
	jwtlib.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func GenerateTokenPair(userID uuid.UUID, email string, role models.UserRole) (*TokenPair, error) {
	cfg := config.AppConfig
	secret := []byte(cfg.JWTSecret)

	// Access token
	accessExpiry := time.Now().Add(time.Duration(cfg.JWTExpireHours) * time.Hour)
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(accessExpiry),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}
	accessToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshExpiry := time.Now().Add(time.Duration(cfg.JWTRefreshExpireHours) * time.Hour)
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(refreshExpiry),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}
	refreshToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	secret := []byte(config.AppConfig.JWTSecret)

	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
