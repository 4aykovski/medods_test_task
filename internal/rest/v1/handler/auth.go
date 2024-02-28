package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/4aykovksi/medods_test_task/internal/repository"
	"github.com/4aykovksi/medods_test_task/internal/services"
	"github.com/4aykovksi/medods_test_task/pkg/lib/api/response"
	"github.com/4aykovksi/medods_test_task/pkg/lib/auth"
)

const (
	GuidNotSpecifiedMsg    = "guid wasn't specified"
	TokenNotSpecifiedMsg   = "refresh token is not specified"
	WrongCredentialsMsg    = "wrong credentials"
	InternalServerErrorMsg = "internal server error"
	refreshCookieName      = "refreshToken"
	refreshCookiePath      = "/api/v1/auth"
)

type authService interface {
	SignIn(ctx context.Context, input services.AuthSignInInput) (*auth.Tokens, error)
	Refresh(ctx context.Context, base64token string) (*auth.Tokens, error)
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
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// SignIn handles sign in requests
// 200 - OK. response contains access and refresh tokens, refresh cookies is set
// 400 - guid is not specified.
// 407 - can't find given guid in database
// 500 - various internal server errors
func (h *AuthHandler) SignIn(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "rest.v1.handler.auth.SignIn"

		// TODO: add middleware to logging request processing
		log := log.With(slog.String("op", op))
		log.Info("start processing request")

		w.Header().Set("Content-Type", "application/json")

		guid := r.URL.Query().Get("guid")
		if guid == "" {
			log.Info("token wasn't specified")

			sendErrorResponse(log, w, GuidNotSpecifiedMsg, http.StatusBadRequest)
			return
		}

		tokens, err := h.authService.SignIn(r.Context(), services.AuthSignInInput{GUID: guid})
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				log.Info("wrong credentials")

				sendErrorResponse(log, w, WrongCredentialsMsg, http.StatusProxyAuthRequired)
				return
			}

			log.Error("internal error on authService.SignIn", slog.String("err", err.Error()))

			sendErrorResponse(log, w, InternalServerErrorMsg, http.StatusInternalServerError)
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

type authRefreshInput struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type authRefreshOutput struct {
	response.Response
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Refresh handles refresh request
// 200 - OK. response contains access and refresh tokens, refresh cookies is set
// 400 - refresh token is not specified.
// 407 - refresh token wasn't find or it's not valid
// 500 - various internal server errors
func (h *AuthHandler) Refresh(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "rest.v1.handler.auth.Refresh"

		// TODO: add middleware to logging request processing
		log := log.With(slog.String("op", op))
		log.Info("start processing request")

		w.Header().Set("Content-Type", "application/json")

		var token string
		cookie, err := r.Cookie(refreshCookieName)
		if err != nil {
			var req authRefreshInput
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error("can't decode request body", slog.String("err", err.Error()))

				sendErrorResponse(log, w, InternalServerErrorMsg, http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(body, &req)
			if err != nil || req.RefreshToken == "" {
				log.Info("refresh_cookie is not specified")

				sendErrorResponse(log, w, TokenNotSpecifiedMsg, http.StatusBadRequest)
				return
			}
			token = req.RefreshToken
		} else {
			token = cookie.Value
		}

		tokens, err := h.authService.Refresh(r.Context(), token)
		if err != nil {
			if errors.Is(err, services.ErrWrongCred) {
				log.Info("wrong credentials")

				sendErrorResponse(log, w, WrongCredentialsMsg, http.StatusProxyAuthRequired)
				return
			}
			log.Error("internal server error", slog.String("err", err.Error()))

			sendErrorResponse(log, w, InternalServerErrorMsg, http.StatusInternalServerError)
			return
		}

		refreshCookie := h.newRefreshCookie(tokens.RefreshToken, tokens.ExpiresIn)
		http.SetCookie(w, refreshCookie)

		log.Info("successfully signed in")

		res := authRefreshOutput{
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

// sendErrorResponse send response with given statusCode as http status code and msg in body
func sendErrorResponse(log *slog.Logger, w http.ResponseWriter, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	res := response.Error(msg)
	jsonRes, err := json.Marshal(res)
	if err != nil {
		log.Error("internal error on marshalling response", slog.String("err", err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonRes)
	return
}
