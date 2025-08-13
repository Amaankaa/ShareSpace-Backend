package usecases

import (
	"context"
	"errors"
	"mime/multipart"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Domain/services"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	utils "github.com/Amaankaa/Blog-Starter-Project/Domain/utils"
)

type UserUsecase struct {
	userRepo          userpkg.IUserRepository
	passwordSvc       userpkg.IPasswordService
	tokenRepo         userpkg.ITokenRepository
	emailVerifier     services.IEmailVerifier
	emailSender       services.IEmailSender
	jwtService        userpkg.IJWTService
	passwordResetRepo userpkg.IPasswordResetRepository
	verificationRepo  userpkg.IVerificationRepository
	cloudinaryService userpkg.ICloudinaryService
}

func NewUserUsecase(
	userRepo userpkg.IUserRepository,
	passwordSvc userpkg.IPasswordService,
	tokenRepo userpkg.ITokenRepository,
	jwtService userpkg.IJWTService,
	emailVerifier services.IEmailVerifier,
	emailSender services.IEmailSender,
	passwordResetRepo userpkg.IPasswordResetRepository,
	verificationRepo userpkg.IVerificationRepository,
	cloudinaryService userpkg.ICloudinaryService,
) *UserUsecase {
	return &UserUsecase{
		userRepo:          userRepo,
		passwordSvc:       passwordSvc,
		tokenRepo:         tokenRepo,
		jwtService:        jwtService,
		emailVerifier:     emailVerifier,
		emailSender:       emailSender,
		passwordResetRepo: passwordResetRepo,
		verificationRepo:  verificationRepo,
		cloudinaryService: cloudinaryService,
	}
}

func (uu *UserUsecase) RegisterUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	// Basic field validation
	if user.Username == "" || user.Email == "" || user.Password == "" || user.Fullname == "" {
		return userpkg.User{}, errors.New("all fields are required")
	}

	// Email format
	if !utils.IsValidEmail(user.Email) {
		return userpkg.User{}, errors.New("invalid email format")
	}

	// Username and email uniqueness
	exists, err := uu.userRepo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return userpkg.User{}, errors.New("failed to check username existence: " + err.Error())
	}
	if exists {
		return userpkg.User{}, errors.New("username already taken")
	}
	
	exists, err = uu.userRepo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return userpkg.User{}, errors.New("failed to check email existence: " + err.Error())
	}
	if exists {
		return userpkg.User{}, errors.New("email already taken")
	}
	
	// Real email check
	isReal, err := uu.emailVerifier.IsRealEmail(user.Email)
	if err != nil {
		return userpkg.User{}, errors.New("failed to verify email: " + err.Error())
	}
	if !isReal {
		return userpkg.User{}, errors.New("email is unreachable")
	}

	// Strong password check
	if !utils.IsStrongPassword(user.Password) {
		return userpkg.User{}, errors.New("password must be at least 8 chars, with upper, lower, number, and special char")
	}


	// Assign role
	count, err := uu.userRepo.CountUsers(ctx)
	if err != nil {
		return userpkg.User{}, err
	}
	if count == 0 {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}

	// Password hashing
	hashed, err := uu.passwordSvc.HashPassword(user.Password)
	if err != nil {
		return userpkg.User{}, err
	}
	user.Password = hashed

	// Create user (with isVerified = false by default)
	createdUser, err := uu.userRepo.CreateUser(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	// Generate and send verification OTP
	otp := utils.GenerateOTP(6)

	// Send verification email
	err = uu.emailSender.SendEmail(user.Email, "Email Verification Code", "Your verification OTP: "+otp)
	if err != nil {
		return userpkg.User{}, errors.New("failed to send verification code")
	}

	// Hash OTP before storing
	hashedOTP, err := uu.passwordSvc.HashPassword(otp)
	if err != nil {
		return userpkg.User{}, errors.New("failed to process verification code")
	}

	// Store verification request
	verification := userpkg.Verification{
		Email:        user.Email,
		OTP:          hashedOTP,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 0,
	}

	err = uu.verificationRepo.StoreVerification(ctx, verification)
	if err != nil {
		return userpkg.User{}, errors.New("failed to store verification code")
	}

	createdUser.Password = "" // scrub before return
	return createdUser, nil
}

