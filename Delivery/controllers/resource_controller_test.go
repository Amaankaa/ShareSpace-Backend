package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	resourcepkg "github.com/Amaankaa/Blog-Starter-Project/Domain/resource"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResourceController Test Suite

type ResourceControllerTestSuite struct {
	suite.Suite
	router              *gin.Engine
	mockResourceUsecase *mocks.ResourceUsecase
	controller          *controllers.ResourceController
}

func TestResourceControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceControllerTestSuite))
}

func (s *ResourceControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockResourceUsecase = mocks.NewResourceUsecase(s.T())
	s.controller = controllers.NewResourceController(s.mockResourceUsecase)
	s.router = gin.New()

	// Middleware to set userID when Authorization header is present
	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Set("userID", "507f1f77bcf86cd799439011")
		}
		c.Next()
	})

	// Routes
	s.router.POST("/resources", s.controller.CreateResource)
	s.router.GET("/resources", s.controller.GetResources)
	s.router.GET("/resources/:id", s.controller.GetResource)
	s.router.PATCH("/resources/:id", s.controller.UpdateResource)
	s.router.DELETE("/resources/:id", s.controller.DeleteResource)
	s.router.POST("/resources/:id/like", s.controller.LikeResource)
	s.router.DELETE("/resources/:id/like", s.controller.UnlikeResource)
	s.router.POST("/resources/:id/bookmark", s.controller.BookmarkResource)
	s.router.DELETE("/resources/:id/bookmark", s.controller.UnbookmarkResource)
	s.router.GET("/resources/search", s.controller.SearchResources)
	s.router.GET("/resources/popular", s.controller.GetPopularResources)
	s.router.GET("/resources/trending", s.controller.GetTrendingResources)
	s.router.GET("/resources/top-rated", s.controller.GetTopRatedResources)
	s.router.GET("/users/:id/resources", s.controller.GetUserResources)
	s.router.GET("/users/:id/resources/liked", s.controller.GetUserLikedResources)
	s.router.GET("/users/:id/resources/bookmarked", s.controller.GetUserBookmarkedResources)
}

func (s *ResourceControllerTestSuite) TearDownTest() {
	s.mockResourceUsecase.AssertExpectations(s.T())
}

func (s *ResourceControllerTestSuite) performRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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

// Tests

func (s *ResourceControllerTestSuite) TestCreateResource_Success() {
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := resourcepkg.CreateResourceRequest{
		Title:       "Test Resource",
		Description: "Desc",
		Content:     "Content",
		Type:        "guide",
		Category:    "Academic Success",
	}

	expected := &resourcepkg.ResourceResponse{ID: primitive.NewObjectID(), Title: "Test Resource"}
	s.mockResourceUsecase.On("CreateResource", mock.Anything, req, userID).Return(expected, nil)

	w := s.performRequest("POST", "/resources", req, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusCreated, w.Code)
}

