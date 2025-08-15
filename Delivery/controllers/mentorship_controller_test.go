package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MentorshipControllerTestSuite struct {
	suite.Suite
	router     *gin.Engine
	mockUC     *mocks.IMentorshipUsecase
	controller *controllers.MentorshipController
}

func TestMentorshipControllerTestSuite(t *testing.T) {
	suite.Run(t, new(MentorshipControllerTestSuite))
}

func (s *MentorshipControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockUC = mocks.NewIMentorshipUsecase(s.T())
	s.controller = controllers.NewMentorshipController(s.mockUC)
	s.router = gin.New()

	// Simple auth shim
	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Set("userID", "507f1f77bcf86cd799439011")
		}
		c.Next()
	})

	// Routes
	s.router.POST("/mentorship/requests", s.controller.SendMentorshipRequest)
	s.router.GET("/mentorship/requests/incoming", s.controller.GetIncomingRequests)
	s.router.GET("/mentorship/requests/outgoing", s.controller.GetOutgoingRequests)
	s.router.POST("/mentorship/requests/:id/respond", s.controller.RespondToRequest)
	s.router.DELETE("/mentorship/requests/:id", s.controller.CancelRequest)
	s.router.GET("/mentorship/connections/:id", s.controller.GetConnection)
	s.router.GET("/mentorship/connections/mentor", s.controller.GetMyMentorships)
	s.router.GET("/mentorship/connections/mentee", s.controller.GetMyMenteerships)
	s.router.GET("/mentorship/connections/active", s.controller.GetActiveConnections)
	s.router.POST("/mentorship/connections/:id/interaction", s.controller.UpdateLastInteraction)
	s.router.POST("/mentorship/connections/:id/pause", s.controller.PauseConnection)
	s.router.POST("/mentorship/connections/:id/resume", s.controller.ResumeConnection)
	s.router.POST("/mentorship/connections/:id/end", s.controller.EndConnection)
	s.router.GET("/mentorship/stats", s.controller.GetMentorshipStats)
	s.router.GET("/mentorship/insights", s.controller.GetMentorshipInsights)
}

func (s *MentorshipControllerTestSuite) TearDownTest() {
	s.mockUC.AssertExpectations(s.T())
}

func (s *MentorshipControllerTestSuite) performRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, url, buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// SendMentorshipRequest
func (s *MentorshipControllerTestSuite) TestSendMentorshipRequest_Success() {
	uid := "507f1f77bcf86cd799439011"
	mentorID := primitive.NewObjectID().Hex()
	s.mockUC.On("SendMentorshipRequest", mock.Anything, uid, mock.MatchedBy(func(arg mentorshippkg.CreateMentorshipRequestDTO) bool { return arg.MentorID.Hex() == mentorID })).Return(mentorshippkg.MentorshipRequestResponse{ID: primitive.NewObjectID()}, nil)
	body := map[string]interface{}{"mentorId": mentorID, "topics": []string{"cv"}, "message": "hi"}
	w := s.performRequest("POST", "/mentorship/requests", body, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusCreated, w.Code)
}

