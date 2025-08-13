package services

type IEmailSender interface {
	SendEmail(to, subject, content string) error
}