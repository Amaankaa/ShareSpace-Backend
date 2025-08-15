package controllers

import (
	"net/http"
	"strconv"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessagingController struct{ uc msgpkg.IMessagingUsecase }

func NewMessagingController(uc msgpkg.IMessagingUsecase) *MessagingController {
	return &MessagingController{uc: uc}
}

func (mc *MessagingController) CreateConversation(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetString("user_id")
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var body struct {
		ParticipantIDs []string `json:"participantIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var ids []primitive.ObjectID
	for _, s := range body.ParticipantIDs {
		if id, err := primitive.ObjectIDFromHex(s); err == nil {
			ids = append(ids, id)
		}
	}
	uid, _ := primitive.ObjectIDFromHex(userID)
	conv, err := mc.uc.CreateConversation(c.Request.Context(), uid, ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, conv)
}

func (mc *MessagingController) GetConversations(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetString("user_id")
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	uid, _ := primitive.ObjectIDFromHex(userID)
	list, err := mc.uc.GetUserConversations(c.Request.Context(), uid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (mc *MessagingController) GetMessages(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		userID = c.GetString("user_id")
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	convIDHex := c.Param("id")
	convID, err := primitive.ObjectIDFromHex(convIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	uid, _ := primitive.ObjectIDFromHex(userID)
	list, err := mc.uc.GetMessages(c.Request.Context(), uid, convID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
