package auth

import "time"

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Time
}
