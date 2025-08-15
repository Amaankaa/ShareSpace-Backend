package usecases_test

import (
	"context"
	"errors"
	"testing"

	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Using generated mocks from mockery

// Test Suite
type PostUsecaseTestSuite struct {
	suite.Suite
	ctx          context.Context
	mockPostRepo *mocks.PostRepository
	mockUserRepo *mocks.IUserRepository
	usecase      *usecases.PostUsecase
}

func TestPostUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(PostUsecaseTestSuite))
}

func (s *PostUsecaseTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockPostRepo = mocks.NewPostRepository(s.T())
	s.mockUserRepo = mocks.NewIUserRepository(s.T())
	s.usecase = usecases.NewPostUsecase(s.mockPostRepo, s.mockUserRepo)
}

func (s *PostUsecaseTestSuite) TearDownTest() {
	s.mockPostRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

// Test CreatePost
func (s *PostUsecaseTestSuite) TestCreatePost_Success() {
	// Arrange
	authorID := primitive.NewObjectID()
	req := postpkg.CreatePostRequest{
		Title:    "Test Post",
		Content:  "This is a test post content",
		Category: "Academic Struggles",
		Tags:     []string{"test", "academic"},
	}

	expectedUser := userpkg.User{
		ID:          authorID,
		DisplayName: "Test User",
		IsMentor:    true,
	}

	expectedPost := &postpkg.Post{
		ID:       primitive.NewObjectID(),
		AuthorID: authorID,
		Title:    "Test Post",
		Content:  "This is a test post content",
		Category: "Academic Struggles",
		Tags:     []string{"test", "academic"},
	}

	s.mockUserRepo.On("FindByID", s.ctx, authorID.Hex()).Return(expectedUser, nil)
	s.mockPostRepo.On("CreatePost", s.ctx, mock.AnythingOfType("postpkg.Post")).Return(expectedPost, nil)

	// Act
	result, err := s.usecase.CreatePost(s.ctx, req, authorID)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal("Test Post", result.Title)
	s.Equal("This is a test post content", result.Content)
	s.Equal("Academic Struggles", result.Category)
	s.Equal([]string{"test", "academic"}, result.Tags)
}

func (s *PostUsecaseTestSuite) TestCreatePost_InvalidCategory() {
	// Arrange
	authorID := primitive.NewObjectID()
	req := postpkg.CreatePostRequest{
		Title:    "Test Post",
		Content:  "This is a test post content",
		Category: "Invalid Category",
		Tags:     []string{"test"},
	}

	// Act
	result, err := s.usecase.CreatePost(s.ctx, req, authorID)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "invalid category")
}

func (s *PostUsecaseTestSuite) TestCreatePost_UserNotFound() {
	// Arrange
	authorID := primitive.NewObjectID()
	req := postpkg.CreatePostRequest{
		Title:    "Test Post",
		Content:  "This is a test post content",
		Category: "Academic Struggles",
		Tags:     []string{"test"},
	}

	s.mockUserRepo.On("FindByID", s.ctx, authorID.Hex()).Return(userpkg.User{}, errors.New("user not found"))

	// Act
	result, err := s.usecase.CreatePost(s.ctx, req, authorID)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to get author")
}

// Test GetPost
func (s *PostUsecaseTestSuite) TestGetPost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	authorID := primitive.NewObjectID()
	viewerID := primitive.NewObjectID()

	expectedPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Test Post",
		Content:  "Test content",
		Category: "Academic Struggles",
	}

	expectedUser := userpkg.User{
		ID:          authorID,
		DisplayName: "Test Author",
		IsMentor:    true,
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(expectedPost, nil)
	s.mockUserRepo.On("FindByID", s.ctx, authorID.Hex()).Return(expectedUser, nil)
	s.mockPostRepo.On("IsPostLikedByUser", s.ctx, postID, viewerID).Return(false, nil)
	s.mockPostRepo.On("IncrementViewCount", mock.Anything, postID).Return(nil)

	// Act
	result, err := s.usecase.GetPost(s.ctx, postID, &viewerID)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal("Test Post", result.Title)
	s.Equal("Test content", result.Content)
	s.Equal("Test Author", result.Author.DisplayName)
	s.False(result.IsLikedByUser)
}

func (s *PostUsecaseTestSuite) TestGetPost_NotFound() {
	// Arrange
	postID := primitive.NewObjectID()

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return((*postpkg.Post)(nil), errors.New("post not found"))

	// Act
	result, err := s.usecase.GetPost(s.ctx, postID, nil)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "post not found")
}

