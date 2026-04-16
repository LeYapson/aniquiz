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
	query := `INSERT INTO tracks (title, anime_name, artist, audio_url, difficulty, mal_id)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  ON CONFLICT DO NOTHING` //evite les doublons

	_, err = conn.Exec(context.Background(), query, track.Title, track.AnimeName, track.Artist, track.AudioURL, track.Difficulty, track.MalID)

	if err != nil {
		return fmt.Errorf("Erreur lors  de l'insertion : %v", err)
	}

	return nil
}

func GetAllTracks() ([]models.Track, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	//On recupere tous ce qu'on a en base
	rows, err := conn.Query(context.Background(),
		"SELECT title, artist, anime_name, audio_url, mal_id FROM tracks")
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la récupération : %v", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var t models.Track
		err := rows.Scan(&t.Title, &t.Artist, &t.AnimeName, &t.AudioURL, &t.MalID)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func GetRandomTrack() (*models.Track, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	var t models.Track
	query := `SELECT id, title, artist, anime_name, audio_url FROM tracks ORDER BY RANDOM() LIMIT 1`

	err = conn.QueryRow(context.Background(), query).Scan(&t.ID, &t.Title, &t.Artist, &t.AnimeName, &t.AudioURL)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
