package handlers

import (
	"net/http"
	"os"
	"unicode"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
)

// Store abstracts the database operations needed by HTTP handlers.
type Store interface {
	GetRandomTrack() (*models.Track, error)
	GetTrackByID(id int) (*models.Track, error)
	GetAllTracks() ([]models.Track, error)
	GetDistinctAnimeNames() ([]string, error)
	CreateUser(username, email, passwordHash string) error
	GetUserByUsernameOrEmail(identifier string) (*models.User, error)
	SaveSpeedrunResult(userID, score int) error
	GetSpeedrunLeaderboard(limit int) ([]models.SpeedrunLeaderboardEntry, error)
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

	router.GET("/api/leaderboard/speedrun", SpeedrunLeaderboardHandler(store))

	router.GET("/animes", func(c *gin.Context) {
		names, err := store.GetDistinctAnimeNames()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			return
		}
		if names == nil {
			names = []string{}
		}
		c.JSON(http.StatusOK, names)
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
				IsSolo        bool   `json:"is_solo"`
				Password      string `json:"password"`
				MaxRounds     int    `json:"max_rounds"`
				RoundDuration int    `json:"round_duration"`
			}
			if err := c.ShouldBindJSON(&body); err != nil || body.RoomID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "room_id requis"})
				return
			}
			// Restrict room IDs to safe characters to prevent log injection or path traversal.
			if len(body.RoomID) > 64 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "room_id trop long (max 64 caractères)"})
				return
			}
			for _, r := range body.RoomID {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
					c.JSON(http.StatusBadRequest, gin.H{"error": "room_id invalide (lettres, chiffres, - et _ uniquement)"})
					return
				}
			}

			game.RoomsMu.Lock()
			defer game.RoomsMu.Unlock()

			if _, exists := game.ActiveRooms[body.RoomID]; exists {
				c.JSON(http.StatusConflict, gin.H{"error": "salon déjà existant"})
				return
			}

			creatorID, _ := c.Get("username")
			room := game.CreateRoom(body.RoomID, creatorID.(string), body.IsSolo)
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

		// Import d'un anime : fonctionnalité ouverte à tous (crowdsourcing, rate-limité).
		protected.POST("/api/admin/import", BatchImportHandler)
		protected.GET("/api/anime/search", AnimeSearchHandler)

		// Le client demande s'il est admin (pour afficher l'onglet Admin).
		protected.GET("/api/me/admin", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{"is_admin": IsAdmin(username.(string))})
		})

		// Opérations lourdes d'administration : réservées aux admins.
		admin := protected.Group("")
		admin.Use(AdminMiddleware())
		{
			admin.POST("/api/admin/seed", SeedHandler)
			admin.GET("/api/admin/seed/status", SeedStatusHandler)
			admin.POST("/api/admin/audio/healthcheck", AudioHealthcheckHandler)
			admin.GET("/api/admin/audio/healthcheck/status", AudioHealthStatusHandler)
		}

		// Retourne les MAL IDs depuis la liste AniList et/ou MAL de l'utilisateur
		protected.GET("/api/me/anime-ids", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			user, err := database.GetUserByID(userID.(int))
			if err != nil || user == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "utilisateur introuvable"})
				return
			}

			seen := make(map[int]bool)
			var ids []int

			if user.AnilistToken != "" && user.AnilistUserID > 0 {
				anilistIDs, err := sourcing.GetAnilistAnimeList(user.AnilistToken, user.AnilistUserID)
				if err == nil {
					for _, id := range anilistIDs {
						if !seen[id] {
							seen[id] = true
							ids = append(ids, id)
						}
					}
				}
			}

			if user.MalToken != "" {
				malIDs, err := sourcing.GetMALAnimeList(user.MalToken)
				if err == nil {
					for _, id := range malIDs {
						if !seen[id] {
							seen[id] = true
							ids = append(ids, id)
						}
					}
				}
			}

			if ids == nil {
				ids = []int{}
			}

			// Combien de ces animés ont réellement des pistes jouables ?
			// Permet à l'UI de prévenir si la librairie ne couvre presque rien
			// de la liste perso de l'utilisateur.
			playableAnime, playableTracks, err := database.CountPlayableForMalIDs(ids)
			if err != nil {
				playableAnime, playableTracks = 0, 0
			}

			c.JSON(http.StatusOK, gin.H{
				"ids":             ids,
				"playable_anime":  playableAnime,
				"playable_tracks": playableTracks,
			})
		})

		// Système d'amis
		protected.GET("/api/friends", ListFriendsHandler)
		protected.GET("/api/friends/requests", ListFriendRequestsHandler)
		protected.POST("/api/friends/request", SendFriendRequestHandler)
		protected.POST("/api/friends/respond", RespondFriendRequestHandler)
		protected.DELETE("/api/friends/:id", RemoveFriendHandler)

		// Cosmétiques (cadre d'avatar)
		protected.PUT("/api/me/cosmetics", SetCosmeticsHandler)

		// Invitations à rejoindre un salon
		protected.GET("/api/invites", ListRoomInvitesHandler)
		protected.POST("/api/invites", SendRoomInviteHandler)
		protected.DELETE("/api/invites/:id", DismissRoomInviteHandler)

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

		// --- SPEED RUN ---
		protected.POST("/api/speedrun/start", StartSpeedrunHandler(store))
		protected.POST("/api/speedrun/answer", AnswerSpeedrunHandler(store))
		protected.POST("/api/speedrun/skip", SkipSpeedrunHandler(store))
		protected.POST("/api/speedrun/finish", FinishSpeedrunHandler(store))

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
