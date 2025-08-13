package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"strings"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecaseTestSuite struct {
	suite.Suite
	ctx                  context.Context
	mockUserRepo         *mocks.IUserRepository
	mockPasswordSvc      *mocks.IPasswordService
	mockTokenRepo        *mocks.ITokenRepository
	mockJWTService       *mocks.IJWTService
	mockEmailVerifier    *mocks.IEmailVerifier
	mockEmailSender      *mocks.IEmailSender
	mockResetRepo        *mocks.IPasswordResetRepository
	mockVerificationRepo *mocks.IVerificationRepository
	mockCloudinaryService *mocks.ICloudinaryService
	usecase              *usecases.UserUsecase
}

func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}

func (s *UserUsecaseTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.mockUserRepo = new(mocks.IUserRepository)
	s.mockPasswordSvc = new(mocks.IPasswordService)
	s.mockTokenRepo = new(mocks.ITokenRepository)
	s.mockJWTService = new(mocks.IJWTService)
	s.mockEmailVerifier = new(mocks.IEmailVerifier)
	s.mockEmailSender = new(mocks.IEmailSender)
	s.mockResetRepo = new(mocks.IPasswordResetRepository)
	s.mockVerificationRepo = new(mocks.IVerificationRepository)
	s.mockCloudinaryService = new(mocks.ICloudinaryService)

	s.usecase = usecases.NewUserUsecase(
		s.mockUserRepo,
		s.mockPasswordSvc,
		s.mockTokenRepo,
		s.mockJWTService,
		s.mockEmailVerifier,
		s.mockEmailSender,
		s.mockResetRepo,
		s.mockVerificationRepo,
		s.mockCloudinaryService,
	)
}

func (s *UserUsecaseTestSuite) TestRegisterFirstUserAsAdmin() {
	// Arrange
	testUser := userpkg.User{
		Username: "adminuser",
		Email:    "admin@example.com",
		Password: "AdminPass123!",
		Fullname: "Admin User",
	}

	// Expected user with admin role set
	expectedUser := testUser
	expectedUser.Role = "admin"

	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(0), nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(true, nil)
	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockPasswordSvc.On("HashPassword", testUser.Password).Return("hashedpassword", nil)
	s.mockUserRepo.On("CreateUser", s.ctx, mock.Anything).Run(func(args mock.Arguments) {
		userArg := args.Get(1).(userpkg.User)
		s.Equal("admin", userArg.Role)
	}).Return(expectedUser, nil)

	// Mock email sending for verification
	s.mockEmailSender.On("SendEmail", testUser.Email, "Email Verification Code", mock.MatchedBy(func(body string) bool {
		return strings.Contains(body, "Your verification OTP:")
	})).Return(nil)

	// Mock OTP hashing and verification storage
	s.mockPasswordSvc.On("HashPassword", mock.Anything).Return("hashedOTP", nil)
	s.mockVerificationRepo.On("StoreVerification", s.ctx, mock.Anything).Return(nil)

	// Act
	result, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, result.Username)
	s.Equal("admin", result.Role)
	s.Empty(result.Password) // Password should be scrubbed
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRegisterSecondUserAsNormal() {
	// Arrange
	testUser := userpkg.User{
		Username: "normaluser",
		Email:    "user@example.com",
		Password: "UserPass123!",
		Fullname: "Normal User",
	}

	// Expected user with user role set
	expectedUser := testUser
	expectedUser.Role = "user"

	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(1), nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(true, nil)
	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockPasswordSvc.On("HashPassword", testUser.Password).Return("hashedpassword", nil)
	s.mockUserRepo.On("CreateUser", s.ctx, mock.Anything).Run(func(args mock.Arguments) {
		userArg := args.Get(1).(userpkg.User)
		s.Equal("user", userArg.Role)
	}).Return(expectedUser, nil)

	// Mock email sending for verification
	s.mockEmailSender.On("SendEmail", testUser.Email, "Email Verification Code", mock.MatchedBy(func(body string) bool {
		return strings.Contains(body, "Your verification OTP:")
	})).Return(nil)

	// Mock OTP hashing and verification storage
	s.mockPasswordSvc.On("HashPassword", mock.Anything).Return("hashedOTP", nil)
	s.mockVerificationRepo.On("StoreVerification", s.ctx, mock.Anything).Return(nil)

	// Act
	result, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, result.Username)
	s.Equal("user", result.Role)
	s.Empty(result.Password)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectsInvalidEmailFormat() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "invalid-email",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("invalid email format", err.Error())
}

