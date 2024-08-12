package helper

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the provided password using bcrypt with a cost factor of 20.
// It returns the hashed password as a byte slice or an error if hashing fails.
func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error("Failed to generate hash of password")
		return nil, err
	}

	return hash, nil
}

// VerifyPassword compares a hashed password with a provided password to check if they match.
func VerifyPassword(hashedPassword string, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))

	return err == nil
}
