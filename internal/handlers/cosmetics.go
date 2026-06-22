package handlers

import (
	"net/http"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/gin-gonic/gin"
)

// frameUnlockLevel : niveau requis pour chaque cadre d'avatar.
// DOIT rester synchronisé avec FRAMES côté frontend (src/cosmetics.js).
var frameUnlockLevel = map[string]int{
	"":         1,
	"bronze":   2,
	"silver":   5,
	"gold":     10,
	"emerald":  15,
	"sapphire": 20,
	"ruby":     30,
	"rainbow":  50,
}

// SetCosmeticsHandler — PUT /api/me/cosmetics {"avatar_frame":"gold"}
// Valide que le cadre existe et que le niveau de l'utilisateur le débloque.
func SetCosmeticsHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var body struct {
		AvatarFrame string `json:"avatar_frame"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar_frame requis"})
		return
	}

	reqLevel, known := frameUnlockLevel[body.AvatarFrame]
	if !known {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cadre inconnu"})
		return
	}

	user, err := database.GetUserByID(userID.(int))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "utilisateur introuvable"})
		return
	}
	if user.Level < reqLevel {
		c.JSON(http.StatusForbidden, gin.H{"error": "cadre non débloqué"})
		return
	}

	if err := database.SetAvatarFrame(user.ID, body.AvatarFrame); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_frame": body.AvatarFrame})
}
