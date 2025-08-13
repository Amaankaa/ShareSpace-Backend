package userpkg

import "context"

// IUserRepository defines user data access operations
type IUserRepository interface {
	FindByID(ctx context.Context, userID string) (User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
	UpdatePasswordByEmail(ctx context.Context, email, hashedPassword string) error
	UpdateUserRoleByID(ctx context.Context, userID, role string) error
	UpdateIsVerifiedByEmail(ctx context.Context, email string, verified bool) error
	UpdateProfile(ctx context.Context, userID string, updates UpdateProfileRequest) (User, error)
    GetUserProfile(ctx context.Context, userID string) (User, error)
	UpdateRoleAndPromoter(ctx context.Context, userID string, role string, promoterID *string) error
}

type ITokenRepository interface {
	StoreToken(ctx context.Context, token Token) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (Token, error)
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteTokensByUserID(ctx context.Context, userID string) error
}

type IPasswordResetRepository interface {
	StoreResetRequest(ctx context.Context, reset PasswordReset) error
	GetResetRequest(ctx context.Context, email string) (PasswordReset, error)
	DeleteResetRequest(ctx context.Context, email string) error
	IncrementAttemptCount(ctx context.Context, email string) error
}

type IVerificationRepository interface {
	StoreVerification(ctx context.Context, v Verification) error
	GetVerification(ctx context.Context, email string) (Verification, error)
	DeleteVerification(ctx context.Context, email string) error
	IncrementAttemptCount(ctx context.Context, email string) error
}
