package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type Manager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewManager(secret string, accessHours, refreshHours int) *Manager {
	return &Manager{
		secret:             []byte(secret),
		accessTokenExpiry:  time.Duration(accessHours) * time.Hour,
		refreshTokenExpiry: time.Duration(refreshHours) * time.Hour,
	}
}

func (m *Manager) GenerateAccessToken(userID uint) (string, error) {
	return m.generateToken(userID, "access", m.accessTokenExpiry)
}

func (m *Manager) GenerateRefreshToken(userID uint) (string, error) {
	return m.generateToken(userID, "refresh", m.refreshTokenExpiry)
}

func (m *Manager) generateToken(userID uint, tokenType string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
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
