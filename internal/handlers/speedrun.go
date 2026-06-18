package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

const speedrunDuration = 5 * time.Minute

type speedrunSession struct {
	UserID       int
	Score        int
	CurrentTrack *models.Track
	ExpiresAt    time.Time
	mu           sync.Mutex
}

var (
	speedrunSessions   = make(map[string]*speedrunSession)
	speedrunSessionsMu sync.Mutex
)

func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func cleanExpiredSessions() {
	speedrunSessionsMu.Lock()
	defer speedrunSessionsMu.Unlock()
	grace := time.Now().Add(-30 * time.Second)
	for id, s := range speedrunSessions {
		if s.ExpiresAt.Before(grace) {
			delete(speedrunSessions, id)
		}
	}
}

// StartSpeedrunHandler POST /api/speedrun/start
func StartSpeedrunHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		go cleanExpiredSessions()

		userID, _ := c.Get("userID")
		uid := userID.(int)

		track, err := store.GetRandomTrack()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de récupérer une piste"})
			return
		}

		sessionID, err := generateSessionID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur interne"})
			return
		}

		session := &speedrunSession{
			UserID:       uid,
			Score:        0,
			CurrentTrack: track,
			ExpiresAt:    time.Now().Add(speedrunDuration),
		}

		speedrunSessionsMu.Lock()
		speedrunSessions[sessionID] = session
		speedrunSessionsMu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"expires_at": session.ExpiresAt,
			"track": gin.H{
				"id":        track.ID,
				"audio_url": track.AudioURL,
			},
		})
	}
}

// AnswerSpeedrunHandler POST /api/speedrun/answer
func AnswerSpeedrunHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
			Answer    string `json:"answer"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id et answer requis"})
			return
		}

		speedrunSessionsMu.Lock()
		session, exists := speedrunSessions[req.SessionID]
		speedrunSessionsMu.Unlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session introuvable"})
			return
		}

		session.mu.Lock()
		defer session.mu.Unlock()

		timeRemaining := time.Until(session.ExpiresAt)
		if timeRemaining <= 0 {
			_ = store.SaveSpeedrunResult(session.UserID, session.Score)
			c.JSON(http.StatusOK, gin.H{
				"finished":       true,
				"score":          session.Score,
				"time_remaining": 0,
			})
			return
		}

		result := game.VerifyAnswer(req.Answer, session.CurrentTrack)
		if !result.IsCorrect {
			c.JSON(http.StatusOK, gin.H{
				"correct":        false,
				"time_remaining": int(timeRemaining.Seconds()),
			})
			return
		}

		session.Score++
		nextTrack, err := store.GetRandomTrack()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de récupérer la piste suivante"})
			return
		}
		session.CurrentTrack = nextTrack

		c.JSON(http.StatusOK, gin.H{
			"correct":        true,
			"score":          session.Score,
			"time_remaining": int(timeRemaining.Seconds()),
			"next_track": gin.H{
				"id":        nextTrack.ID,
				"audio_url": nextTrack.AudioURL,
			},
		})
	}
}

// SkipSpeedrunHandler POST /api/speedrun/skip
func SkipSpeedrunHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id requis"})
			return
		}

		speedrunSessionsMu.Lock()
		session, exists := speedrunSessions[req.SessionID]
		speedrunSessionsMu.Unlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session introuvable"})
			return
		}

		session.mu.Lock()
		defer session.mu.Unlock()

		timeRemaining := time.Until(session.ExpiresAt)
		if timeRemaining <= 0 {
			_ = store.SaveSpeedrunResult(session.UserID, session.Score)
			c.JSON(http.StatusOK, gin.H{
				"finished":       true,
				"score":          session.Score,
				"time_remaining": 0,
			})
			return
		}

		nextTrack, err := store.GetRandomTrack()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de récupérer la piste suivante"})
			return
		}
		session.CurrentTrack = nextTrack

		c.JSON(http.StatusOK, gin.H{
			"skipped":        true,
			"score":          session.Score,
			"time_remaining": int(timeRemaining.Seconds()),
			"next_track": gin.H{
				"id":        nextTrack.ID,
				"audio_url": nextTrack.AudioURL,
			},
		})
	}
}

// FinishSpeedrunHandler POST /api/speedrun/finish
func FinishSpeedrunHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id requis"})
			return
		}

		speedrunSessionsMu.Lock()
		session, exists := speedrunSessions[req.SessionID]
		if exists {
			delete(speedrunSessions, req.SessionID)
		}
		speedrunSessionsMu.Unlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session introuvable ou déjà terminée"})
			return
		}

		session.mu.Lock()
		score := session.Score
		uid := session.UserID
		session.mu.Unlock()

		if err := store.SaveSpeedrunResult(uid, score); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de sauvegarder le score"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"score": score})
	}
}

// SpeedrunLeaderboardHandler GET /api/leaderboard/speedrun
func SpeedrunLeaderboardHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := store.GetSpeedrunLeaderboard(50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire le classement"})
			return
		}
		if entries == nil {
			entries = []models.SpeedrunLeaderboardEntry{}
		}
		c.JSON(http.StatusOK, entries)
	}
}
