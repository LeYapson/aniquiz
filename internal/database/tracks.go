package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/jackc/pgx/v5"
)

// ErrNoTrack signale qu'aucune piste jouable ne correspond aux filtres demandés.
// Permet au moteur de jeu de réagir proprement (repli / abandon) plutôt que de
// rester bloqué sur une erreur brute pgx.ErrNoRows.
var ErrNoTrack = errors.New("aucune piste ne correspond aux filtres")

func SaveTrack(track models.Track) error {
	// ON CONFLICT sur (mal_id, title, track_type) nécessite la contrainte :
	// ALTER TABLE tracks ADD CONSTRAINT tracks_unique_track UNIQUE (mal_id, title, track_type);
	query := `INSERT INTO tracks (title, anime_name, artist, audio_url, difficulty, mal_id, track_type, anime_year, anime_titles)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  ON CONFLICT (mal_id, title, track_type)
			  DO UPDATE SET anime_titles = EXCLUDED.anime_titles`

	_, err := Pool.Exec(context.Background(), query,
		track.Title, track.AnimeName, track.Artist, track.AudioURL,
		track.Difficulty, track.MalID, track.TrackType, track.AnimeYear, track.AltTitles,
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

// GetDistinctAnimeNames returns a sorted list of unique anime titles for the
// answer autocomplete — including alternative titles (English name, synonyms)
// so a player can type "How NOT to Summon a Demon King" and get a suggestion.
func GetDistinctAnimeNames() ([]string, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT DISTINCT t FROM (
			SELECT anime_name AS t FROM tracks
			UNION
			SELECT unnest(anime_titles) AS t FROM tracks
		) titles
		WHERE t IS NOT NULL AND t != ''
		ORDER BY t`)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des animes : %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
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
		       difficulty, track_type, anime_year, mal_id, anime_titles
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
			&t.Difficulty, &t.TrackType, &t.AnimeYear, &t.MalID, &t.AltTitles)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoTrack
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CountPlayableForMalIDs retourne, pour une liste de MAL IDs (la liste perso de
// l'utilisateur), le nombre d'animés et de pistes réellement jouables présents
// dans la librairie. Sert à prévenir l'utilisateur avant qu'il ne lance une
// partie filtrée qui n'aurait (presque) aucune piste disponible.
func CountPlayableForMalIDs(ids []int) (animeCount int, trackCount int, err error) {
	if len(ids) == 0 {
		return 0, 0, nil
	}
	malIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		malIDs = append(malIDs, int32(id))
	}
	query := `
		SELECT COUNT(DISTINCT mal_id), COUNT(*)
		FROM tracks
		WHERE audio_url != 'not_found'
		  AND mal_id = ANY($1::int[])`
	err = Pool.QueryRow(context.Background(), query, malIDs).Scan(&animeCount, &trackCount)
	if err != nil {
		return 0, 0, fmt.Errorf("comptage des pistes jouables échoué : %w", err)
	}
	return animeCount, trackCount, nil
}

// AudioRef identifie l'URL audio d'une piste, pour la vérification de validité.
type AudioRef struct {
	ID       int
	AudioURL string
}

// GetTracksWithAudio retourne les pistes ayant une URL audio renseignée
// (autre que 'not_found'), afin d'en vérifier la disponibilité.
func GetTracksWithAudio() ([]AudioRef, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT id, audio_url FROM tracks WHERE audio_url != 'not_found' AND audio_url != ''`)
	if err != nil {
		return nil, fmt.Errorf("récupération des pistes audio échouée : %w", err)
	}
	defer rows.Close()

	var refs []AudioRef
	for rows.Next() {
		var r AudioRef
		if err := rows.Scan(&r.ID, &r.AudioURL); err != nil {
			return nil, err
		}
		refs = append(refs, r)
	}
	return refs, nil
}

// MarkAudioNotFound marque une piste comme sans audio jouable (lien mort).
// Elle est alors automatiquement exclue de GetRandomTrackFiltered.
func MarkAudioNotFound(id int) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE tracks SET audio_url = 'not_found' WHERE id = $1`, id)
	return err
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
