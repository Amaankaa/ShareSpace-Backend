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
	// Comments on posts (protected)
	protected.POST("/posts/:id/comments", controller.CommentController.CreateComment)
	protected.PATCH("/comments/:commentId", controller.CommentController.UpdateComment)
	protected.DELETE("/comments/:commentId", controller.CommentController.DeleteComment)

	// Posts routes (public - can be viewed without authentication, but with optional user context)
	r.GET("/posts", controller.PostController.GetPosts)
	r.GET("/posts/search", controller.PostController.SearchPosts)
	r.GET("/posts/popular", controller.PostController.GetPopularPosts)
	r.GET("/posts/trending-tags", controller.PostController.GetTrendingTags)
	r.GET("/posts/:id", controller.PostController.GetPost)
	r.GET("/posts/:id/comments", controller.CommentController.GetComments)
	r.GET("/posts/category/:category", controller.PostController.GetPostsByCategory)
	r.GET("/users/:userId/posts", controller.PostController.GetUserPosts)

	// Resources routes (protected)
	protected.POST("/resources", controller.ResourceController.CreateResource)
	protected.PATCH("/resources/:id", controller.ResourceController.UpdateResource)
	protected.DELETE("/resources/:id", controller.ResourceController.DeleteResource)
	protected.POST("/resources/:id/like", controller.ResourceController.LikeResource)
	protected.DELETE("/resources/:id/like", controller.ResourceController.UnlikeResource)
	protected.POST("/resources/:id/bookmark", controller.ResourceController.BookmarkResource)
	protected.DELETE("/resources/:id/bookmark", controller.ResourceController.UnbookmarkResource)
	protected.GET("/resources/:id/analytics", controller.ResourceController.GetResourceAnalytics)
	protected.POST("/resources/:id/report", controller.ResourceController.ReportResource)

	// Resources routes (public)
	r.GET("/resources", controller.ResourceController.GetResources)
	r.GET("/resources/search", controller.ResourceController.SearchResources)
	r.GET("/resources/popular", controller.ResourceController.GetPopularResources)
	r.GET("/resources/trending", controller.ResourceController.GetTrendingResources)
	r.GET("/resources/top-rated", controller.ResourceController.GetTopRatedResources)
	r.GET("/resources/:id", controller.ResourceController.GetResource)
	r.GET("/users/:userId/resources", controller.ResourceController.GetUserResources)
	r.GET("/users/:userId/resources/liked", controller.ResourceController.GetUserLikedResources)
	r.GET("/users/:userId/resources/bookmarked", controller.ResourceController.GetUserBookmarkedResources)

	// Mentorship routes (protected)
	if controller.MentorshipController != nil {
		protected.POST("/mentorship/requests", controller.MentorshipController.SendMentorshipRequest)
		protected.GET("/mentorship/requests/incoming", controller.MentorshipController.GetIncomingRequests)
		protected.GET("/mentorship/requests/outgoing", controller.MentorshipController.GetOutgoingRequests)
		protected.POST("/mentorship/requests/:id/respond", controller.MentorshipController.RespondToRequest)
		protected.DELETE("/mentorship/requests/:id", controller.MentorshipController.CancelRequest)
		protected.GET("/mentorship/connections/:id", controller.MentorshipController.GetConnection)
		protected.GET("/mentorship/connections/mentor", controller.MentorshipController.GetMyMentorships)
		protected.GET("/mentorship/connections/mentee", controller.MentorshipController.GetMyMenteerships)
		protected.GET("/mentorship/connections/active", controller.MentorshipController.GetActiveConnections)
		protected.POST("/mentorship/connections/:id/interaction", controller.MentorshipController.UpdateLastInteraction)
		protected.POST("/mentorship/connections/:id/pause", controller.MentorshipController.PauseConnection)
		protected.POST("/mentorship/connections/:id/resume", controller.MentorshipController.ResumeConnection)
		protected.POST("/mentorship/connections/:id/end", controller.MentorshipController.EndConnection)
		protected.GET("/mentorship/stats", controller.MentorshipController.GetMentorshipStats)
		protected.GET("/mentorship/insights", controller.MentorshipController.GetMentorshipInsights)
	}
	r.GET("/users/:userId/resources/stats", controller.ResourceController.GetUserResourceStats)

	// Messaging routes (protected) and WebSocket endpoint
	if controller.MessagingController != nil {
		protected.POST("/conversations", controller.MessagingController.CreateConversation)
		protected.GET("/conversations", controller.MessagingController.GetConversations)
		protected.GET("/conversations/:id/messages", controller.MessagingController.GetMessages)
	}

	// Admin routes for user promotion and demotion
	admin := protected.Group("")
	admin.Use(authMiddleware.AdminOnly())
	admin.PUT("/user/:id/promote", controller.PromoteUser)
	admin.PUT("/user/:id/demote", controller.DemoteUser)
	// Admin-only resource verification
	admin.POST("/resources/:id/verify", controller.ResourceController.VerifyResource)

	return r
}