func (uu *UserUsecase) LoginUser(ctx context.Context, login, password string) (userpkg.User, string, string, error) {
	user, err := uu.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return userpkg.User{}, "", "", errors.New("invalid credentials")
	}
	// Prevent login if email not verified
	if !user.IsVerified {
		return userpkg.User{}, "", "", errors.New("email not verified")
	}

	if err := uu.passwordSvc.ComparePassword(user.Password, password); err != nil {
		return userpkg.User{}, "", "", errors.New("invalid credentials")
	}

	// Generate tokens
	tokenRes, err := uu.jwtService.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return userpkg.User{}, "", "", err
	}

	// Store tokens
	err = uu.tokenRepo.StoreToken(ctx, userpkg.Token{
		UserID:       user.ID,
		AccessToken:  tokenRes.AccessToken,
		RefreshToken: tokenRes.RefreshToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    tokenRes.RefreshExpiresAt,
	})
	if err != nil {
		return userpkg.User{}, "", "", err
	}

	user.Password = ""
	return user, tokenRes.AccessToken, tokenRes.RefreshToken, nil
}

func (uu *UserUsecase) RefreshToken(ctx context.Context, refreshToken string) (userpkg.TokenResult, error) {
	claims, err := uu.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return userpkg.TokenResult{}, errors.New("invalid or expired refresh token")
	}

	userID, ok := claims["_id"].(string)
	if !ok {
		return userpkg.TokenResult{}, errors.New("invalid token payload")
	}

	// Check if token is stored in DB
	stored, err := uu.tokenRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return userpkg.TokenResult{}, errors.New("refresh token not recognized")
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return userpkg.TokenResult{}, errors.New("refresh token expired")
	}

	// Fetch user info (optional, for roles/username)
	user, err := uu.userRepo.FindByID(ctx, userID)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	// Generate new tokens
	tokens, err := uu.jwtService.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	// Store new refresh token, remove old
	_ = uu.tokenRepo.DeleteByRefreshToken(ctx, refreshToken)
	_ = uu.tokenRepo.StoreToken(ctx, userpkg.Token{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.RefreshExpiresAt,
		CreatedAt:    time.Now(),
	})

	return tokens, nil
}

func (u *UserUsecase) SendResetOTP(ctx context.Context, email string) error {
	exists, _ := u.userRepo.ExistsByEmail(ctx, email)
	if !exists {
		return errors.New("email not registered")
	}

	otp := utils.GenerateOTP(6)

	err := u.emailSender.SendEmail(email, "Your OTP Code", "Your OTP: "+otp)
	if err != nil {
		return err
	}

	//Hash OTP before storing
	hashedOTP, err := u.passwordSvc.HashPassword(otp)
	if err != nil {
		return err
	}

	reset := userpkg.PasswordReset{
		Email:        email,
		OTP:          hashedOTP,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 0,
	}
	return u.passwordResetRepo.StoreResetRequest(ctx, reset)
}

func (u *UserUsecase) VerifyOTP(ctx context.Context, email, otp string) error {
	stored, err := u.passwordResetRepo.GetResetRequest(ctx, email)
	if err != nil {
		return errors.New("no reset request found")
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
		return errors.New("OTP expired")
	}

	if stored.AttemptCount >= 5 {
		_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
		return errors.New("too many invalid attempts — OTP expired")
	}

	if u.passwordSvc.ComparePassword(stored.OTP, otp) != nil {
		// increment attempt count
		_ = u.passwordResetRepo.IncrementAttemptCount(ctx, email)
		return errors.New("invalid OTP")
	}

	// OTP is valid — delete it
	_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
	return nil
}

func (u *UserUsecase) ResetPassword(ctx context.Context, email, newPassword string) error {
	hashed, err := u.passwordSvc.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return u.userRepo.UpdatePasswordByEmail(ctx, email, hashed)
}

func (u *UserUsecase) Logout(ctx context.Context, userID string) error {
	return u.tokenRepo.DeleteTokensByUserID(ctx, userID)
}

