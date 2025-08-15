package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentControllerTestSuite struct {
	suite.Suite
	router     *gin.Engine
	mockUC     *mocks.ICommentUsecase
	controller *controllers.CommentController
}

func TestCommentControllerTestSuite(t *testing.T) { suite.Run(t, new(CommentControllerTestSuite)) }

func (s *CommentControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockUC = &mocks.ICommentUsecase{}
	s.controller = controllers.NewCommentController(s.mockUC)
	s.router = gin.New()

	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Set("userID", "507f1f77bcf86cd799439011")
		}
		c.Next()
	})

	s.router.POST("/posts/:id/comments", s.controller.CreateComment)
	s.router.GET("/posts/:id/comments", s.controller.GetComments)
	s.router.DELETE("/comments/:commentId", s.controller.DeleteComment)
}

func (s *CommentControllerTestSuite) TearDownTest() { s.mockUC.AssertExpectations(s.T()) }

func (s *CommentControllerTestSuite) performRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *CommentControllerTestSuite) TestCreateComment_Success() {
	postID := primitive.NewObjectID()
	uid, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	req := commentpkg.CreateCommentRequest{Content: "Nice post"}
	s.mockUC.On("CreateComment", mock.Anything, postID, req, uid).Return(&commentpkg.CommentResponse{ID: primitive.NewObjectID(), PostID: postID, Content: "Nice post"}, nil).Once()
	w := s.performRequest("POST", "/posts/"+postID.Hex()+"/comments", req, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusCreated, w.Code)
}

func (s *CommentControllerTestSuite) TestCreateComment_Unauthorized() {
	postID := primitive.NewObjectID()
	req := commentpkg.CreateCommentRequest{Content: "Hi"}
	w := s.performRequest("POST", "/posts/"+postID.Hex()+"/comments", req, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *CommentControllerTestSuite) TestGetComments_Success() {
	postID := primitive.NewObjectID()
	s.mockUC.On("GetComments", mock.Anything, postID, mock.AnythingOfType("comment.CommentPagination")).Return(&commentpkg.CommentListResponse{Comments: []commentpkg.CommentResponse{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 1}, nil).Once()
	w := s.performRequest("GET", "/posts/"+postID.Hex()+"/comments", nil, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *CommentControllerTestSuite) TestDeleteComment_Success() {
	cid := primitive.NewObjectID()
	uid, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	s.mockUC.On("DeleteComment", mock.Anything, cid, uid).Return(nil).Once()
	w := s.performRequest("DELETE", "/comments/"+cid.Hex(), nil, map[string]string{"Authorization": "Bearer token"})
	s.Equal(http.StatusOK, w.Code)
}