func (s *ResourceControllerTestSuite) TestCreateResource_Unauthorized() {
	req := resourcepkg.CreateResourceRequest{Title: "Test", Description: "Desc", Content: "Content", Type: "guide", Category: "Academic Success"}
	w := s.performRequest("POST", "/resources", req, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *ResourceControllerTestSuite) TestCreateResource_InvalidJSON() {
	// Build a manual request with invalid JSON
	req, _ := http.NewRequest("POST", "/resources", bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetResource_Success() {
	id := primitive.NewObjectID()
	expected := &resourcepkg.ResourceResponse{ID: id, Title: "T"}
	s.mockResourceUsecase.On("GetResource", mock.Anything, id, (*primitive.ObjectID)(nil)).Return(expected, nil)
	w := s.performRequest("GET", "/resources/"+id.Hex(), nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetResource_InvalidID() {
	w := s.performRequest("GET", "/resources/invalid", nil, nil)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestUpdateResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := resourcepkg.UpdateResourceRequest{Title: "New"}
	expected := &resourcepkg.ResourceResponse{ID: id, Title: "New"}
	s.mockResourceUsecase.On("UpdateResource", mock.Anything, id, req, userID).Return(expected, nil)
	w := s.performRequest("PATCH", "/resources/"+id.Hex(), req, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestUpdateResource_InvalidID() {
	w := s.performRequest("PATCH", "/resources/invalid-id", resourcepkg.UpdateResourceRequest{Title: "X"}, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestUpdateResource_Forbidden() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := resourcepkg.UpdateResourceRequest{Title: "New"}
	s.mockResourceUsecase.On("UpdateResource", mock.Anything, id, req, userID).Return(nil, errors.New("unauthorized: only the creator can update this resource")).Once()
	w := s.performRequest("PATCH", "/resources/"+id.Hex(), req, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *ResourceControllerTestSuite) TestUpdateResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := resourcepkg.UpdateResourceRequest{Title: "New"}
	s.mockResourceUsecase.On("UpdateResource", mock.Anything, id, req, userID).Return(nil, errors.New("resource not found")).Once()
	w := s.performRequest("PATCH", "/resources/"+id.Hex(), req, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestDeleteResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("DeleteResource", mock.Anything, id, userID).Return(nil)
	w := s.performRequest("DELETE", "/resources/"+id.Hex(), nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestDeleteResource_InvalidID() {
	w := s.performRequest("DELETE", "/resources/invalid-id", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestDeleteResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("DeleteResource", mock.Anything, id, userID).Return(errors.New("resource not found")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex(), nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestDeleteResource_Forbidden() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("DeleteResource", mock.Anything, id, userID).Return(errors.New("unauthorized: only the creator can delete this resource")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex(), nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *ResourceControllerTestSuite) TestLikeResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("LikeResource", mock.Anything, id, userID).Return(nil)
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestLikeResource_InvalidID() {
	w := s.performRequest("POST", "/resources/invalid-id/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestLikeResource_Unauthorized() {
	id := primitive.NewObjectID()
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/like", nil, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *ResourceControllerTestSuite) TestLikeResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("LikeResource", mock.Anything, id, userID).Return(errors.New("resource not found")).Once()
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestLikeResource_AlreadyLiked() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("LikeResource", mock.Anything, id, userID).Return(errors.New("resource already liked by user")).Once()
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusConflict, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnlikeResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnlikeResource", mock.Anything, id, userID).Return(nil)
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnlikeResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnlikeResource", mock.Anything, id, userID).Return(errors.New("resource not found")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnlikeResource_NotLiked() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnlikeResource", mock.Anything, id, userID).Return(errors.New("resource not liked by user")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/like", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusConflict, w.Code)
}

func (s *ResourceControllerTestSuite) TestBookmarkResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("BookmarkResource", mock.Anything, id, userID).Return(nil)
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestBookmarkResource_AlreadyBookmarked() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("BookmarkResource", mock.Anything, id, userID).Return(errors.New("resource already bookmarked by user")).Once()
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusConflict, w.Code)
}

func (s *ResourceControllerTestSuite) TestBookmarkResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("BookmarkResource", mock.Anything, id, userID).Return(errors.New("resource not found")).Once()
	w := s.performRequest("POST", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnbookmarkResource_Success() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnbookmarkResource", mock.Anything, id, userID).Return(nil)
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnbookmarkResource_NotBookmarked() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnbookmarkResource", mock.Anything, id, userID).Return(errors.New("resource not bookmarked by user")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusConflict, w.Code)
}

func (s *ResourceControllerTestSuite) TestUnbookmarkResource_NotFound() {
	id := primitive.NewObjectID()
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockResourceUsecase.On("UnbookmarkResource", mock.Anything, id, userID).Return(errors.New("resource not found")).Once()
	w := s.performRequest("DELETE", "/resources/"+id.Hex()+"/bookmark", nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ResourceControllerTestSuite) TestSearchResources_Success() {
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{{ID: primitive.NewObjectID(), Title: "R"}}, Total: 1}
	s.mockResourceUsecase.On("SearchResources", mock.Anything, "q", mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination"), (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/resources/search?q=q", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestSearchResources_MissingQuery() {
	w := s.performRequest("GET", "/resources/search", nil, nil)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestSearchResources_InternalError() {
	s.mockResourceUsecase.On("SearchResources", mock.Anything, "q", mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination"), (*primitive.ObjectID)(nil)).Return(nil, errors.New("boom")).Once()
	w := s.performRequest("GET", "/resources/search?q=q", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetPopularResources_Success() {
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetPopularResources", mock.Anything, 20, "week", (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/resources/popular", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetPopularResources_InternalError() {
	s.mockResourceUsecase.On("GetPopularResources", mock.Anything, 20, "week", (*primitive.ObjectID)(nil)).Return(nil, errors.New("err")).Once()
	w := s.performRequest("GET", "/resources/popular", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetTrendingResources_Success() {
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetTrendingResources", mock.Anything, 20, (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/resources/trending", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetTrendingResources_InternalError() {
	s.mockResourceUsecase.On("GetTrendingResources", mock.Anything, 20, (*primitive.ObjectID)(nil)).Return(nil, errors.New("err")).Once()
	w := s.performRequest("GET", "/resources/trending", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetTopRatedResources_Success() {
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetTopRatedResources", mock.Anything, 20, "", (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/resources/top-rated", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetTopRatedResources_InternalError() {
	s.mockResourceUsecase.On("GetTopRatedResources", mock.Anything, 20, "", (*primitive.ObjectID)(nil)).Return(nil, errors.New("err")).Once()
	w := s.performRequest("GET", "/resources/top-rated", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetResources_Success() {
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}, Total: 0, Page: 1, PageSize: 20}
	s.mockResourceUsecase.On("GetResources", mock.Anything, mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination"), (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/resources", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetResources_InternalError() {
	s.mockResourceUsecase.On("GetResources", mock.Anything, mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination"), (*primitive.ObjectID)(nil)).Return(nil, errors.New("err")).Once()
	w := s.performRequest("GET", "/resources", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserResources_Success() {
	userID := primitive.NewObjectID()
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetUserResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination"), (*primitive.ObjectID)(nil)).Return(resp, nil)
	w := s.performRequest("GET", "/users/"+userID.Hex()+"/resources", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserResources_InvalidUserID() {
	w := s.performRequest("GET", "/users/invalid/resources", nil, nil)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserLikedResources_Success() {
	userID := primitive.NewObjectID()
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetUserLikedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(resp, nil)
	w := s.performRequest("GET", "/users/"+userID.Hex()+"/resources/liked", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserLikedResources_InvalidUserID() {
	w := s.performRequest("GET", "/users/invalid/resources/liked", nil, nil)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserBookmarkedResources_Success() {
	userID := primitive.NewObjectID()
	resp := &resourcepkg.ResourceListResponse{Resources: []resourcepkg.ResourceResponse{}}
	s.mockResourceUsecase.On("GetUserBookmarkedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(resp, nil)
	w := s.performRequest("GET", "/users/"+userID.Hex()+"/resources/bookmarked", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *ResourceControllerTestSuite) TestGetUserBookmarkedResources_InvalidUserID() {
	w := s.performRequest("GET", "/users/invalid/resources/bookmarked", nil, nil)
	s.Equal(http.StatusBadRequest, w.Code)
}