func (s *UserUsecaseTestSuite) TestRejectWeakPassword() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "weak",
		Fullname: "Test User",
	}

	// Mock the checks that happen before password validation
	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("password must be at least 8 chars, with upper, lower, number, and special char", err.Error())
}

func (s *UserUsecaseTestSuite) TestRejectDuplicateUsername() {
	// Arrange
	testUser := userpkg.User{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(true, nil)
	s.mockEmailVerifier.On("IsRealEmail", mock.Anything).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("username already taken", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectDuplicateEmail() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "existing@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(true, nil)
	s.mockEmailVerifier.On("IsRealEmail", mock.Anything).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("email already taken", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectEmptyFields() {
	testCases := []struct {
		name  string
		user  userpkg.User
		field string
	}{
		{"EmptyUsername", userpkg.User{Email: "test@example.com", Password: "Pass123!", Fullname: "Test"}, "username"},
		{"EmptyEmail", userpkg.User{Username: "testuser", Password: "Pass123!", Fullname: "Test"}, "email"},
		{"EmptyPassword", userpkg.User{Username: "testuser", Email: "test@example.com", Fullname: "Test"}, "password"},
		{"EmptyFullname", userpkg.User{Username: "testuser", Email: "test@example.com", Password: "Pass123!"}, "fullname"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Act
			_, err := s.usecase.RegisterUser(s.ctx, tc.user)

			// Assert
			s.Error(err)
			s.Equal("all fields are required", err.Error())
		})
	}
}

func (s *UserUsecaseTestSuite) TestFailsIfEmailVerifierErrors() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(false, errors.New("verification service down"))

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("failed to verify email: verification service down", err.Error())
	s.mockEmailVerifier.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_Success() {
	// Arrange
	login := "testuser"
	password := "ValidPass123!"
	hashedPassword := "hashedpassword"
	userID := primitive.NewObjectID()

	testUser := userpkg.User{
	   ID:         userID,
	   Username:   login,
	   Password:   hashedPassword,
	   Role:       "user",
		// mark user as verified to pass verification check
		IsVerified: true,
	}

	tokenRes := userpkg.TokenResult{
		AccessToken:      "access_token",
		RefreshToken:     "refresh_token",
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(testUser, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedPassword, password).Return(nil)
	s.mockJWTService.On("GenerateToken", userID.Hex(), login, "user").Return(tokenRes, nil)
	s.mockTokenRepo.On("StoreToken", s.ctx, mock.Anything).Return(nil)

	// Act
	user, accessToken, refreshToken, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, user.Username)
	s.Equal(tokenRes.AccessToken, accessToken)
	s.Equal(tokenRes.RefreshToken, refreshToken)
	s.Empty(user.Password)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_NotFound() {
	// Arrange
	login := "nonexistent"
	password := "password"

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(userpkg.User{}, errors.New("not found"))

	// Act
	_, _, _, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_WrongPassword() {
	// Arrange
	login := "testuser"
	password := "wrongpassword"
	hashedPassword := "hashedpassword"

	testUser := userpkg.User{
	   Username:   login,
	   Password:   hashedPassword,
		// mark user as verified to bypass verification check
		IsVerified: true,
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(testUser, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedPassword, password).Return(errors.New("mismatch"))

	// Act
	_, _, _, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_EmailNotFound() {
	// Arrange
	email := "nonexistent@example.com"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(false, nil)

	// Act
	err := s.usecase.SendResetOTP(s.ctx, email)

	// Assert
	s.Error(err)
	s.Equal("email not registered", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_Success() {
	// Arrange
	email := "user@example.com"
	hashedOTP := "hashedOTP"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(true, nil)
	s.mockEmailSender.On("SendEmail", "user@example.com", "Your OTP Code", mock.Anything).Return(nil)
	s.mockPasswordSvc.
		On("HashPassword", mock.Anything).
		Return(hashedOTP, nil)

	s.mockResetRepo.On("StoreResetRequest", s.ctx, mock.Anything).Return(nil)

	// Act
	err := s.usecase.SendResetOTP(s.ctx, email)

	// Assert
	s.NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockEmailSender.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Success() {
	// Arrange
	email := "user@example.com"
	otp := "123456"
	hashedOTP := "hashedOTP"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       hashedOTP,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedOTP, otp).Return(nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.NoError(err)
	s.mockResetRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Expired() {
	// Arrange
	email := "user@example.com"
	otp := "123456"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       "hashedOTP",
		ExpiresAt: time.Now().Add(-10 * time.Minute), // Already expired
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("OTP expired", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_MaxAttempts() {
	// Arrange
	email := "user@example.com"
	otp := "123456"

	storedReset := userpkg.PasswordReset{
		Email:        email,
		OTP:          "hashedOTP",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 5, // Max attempts
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("too many invalid attempts — OTP expired", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_InvalidOTP() {
	// Arrange
	email := "user@example.com"
	otp := "wrong123"
	hashedOTP := "hashedOTP"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       hashedOTP,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedOTP, otp).Return(errors.New("mismatch"))
	s.mockResetRepo.On("IncrementAttemptCount", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("invalid OTP", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestResetPassword_Success() {
	// Arrange
	email := "user@example.com"
	newPassword := "NewPass123!"
	hashedPassword := "hashedNewPassword"

	s.mockPasswordSvc.On("HashPassword", newPassword).Return(hashedPassword, nil)
	s.mockUserRepo.On("UpdatePasswordByEmail", s.ctx, email, hashedPassword).Return(nil)

	// Act
	err := s.usecase.ResetPassword(s.ctx, email, newPassword)

	// Assert
	s.NoError(err)
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_Success() {
	// Arrange
	refreshToken := "valid_refresh_token"
	userID := primitive.NewObjectID()
	username := "testuser"
	role := "user"

	claims := map[string]interface{}{
		"_id":      userID.Hex(),
		"username": username,
		"role":     role,
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	user := userpkg.User{
		ID:       userID,
		Username: username,
		Role:     role,
	}

	newTokens := userpkg.TokenResult{
		AccessToken:      "new_access_token",
		RefreshToken:     "new_refresh_token",
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.mockJWTService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, refreshToken).Return(storedToken, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(user, nil)
	s.mockJWTService.On("GenerateToken", userID.Hex(), username, role).Return(newTokens, nil)
	s.mockTokenRepo.On("DeleteByRefreshToken", s.ctx, refreshToken).Return(nil)
	s.mockTokenRepo.On("StoreToken", s.ctx, mock.Anything).Return(nil)

	// Act
	result, err := s.usecase.RefreshToken(s.ctx, refreshToken)

	// Assert
	s.NoError(err)
	s.Equal(newTokens.AccessToken, result.AccessToken)
	s.Equal(newTokens.RefreshToken, result.RefreshToken)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_InvalidToken() {
	// Arrange
	invalidToken := "invalid_token"

	s.mockJWTService.On("ValidateToken", invalidToken).Return(nil, errors.New("invalid token"))

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, invalidToken)

	// Assert
	s.Error(err)
	s.Equal("invalid or expired refresh token", err.Error())
	s.mockJWTService.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_TokenExpired() {
	// Arrange
	expiredToken := "expired_token"
	userID := primitive.NewObjectID()

	claims := map[string]interface{}{
		"_id": userID.Hex(),
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: expiredToken,
		ExpiresAt:    time.Now().Add(-24 * time.Hour), // Already expired
	}

	s.mockJWTService.On("ValidateToken", expiredToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, expiredToken).Return(storedToken, nil)

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, expiredToken)

	// Assert
	s.Error(err)
	s.Equal("refresh token expired", err.Error())
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_UserNotFound() {
	// Arrange
	refreshToken := "valid_token"
	userID := primitive.NewObjectID()

	claims := map[string]interface{}{
		"_id": userID.Hex(),
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	s.mockJWTService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, refreshToken).Return(storedToken, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(userpkg.User{}, errors.New("not found"))

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, refreshToken)

	// Assert
	s.Error(err)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLogout_Success() {
	// Arrange
	userID := primitive.NewObjectID().Hex()

	s.mockTokenRepo.
		On("DeleteTokensByUserID", mock.Anything, userID). // <-- Fix here
		Return(nil)

	// Act
	err := s.usecase.Logout(s.ctx, userID)

	// Assert
	s.NoError(err)
	s.mockTokenRepo.AssertCalled(s.T(), "DeleteTokensByUserID", mock.Anything, userID) // <-- Fix here
}

func (s *UserUsecaseTestSuite) TestLogout_FailureFromTokenRepo() {
	// Arrange
	userID := primitive.NewObjectID().Hex()
	expectedErr := errors.New("failed to delete tokens")

	s.mockTokenRepo.
		On("DeleteTokensByUserID", mock.Anything, userID). // <-- Fix here
		Return(expectedErr)

	// Act
	err := s.usecase.Logout(s.ctx, userID)

	// Assert
	s.EqualError(err, expectedErr.Error())
	s.mockTokenRepo.AssertCalled(s.T(), "DeleteTokensByUserID", mock.Anything, userID) // <-- Fix here
}

// TestPromoteUser_CallsRepo ensures PromoteUser calls the repository
func (s *UserUsecaseTestSuite) TestPromoteUser_CallsRepo() {
	targetID := "user123"
	actorID := "admin999"
	s.mockUserRepo.On("FindByID", s.ctx, targetID).Return(userpkg.User{Role: "user"}, nil)
	s.mockUserRepo.On("FindByID", s.ctx, actorID).Return(userpkg.User{}, nil)
	s.mockUserRepo.On("UpdateRoleAndPromoter", s.ctx, targetID, "admin", mock.AnythingOfType("*string")).Run(func(args mock.Arguments) {
		// verify pointer value matches actorID
		p := args.Get(3).(*string)
		s.Equal(actorID, *p)
	}).Return(nil)
	err := s.usecase.PromoteUser(s.ctx, targetID, actorID)
	s.NoError(err)
	s.mockUserRepo.AssertCalled(s.T(), "UpdateRoleAndPromoter", s.ctx, targetID, "admin", mock.AnythingOfType("*string"))
}

// TestDemoteUser_CallsRepo ensures DemoteUser calls the repository
func (s *UserUsecaseTestSuite) TestDemoteUser_CallsRepo() {
	targetID := "user456"
	actorID := "admin999"
	s.mockUserRepo.On("FindByID", s.ctx, targetID).Return(userpkg.User{}, nil)
	s.mockUserRepo.On("FindByID", s.ctx, actorID).Return(userpkg.User{}, nil)
	s.mockUserRepo.On("UpdateRoleAndPromoter", s.ctx, targetID, "user", (*string)(nil)).Return(nil)
	err := s.usecase.DemoteUser(s.ctx, targetID, actorID)
	s.NoError(err)
	s.mockUserRepo.AssertCalled(s.T(), "UpdateRoleAndPromoter", s.ctx, targetID, "user", (*string)(nil))
}

// TestSendVerificationOTP_Success ensures registration OTP is stored and sent
func (s *UserUsecaseTestSuite) TestSendVerificationOTP_Success() {
	email := "reg@example.com"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(true, nil)
	// capture generated otp
	s.mockEmailSender.On("SendEmail", email, "Your verification code", mock.MatchedBy(func(body string) bool {
		return strings.Contains(body, "Your code:")
	})).Return(nil)
	s.mockPasswordSvc.On("HashPassword", mock.Anything).Return("hashedOTP", nil)
	s.mockVerificationRepo.On("StoreVerification", s.ctx, mock.Anything).Return(nil)

	err := s.usecase.SendVerificationOTP(s.ctx, email)
	s.NoError(err)
	s.mockVerificationRepo.AssertCalled(s.T(), "StoreVerification", s.ctx, mock.Anything)
}

// TestVerifyUser_Success flips isVerified and deletes record
func (s *UserUsecaseTestSuite) TestVerifyUser_Success() {
	email := "reg@example.com"
	otp := "654321"
	record := userpkg.Verification{Email: email, OTP: "hashedOTP", ExpiresAt: time.Now().Add(10 * time.Minute), AttemptCount: 0}

	s.mockVerificationRepo.On("GetVerification", s.ctx, email).Return(record, nil)
	s.mockPasswordSvc.On("ComparePassword", record.OTP, otp).Return(nil)
	s.mockVerificationRepo.On("DeleteVerification", s.ctx, email).Return(nil)
	s.mockUserRepo.On("UpdateIsVerifiedByEmail", s.ctx, email, true).Return(nil)

	err := s.usecase.VerifyUser(s.ctx, email, otp)
	s.NoError(err)
	s.mockUserRepo.AssertCalled(s.T(), "UpdateIsVerifiedByEmail", s.ctx, email, true)
}

// TestVerifyUser_Expired returns error and deletes
func (s *UserUsecaseTestSuite) TestVerifyUser_Expired() {
	email := "reg@example.com"
	record := userpkg.Verification{Email: email, OTP: "hash", ExpiresAt: time.Now().Add(-1 * time.Minute), AttemptCount: 0}

	s.mockVerificationRepo.On("GetVerification", s.ctx, email).Return(record, nil)
	s.mockVerificationRepo.On("DeleteVerification", s.ctx, email).Return(nil)

	err := s.usecase.VerifyUser(s.ctx, email, "any")
	s.Error(err)
	s.Contains(err.Error(), "verification code expired")
}

// TestVerifyUser_InvalidCode increments attempt counter
func (s *UserUsecaseTestSuite) TestVerifyUser_InvalidCode() {
	email := "reg@example.com"
	record := userpkg.Verification{Email: email, OTP: "hash", ExpiresAt: time.Now().Add(10 * time.Minute), AttemptCount: 1}

	s.mockVerificationRepo.On("GetVerification", s.ctx, email).Return(record, nil)
	s.mockPasswordSvc.On("ComparePassword", record.OTP, "wrong").Return(errors.New("mismatch"))
	s.mockVerificationRepo.On("IncrementAttemptCount", s.ctx, email).Return(nil)

	err := s.usecase.VerifyUser(s.ctx, email, "wrong")
	s.Error(err)
	s.Equal("invalid code", err.Error())
	s.mockVerificationRepo.AssertCalled(s.T(), "IncrementAttemptCount", s.ctx, email)
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_Success() {
	userID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{
		Fullname:       "New Name",
		Bio:            "A new bio",
		ProfilePicture: "pic.jpg",
		ContactInfo:    userpkg.ContactInfo{Phone: "+1234567890", Website: "https://mysite.com"},
	}
	expectedUser := userpkg.User{
		ID:             primitive.NewObjectID(),
		Fullname:       updates.Fullname,
		Bio:            updates.Bio,
		ProfilePicture: updates.ProfilePicture,
		ContactInfo:    updates.ContactInfo,
	}
	s.mockUserRepo.On("UpdateProfile", s.ctx, userID, updates).Return(expectedUser, nil)

	user, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.NoError(err)
	s.Equal(updates.Fullname, user.Fullname)
	s.Equal(updates.Bio, user.Bio)
	s.Equal(updates.ProfilePicture, user.ProfilePicture)
	s.Equal(updates.ContactInfo, user.ContactInfo)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_FullnameTooShort() {
	userID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{Fullname: "A"}
	_, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.Error(err)
	s.Equal("fullname must be at least 2 characters", err.Error())
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_BioTooLong() {
	userID := primitive.NewObjectID().Hex()
	bio := strings.Repeat("a", 501)
	updates := userpkg.UpdateProfileRequest{Bio: bio}
	_, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.Error(err)
	s.Equal("bio cannot exceed 500 characters", err.Error())
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_InvalidPhone() {
	userID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{
		ContactInfo: userpkg.ContactInfo{Phone: "invalid"},
	}
	_, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.Error(err)
	s.Equal("invalid phone number format", err.Error())
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_InvalidWebsite() {
	userID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{
		ContactInfo: userpkg.ContactInfo{Website: "not-a-url"},
	}
	_, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.Error(err)
	s.Equal("invalid website URL", err.Error())
}

func (s *UserUsecaseTestSuite) TestUpdateProfile_RepoError() {
	userID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{Fullname: "Valid Name"}
	s.mockUserRepo.On("UpdateProfile", s.ctx, userID, updates).Return(userpkg.User{}, errors.New("repo error"))
	_, err := s.usecase.UpdateProfile(s.ctx, userID, updates, nil, "")
	s.Error(err)
	s.Equal("repo error", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestGetUserProfile_Success() {
	// Arrange
	userID := "valid-user-id"
	expectedUser := userpkg.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Email:    "test@example.com",
		Fullname: "Test User",
	}
	s.mockUserRepo.On("GetUserProfile", s.ctx, userID).Return(expectedUser, nil)

	// Act
	actualUser, err := s.usecase.GetUserProfile(s.ctx, userID)

	// Assert
	s.NoError(err, "Expected no error for valid user ID")
	s.Equal(expectedUser, actualUser, "Expected user details to match")
	s.mockUserRepo.AssertCalled(s.T(), "GetUserProfile", s.ctx, userID)
}

func (s *UserUsecaseTestSuite) TestGetUserProfile_NotFound() {
	// Arrange
	userID := "non-existent-user-id"
	s.mockUserRepo.On("GetUserProfile", s.ctx, userID).Return(userpkg.User{}, errors.New("user not found"))

	// Act
	actualUser, err := s.usecase.GetUserProfile(s.ctx, userID)

	// Assert
	s.Error(err, "Expected error for non-existent user ID")
	s.Contains(err.Error(), "user not found", "Expected error message to contain 'user not found'")
	s.Equal(userpkg.User{}, actualUser, "Expected empty user object for non-existent user ID")
	s.mockUserRepo.AssertCalled(s.T(), "GetUserProfile", s.ctx, userID)
}
