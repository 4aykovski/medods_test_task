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

type UserRepository struct {
	db *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db: db.Collection(usersCollection),
	}
}

func (repo *UserRepository) FindByGUID(ctx context.Context, guid string) (*model.User, error) {
	const op = "internal.repository.mongorepos.user.FindByGUID"

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
