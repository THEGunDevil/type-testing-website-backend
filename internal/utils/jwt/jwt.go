package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	ID           string `json:"id"`
	Role         string `json:"role"`
	TokenVersion int32  `json:"token_version"`
	jwt.RegisteredClaims
}
type JWTManager struct {
	accessSecret         []byte
	refreshSecret        []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTManager(
	accessSecret string,
	refreshSecret string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) (*JWTManager, error) {

	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		return nil, errors.New("JWT secrets must be at least 32 bytes")
	}

	if accessSecret == refreshSecret {
		return nil, errors.New("access and refresh secrets must be different")
	}

	return &JWTManager{
		accessSecret:         []byte(accessSecret),
		refreshSecret:        []byte(refreshSecret),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(
	userID string,
	role string,
	tokenVersion int32,
) (string, error) {

	if len(m.accessSecret) == 0 {
		return "", errors.New("access secret not initialized")
	}

	now := time.Now()

	claims := MyCustomClaims{
		ID:           userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "typetype",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}
func (m *JWTManager) GenerateRefreshToken(
	userID string,
	role string,
	tokenVersion int32,
) (string, error) {

	if len(m.refreshSecret) == 0 {
		return "", errors.New("refresh secret not initialized")
	}

	now := time.Now()

	claims := MyCustomClaims{
		ID:           userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "typetype",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) verifyToken(
	tokenStr string,
	secret []byte,
) (*MyCustomClaims, error) {

	if len(secret) == 0 {
		return nil, errors.New("secret not initialized")
	}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&MyCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return secret, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}

		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*MyCustomClaims)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	if claims.Issuer != "typetype" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}
func (m *JWTManager) VerifyAccessToken(
	tokenStr string,
) (*MyCustomClaims, error) {

	return m.verifyToken(tokenStr, m.accessSecret)
}

func (m *JWTManager) VerifyRefreshToken(
	tokenStr string,
) (*MyCustomClaims, error) {

	return m.verifyToken(tokenStr, m.refreshSecret)
}
