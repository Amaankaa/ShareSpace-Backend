package services

type IEmailVerifier interface {
    IsRealEmail(email string) (bool, error)
}