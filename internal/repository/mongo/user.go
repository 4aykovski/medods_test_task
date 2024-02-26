package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aykovksi/medods_test_task/internal/model"
	"github.com/4aykovksi/medods_test_task/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db: db.Collection(usersCollection),
	}
}

func (repo *UserRepository) Insert(ctx context.Context, user *model.User) error {
	const op = "internal.repository.mongo.user.Insert"

	_, err := repo.db.InsertOne(ctx, user)
	if err != nil {
		var mongoErr mongo.WriteError
		if errors.As(err, &mongoErr) {
			if mongoErr.Code == 11000 {
				return repository.ErrUserAlreadyExists
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UserRepository) FindByGUID(ctx context.Context, guid string) (*model.User, error) {
	const op = "internal.repository.mongo.user.FindByGUID"

	filter := bson.D{{"guid", guid}}

	var user model.User
	err := repo.db.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
