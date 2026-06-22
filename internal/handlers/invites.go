package handlers

import (
	"net/http"
	"strconv"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

// SendRoomInviteHandler — POST /api/invites {"to_user_id":1,"room_id":"abc"}
// Invite un ami à rejoindre le salon courant. Refuse si les deux ne sont pas
// amis ou si le salon n'existe pas / est solo.
func SendRoomInviteHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(int)

	var body struct {
		ToUserID int    `json:"to_user_id"`
		RoomID   string `json:"room_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ToUserID == 0 || body.RoomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_user_id et room_id requis"})
		return
	}
	if body.ToUserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "impossible de s'inviter soi-même"})
		return
	}

	friends, err := database.AreFriends(uid, body.ToUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if !friends {
		c.JSON(http.StatusForbidden, gin.H{"error": "vous n'êtes pas amis"})
		return
	}

	// Le salon doit exister et ne pas être un solo ; on en récupère le mot de
	// passe pour que l'invité puisse rejoindre même un salon privé.
	game.RoomsMu.Lock()
	room, exists := game.ActiveRooms[body.RoomID]
	game.RoomsMu.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "salon introuvable"})
		return
	}
	room.Mu.Lock()
	pwd := room.Password
	isSolo := room.IsSolo
	room.Mu.Unlock()
	if isSolo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "salon non invitable"})
		return
	}

	if err := database.CreateRoomInvite(uid, body.ToUserID, body.RoomID, pwd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation envoyée"})
}

// ListRoomInvitesHandler — GET /api/invites
// Ne renvoie que les invitations vers des salons encore actifs.
func ListRoomInvitesHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	invites, err := database.ListRoomInvites(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	game.RoomsMu.Lock()
	active := make(map[string]bool, len(game.ActiveRooms))
	for id := range game.ActiveRooms {
		active[id] = true
	}
	game.RoomsMu.Unlock()

	out := make([]models.RoomInvite, 0, len(invites))
	for _, inv := range invites {
		if active[inv.RoomID] {
			out = append(out, inv)
		}
	}
	c.JSON(http.StatusOK, out)
}

// DismissRoomInviteHandler — DELETE /api/invites/:id
func DismissRoomInviteHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalide"})
		return
	}
	if err := database.DeleteRoomInvite(userID.(int), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
