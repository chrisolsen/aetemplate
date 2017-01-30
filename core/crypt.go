package core

import (
	"golang.org/x/crypto/bcrypt"
)

// Crypt .
type Crypt struct{}

// Encrypt converts the raw password to a brcypt hash
func (c *Crypt) Encrypt(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(b), err
}

// Validate checks that the saved hash and raw password hash match
func (c *Crypt) Validate(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
