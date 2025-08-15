package controllers

import (
    "context"
    "net/http"
    "strconv"
    "time"

    mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type MentorshipController struct {
    usecase mentorshippkg.IMentorshipUsecase
}

func NewMentorshipController(usecase mentorshippkg.IMentorshipUsecase) *MentorshipController {
    return &MentorshipController{usecase: usecase}
}

// POST /mentorship/requests
func (mc *MentorshipController) SendMentorshipRequest(c *gin.Context) {
    var body struct {
        MentorID string   `json:"mentorId"`
        Message  string   `json:"message"`
        Topics   []string `json:"topics"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    mentorOID, err := primitive.ObjectIDFromHex(body.MentorID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mentor ID"})
        return
    }
    dto := mentorshippkg.CreateMentorshipRequestDTO{MentorID: mentorOID, Message: body.Message, Topics: body.Topics}
    if err := dto.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    resp, err := mc.usecase.SendMentorshipRequest(ctx, uid, dto)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, resp)
}

// GET /mentorship/requests/incoming
func (mc *MentorshipController) GetIncomingRequests(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    res, err := mc.usecase.GetIncomingRequests(ctx, uid, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

// GET /mentorship/requests/outgoing
func (mc *MentorshipController) GetOutgoingRequests(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    res, err := mc.usecase.GetOutgoingRequests(ctx, uid, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

// POST /mentorship/requests/:id/respond
func (mc *MentorshipController) RespondToRequest(c *gin.Context) {
    reqID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    var body mentorshippkg.RespondToRequestDTO
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    resp, err := mc.usecase.RespondToRequest(ctx, reqID, uid, body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp)
}

// DELETE /mentorship/requests/:id
func (mc *MentorshipController) CancelRequest(c *gin.Context) {
    reqID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    if err := mc.usecase.CancelRequest(ctx, reqID, uid); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Request cancelled"})
}

// GET /mentorship/connections/:id
func (mc *MentorshipController) GetConnection(c *gin.Context) {
    connID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    resp, err := mc.usecase.GetMentorshipConnection(ctx, connID, uid)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp)
}

// GET /mentorship/connections/mentor
func (mc *MentorshipController) GetMyMentorships(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    res, err := mc.usecase.GetMyMentorships(ctx, uid, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

// GET /mentorship/connections/mentee
func (mc *MentorshipController) GetMyMenteerships(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    res, err := mc.usecase.GetMyMenteerships(ctx, uid, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

// GET /mentorship/connections/active
func (mc *MentorshipController) GetActiveConnections(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    res, err := mc.usecase.GetActiveConnections(ctx, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

// POST /mentorship/connections/:id/interaction
func (mc *MentorshipController) UpdateLastInteraction(c *gin.Context) {
    connID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    if err := mc.usecase.UpdateLastInteraction(ctx, connID, uid); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Interaction updated"})
}

// POST /mentorship/connections/:id/pause
func (mc *MentorshipController) PauseConnection(c *gin.Context) {
    connID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    if err := mc.usecase.PauseConnection(ctx, connID, uid); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Connection paused"})
}

// POST /mentorship/connections/:id/resume
func (mc *MentorshipController) ResumeConnection(c *gin.Context) {
    connID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    if err := mc.usecase.ResumeConnection(ctx, connID, uid); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Connection resumed"})
}

// POST /mentorship/connections/:id/end
func (mc *MentorshipController) EndConnection(c *gin.Context) {
    connID := c.Param("id")
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    var body mentorshippkg.EndConnectionDTO
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }
    if err := body.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    if err := mc.usecase.EndConnection(ctx, connID, uid, body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Connection ended"})
}

// GET /mentorship/stats
func (mc *MentorshipController) GetMentorshipStats(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    stats, err := mc.usecase.GetMentorshipStats(ctx, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, stats)
}

// GET /mentorship/insights
func (mc *MentorshipController) GetMentorshipInsights(c *gin.Context) {
    uid := c.GetString("userID")
    if uid == "" {
        uid = c.GetString("user_id")
    }
    if uid == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    insights, err := mc.usecase.GetMentorshipInsights(ctx, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, insights)
}
