package routers

import (
	"net/http"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	infrastructure "github.com/Amaankaa/Blog-Starter-Project/Infrastructure"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.Controller, authMiddleware *infrastructure.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"service":   "sharespace-backend",
		})
	})

	// Public routes
	r.POST("/register", controller.Register)
	r.POST("/verify-user", controller.VerifyUser) // Registration verification (separate from password-reset OTP)
	r.POST("/login", controller.Login)
	r.POST("/forgot-password", controller.ForgotPassword)
	r.POST("/verify-otp", controller.VerifyOTP)
	r.POST("/reset-password", controller.ResetPassword)

	// Protected routes
	protected := r.Group("")
	protected.Use(authMiddleware.AuthMiddleware())

	//User routes
	protected.POST("/logout", controller.Logout)
	protected.GET("/profile", controller.GetProfile)
	protected.PUT("/profile", controller.UpdateProfile)

	// Posts routes (protected)
	protected.POST("/posts", controller.PostController.CreatePost)
	protected.PATCH("/posts/:id", controller.PostController.UpdatePost)
	protected.DELETE("/posts/:id", controller.PostController.DeletePost)
	protected.POST("/posts/:id/like", controller.PostController.LikePost)
	protected.DELETE("/posts/:id/like", controller.PostController.UnlikePost)

	// Posts routes (public - can be viewed without authentication, but with optional user context)
	r.GET("/posts", controller.PostController.GetPosts)
	r.GET("/posts/search", controller.PostController.SearchPosts)
	r.GET("/posts/popular", controller.PostController.GetPopularPosts)
	r.GET("/posts/trending-tags", controller.PostController.GetTrendingTags)
	r.GET("/posts/:id", controller.PostController.GetPost)
	r.GET("/posts/category/:category", controller.PostController.GetPostsByCategory)
	r.GET("/users/:userId/posts", controller.PostController.GetUserPosts)

	// Admin routes for user promotion and demotion
	admin := protected.Group("")
	admin.Use(authMiddleware.AdminOnly())
	admin.PUT("/user/:id/promote", controller.PromoteUser)
	admin.PUT("/user/:id/demote", controller.DemoteUser)

	return r
}
