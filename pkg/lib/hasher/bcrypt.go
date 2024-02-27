package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (b *BcryptHasher) Hash(input string) (string, error) {
	const op = "pkg.lib.hasher.bcrypt.Hash"

	hash, err := bcrypt.GenerateFromPassword([]byte(input), 12)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return string(hash), nil
}

func (b *BcryptHasher) CompareHash(hash string, input string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input))
	if err != nil {
		return false
	}

	return true
}
