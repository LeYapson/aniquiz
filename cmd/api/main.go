package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	// Charge .env si présent (ignoré en production où les vars sont déjà exportées)
	_ = godotenv.Load()

	// 1 - Connexion à la base de données
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 1b - Migrations de schéma (idempotentes, sûres à relancer)
	if err := database.Migrate(); err != nil {
		log.Fatalf("Erreur migration : %v", err)
	}

	// 2 - Router avec les routes testables (ping, rooms, quiz/next, quiz/answer)
	store := &handlers.PgStore{}
	router := handlers.NewRouter(store)

	// 3 - Routes spécifiques au serveur (WebSocket, scraping, debug)

	router.GET("/ws", func(c *gin.Context) {
		roomID := c.Query("room")
		password := c.Query("password")
		tokenString := c.Query("token")

		// 1. Validation du token JWT
		claims, err := handlers.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou manquant"})
			return
		}

		game.RoomsMu.Lock()
		room, exists := game.ActiveRooms[roomID]
		game.RoomsMu.Unlock()

		fmt.Printf("--- Tentative de connexion WebSocket ---\n")
		fmt.Printf("RoomID: %s | User: %s | Existe: %t\n", roomID, claims.Username, exists)

		// 2. Le salon doit exister
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Le salon n'existe pas. Veuillez le créer d'abord."})
			return
		}

		// 3. Vérification du mot de passe si le salon est privé
		room.Mu.Lock()
		if room.IsPrivate && room.Password != password {
			room.Mu.Unlock()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe incorrect ou requis pour ce salon"})
			return
		}
		room.Mu.Unlock()

		// 4. Upgrade WebSocket
		wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Erreur Upgrade WS: %v", err)
			return
		}

		client := &game.Client{
			ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
			UserID:   claims.UserID,
			Username: claims.Username,
			Conn:     wsConn,
			Room:     room,
			Send:     make(chan []byte, 256),
		}

		room.Register <- client
		go client.ReadPump()
		go client.WritePump()
	})

	router.GET("/anime/:id", func(c *gin.Context) {
		animeId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anime ID"})
			return
		}
		info, err := sourcing.ProcessAndSaveAnime(animeId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Anime traité et musiques sauvegardées !",
			"anime":   info.Title,
		})
	})

	router.GET("/test-audio", func(c *gin.Context) {
		tracks, err := database.GetAllTracks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			return
		}

		html := `<html><head><title>AniQuiz - Audio Test</title>
		<style>body{font-family:sans-serif;background:#1a1a1a;color:white;padding:20px}
		.track-card{background:#333;padding:15px;border-radius:8px;margin-bottom:10px;border-left:5px solid #e91e63}
		audio{width:100%;margin-top:10px}.anime-title{color:#ff9800;font-weight:bold}</style>
		</head><body><h1>🎵 Testeur de Sourcing Audio</h1>
		<p>Nombre de pistes en base : ` + fmt.Sprint(len(tracks)) + `</p>`

		for _, t := range tracks {
			html += fmt.Sprintf(`<div class="track-card">
			<span class="anime-title">[%d] %s</span><br>
			<strong>%s</strong> - %s<br>
			<audio controls><source src="%s" type="video/webm"><source src="%s" type="audio/mpeg"></audio>
			</div>`, t.MalID, t.AnimeName, t.Title, t.Artist, t.AudioURL, t.AudioURL)
		}
		html += "</body></html>"
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// 4 - OAuth AniList
	// GET /api/auth/anilist — redirige l'utilisateur vers la page d'autorisation AniList
	router.GET("/api/auth/anilist", func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token requis"})
			return
		}
		// Valider le JWT avant de lancer l'OAuth
		if _, err := handlers.ValidateToken(tokenString); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
			return
		}
		// Le JWT est passé comme state : AniList le renverra dans le callback
		authURL := sourcing.BuildAuthURL(tokenString)
		c.Redirect(http.StatusFound, authURL)
	})

	// GET /api/auth/anilist/callback — AniList redirige ici après autorisation
	router.GET("/api/auth/anilist/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state") // contient le JWT de l'utilisateur
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		if code == "" || state == "" {
			c.Redirect(http.StatusFound, frontendURL+"?anilist=error&reason=missing_params")
			return
		}

		// Valider le JWT récupéré depuis le state
		claims, err := handlers.ValidateToken(state)
		if err != nil {
			c.Redirect(http.StatusFound, frontendURL+"?anilist=error&reason=invalid_token")
			return
		}

		// Échanger le code contre un access token AniList
		accessToken, err := sourcing.ExchangeCode(code)
		if err != nil {
			log.Printf("Erreur échange code AniList: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?anilist=error&reason=exchange_failed")
			return
		}

		// Récupérer le profil AniList
		profile, err := sourcing.GetAnilistProfile(accessToken)
		if err != nil {
			log.Printf("Erreur profil AniList: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?anilist=error&reason=profile_failed")
			return
		}

		// Sauvegarder en base
		if err := database.UpdateUserAnilist(claims.UserID, profile.ID, profile.Username, accessToken); err != nil {
			log.Printf("Erreur sauvegarde AniList: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?anilist=error&reason=db_failed")
			return
		}

		c.Redirect(http.StatusFound, frontendURL+"?anilist=success&username="+profile.Username)
	})

	// 5 - OAuth MyAnimeList (PKCE)
	// GET /api/auth/mal — redirige vers MAL avec code_challenge PKCE
	router.GET("/api/auth/mal", func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token requis"})
			return
		}
		if _, err := handlers.ValidateToken(tokenString); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
			return
		}
		authURL, _, err := sourcing.BuildMALAuthURL(tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de générer l'URL MAL"})
			return
		}
		c.Redirect(http.StatusFound, authURL)
	})

	// GET /api/auth/mal/callback — MAL redirige ici après autorisation
	router.GET("/api/auth/mal/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		if code == "" || state == "" {
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=missing_params")
			return
		}

		// Décoder le state pour récupérer JWT + code_verifier
		jwtToken, verifier, err := sourcing.DecodeMalState(state)
		if err != nil {
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=invalid_state")
			return
		}

		claims, err := handlers.ValidateToken(jwtToken)
		if err != nil {
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=invalid_token")
			return
		}

		// Échanger le code contre un access token MAL (avec le verifier PKCE)
		accessToken, err := sourcing.ExchangeMALCode(code, verifier)
		if err != nil {
			log.Printf("Erreur échange code MAL: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=exchange_failed")
			return
		}

		// Récupérer le profil MAL
		profile, err := sourcing.GetMALProfile(accessToken)
		if err != nil {
			log.Printf("Erreur profil MAL: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=profile_failed")
			return
		}

		// Sauvegarder en base
		if err := database.UpdateUserMAL(claims.UserID, profile.ID, profile.Username, accessToken); err != nil {
			log.Printf("Erreur sauvegarde MAL: %v", err)
			c.Redirect(http.StatusFound, frontendURL+"?mal=error&reason=db_failed")
			return
		}

		c.Redirect(http.StatusFound, frontendURL+"?mal=success&username="+profile.Username)
	})

	// 6 - Fichiers statiques du frontend (build Vue)
	// En production l'image Docker copie dist/ dans ./static
	if _, err := os.Stat("./static"); err == nil {
		router.Static("/assets", "./static/assets")
		router.StaticFile("/favicon.ico", "./static/favicon.ico")
		// SPA fallback : toute route inconnue sert index.html
		router.NoRoute(func(c *gin.Context) {
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "route introuvable"})
				return
			}
			c.File("./static/index.html")
		})
	}

	// 7 - Démarrage du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Serveur lancé sur http://localhost:%s\n", port)
	router.Run(":" + port)
}
