package domain

import (
	"math/rand"
	"time"
)

func GenerateOTP(length int) string {
	const charset = "0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // local generator

	otp := make([]byte, length)
	for i := range otp {
		otp[i] = charset[r.Intn(len(charset))]
	}
	return string(otp)
}