package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Manager struct {
	secret string
}

func NewManager(secret string) *Manager {
	return &Manager{
		secret: secret,
	}
}

func (m *Manager) CreateTokensPair(userId string, accessTokenTtl, refreshTokenTtl time.Duration) (*Tokens, error) {
	const op = "pkg.lib.auth.token_manager.CreateTokensPair"

	accessToken, err := m.newJWT(userId, accessTokenTtl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := m.newRefreshToken(userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(refreshTokenTtl),
	}, nil
}

func (m *Manager) Parse(inputToken string) (string, error) {
	const op = "lib.token-manager.token_manager.Parse"

	token, err := jwt.Parse(inputToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("%s: can't get user claims from token", op)
	}

	return claims["sub"].(string), nil
}

func (m *Manager) newJWT(userId string, ttl time.Duration) (string, error) {
	const op = "pkg.lib.auth.token_manager.newJWT"

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   userId,
	})

	completeToken, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return completeToken, nil
}

func (m *Manager) newRefreshToken(userId string) (string, error) {
	const op = "pkg.lib.auth.token_manager.newRefreshToken"

	b := make([]byte, 7)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%s-%x", userId, b), nil
}
