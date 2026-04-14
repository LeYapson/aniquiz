package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/LeYapson/aniquiz/internal/database"

)

func main() {
	// 1 - Connexion à la base de données
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err);
	}
	defer conn.Close(context.Background())
	fmt.Println("✅ Connecté avec succès à la base DBngin !")
}