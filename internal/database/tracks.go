package database

import (
	"context"
	"fmt"

	"github.com/LeYapson/aniquiz/internal/models"
)

func SaveTrack(track models.Track) error {
	// ON CONFLICT sur (mal_id, title, track_type) nécessite la contrainte :
	// ALTER TABLE tracks ADD CONSTRAINT tracks_unique_track UNIQUE (mal_id, title, track_type);
	query := `INSERT INTO tracks (title, anime_name, artist, audio_url, difficulty, mal_id, track_type, anime_year)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			  ON CONFLICT (mal_id, title, track_type) DO NOTHING`

	_, err := Pool.Exec(context.Background(), query,
		track.Title, track.AnimeName, track.Artist, track.AudioURL,
		track.Difficulty, track.MalID, track.TrackType, track.AnimeYear,
	)
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
	return GetRandomTrackFiltered(models.TrackFilters{})
}

// GetRandomTrackFiltered retourne une piste aléatoire selon les filtres de la salle.
// Les filtres à zéro/vide/nil sont ignorés (pas de restriction).
func GetRandomTrackFiltered(f models.TrackFilters) (*models.Track, error) {
	var t models.Track
	// Convertir []int en []int32 pour pgx
	var malIDs []int32
	for _, id := range f.MalIDs {
		malIDs = append(malIDs, int32(id))
	}

	query := `
		SELECT id, title, artist, anime_name, audio_url,
		       difficulty, track_type, anime_year, mal_id
		FROM tracks
		WHERE audio_url != 'not_found'
		  AND ($1 = '' OR track_type = $1)
		  AND ($2 = 0  OR anime_year >= $2)
		  AND ($3 = 0  OR anime_year <= $3)
		  AND ($4::int[] IS NULL OR mal_id = ANY($4::int[]))
		ORDER BY RANDOM()
		LIMIT 1
	`
	var malIDsParam interface{} = nil
	if len(malIDs) > 0 {
		malIDsParam = malIDs
	}
	err := Pool.QueryRow(context.Background(), query, f.TrackType, f.MinYear, f.MaxYear, malIDsParam).
		Scan(&t.ID, &t.Title, &t.Artist, &t.AnimeName, &t.AudioURL,
			&t.Difficulty, &t.TrackType, &t.AnimeYear, &t.MalID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// IsAnimeImported retourne true si des pistes existent déjà pour cet anime.
func IsAnimeImported(malID int) (bool, error) {
	var count int
	err := Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM tracks WHERE mal_id = $1", malID).Scan(&count)
	return count > 0, err
}

func GetTrackByID(id int) (*models.Track, error) {
	var t models.Track
	err := Pool.QueryRow(context.Background(),
		"SELECT id, title, artist, anime_name FROM tracks WHERE id = $1", id).Scan(
		&t.ID, &t.Title, &t.Artist, &t.AnimeName,
	)
	return &t, err
}
