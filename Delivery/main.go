package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/routers"
	infrastructure "github.com/Amaankaa/Blog-Starter-Project/Infrastructure"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables (optional for production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set in environment")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	db := client.Database("blog_db")
	userCollection := db.Collection("users")
	tokenCollection := db.Collection("tokens")
	passwordResetCollection := db.Collection("password_resets")
	verificationCollection := db.Collection("verifications")

	// Initialize infrastructure services
	passwordService := infrastructure.NewPasswordService()
	jwtService := infrastructure.NewJWTService()

	emailVerifier, err := infrastructure.NewEmailListVerifyVerifier()
	if err != nil {
		log.Fatalf("Failed to initialize email verifier: %v", err)
	}
	emailSender := infrastructure.NewBrevoEmailSender()

	// Cloudinary configuration
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	
	if cloudName == "" || cloudAPIKey == "" || cloudAPISecret == "" {
		log.Fatal("Cloudinary credentials not set in environment")
	}

	// Services
	cloudinaryService, err := infrastructure.NewCloudinaryService(cloudName, cloudAPIKey, cloudAPISecret)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary service: %v", err)
	}

	//Repositories: only take collection (not services)
	userRepo := repositories.NewUserRepository(userCollection)
	tokenRepo := repositories.NewTokenRepository(tokenCollection)
	passwordResetRepo := repositories.NewPasswordResetRepo(passwordResetCollection, userCollection)
	//AI configuration
	aiAPIKey := os.Getenv("GEMINI_API_KEY")
	if aiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY not set in environment")
	}
	aiAPIURL := os.Getenv("GEMINI_API_URL")
	if aiAPIURL == "" {
		log.Fatal("GEMINI_API_URL not set in environment")
	}

	//Usecase: handles business logic, gets all dependencies
	verificationRepo := repositories.NewVerificationRepo(verificationCollection)
	userUsecase := usecases.NewUserUsecase(
		userRepo,
		passwordService,
		tokenRepo,
		jwtService,
		emailVerifier,
		emailSender,
		passwordResetRepo,
		verificationRepo,
		cloudinaryService,
	)

	//Controller
	controller := controllers.NewController(userUsecase)

	// Initialize AuthMiddleware
	authMiddleware := infrastructure.NewAuthMiddleware(jwtService)

	//Router
	r := routers.SetupRouter(controller, authMiddleware)

	//Start Server
	log.Println("Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