func (uu *UserUsecase) PromoteUser(ctx context.Context, targetUserID string, actorUserID string) error {
	if targetUserID == actorUserID {
		return errors.New("cannot promote yourself")
	}
	// Load users
	target, err := uu.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	actor, err := uu.userRepo.FindByID(ctx, actorUserID)
	if err != nil {
		return err
	}
	// Block acting on your promoter
	if !actor.PromotedBy.IsZero() && actor.PromotedBy.Hex() == target.ID.Hex() {
		return errors.New("cannot act on your promoter")
	}
	if target.Role == "admin" {
		return nil
	}
	if err := uu.userRepo.UpdateRoleAndPromoter(ctx, targetUserID, "admin", &actorUserID); err != nil {
		return err
	}
	return nil
}

func (uu *UserUsecase) DemoteUser(ctx context.Context, targetUserID string, actorUserID string) error {
	// Load users
	target, err := uu.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	actor, err := uu.userRepo.FindByID(ctx, actorUserID)
	if err != nil {
		return err
	}
	// Block acting on your promoter
	if !actor.PromotedBy.IsZero() && actor.PromotedBy.Hex() == target.ID.Hex() {
		return errors.New("cannot act on your promoter")
	}
	if err := uu.userRepo.UpdateRoleAndPromoter(ctx, targetUserID, "user", nil); err != nil {
		return err
	}
	return nil
}

func (u *UserUsecase) SendVerificationOTP(ctx context.Context, email string) error {
	exists, _ := u.userRepo.ExistsByEmail(ctx, email)
	if !exists {
		return errors.New("email not registered")
	}

	otp := utils.GenerateOTP(6)
	if err := u.emailSender.SendEmail(email, "Your verification code", "Your code: "+otp); err != nil {
		return err
	}

	hashed, err := u.passwordSvc.HashPassword(otp)
	if err != nil {
		return err
	}

	v := userpkg.Verification{
		Email:        email,
		OTP:          hashed,
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		AttemptCount: 0,
	}
	return u.verificationRepo.StoreVerification(ctx, v)
}

func (u *UserUsecase) VerifyUser(ctx context.Context, email, otp string) error {
	v, err := u.verificationRepo.GetVerification(ctx, email)
	if err != nil {
		return errors.New("no verification found")
	}
	if time.Now().After(v.ExpiresAt) {
		_ = u.verificationRepo.DeleteVerification(ctx, email)
		return errors.New("verification code expired")
	}
	if v.AttemptCount >= 5 {
		_ = u.verificationRepo.DeleteVerification(ctx, email)
		return errors.New("too many invalid attempts")
	}
	if u.passwordSvc.ComparePassword(v.OTP, otp) != nil {
		_ = u.verificationRepo.IncrementAttemptCount(ctx, email)
		return errors.New("invalid code")
	}
	// flip user verified
	if err := u.userRepo.UpdateIsVerifiedByEmail(ctx, email, true); err != nil {
		return err
	}
	_ = u.verificationRepo.DeleteVerification(ctx, email)
	return nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID string, updates userpkg.UpdateProfileRequest, file multipart.File, filename string) (userpkg.User, error) {

	if updates.Fullname != "" && len(updates.Fullname) < 2 {
        return userpkg.User{}, errors.New("fullname must be at least 2 characters")
    }
    if updates.Bio != "" && len(updates.Bio) > 500 {
        return userpkg.User{}, errors.New("bio cannot exceed 500 characters")
    }
    if updates.ContactInfo.Phone != "" && !IsValidPhone(updates.ContactInfo.Phone) {
        return userpkg.User{}, errors.New("invalid phone number format")
    }
    if updates.ContactInfo.Website != "" && !IsValidURL(updates.ContactInfo.Website) {
		return userpkg.User{}, errors.New("invalid website URL")
    }

	if file != nil && filename != "" {
		imageURL, err := u.cloudinaryService.UploadImage(ctx, file, filename)
		if err != nil {
			return userpkg.User{}, err
		}
		updates.ProfilePicture = imageURL
	}
    
	return u.userRepo.UpdateProfile(ctx, userID, updates)
}

// Helper function
func IsValidPhone(phone string) bool {
    phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
    return phoneRegex.MatchString(phone)
}

func IsValidURL(rawurl string) bool {
	_, err := url.Parse(rawurl)
	return err == nil && (strings.HasPrefix(rawurl, "http://") || strings.HasPrefix(rawurl, "https://"))
}

func (u *UserUsecase) GetUserProfile(ctx context.Context, userID string) (userpkg.User, error) {
	return u.userRepo.GetUserProfile(ctx, userID)
}
