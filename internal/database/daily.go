package database

import (
	"context"
	"time"

	"github.com/LeYapson/aniquiz/internal/models"
)

// Époque de référence pour le numéro de jour (déterministe).
var dailyEpoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func dailyDayNum(date time.Time) int {
	return int(date.UTC().Truncate(24*time.Hour).Sub(dailyEpoch).Hours() / 24)
}

// GetDailyTrack retourne la piste du jour de façon déterministe :
// le numéro du jour modulo le nombre total de pistes garantit que
// tout le monde reçoit le même extrait, et que la rotation couvre
// l'ensemble de la bibliothèque.
func GetDailyTrack(date time.Time) (*models.Track, error) {
	dayNum := dailyDayNum(date)
	var t models.Track
	// Filtre les pistes mortes et shuffles via md5 pour éviter les clusters d'animes.
	err := Pool.QueryRow(context.Background(), `
		SELECT id, title, artist, anime_name, audio_url, difficulty, track_type
		FROM tracks
		WHERE audio_url != 'not_found' AND audio_url != ''
		ORDER BY md5(id::text || 'aniquiz-daily')
		LIMIT 1
		OFFSET ($1 % (SELECT COUNT(*) FROM tracks WHERE audio_url != 'not_found' AND audio_url != ''))`,
		dayNum,
	).Scan(&t.ID, &t.Title, &t.Artist, &t.AnimeName, &t.AudioURL, &t.Difficulty, &t.TrackType)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DailyStartFraction retourne la fraction de départ déterministe [0, 0.5[.
func DailyStartFraction(date time.Time) float64 {
	return float64(dailyDayNum(date)%50) / 100.0
}

type DailyResult struct {
	Found    bool      `json:"found"`
	TimeMs   int       `json:"time_ms"`
	PlayedAt time.Time `json:"played_at"`
}

func GetDailyResult(userID int, date time.Time) (*DailyResult, error) {
	var r DailyResult
	err := Pool.QueryRow(context.Background(),
		`SELECT found, time_ms, played_at FROM daily_results WHERE user_id=$1 AND date=$2`,
		userID, date.UTC().Format("2006-01-02"),
	).Scan(&r.Found, &r.TimeMs, &r.PlayedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func SaveDailyResult(userID int, date time.Time, found bool, timeMs int) error {
	_, err := Pool.Exec(context.Background(), `
		INSERT INTO daily_results (user_id, date, found, time_ms)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, date) DO NOTHING`,
		userID, date.UTC().Format("2006-01-02"), found, timeMs,
	)
	return err
}

type DailyLeaderboardEntry struct {
	Username string `json:"username"`
	TimeMs   int    `json:"time_ms"`
	Found    bool   `json:"found"`
}

func GetDailyLeaderboard(date time.Time, limit int) ([]DailyLeaderboardEntry, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT u.username, dr.time_ms, dr.found
		FROM daily_results dr
		JOIN users u ON u.id = dr.user_id
		WHERE dr.date = $1
		ORDER BY dr.found DESC, dr.time_ms ASC
		LIMIT $2`,
		date.UTC().Format("2006-01-02"), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []DailyLeaderboardEntry
	for rows.Next() {
		var e DailyLeaderboardEntry
		if err := rows.Scan(&e.Username, &e.TimeMs, &e.Found); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
