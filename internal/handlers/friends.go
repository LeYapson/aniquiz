package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

// SendFriendRequestHandler — POST /api/friends/request  {"username": "..."}
func SendFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var body struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username requis"})
		return
	}

	err := database.SendFriendRequest(userID.(int), body.Username)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": "Demande envoyée"})
	case errors.Is(err, database.ErrFriendUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, database.ErrFriendSelf),
		errors.Is(err, database.ErrAlreadyFriends),
		errors.Is(err, database.ErrRequestExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
	}
}

// RespondFriendRequestHandler — POST /api/friends/respond {"request_id":1,"accept":true}
func RespondFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var body struct {
		RequestID int  `json:"request_id"`
		Accept    bool `json:"accept"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RequestID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id requis"})
		return
	}

	err := database.RespondFriendRequest(userID.(int), body.RequestID, body.Accept)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	case errors.Is(err, database.ErrRequestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
	}
}

// RemoveFriendHandler — DELETE /api/friends/:id
func RemoveFriendHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	friendID, err := strconv.Atoi(c.Param("id"))
	if err != nil || friendID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalide"})
		return
	}

	if err := database.RemoveFriend(userID.(int), friendID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ami retiré"})
}

// ListFriendsHandler — GET /api/friends
func ListFriendsHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	friends, err := database.ListFriends(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if friends == nil {
		friends = []models.Friend{}
	}
	c.JSON(http.StatusOK, friends)
}

// ListFriendRequestsHandler — GET /api/friends/requests
func ListFriendRequestsHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	reqs, err := database.ListPendingRequests(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if reqs == nil {
		reqs = []models.FriendRequest{}
	}
	c.JSON(http.StatusOK, reqs)
}
