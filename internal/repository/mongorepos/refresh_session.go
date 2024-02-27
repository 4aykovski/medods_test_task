package mongorepos

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aykovksi/medods_test_task/internal/model"
	"github.com/4aykovksi/medods_test_task/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshSessionRepository struct {
	db *mongo.Collection
}

func NewRefreshSessionsRepository(db *mongo.Database) *RefreshSessionRepository {
	return &RefreshSessionRepository{
		db: db.Collection(refreshSessionCollection),
	}
}

func (repo *RefreshSessionRepository) Insert(ctx context.Context, session model.RefreshSession) error {
	const op = "internal.repository.mongorepos.refresh_session.Insert"

	_, err := repo.db.InsertOne(ctx, session)
	if err != nil {
		var mongoErr mongo.WriteError
		if errors.As(err, &mongoErr) {
			if mongoErr.Code == 11000 {
				return repository.ErrSessionAlreadyExists
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *RefreshSessionRepository) FindByGUID(ctx context.Context, GUID string) (*model.RefreshSession, error) {
	const op = "internal.repository.mongorepos.refresh_session.FindByToken"

	filter := bson.D{{"guid", GUID}}

	var session model.RefreshSession
	err := repo.db.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrSessionNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &session, nil
}

func (repo *RefreshSessionRepository) FindAllUserSessions(ctx context.Context, GUID string) ([]model.RefreshSession, error) {
	const op = "internal.repository.mongorepos.refresh_session.FindAllUserSessions"

	filter := bson.D{{"guid", GUID}}

	var sessions []model.RefreshSession
	cursor, err := repo.db.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if cursor.RemainingBatchLength() == 0 {
		return nil, repository.ErrUserSessionsNotFound
	}

	err = cursor.All(ctx, &sessions)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, err
}

func (repo *RefreshSessionRepository) DeleteByToken(ctx context.Context, token string) error {
	const op = "internal.repository.mongorepos.refresh_session.DeleteByToken"

	filter := bson.D{{"refresh_token", token}}

	result, err := repo.db.DeleteOne(ctx, filter)
	if result.DeletedCount == 0 {
		return repository.ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
