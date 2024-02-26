package services

import (
	"context"

	"github.com/4aykovksi/medods_test_task/internal/model"
)

type userRepository interface {
	FindByGUID(ctx context.Context, guid string) (*model.User, error)
}

type refreshSessionService interface {
}

type UserService struct {
	UserRepo              userRepository
	RefreshSessionService refreshSessionService
}

func NewUserService(
	userRepo userRepository,
	sessionService refreshSessionService,
) *UserService {
	return &UserService{
		UserRepo:              userRepo,
		RefreshSessionService: sessionService,
	}
}

type userSignInInput struct {
	GUID string
}

func (u *UserService) SignIn(ctx context.Context, input userSignInInput) error {
	const op = "internal.services.user.SignIn"

}
