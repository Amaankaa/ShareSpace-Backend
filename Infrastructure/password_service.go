package infrastructure

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct{}

func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

func (ps *PasswordService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (ps *PasswordService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}