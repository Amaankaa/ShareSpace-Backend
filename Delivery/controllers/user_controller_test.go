package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	mock_user "github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	router *gin.Engine
	mockUC *mock_user.IUserUsecase
}

func (s *ControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockUC = new(mock_user.IUserUsecase)
	ctrl := controllers.NewController(s.mockUC)
	s.router = gin.Default()
	s.router.POST("/register", ctrl.Register)
	// Registration verification endpoint
	s.router.POST("/verify-user", ctrl.VerifyUser)
	s.router.POST("/login", ctrl.Login)
	s.router.POST("/forgot-password", ctrl.ForgotPassword)
	s.router.POST("/refresh", ctrl.RefreshToken)
	s.router.POST("/verify-otp", ctrl.VerifyOTP)
	s.router.POST("/reset-password", ctrl.ResetPassword)
	// Inject a dummy user_id into context to simulate authenticated admin
	addActor := func(c *gin.Context) {
		c.Set("user_id", "admin999")
		c.Next()
	}
	s.router.PUT("/user/:id/promote", addActor, ctrl.PromoteUser)
	s.router.PUT("/user/:id/demote", addActor, ctrl.DemoteUser)
}

func (s *ControllerTestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var b []byte
	var err error
	if body != nil {
		b, err = json.Marshal(body)
	}

	if err != nil {
		s.FailNow("Failed to marshal body", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// === TESTS ===

func (s *ControllerTestSuite) TestRefreshToken_Success() {
	expected := userpkg.TokenResult{
		AccessToken:      "new-access",
		RefreshToken:     "new-refresh",
		AccessExpiresAt:  time.Now().Add(1 * time.Hour),
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mockUC.On("RefreshToken", mock.Anything, "valid-refresh-token").
		Return(expected, nil)

	w := s.performRequest("POST", "/refresh", map[string]string{"refresh_token": "valid-refresh-token"})

	s.Equal(http.StatusOK, w.Code)
	var res userpkg.TokenResult
	s.NoError(json.Unmarshal(w.Body.Bytes(), &res))
	s.Equal(expected.AccessToken, res.AccessToken)
	s.Equal(expected.RefreshToken, res.RefreshToken)
	s.False(res.AccessExpiresAt.IsZero())
	s.False(res.RefreshExpiresAt.IsZero())
}

func (s *ControllerTestSuite) TestRefreshToken_MissingToken() {
	w := s.performRequest("POST", "/refresh", map[string]string{})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestRefreshToken_InvalidToken() {
	s.mockUC.On("RefreshToken", mock.Anything, "bad-token").
		Return(userpkg.TokenResult{}, errors.New("unauthorized"))

	w := s.performRequest("POST", "/refresh", map[string]string{"refresh_token": "bad-token"})
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestVerifyOTP_Success() {
	s.mockUC.On("VerifyOTP", mock.Anything, "test@example.com", "123456").Return(nil)
	w := s.performRequest("POST", "/verify-otp", map[string]string{"email": "test@example.com", "otp": "123456"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ControllerTestSuite) TestVerifyOTP_InvalidJSON() {
	req := httptest.NewRequest("POST", "/verify-otp", bytes.NewBuffer([]byte(`{"email":`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestVerifyOTP_WrongOTP() {
	s.mockUC.On("VerifyOTP", mock.Anything, "test@example.com", "wrong").Return(errors.New("invalid otp"))
	w := s.performRequest("POST", "/verify-otp", map[string]string{"email": "test@example.com", "otp": "wrong"})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestResetPassword_Success() {
	s.mockUC.On("ResetPassword", mock.Anything, "test@example.com", "NewPass@123").Return(nil)
	w := s.performRequest("POST", "/reset-password", map[string]string{"email": "test@example.com", "new_password": "NewPass@123"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ControllerTestSuite) TestResetPassword_InvalidJSON() {
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewBuffer([]byte(`{"email":`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestResetPassword_WeakPassword() {
	s.mockUC.On("ResetPassword", mock.Anything, "test@example.com", "123").Return(errors.New("weak password"))
	w := s.performRequest("POST", "/reset-password", map[string]string{"email": "test@example.com", "new_password": "123"})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestRegister_WeakPassword() {
	input := userpkg.User{Username: "user1", Email: "test@example.com", Password: "123", Fullname: "Test User"}
	s.mockUC.On("RegisterUser", mock.Anything, input).Return(userpkg.User{}, errors.New("password must be at least 8 chars"))
	w := s.performRequest("POST", "/register", input)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestRegister_DuplicateEmail() {
	input := userpkg.User{Username: "user1", Email: "taken@example.com", Password: "ValidPass@123", Fullname: "Test User"}
	s.mockUC.On("RegisterUser", mock.Anything, input).Return(userpkg.User{}, errors.New("email already taken"))
	w := s.performRequest("POST", "/register", input)
	s.Equal(http.StatusBadRequest, w.Code)
}

// TestRegister_Success ensures successful registration (verification is embedded in RegisterUser)
func (s *ControllerTestSuite) TestRegister_Success() {
	input := userpkg.User{Username: "newuser", Email: "new@example.com", Password: "Pass@1234", Fullname: "New User"}
	expected := userpkg.User{Username: "newuser", Email: "new@example.com", Fullname: "New User"}
	s.mockUC.On("RegisterUser", mock.Anything, input).Return(expected, nil)

	w := s.performRequest("POST", "/register", input)
	s.Equal(http.StatusCreated, w.Code)
}

// TestRegister_VerificationFail returns 400 if OTP send fails (embedded in RegisterUser)
func (s *ControllerTestSuite) TestRegister_VerificationFail() {
	input := userpkg.User{Username: "newuser", Email: "new@example.com", Password: "Pass@1234", Fullname: "New User"}
	// RegisterUser should return an error when email sending fails (this is embedded in RegisterUser)
	s.mockUC.On("RegisterUser", mock.Anything, input).Return(userpkg.User{}, errors.New("failed to send verification code"))

	w := s.performRequest("POST", "/register", input)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ControllerTestSuite) TestLogin_InvalidCredentials() {
	s.mockUC.On("LoginUser", mock.Anything, "user1", "wrongpass").
		Return(userpkg.User{}, "", "", errors.New("invalid credentials"))

	w := s.performRequest("POST", "/login", map[string]string{"login": "user1", "password": "wrongpass"})
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestLogin_Unverified() {
	// usecase.LoginUser returns error "email not verified"
	s.mockUC.On("LoginUser", mock.Anything, "user1", "pass").Return(userpkg.User{}, "", "", errors.New("email not verified"))

	w := s.performRequest("POST", "/login", map[string]string{"login": "user1", "password": "pass"})
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *ControllerTestSuite) TestPromoteUser_Success() {
	id := "user123"
	s.mockUC.On("PromoteUser", mock.Anything, id, "admin999").Return(nil)
	req := httptest.NewRequest(http.MethodPut, "/user/"+id+"/promote", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "user promoted")
	s.mockUC.AssertCalled(s.T(), "PromoteUser", mock.Anything, id, "admin999")
}

func (s *ControllerTestSuite) TestPromoteUser_Error() {
	id := "user123"
	errMock := errors.New("fail to promote")
	s.mockUC.On("PromoteUser", mock.Anything, id, "admin999").Return(errMock)
	req := httptest.NewRequest(http.MethodPut, "/user/"+id+"/promote", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "fail to promote")
}

func (s *ControllerTestSuite) TestDemoteUser_Success() {
	id := "user456"
	s.mockUC.On("DemoteUser", mock.Anything, id, "admin999").Return(nil)
	req := httptest.NewRequest(http.MethodPut, "/user/"+id+"/demote", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "user demoted")
	s.mockUC.AssertCalled(s.T(), "DemoteUser", mock.Anything, id, "admin999")
}

func (s *ControllerTestSuite) TestDemoteUser_Error() {
	id := "user456"
	errMock := errors.New("fail to demote")
	s.mockUC.On("DemoteUser", mock.Anything, id, "admin999").Return(errMock)
	req := httptest.NewRequest(http.MethodPut, "/user/"+id+"/demote", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "fail to demote")
}

func (s *ControllerTestSuite) TestGetProfile_Success() {
	// Arrange
	expectedUser := userpkg.User{
		Username: "testuser",
		Email:    "test@example.com",
		Fullname: "Test User",
	}
	s.mockUC.On("GetUserProfile", mock.Anything, "valid-user-id").Return(expectedUser, nil)

	s.router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", "valid-user-id")
		c.Next()
	}, controllers.NewController(s.mockUC).GetProfile)

	// Act
	w := s.performRequest("GET", "/profile", nil)

	// Assert
	s.Equal(http.StatusOK, w.Code)
	var actualUser userpkg.User
	s.NoError(json.Unmarshal(w.Body.Bytes(), &actualUser))
	s.Equal(expectedUser, actualUser)
	s.mockUC.AssertCalled(s.T(), "GetUserProfile", mock.Anything, "valid-user-id")
}

func (s *ControllerTestSuite) TestGetProfile_Unauthorized() {
	// Arrange
	s.router.GET("/profile", controllers.NewController(s.mockUC).GetProfile)

	// Act
	w := s.performRequest("GET", "/profile", nil)

	// Assert
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "Unauthorized")
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
