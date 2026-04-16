package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/game"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/LeYapson/aniquiz/internal/sourcing"

	"github.com/gin-gonic/gin"
)

func main() {

	type AnswerRequest struct {
		TrackID int `json:"track_id"`
		Answer string `json:"answer"`
	}



	// 1 - Connexion à la base de données
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err);
	}
	defer conn.Close(context.Background())
	

	// 2 - Initialisation du router Gin
	router := gin.Default()

	// 3 - Définition des routes
	router.GET("/ping", func(c *gin.Context)  {
		c.JSON(http.StatusOK, gin.H {
			"message": "pong",
			"status": "connected to DB",
		})
	})

	router.GET("/anime/:id", func(c *gin.Context) {
		animeId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H {
				"error": "Invalid anime ID",
			})
			return
		}
		info, err := sourcing.ProcessAndSaveAnime(animeId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H {
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H {
			"message": "Anime traité et musiques sauvegardées ! ",
			"anime": info.Title,
		})
	})

	router.GET("/test-audio", func(c *gin.Context)  {
		tracks, err := database.GetAllTracks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			c.String(500, "Erreur DB : %v", err)
			return
		}

		// Un peu de CSS pour ne pas saigner des yeux
    html := `
    <html>
        <head>
            <title>AniQuiz - Audio Test</title>
            <style>
                body { font-family: sans-serif; background: #1a1a1a; color: white; padding: 20px; }
                .track-card { background: #333; padding: 15px; border-radius: 8px; margin-bottom: 10px; border-left: 5px solid #e91e63; }
                audio { width: 100%; margin-top: 10px; }
                .anime-title { color: #ff9800; font-weight: bold; }
            </style>
        </head>
        <body>
            <h1>🎵 Testeur de Sourcing Audio</h1>
            <p>Nombre de pistes en base : ` + fmt.Sprint(len(tracks)) + `</p>`

    for _, t := range tracks {
        html += fmt.Sprintf(`
            <div class="track-card">
                <span class="anime-title">[%d] %s</span><br>
                <strong>%s</strong> - %s<br>
                <audio controls>
                    <source src="%s" type="video/webm">
                    <source src="%s" type="audio/mpeg">
                    Votre navigateur ne supporte pas le lecteur.
                </audio>
            </div>`, t.MalID, t.AnimeName, t.Title, t.Artist, t.AudioURL, t.AudioURL)
    }

    html += "</body></html>"
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	router.GET("/quiz/next", func(c *gin.Context)  {
		track, err := database.GetRandomTrack()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de lire la DB"})
			return
		}

		//On prépare la question (on cache la réponse)
		question := models.QuizQuestion{
			AudioURL: track.AudioURL,
		}

		// optionnel : on peut sotcker l'id de la réponse en session
		//ou envoyer l'ID crypté pour vérifier la réponse plus tard

		c.JSON(http.StatusOK, gin.H{
			"question": question,
			"debug_id": track.ID, // à retirer en prod, c'est juste pour vérifier que ça marche
		})
		
	})

	router.POST("/quiz/answer", func(c *gin.Context) {
		var req AnswerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		//1 recuperer la vrai réponse en DB
		track, err := database.GetTrackByID(req.TrackID)
		if err != nil {
			c.JSON(404, gin.H{"error": "Musique introuvable"})
			return
		}

		//2 comparer la réponse de l'utilisateur avec la vrai réponse
		success := game.IsCorrect(req.Answer, track.AnimeName)

		c.JSON(200, gin.H{
			"correct": success,
			"expected": track.AnimeName,
			"your answer": req.Answer,
		})
	})

	router.POST("/rooms", func(c *gin.Context) {
    // Générer un ID simple (ex: "ABCD")
    roomID := "ROOM123" 
    room := game.CreateRoom(roomID)
    
    c.JSON(200, gin.H{
        "message": "Salon créé",
        "room_id": room.ID,
    })
})

	// 4 - Démarrage du serveur
	fmt.Println("Serveur lancé sur le http://localhost:8080")
	router.Run(":8080")
	
}