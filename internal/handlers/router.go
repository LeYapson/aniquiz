package handlers

import (
	"net/http"

	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

// Store abstracts the database operations needed by HTTP handlers.
type Store interface {
	GetRandomTrack() (*models.Track, error)
	GetTrackByID(id int) (*models.Track, error)
	GetAllTracks() ([]models.Track, error)
}

// AnswerRequest is the expected body for POST /quiz/answer.
type AnswerRequest struct {
	TrackID int    `json:"track_id"`
	Answer  string `json:"answer"`
}

func ListRoomsHandler(c *gin.Context) {
	rooms := game.GetPublicRooms()
	c.JSON(http.StatusOK, rooms)
}

// NewRouter builds the Gin engine with all testable routes wired up.
func NewRouter(store Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Ton URL Front-end
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Gérer les requêtes de pré-vérification (Preflight) envoyées par le navigateur
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/rooms", ListRoomsHandler)

	// --- NOUVEL ENDPOINT POUR L'AUTO-COMPLÉTION ---
	router.GET("/animes", func(c *gin.Context) {
		// Utilisation du store injecté dans NewRouter
		tracks, err := store.GetAllTracks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			return
		}

		animeSet := make(map[string]struct{})
		for _, track := range tracks {
			if track.AnimeName != "" {
				animeSet[track.AnimeName] = struct{}{}
			}
		}

		animeNames := make([]string, 0, len(animeSet))
		for name := range animeSet {
			animeNames = append(animeNames, name)
		}

		c.JSON(http.StatusOK, animeNames)
	})

	router.POST("/rooms", func(c *gin.Context) {
		var body struct {
			RoomID        string `json:"room_id"`
			IsPrivate     bool   `json:"is_private"`
			Password      string `json:"password"`
			CreatorID     string `json:"creator_id"` // Reçu du front ou généré
			MaxRounds     int    `json:"max_rounds"`
			RoundDuration int    `json:"round_duration"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.RoomID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room_id requis"})
			return
		}

		game.RoomsMu.Lock()
		defer game.RoomsMu.Unlock()

		if _, exists := game.ActiveRooms[body.RoomID]; exists {
			c.JSON(http.StatusConflict, gin.H{"error": "salon déjà existant"})
			return
		}

		// Si aucune ID de créateur n'est fournie, on peut en générer une temporaire
		creatorID := body.CreatorID
		if creatorID == "" {
			creatorID = "admin-" + body.RoomID // Sécurité par défaut simple
		}

		room := game.CreateRoom(body.RoomID, creatorID)

		room.IsPrivate = body.IsPrivate
		room.Password = body.Password

		if body.MaxRounds > 0 {
			room.MaxRounds = body.MaxRounds
		}

		if body.RoundDuration > 0 {
			room.RoundDuration = body.RoundDuration
		}

		go room.Run()
		game.ActiveRooms[body.RoomID] = room

		c.JSON(http.StatusCreated, gin.H{"message": "Salon créé", "room_id": room.ID, "creator_id": creatorID})
	})

	router.GET("/quiz/next", func(c *gin.Context) {
		track, err := store.GetRandomTrack()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"question": models.QuizQuestion{AudioURL: track.AudioURL},
			"debug_id": track.ID,
		})
	})

	router.POST("/quiz/answer", func(c *gin.Context) {
		var req AnswerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		track, err := store.GetTrackByID(req.TrackID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Musique introuvable"})
			return
		}

		result := game.VerifyAnswer(req.Answer, track)
		c.JSON(http.StatusOK, gin.H{
			"correct":     result.IsCorrect,
			"points":      result.Points,
			"message":     result.Message,
			"expected":    track.AnimeName,
			"your_answer": req.Answer,
		})
	})

	return router
}