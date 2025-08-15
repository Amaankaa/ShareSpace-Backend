package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Using generated mocks from mockery

// Test Suite
type PostControllerTestSuite struct {
	suite.Suite
	router          *gin.Engine
	mockPostUsecase *mocks.PostUsecase
	controller      *controllers.PostController
}

func TestPostControllerTestSuite(t *testing.T) {
	suite.Run(t, new(PostControllerTestSuite))
}

func (s *PostControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockPostUsecase = mocks.NewPostUsecase(s.T())
	s.controller = controllers.NewPostController(s.mockPostUsecase)
	s.router = gin.New()

	// Add middleware to set userID for authenticated routes
	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Set("userID", "507f1f77bcf86cd799439011") // Valid ObjectID
		}
		c.Next()
	})

	// Setup routes
	s.router.POST("/posts", s.controller.CreatePost)
	s.router.GET("/posts/:id", s.controller.GetPost)
	s.router.PATCH("/posts/:id", s.controller.UpdatePost)
	s.router.DELETE("/posts/:id", s.controller.DeletePost)
	s.router.POST("/posts/:id/like", s.controller.LikePost)
	s.router.DELETE("/posts/:id/like", s.controller.UnlikePost)
	s.router.GET("/posts", s.controller.GetPosts)
	s.router.GET("/posts/search", s.controller.SearchPosts)
	s.router.GET("/posts/popular", s.controller.GetPopularPosts)
	s.router.GET("/posts/trending-tags", s.controller.GetTrendingTags)
}

func (s *PostControllerTestSuite) TearDownTest() {
	s.mockPostUsecase.AssertExpectations(s.T())
}

func (s *PostControllerTestSuite) performRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// Test CreatePost
func (s *PostControllerTestSuite) TestCreatePost_Success() {
	// Arrange
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := postpkg.CreatePostRequest{
		Title:    "Test Post",
		Content:  "Test content",
		Category: "Academic Struggles",
		Tags:     []string{"test"},
	}

	expectedResponse := &postpkg.PostResponse{
		ID:       primitive.NewObjectID(),
		Title:    "Test Post",
		Content:  "Test content",
		Category: "Academic Struggles",
		Tags:     []string{"test"},
	}

	s.mockPostUsecase.On("CreatePost", mock.Anything, req, userID).Return(expectedResponse, nil)

	// Act
	w := s.performRequest("POST", "/posts", req, map[string]string{"Authorization": "Bearer token"})

	// Assert
	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Post created successfully", response["message"])
	s.NotNil(response["post"])
}

func (s *PostControllerTestSuite) TestCreatePost_Unauthorized() {
	// Arrange
	req := postpkg.CreatePostRequest{
		Title:    "Test Post",
		Content:  "Test content",
		Category: "Academic Struggles",
	}

	// Act (no Authorization header)
	w := s.performRequest("POST", "/posts", req, nil)

	// Assert
	s.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("User not authenticated", response["error"])
}

func (s *PostControllerTestSuite) TestCreatePost_InvalidJSON() {
	// Act
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"].(string), "Invalid request format")
}

// Test GetPost
func (s *PostControllerTestSuite) TestGetPost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	expectedResponse := &postpkg.PostResponse{
		ID:      postID,
		Title:   "Test Post",
		Content: "Test content",
	}

	s.mockPostUsecase.On("GetPost", mock.Anything, postID, (*primitive.ObjectID)(nil)).Return(expectedResponse, nil)

	// Act
	w := s.performRequest("GET", "/posts/"+postID.Hex(), nil, nil)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response postpkg.PostResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Test Post", response.Title)
}

func (s *PostControllerTestSuite) TestGetPost_InvalidID() {
	// Act
	w := s.performRequest("GET", "/posts/invalid-id", nil, nil)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Invalid post ID", response["error"])
}

// Test UpdatePost
func (s *PostControllerTestSuite) TestUpdatePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := postpkg.UpdatePostRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	expectedResponse := &postpkg.PostResponse{
		ID:      postID,
		Title:   "Updated Title",
		Content: "Updated content",
	}

	s.mockPostUsecase.On("UpdatePost", mock.Anything, postID, req, userID).Return(expectedResponse, nil)

	// Act
	w := s.performRequest("PATCH", "/posts/"+postID.Hex(), req, map[string]string{"Authorization": "Bearer token"})

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Post updated successfully", response["message"])
}

// Test DeletePost
func (s *PostControllerTestSuite) TestDeletePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")

	s.mockPostUsecase.On("DeletePost", mock.Anything, postID, userID).Return(nil)

	// Act
	w := s.performRequest("DELETE", "/posts/"+postID.Hex(), nil, map[string]string{"Authorization": "Bearer token"})

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Post deleted successfully", response["message"])
}

// Test LikePost
func (s *PostControllerTestSuite) TestLikePost_Success() {
	// Arrange
	postID := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")

	s.mockPostUsecase.On("LikePost", mock.Anything, postID, userID).Return(nil)

	// Act
	w := s.performRequest("POST", "/posts/"+postID.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Post liked successfully", response["message"])
}

// Test GetPosts
func (s *PostControllerTestSuite) TestGetPosts_Success() {
	// Arrange
	expectedResponse := &postpkg.PostListResponse{
		Posts: []postpkg.PostResponse{
			{ID: primitive.NewObjectID(), Title: "Post 1"},
			{ID: primitive.NewObjectID(), Title: "Post 2"},
		},
		Total:      2,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	s.mockPostUsecase.On("GetPosts", mock.Anything, mock.AnythingOfType("postpkg.PostFilter"), mock.AnythingOfType("postpkg.PostPagination"), (*primitive.ObjectID)(nil)).Return(expectedResponse, nil)

	// Act
	w := s.performRequest("GET", "/posts", nil, nil)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response postpkg.PostListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(2, len(response.Posts))
	s.Equal(int64(2), response.Total)
}

// Test SearchPosts
func (s *PostControllerTestSuite) TestSearchPosts_Success() {
	// Arrange
	expectedResponse := &postpkg.PostListResponse{
		Posts: []postpkg.PostResponse{
			{ID: primitive.NewObjectID(), Title: "Search Result"},
		},
		Total: 1,
	}

	s.mockPostUsecase.On("SearchPosts", mock.Anything, "test query", mock.AnythingOfType("postpkg.PostFilter"), mock.AnythingOfType("postpkg.PostPagination"), (*primitive.ObjectID)(nil)).Return(expectedResponse, nil)

	// Act
	w := s.performRequest("GET", "/posts/search?q=test+query", nil, nil)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response postpkg.PostListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(1, len(response.Posts))
}

func (s *PostControllerTestSuite) TestSearchPosts_MissingQuery() {
	// Act
	w := s.performRequest("GET", "/posts/search", nil, nil)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Search query is required", response["error"])
}
