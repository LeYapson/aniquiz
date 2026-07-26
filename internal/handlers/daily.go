package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// today retourne la date du jour en UTC, tronquée à minuit.
func today() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// DailyHandler — GET /api/daily
// Retourne la piste du jour + les choix QCM + le résultat si déjà joué.
func DailyHandler(c *gin.Context) {
	userID := c.GetInt("userID")
	date := today()

	track, err := database.GetDailyTrack(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de charger le quiz du jour"})
		return
	}

	startFraction := database.DailyStartFraction(date)

	// Génère 3 mauvaises réponses + la bonne, mélangées.
	wrong, _ := database.GetRandomAnimeNames(track.AnimeName, 3)
	choices := append(wrong, track.AnimeName)
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

	// Vérifie si l'utilisateur a déjà joué aujourd'hui.
	result, err := database.GetDailyResult(userID, date)
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur base de données"})
		return
	}

	resp := gin.H{
		"date":           date.Format("2006-01-02"),
		"audio_url":      track.AudioURL,
		"start_fraction": startFraction,
		"choices":        choices,
		"already_played": result != nil,
	}

	if result != nil {
		resp["result"] = result
		resp["answer"] = track.AnimeName
		resp["title"] = track.Title
		resp["artist"] = track.Artist
		resp["video_url"] = track.AudioURL
	}

	c.JSON(http.StatusOK, resp)
}

// DailySubmitHandler — POST /api/daily/submit
func DailySubmitHandler(c *gin.Context) {
	userID := c.GetInt("userID")
	date := today()

	// Anti-rejeu : une seule tentative par jour.
	existing, _ := database.GetDailyResult(userID, date)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "déjà joué aujourd'hui"})
		return
	}

	var body struct {
		Answer string `json:"answer" binding:"required"`
		TimeMs int    `json:"time_ms"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "réponse invalide"})
		return
	}

	track, err := database.GetDailyTrack(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de charger la piste"})
		return
	}

	result := game.VerifyAnswerMode(body.Answer, track, game.GuessModeAnime)
	found := result.Points > 0

	// Borne défensive : on n'accepte pas un time_ms négatif ou irréaliste.
	timeMs := body.TimeMs
	if timeMs < 0 {
		timeMs = 0
	}
	// Un round dure au max 60 s — on clamp à 60 000 ms pour éviter la triche.
	if timeMs > 60_000 {
		timeMs = 60_000
	}

	if err := database.SaveDailyResult(userID, date, found, timeMs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de sauvegarder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"correct":   found,
		"answer":    track.AnimeName,
		"title":     track.Title,
		"artist":    track.Artist,
		"video_url": track.AudioURL,
	})
}

// DailyLeaderboardHandler — GET /api/daily/leaderboard
func DailyLeaderboardHandler(c *gin.Context) {
	entries, err := database.GetDailyLeaderboard(today(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de charger le classement"})
		return
	}
	if entries == nil {
		entries = []database.DailyLeaderboardEntry{}
	}
	// Masque les ms si le joueur n'a pas trouvé.
	for i := range entries {
		if !entries[i].Found {
			entries[i].TimeMs = 0
		}
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, entries)
}

