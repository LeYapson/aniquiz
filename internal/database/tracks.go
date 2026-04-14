package database

import (
	"context"
	"fmt"
	"github.com/LeYapson/aniquiz/internal/models"
)

func SaveTrack(track models.Track) error {
	conn, err := Connect()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	// La requete SQL avec des placeholder ($1, $2, $3) pour eviter les injections SQL
	query := `INSERT INTO tracks (title, anime_name, artist, audio_url, difficulty)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err = conn.Exec(context.Background(), query, track.Title, track.AnimeName, track.Artist, track.AudioURL, track.Difficulty)

	if err != nil {
		return fmt.Errorf("Erreur lors  de l'insertion : %v", err)
	}

	return nil
}