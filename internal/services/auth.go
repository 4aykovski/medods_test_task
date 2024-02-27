package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/4aykovksi/medods_test_task/internal/model"
	"github.com/4aykovksi/medods_test_task/internal/repository"
	"github.com/4aykovksi/medods_test_task/pkg/lib/auth"
)

type userRepository interface {
	FindByGUID(ctx context.Context, guid string) (*model.User, error)
}

type refreshSessionService interface {
	CreateRefreshSession(ctx context.Context, GUID string, token string, ttl time.Duration) error
	ValidateRefreshSession(ctx context.Context, GUID string, token string) error
}

type tokenManager interface {
	CreateTokensPair(userId string, accessTokenTtl, refreshTokenTtl time.Duration) (*auth.Tokens, error)
	Parse(inputToken string) (string, error)
}

type hasher interface {
	Hash(input string) (string, error)
	CompareHash(hash string, input string) bool
}

type AuthService struct {
	userRepo              userRepository
	refreshSessionService refreshSessionService

	tokenManager tokenManager
	hasher       hasher

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo userRepository,
	sessionService refreshSessionService,
	manager tokenManager,
	hasher hasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:              userRepo,
		refreshSessionService: sessionService,
		tokenManager:          manager,
		hasher:                hasher,
		accessTokenTTL:        accessTokenTTL,
		refreshTokenTTL:       refreshTokenTTL,
	}
}

type AuthSignInInput struct {
	GUID string
}

func (service *AuthService) SignIn(ctx context.Context, input AuthSignInInput) (*auth.Tokens, error) {
	const op = "internal.services.auth.SignIn"

	user, err := service.userRepo.FindByGUID(ctx, input.GUID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := service.getTokensPair(ctx, user.GUID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionAlreadyExists) {
			return nil, repository.ErrSessionAlreadyExists
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (service *AuthService) Refresh(ctx context.Context, base64token string) (*auth.Tokens, error) {
	const op = "internal.services.auth.Refresh"

	token, err := service.decodeBase64Token(base64token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	GUID := string(token[0 : len(token)-15])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = service.refreshSessionService.ValidateRefreshSession(ctx, GUID, string(token))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := service.getTokensPair(ctx, GUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (service *AuthService) getTokensPair(ctx context.Context, GUID string) (*auth.Tokens, error) {
	const op = "internal.services.auth.getTokensPair"

	tokens, err := service.tokenManager.CreateTokensPair(GUID, service.accessTokenTTL, service.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	hashedRefreshToken, err := service.hasher.Hash(tokens.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = service.refreshSessionService.CreateRefreshSession(ctx, GUID, hashedRefreshToken, service.refreshTokenTTL)
	if err != nil {
		if errors.Is(err, repository.ErrSessionAlreadyExists) {
			return nil, repository.ErrSessionAlreadyExists
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens.RefreshToken = base64.StdEncoding.EncodeToString([]byte(tokens.RefreshToken))

	return tokens, nil
}

func (service *AuthService) decodeBase64Token(base64token string) ([]byte, error) {
	const op = "internal.services.auth.decodeBase64Token"

	token, err := base64.StdEncoding.DecodeString(base64token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
