package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/4aykovksi/medods_test_task/internal/model"
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

type UserService struct {
	userRepo              userRepository
	refreshSessionService refreshSessionService

	tokenManager tokenManager
	hasher       hasher

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewUserService(
	userRepo userRepository,
	sessionService refreshSessionService,
	manager tokenManager,
	hasher hasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:              userRepo,
		refreshSessionService: sessionService,
		tokenManager:          manager,
		hasher:                hasher,
		accessTokenTTL:        accessTokenTTL,
		refreshTokenTTL:       refreshTokenTTL,
	}
}

type userSignInInput struct {
	GUID string
}

func (service *UserService) SignIn(ctx context.Context, input userSignInInput) (*auth.Tokens, error) {
	const op = "internal.services.user.SignIn"

	user, err := service.userRepo.FindByGUID(ctx, input.GUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := service.getTokensPair(ctx, user.GUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (service *UserService) Refresh(ctx context.Context, base64token string) (*auth.Tokens, error) {
	const op = "internal.services.user.Refresh"

	token, err := service.decodeBase64Token(base64token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	GUID, err := service.tokenManager.Parse(token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = service.refreshSessionService.ValidateRefreshSession(ctx, GUID, token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := service.getTokensPair(ctx, GUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (service *UserService) getTokensPair(ctx context.Context, GUID string) (*auth.Tokens, error) {
	const op = "internal.services.user.getTokensPair"

	tokens, err := service.tokenManager.CreateTokensPair(GUID, service.accessTokenTTL, service.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	hashedRefreshToken, err := service.hasher.Hash(tokens.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = service.refreshSessionService.CreateRefreshSession(ctx, hashedRefreshToken, GUID, service.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens.RefreshToken = base64.StdEncoding.EncodeToString([]byte(tokens.RefreshToken))

	return tokens, nil
}

func (service *UserService) decodeBase64Token(base64token string) (string, error) {
	const op = "internal.services.user.decodeBase64Token"

	token, err := base64.StdEncoding.DecodeString(base64token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return string(token), nil
}
