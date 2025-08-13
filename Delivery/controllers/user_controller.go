package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	userUsecase userpkg.IUserUsecase
}

func NewController(userUsecase userpkg.IUserUsecase) *Controller {
	return &Controller{
		userUsecase: userUsecase,
	}
}

// User Controllers
func (ctrl *Controller) Register(c *gin.Context) {
	var user userpkg.User

	// 1. Parse JSON input
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Create context with timeout (e.g. 20s)
	//We needed a longer timeout to verify the emails validity
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	// 3. Call the usecase (now includes OTP sending)
	createdUser, err := ctrl.userUsecase.RegisterUser(ctx, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Success response with verification instruction
	response := gin.H{
		"message": "Registration successful! Please check your email for verification code.",
		"user":    createdUser,
		"note":    "You must verify your email before you can login.",
	}
	c.JSON(http.StatusCreated, response)
}

func (ctrl *Controller) Login(c *gin.Context) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	user, accessToken, refreshToken, err := ctrl.userUsecase.LoginUser(ctx, input.Login, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (ctrl *Controller) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	newTokens, err := ctrl.userUsecase.RefreshToken(ctx, body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newTokens)
}

func (ctrl *Controller) ForgotPassword(c *gin.Context) {
	var req struct{ Email string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := ctrl.userUsecase.SendResetOTP(ctx, req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent"})
}

func (ctrl *Controller) VerifyOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := ctrl.userUsecase.VerifyOTP(ctx, req.Email, req.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified, you can reset your password."})
}

func (ctrl *Controller) ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := ctrl.userUsecase.ResetPassword(ctx, req.Email, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (ctrl *Controller) Logout(c *gin.Context) {
	userID := c.GetString("userID") // assuming middleware sets this
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	err := ctrl.userUsecase.Logout(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Logout failed"})
		return
	}
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "localhost"
	}

	// Optionally clear tokens from client
	c.SetCookie("access_token", "", -1, "/", cookieDomain, false, true)
	c.SetCookie("refresh_token", "", -1, "/", cookieDomain, false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (ctrl *Controller) PromoteUser(c *gin.Context) {
	userID := c.Param("id")
	actorID, ok := c.Get("user_id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	if err := ctrl.userUsecase.PromoteUser(ctx, userID, actorID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user promoted"})
}

func (ctrl *Controller) DemoteUser(c *gin.Context) {
	userID := c.Param("id")
	actorID, ok := c.Get("user_id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	if err := ctrl.userUsecase.DemoteUser(ctx, userID, actorID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user demoted"})
}

func (ctrl *Controller) VerifyUser(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := ctrl.userUsecase.VerifyUser(ctx, req.Email, req.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User verified"})
}

func (ctrl *Controller) GetProfile(c *gin.Context) {

    userID := c.GetString("user_id")

    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    user, err := ctrl.userUsecase.GetUserProfile(ctx, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    c.JSON(http.StatusOK, user)
}

func (ctrl *Controller) UpdateProfile(c *gin.Context) {
    userID := c.GetString("user_id")
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Parse multipart form (10MB limit)
    err := c.Request.ParseMultipartForm(10 << 20)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }

    var updates userpkg.UpdateProfileRequest

    // Get text fields
    if fullname := c.PostForm("fullname"); fullname != "" {
        updates.Fullname = fullname
    }
    if bio := c.PostForm("bio"); bio != "" {
        updates.Bio = bio
    }
    if phone := c.PostForm("phone"); phone != "" {
        updates.ContactInfo.Phone = phone
    }
    if website := c.PostForm("website"); website != "" {
        updates.ContactInfo.Website = website
    }
    if twitter := c.PostForm("twitter"); twitter != "" {
        updates.ContactInfo.Twitter = twitter
    }
    if linkedin := c.PostForm("linkedin"); linkedin != "" {
        updates.ContactInfo.LinkedIn = linkedin
    }

    // Handle file upload
    file, header, err := c.Request.FormFile("profilePicture")
	if err != nil {
		file = nil
		header = nil
	}

	var filename string
	if header != nil {
		filename = header.Filename
	} else {
		filename = ""
	}

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    updatedUser, err := ctrl.userUsecase.UpdateProfile(ctx, userID, updates, file, filename)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, updatedUser)
}
