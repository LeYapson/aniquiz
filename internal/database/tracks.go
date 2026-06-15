package database

import (
	"context"
	"fmt"

	"github.com/LeYapson/aniquiz/internal/models"
)

func SaveTrack(track models.Track) error {
	query := `INSERT INTO tracks (title, anime_name, artist, audio_url, difficulty, mal_id)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  ON CONFLICT DO NOTHING`

	_, err := Pool.Exec(context.Background(), query, track.Title, track.AnimeName, track.Artist, track.AudioURL, track.Difficulty, track.MalID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'insertion : %v", err)
	}
	return nil
}

func GetAllTracks() ([]models.Track, error) {
	rows, err := Pool.Query(context.Background(),
		"SELECT title, artist, anime_name, audio_url, mal_id FROM tracks")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération : %v", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var t models.Track
		if err := rows.Scan(&t.Title, &t.Artist, &t.AnimeName, &t.AudioURL, &t.MalID); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func GetRandomTrack() (*models.Track, error) {
	var t models.Track
	query := `SELECT id, title, artist, anime_name, audio_url FROM tracks ORDER BY RANDOM() LIMIT 1`

	err := Pool.QueryRow(context.Background(), query).Scan(&t.ID, &t.Title, &t.Artist, &t.AnimeName, &t.AudioURL)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetTrackByID(id int) (*models.Track, error) {
	var t models.Track
	err := Pool.QueryRow(context.Background(),
		"SELECT id, title, artist, anime_name FROM tracks WHERE id = $1", id).Scan(
		&t.ID, &t.Title, &t.Artist, &t.AnimeName,
	)
	return &t, err
}
