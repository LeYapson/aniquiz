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

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/rooms", ListRoomsHandler)
	router.POST("/rooms", func(c *gin.Context) {
		var body struct {
			RoomID string `json:"room_id"`
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

		room := game.CreateRoom(body.RoomID, "creator-id-placeholder") // On peut ajouter un vrai creatorID plus tard
		go room.Run()
		game.ActiveRooms[body.RoomID] = room

		c.JSON(http.StatusCreated, gin.H{"message": "Salon créé", "room_id": room.ID})
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
