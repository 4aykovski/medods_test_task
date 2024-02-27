package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/4aykovksi/medods_test_task/internal/rest/v1/handler"
	"github.com/4aykovksi/medods_test_task/internal/services"
	"github.com/4aykovksi/medods_test_task/pkg/lib/auth"
)

type authService interface {
	SignIn(ctx context.Context, input services.AuthSignInInput) (*auth.Tokens, error)
	Refresh(ctx context.Context, base64token string) (*auth.Tokens, error)
}

func NewRouter(
	log *slog.Logger,
	authService authService,
) *http.ServeMux {

	var (
		mux         = http.NewServeMux()
		authHandler = handler.NewAuthHandler(authService)
	)

	mux.HandleFunc("/api/v1/auth/signIn", authHandler.SignIn(log))
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh(log))

	return mux
}
