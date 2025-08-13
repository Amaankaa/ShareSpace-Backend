package userpkg

import (
	"context"
	"mime/multipart"
)

type IUserUsecase interface {
	RegisterUser(ctx context.Context, user User) (User, error)
	Logout(ctx context.Context, userID string) error
	LoginUser(ctx context.Context, login string, password string) (User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (TokenResult, error)
	SendResetOTP(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email, otp string) error
	ResetPassword(ctx context.Context, email, newPassword string) error
	PromoteUser(ctx context.Context, targetUserID string, actorUserID string) error
	DemoteUser(ctx context.Context, targetUserID string, actorUserID string) error
	SendVerificationOTP(ctx context.Context, email string) error
	VerifyUser(ctx context.Context, email, otp string) error
	UpdateProfile(ctx context.Context, userID string, updates UpdateProfileRequest, file multipart.File, filename string) (User, error)
    GetUserProfile(ctx context.Context, userID string) (User, error)
}

// User Infrastructure interfaces
type IJWTService interface {
	GenerateToken(userID, username, role string) (TokenResult, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

// PasswordService interface defines password operations
type IPasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

type ICloudinaryService interface {
    UploadImage(ctx context.Context, file multipart.File, filename string) (string, error)
}