func (s *MentorshipControllerTestSuite) TestSendMentorshipRequest_Unauthorized() {
	mentorID := primitive.NewObjectID().Hex()
	body := map[string]interface{}{"mentorId": mentorID, "topics": []string{"cv"}}
	w := s.performRequest("POST", "/mentorship/requests", body, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *MentorshipControllerTestSuite) TestSendMentorshipRequest_InvalidMentorID() {
	body := map[string]interface{}{"mentorId": "bad", "topics": []string{"cv"}}
	w := s.performRequest("POST", "/mentorship/requests", body, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusBadRequest, w.Code)
}

// Incoming/Outgoing
func (s *MentorshipControllerTestSuite) TestGetIncomingRequests_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetIncomingRequests", mock.Anything, uid, 20, 0).Return([]mentorshippkg.MentorshipRequestResponse{}, nil)
	w := s.performRequest("GET", "/mentorship/requests/incoming", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetIncomingRequests_Unauthorized() {
	w := s.performRequest("GET", "/mentorship/requests/incoming", nil, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetOutgoingRequests_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetOutgoingRequests", mock.Anything, uid, 20, 0).Return([]mentorshippkg.MentorshipRequestResponse{}, nil)
	w := s.performRequest("GET", "/mentorship/requests/outgoing", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

// Respond / Cancel
func (s *MentorshipControllerTestSuite) TestRespondToRequest_Success() {
	uid := "507f1f77bcf86cd799439011"
	reqID := "req123"
	dto := mentorshippkg.RespondToRequestDTO{Accept: true}
	s.mockUC.On("RespondToRequest", mock.Anything, reqID, uid, dto).Return(mentorshippkg.MentorshipRequestResponse{ID: primitive.NewObjectID()}, nil)
	w := s.performRequest("POST", "/mentorship/requests/"+reqID+"/respond", dto, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestRespondToRequest_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/mentorship/requests/xyz/respond", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "t")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *MentorshipControllerTestSuite) TestCancelRequest_Success() {
	uid := "507f1f77bcf86cd799439011"
	reqID := "abc"
	s.mockUC.On("CancelRequest", mock.Anything, reqID, uid).Return(nil)
	w := s.performRequest("DELETE", "/mentorship/requests/"+reqID, nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

// Connection fetch/lists
func (s *MentorshipControllerTestSuite) TestGetConnection_Success() {
	uid := "507f1f77bcf86cd799439011"
	connID := "c1"
	s.mockUC.On("GetMentorshipConnection", mock.Anything, connID, uid).Return(mentorshippkg.MentorshipConnectionResponse{ID: primitive.NewObjectID()}, nil)
	w := s.performRequest("GET", "/mentorship/connections/"+connID, nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetMyMentorships_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetMyMentorships", mock.Anything, uid, 20, 0).Return([]mentorshippkg.MentorshipConnectionResponse{}, nil)
	w := s.performRequest("GET", "/mentorship/connections/mentor", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetMyMenteerships_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetMyMenteerships", mock.Anything, uid, 20, 0).Return([]mentorshippkg.MentorshipConnectionResponse{}, nil)
	w := s.performRequest("GET", "/mentorship/connections/mentee", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetActiveConnections_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetActiveConnections", mock.Anything, uid).Return([]mentorshippkg.MentorshipConnectionResponse{}, nil)
	w := s.performRequest("GET", "/mentorship/connections/active", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

// Mutations on connection state
func (s *MentorshipControllerTestSuite) TestUpdateLastInteraction_Success() {
	uid := "507f1f77bcf86cd799439011"
	connID := "c1"
	s.mockUC.On("UpdateLastInteraction", mock.Anything, connID, uid).Return(nil)
	w := s.performRequest("POST", "/mentorship/connections/"+connID+"/interaction", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestPauseResumeEnd_Success() {
	uid := "507f1f77bcf86cd799439011"
	connID := "c2"
	s.mockUC.On("PauseConnection", mock.Anything, connID, uid).Return(nil).Once()
	w1 := s.performRequest("POST", "/mentorship/connections/"+connID+"/pause", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w1.Code)
	s.mockUC.On("ResumeConnection", mock.Anything, connID, uid).Return(nil).Once()
	w2 := s.performRequest("POST", "/mentorship/connections/"+connID+"/resume", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w2.Code)
	payload := mentorshippkg.EndConnectionDTO{Reason: "done"}
	s.mockUC.On("EndConnection", mock.Anything, connID, uid, payload).Return(nil).Once()
	w3 := s.performRequest("POST", "/mentorship/connections/"+connID+"/end", payload, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w3.Code)
}

func (s *MentorshipControllerTestSuite) TestEndConnection_InvalidBody() {
	req, _ := http.NewRequest("POST", "/mentorship/connections/x/end", bytes.NewBufferString("{"))
	req.Header.Set("Authorization", "t")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

// Stats / Insights
func (s *MentorshipControllerTestSuite) TestGetMentorshipStats_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetMentorshipStats", mock.Anything, uid).Return(mentorshippkg.MentorshipStats{}, nil)
	w := s.performRequest("GET", "/mentorship/stats", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

func (s *MentorshipControllerTestSuite) TestGetMentorshipInsights_Success() {
	uid := "507f1f77bcf86cd799439011"
	s.mockUC.On("GetMentorshipInsights", mock.Anything, uid).Return(mentorshippkg.MentorshipInsights{}, nil)
	w := s.performRequest("GET", "/mentorship/insights", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusOK, w.Code)
}

// A couple of failure samples for mutation endpoints
func (s *MentorshipControllerTestSuite) TestPauseConnection_Unauthorized() {
	w := s.performRequest("POST", "/mentorship/connections/x/pause", nil, nil)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *MentorshipControllerTestSuite) TestUpdateLastInteraction_ErrorFromUsecase() {
	uid := "507f1f77bcf86cd799439011"
	connID := "c9"
	s.mockUC.On("UpdateLastInteraction", mock.Anything, connID, uid).Return(errors.New("boom")).Once()
	w := s.performRequest("POST", "/mentorship/connections/"+connID+"/interaction", nil, map[string]string{"Authorization": "t"})
	s.Equal(http.StatusBadRequest, w.Code)
}
