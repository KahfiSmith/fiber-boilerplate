package jwt

import (
	"errors"
	"fmt"
	"time"

	"fiber-boilerplate/src/config"

	golangjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

type TokenService struct {
	cfg config.AuthConfig
}

func NewTokenService(cfg config.AuthConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

type AccessTokenClaims struct {
	Sub       uint   `json:"sub"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	TokenType string `json:"token_type"`
	golangjwt.RegisteredClaims
}

func (s *TokenService) GenerateAccessToken(userID uint, email string, role string, sessionID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)

	claims := AccessTokenClaims{
		Sub:       userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		TokenType: "access",
		RegisteredClaims: golangjwt.RegisteredClaims{
			Issuer:    s.cfg.JWTIssuer,
			Audience:  golangjwt.ClaimStrings{s.cfg.JWTAudience},
			ExpiresAt: golangjwt.NewNumericDate(expiresAt),
			IssuedAt:  golangjwt.NewNumericDate(now),
			NotBefore: golangjwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTAccessSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := golangjwt.ParseWithClaims(tokenString, claims, func(token *golangjwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != golangjwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTAccessSecret), nil
	})

	if err != nil {
		if errors.Is(err, golangjwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.TokenType != "access" {
		return nil, ErrTokenInvalid
	}

	if claims.Sub == 0 || claims.SessionID == "" {
		return nil, ErrTokenInvalid
	}

	iss, _ := claims.GetIssuer()
	if iss != s.cfg.JWTIssuer {
		return nil, ErrTokenInvalid
	}

	aud, _ := claims.GetAudience()
	if len(aud) == 0 || aud[0] != s.cfg.JWTAudience {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
