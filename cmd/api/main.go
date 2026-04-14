package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LeYapson/aniquiz/internal/database"

)

func main() {
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err);
	}
	defer conn.Close(context.Background())
	fmt.Println("✅ Connecté avec succès à la base DBngin !")
}