package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// announceClient : timeout court pour le POST vers le webhook Discord.
var announceClient = &http.Client{Timeout: 10 * time.Second}

// tagColor associe une couleur d'embed Discord (décimal) à un tag de news.
func tagColor(tag string) int {
	switch tag {
	case "Feature":
		return 0x22c55e // vert
	case "Fix":
		return 0xf97316 // orange
	default: // Annonce, etc.
		return 0x3b82f6 // bleu
	}
}

// AnnounceHandler — POST /api/admin/announce
// Publie une news sur le salon Discord via un webhook (DISCORD_NEWS_WEBHOOK_URL).
func AnnounceHandler(c *gin.Context) {
	webhook := os.Getenv("DISCORD_NEWS_WEBHOOK_URL")
	if webhook == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook Discord non configuré (DISCORD_NEWS_WEBHOOK_URL)"})
		return
	}

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Tag     string `json:"tag"`
		Date    string `json:"date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title requis"})
		return
	}

	// Discord : description d'embed limitée à 4096 caractères.
	desc := body.Content
	if len([]rune(desc)) > 4000 {
		desc = string([]rune(desc)[:4000]) + "…"
	}

	footer := "AniQuiz"
	if body.Date != "" {
		footer = "AniQuiz · " + body.Date
	}

	payload := map[string]interface{}{
		"username": "AniQuiz",
		"embeds": []map[string]interface{}{
			{
				"title":       "📰 " + body.Title,
				"description": desc,
				"color":       tagColor(body.Tag),
				"footer":      map[string]string{"text": footer},
			},
		},
	}
	raw, _ := json.Marshal(payload)

	resp, err := announceClient.Post(webhook, "application/json", bytes.NewReader(raw))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "échec de l'envoi à Discord"})
		return
	}
	defer resp.Body.Close()

	// Discord renvoie 204 No Content en cas de succès.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Discord a refusé la publication"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Publié sur Discord"})
}
