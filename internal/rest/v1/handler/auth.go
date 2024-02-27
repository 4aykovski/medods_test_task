package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/4aykovksi/medods_test_task/internal/repository"
	"github.com/4aykovksi/medods_test_task/internal/services"
	"github.com/4aykovksi/medods_test_task/pkg/lib/api/response"
	"github.com/4aykovksi/medods_test_task/pkg/lib/auth"
)

const (
	TokenNotSpecifiedMsg   = "token wasn't specified"
	WrongCredentialsMsg    = "wrong credentials"
	InternalServerErrorMsg = "internal server error"
	refreshCookieName      = "refreshToken"
	refreshCookiePath      = "/api/v1/auth"
)

type authService interface {
	SignIn(ctx context.Context, input services.AuthSignInInput) (*auth.Tokens, error)
}

type AuthHandler struct {
	authService authService
}

func NewAuthHandler(
	authService authService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type authSignInOutput struct {
	response.Response
	AccessToken  string
	RefreshToken string
}

func (h *AuthHandler) SignIn(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: add middleware to logging request processing
		log.Info("start processing request")

		w.Header().Set("Content-Type", "application/json")

		guid := r.URL.Query().Get("guid")
		if guid == "" {
			log.Info("token wasn't specified")

			w.WriteHeader(http.StatusBadRequest)
			res := response.Error(TokenNotSpecifiedMsg)
			jsonRes, err := json.Marshal(res)
			if err != nil {
				log.Error("internal error on marshalling response", slog.String("err", err.Error()))

				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(jsonRes)
			return
		}

		tokens, err := h.authService.SignIn(r.Context(), services.AuthSignInInput{GUID: guid})
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				log.Info("wrong credentials")

				w.WriteHeader(http.StatusProxyAuthRequired)
				res := response.Error(WrongCredentialsMsg)
				jsonRes, err := json.Marshal(res)
				if err != nil {
					log.Error("internal error on marshalling response", slog.String("err", err.Error()))

					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(jsonRes)
				return
			}

			log.Error("internal error on authService.SignIn", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusInternalServerError)
			res := response.Error(InternalServerErrorMsg)
			jsonRes, err := json.Marshal(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(jsonRes)
			return
		}

		refreshCookie := h.newRefreshCookie(tokens.RefreshToken, tokens.ExpiresIn)
		http.SetCookie(w, refreshCookie)

		log.Info("successfully signed in", slog.String("login", guid))

		res := authSignInOutput{
			Response:     response.OK(),
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}
		jsonRes, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(jsonRes)
		return
	}
}

func (h *AuthHandler) Refresh(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}

}

func (h *AuthHandler) newRefreshCookie(refreshToken string, time time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Expires:  time,
		Path:     refreshCookiePath,
		Secure:   false,
		HttpOnly: true,
		SameSite: 3,
	}
}
