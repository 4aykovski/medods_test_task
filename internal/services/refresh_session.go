package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/4aykovksi/medods_test_task/internal/model"
	"github.com/4aykovksi/medods_test_task/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrWrongCred = errors.New("wrong credentials")
)

type refreshSessionRepository interface {
	Insert(ctx context.Context, session model.RefreshSession) error
	DeleteByToken(ctx context.Context, token string) error
	FindByGUID(ctx context.Context, token string) (*model.RefreshSession, error)
	FindAllUserSessions(ctx context.Context, GUID string) ([]model.RefreshSession, error)
}

type RefreshSessionService struct {
	refreshSessionRepo refreshSessionRepository

	hasher hasher

	maxSessionCount int
}

func NewRefreshSessionService(
	repository refreshSessionRepository,
	hasher hasher,
	maxSessionCount int,
) *RefreshSessionService {
	return &RefreshSessionService{
		refreshSessionRepo: repository,
		hasher:             hasher,
		maxSessionCount:    maxSessionCount,
	}
}

func (service *RefreshSessionService) CreateRefreshSession(ctx context.Context, GUID string, token string, ttl time.Duration) error {
	const op = "internal.services.refresh_session.CreateRefreshSession"

	err := service.refreshSessionRepo.DeleteByToken(ctx, token)
	if err != nil && !errors.Is(err, repository.ErrSessionNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	} else if errors.Is(err, repository.ErrSessionNotFound) {
		err = service.deleteExcessSession(ctx, GUID)
		if err != nil && !errors.Is(err, repository.ErrUserSessionsNotFound) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	session := model.RefreshSession{
		RefreshToken: token,
		GUID:         GUID,
		ExpiresIn:    primitive.NewDateTimeFromTime(time.Now().Add(ttl)),
	}

	err = service.refreshSessionRepo.Insert(ctx, session)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (service *RefreshSessionService) ValidateRefreshSession(ctx context.Context, GUID string, token string) error {
	const op = "internal.services.refresh_session.ValidateRefreshSession"

	session, err := service.refreshSessionRepo.FindByGUID(ctx, GUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ok := service.hasher.CompareHash(session.RefreshToken, token)
	if !ok {
		return ErrWrongCred
	}

	err = service.refreshSessionRepo.DeleteByToken(ctx, token)
	if err != nil && !errors.Is(err, repository.ErrSessionNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	ok = service.isSessionExpired(session)
	if !ok {
		return ErrWrongCred
	}

	return nil
}

func (service *RefreshSessionService) deleteExcessSession(ctx context.Context, GUID string) error {
	const op = "internal.services.refresh_session.deleteExcessSession"

	sessions, err := service.refreshSessionRepo.FindAllUserSessions(ctx, GUID)
	if err != nil && !errors.Is(err, repository.ErrUserSessionsNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(sessions) > service.maxSessionCount {
		sort.Slice(sessions, func(i, j int) bool { return sessions[i].ExpiresIn.Time().Before(sessions[j].ExpiresIn.Time()) })

		err = service.refreshSessionRepo.DeleteByToken(ctx, sessions[0].RefreshToken)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (service *RefreshSessionService) isSessionExpired(session *model.RefreshSession) bool {
	return session.ExpiresIn.Time().After(time.Now())
}
