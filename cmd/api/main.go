package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
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

	// 4 - Démarrage du serveur
	fmt.Println("Serveur lancé sur le http://localhost:8080")
	router.Run(":8080")
}