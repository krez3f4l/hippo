package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(pwd string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), h.cost)
	return string(hashedBytes), err
}

func (h *BcryptHasher) Compare(hash, pwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
}