// Test UpdatePost
func (s *PostUsecaseTestSuite) TestUpdatePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	authorID := primitive.NewObjectID()

	req := postpkg.UpdatePostRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	existingPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Original Title",
		Content:  "Original content",
		Category: "Academic Struggles",
	}

	updatedPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Updated Title",
		Content:  "Updated content",
		Category: "Academic Struggles",
	}

	expectedUser := userpkg.User{
		ID:          authorID,
		DisplayName: "Test Author",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)
	s.mockPostRepo.On("UpdatePost", s.ctx, postID, mock.AnythingOfType("postpkg.Post")).Return(updatedPost, nil)
	s.mockUserRepo.On("FindByID", s.ctx, authorID.Hex()).Return(expectedUser, nil)

	// Act
	result, err := s.usecase.UpdatePost(s.ctx, postID, req, authorID)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal("Updated Title", result.Title)
	s.Equal("Updated content", result.Content)
}

func (s *PostUsecaseTestSuite) TestUpdatePost_Unauthorized() {
	// Arrange
	postID := primitive.NewObjectID()
	authorID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	req := postpkg.UpdatePostRequest{
		Title: "Updated Title",
	}

	existingPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Original Title",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)

	// Act
	result, err := s.usecase.UpdatePost(s.ctx, postID, req, otherUserID)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "unauthorized")
}

// Test DeletePost
func (s *PostUsecaseTestSuite) TestDeletePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	authorID := primitive.NewObjectID()

	existingPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Test Post",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)
	s.mockPostRepo.On("DeletePost", s.ctx, postID).Return(nil)

	// Act
	err := s.usecase.DeletePost(s.ctx, postID, authorID)

	// Assert
	s.NoError(err)
}

func (s *PostUsecaseTestSuite) TestDeletePost_Unauthorized() {
	// Arrange
	postID := primitive.NewObjectID()
	authorID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	existingPost := &postpkg.Post{
		ID:       postID,
		AuthorID: authorID,
		Title:    "Test Post",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)

	// Act
	err := s.usecase.DeletePost(s.ctx, postID, otherUserID)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "unauthorized")
}

// Test LikePost
func (s *PostUsecaseTestSuite) TestLikePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	existingPost := &postpkg.Post{
		ID:    postID,
		Title: "Test Post",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)
	s.mockPostRepo.On("IsPostLikedByUser", s.ctx, postID, userID).Return(false, nil)
	s.mockPostRepo.On("LikePost", s.ctx, postID, userID).Return(nil)

	// Act
	err := s.usecase.LikePost(s.ctx, postID, userID)

	// Assert
	s.NoError(err)
}

func (s *PostUsecaseTestSuite) TestLikePost_AlreadyLiked() {
	// Arrange
	postID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	existingPost := &postpkg.Post{
		ID:    postID,
		Title: "Test Post",
	}

	s.mockPostRepo.On("GetPostByID", s.ctx, postID).Return(existingPost, nil)
	s.mockPostRepo.On("IsPostLikedByUser", s.ctx, postID, userID).Return(true, nil)

	// Act
	err := s.usecase.LikePost(s.ctx, postID, userID)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "already liked")
}

// Test ValidatePostCategory
func (s *PostUsecaseTestSuite) TestValidatePostCategory_Valid() {
	// Act
	err := s.usecase.ValidatePostCategory("Academic Struggles")

	// Assert
	s.NoError(err)
}

func (s *PostUsecaseTestSuite) TestValidatePostCategory_Invalid() {
	// Act
	err := s.usecase.ValidatePostCategory("Invalid Category")

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "invalid category")
}

// Test ValidateMediaLinks
func (s *PostUsecaseTestSuite) TestValidateMediaLinks_Valid() {
	// Arrange
	mediaLinks := []postpkg.MediaLink{
		{Type: "image", URL: "https://example.com/image.jpg"},
		{Type: "video", URL: "https://example.com/video.mp4"},
	}

	// Act
	err := s.usecase.ValidateMediaLinks(mediaLinks)

	// Assert
	s.NoError(err)
}

func (s *PostUsecaseTestSuite) TestValidateMediaLinks_InvalidType() {
	// Arrange
	mediaLinks := []postpkg.MediaLink{
		{Type: "invalid", URL: "https://example.com/file"},
	}

	// Act
	err := s.usecase.ValidateMediaLinks(mediaLinks)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "invalid media type")
}

func (s *PostUsecaseTestSuite) TestValidateMediaLinks_EmptyURL() {
	// Arrange
	mediaLinks := []postpkg.MediaLink{
		{Type: "image", URL: ""},
	}

	// Act
	err := s.usecase.ValidateMediaLinks(mediaLinks)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "URL cannot be empty")
}
