package handlers

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/gin-gonic/gin"
)

// BotAuthMiddleware protège les routes /api/bot/* via une clé partagée
// (en-tête X-Bot-Key == env BOT_API_KEY). Machine-to-machine : pas de JWT.
func BotAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := os.Getenv("BOT_API_KEY")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "bot API non configurée"})
			return
		}
		got := c.GetHeader("X-Bot-Key")
		// Comparaison à temps constant pour éviter les attaques temporelles.
		if subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "clé bot invalide"})
			return
		}
		c.Next()
	}
}

// BotProfileHandler — GET /api/bot/profile?discord_id=...
// Renvoie les stats du joueur lié à cet id Discord (404 si non lié).
func BotProfileHandler(c *gin.Context) {
	discordID := c.Query("discord_id")
	if discordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discord_id requis"})
		return
	}
	user, err := database.GetUserByDiscordID(discordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compte non lié"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"level":    user.Level,
		"xp":       user.Xp,
	})
}

// BotFriendRequestHandler — POST /api/bot/friends/request {discord_id, target}
// Envoie une demande d'ami au nom du joueur lié à cet id Discord.
func BotFriendRequestHandler(c *gin.Context) {
	var body struct {
		DiscordID string `json:"discord_id"`
		Target    string `json:"target"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.DiscordID == "" || body.Target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discord_id et target requis"})
		return
	}

	user, err := database.GetUserByDiscordID(body.DiscordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "compte non lié"})
		return
	}

	switch err := database.SendFriendRequest(user.ID, body.Target); {
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
