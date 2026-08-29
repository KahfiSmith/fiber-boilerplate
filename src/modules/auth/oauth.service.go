package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/jwt"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/dto"
	"fiber-boilerplate/src/modules/auth/types"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	oauthStateKeyPrefix = "oauth_state:"
	oauthStateTTL       = 10 * time.Minute
)

type OAuthService struct {
	cfg          config.OAuthConfig
	authCfg      config.AuthConfig
	repo         *AuthRepository
	refreshRepo  *RefreshRepository
	tokenService *jwt.TokenService
	provider     *oidc.Provider
	oauthConfig  *oauth2.Config
}

func NewOAuthService(oauthCfg config.OAuthConfig, authCfg config.AuthConfig, repo *AuthRepository, refreshRepo *RefreshRepository, tokenService *jwt.TokenService) *OAuthService {
	return &OAuthService{
		cfg:          oauthCfg,
		authCfg:      authCfg,
		repo:         repo,
		refreshRepo:  refreshRepo,
		tokenService: tokenService,
	}
}

func (s *OAuthService) Enabled() bool {
	return s.cfg.GoogleEnabled &&
		s.cfg.GoogleClientID != "" &&
		s.cfg.GoogleClientSecret != "" &&
		s.cfg.GoogleRedirectURL != ""
}

func (s *OAuthService) initProvider() error {
	if s.provider != nil {
		return nil
	}

	discoveryURL := s.cfg.GoogleDiscoveryURL
	if discoveryURL == "" {
		discoveryURL = "https://accounts.google.com/.well-known/openid-configuration"
	}

	provider, err := oidc.NewProvider(context.Background(), discoveryURL)
	if err != nil {
		return fmt.Errorf("oidc provider discovery: %w", err)
	}

	s.provider = provider
	s.oauthConfig = &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return nil
}

func (s *OAuthService) NewState(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	key := oauthStateKeyPrefix + state
	if err := redisclient.Client.Set(ctx, key, "1", oauthStateTTL).Err(); err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}
	return state, nil
}

func (s *OAuthService) ConsumeState(ctx context.Context, state string) error {
	if state == "" {
		return exceptions.BadRequest("Missing OAuth state")
	}
	key := oauthStateKeyPrefix + state
	n, err := redisclient.Client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("consume oauth state: %w", err)
	}
	if n == 0 {
		return exceptions.Unauthorized("INVALID_OAUTH_STATE", "Invalid or expired OAuth state")
	}
	return nil
}

func (s *OAuthService) AuthURL(state string) (string, error) {
	if !s.Enabled() {
		return "", exceptions.BadRequest("Google OAuth is not enabled")
	}
	if err := s.initProvider(); err != nil {
		return "", err
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, code string) (dto.AuthResponse, string, error) {
	if code == "" {
		return dto.AuthResponse{}, "", exceptions.BadRequest("Missing authorization code")
	}
	if err := s.initProvider(); err != nil {
		return dto.AuthResponse{}, "", err
	}

	oauth2Token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_OAUTH_CODE", "Failed to exchange authorization code")
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_OAUTH_TOKEN", "No id_token in OAuth response")
	}

	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.cfg.GoogleClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_OAUTH_TOKEN", "Failed to verify id_token")
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("parse id_token claims: %w", err)
	}
	if claims.Sub == "" || claims.Email == "" {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_OAUTH_TOKEN", "id_token missing sub or email")
	}

	provider := "google"
	user, err := s.repo.FindByOAuth(provider, claims.Sub)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("find oauth user: %w", err)
	}
	if user == nil {
		user = &types.User{
			Name:            claims.Name,
			Email:           claims.Email,
			PasswordHash:    nil,
			Role:            "user",
			IsEmailVerified: true,
			OAuthProvider:   &provider,
			OAuthSubject:    &claims.Sub,
		}
		if err := s.repo.Create(user); err != nil {
			user, err = s.repo.FindByOAuth(provider, claims.Sub)
			if err != nil || user == nil {
				return dto.AuthResponse{}, "", fmt.Errorf("create oauth user: %w", err)
			}
		}
	}

	return s.createSession(ctx, user, claims.Email)
}

func (s *OAuthService) createSession(ctx context.Context, user *types.User, email string) (dto.AuthResponse, string, error) {
	sessionID := uuid.New().String()
	familyID := uuid.New().String()

	accessToken, _, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role, sessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashRefreshToken(rawRefreshToken, s.authCfg.RefreshTokenHMACKey)
	now := time.Now()

	meta := SessionMetadata{
		SessionID:        sessionID,
		FamilyID:         familyID,
		UserID:           user.ID,
		CurrentTokenHash: tokenHash,
		IssuedAt:         now,
		ExpiresAt:        now.Add(s.authCfg.RefreshTokenTTL),
		LastUsedAt:       now,
		IPAddress:        "",
		UserAgent:        "",
	}
	if err := s.refreshRepo.CreateSession(ctx, meta, s.authCfg.RefreshTokenTTL); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("create redis session: %w", err)
	}

	return dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(s.authCfg.AccessTokenTTL.Seconds()),
		User:        *user,
	}, rawRefreshToken, nil
}

var ErrOAuthNotEnabled = errors.New("google oauth is not enabled")
