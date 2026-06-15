package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	// 1 - Connexion à la base de données
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

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

	// 4 - Démarrage du serveur
	fmt.Println("Serveur lancé sur http://localhost:8080")
	router.Run(":8080")
}
