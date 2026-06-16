package handlers

import (
	"net/http"
	"os"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

// Store abstracts the database operations needed by HTTP handlers.
type Store interface {
	GetRandomTrack() (*models.Track, error)
	GetTrackByID(id int) (*models.Track, error)
	GetAllTracks() ([]models.Track, error)
	CreateUser(username, email, passwordHash string) error
	GetUserByUsernameOrEmail(identifier string) (*models.User, error)
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

	allowedOrigin := os.Getenv("FRONTEND_URL")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// --- ROUTES PUBLIQUES ---
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/health", func(c *gin.Context) {
		if err := database.Pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/rooms", ListRoomsHandler)

	router.GET("/api/leaderboard", func(c *gin.Context) {
		entries, err := database.GetLeaderboard(50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire le classement"})
			return
		}
		if entries == nil {
			entries = []models.LeaderboardEntry{}
		}
		c.JSON(http.StatusOK, entries)
	})

	router.GET("/animes", func(c *gin.Context) {
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

	authLimited := router.Group("/")
	authLimited.Use(AuthRateLimiter())
	{
		authLimited.POST("/api/auth/register", RegisterHandler(store))
		authLimited.POST("/api/auth/login", LoginHandler(store))
	}

	// --- ROUTES PROTÉGÉES PAR JWT ---
	protected := router.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/rooms", func(c *gin.Context) {
			var body struct {
				RoomID        string `json:"room_id"`
				IsPrivate     bool   `json:"is_private"`
				Password      string `json:"password"`
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

			creatorID, _ := c.Get("username")
			room := game.CreateRoom(body.RoomID, creatorID.(string))
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

		protected.GET("/quiz/next", func(c *gin.Context) {
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

		protected.POST("/api/admin/import", BatchImportHandler)
		protected.GET("/api/anime/search", AnimeSearchHandler)

		protected.GET("/api/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			user, err := database.GetUserByID(userID.(int))
			if err != nil || user == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "utilisateur introuvable"})
				return
			}
			c.JSON(http.StatusOK, user)
		})

		protected.GET("/api/history", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			results, err := database.GetUserHistory(userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire l'historique"})
				return
			}
			if results == nil {
				results = []models.GameResult{}
			}
			c.JSON(http.StatusOK, results)
		})

		protected.POST("/quiz/answer", func(c *gin.Context) {
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
	}

	return router
}
