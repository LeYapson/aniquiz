package models

import "time"

type User struct {
	ID              int       `json:"id" db:"id"`
	Username        string    `json:"username" db:"username"`
	Email           string    `json:"email" db:"email"`
	PasswordHash    string    `json:"-" db:"password_hash"` // Le tiret "-" cache le hash dans les réponses JSON
	AvatarURL       string    `json:"avatar_url" db:"avatar_url"`
	Xp              int       `json:"xp" db:"xp"`
	Level           int       `json:"level" db:"level"`
	AnilistUsername string    `json:"anilist_username" db:"anilist_username"`
	AnilistUserID   int       `json:"anilist_user_id" db:"anilist_user_id"`
	AnilistToken    string    `json:"-" db:"anilist_token"`
	MalUsername     string    `json:"mal_username" db:"mal_username"`
	MalUserID       int       `json:"mal_user_id" db:"mal_user_id"`
	MalToken        string    `json:"-" db:"mal_token"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Track struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	AnimeName  string `json:"anime_name"`
	Artist     string `json:"artist"`
	AudioURL   string `json:"audio_url"`
	Difficulty int    `json:"difficulty"`
	MalID      int    `json:"mal_id"`
	TrackType  string `json:"track_type"` // "OP" ou "ED"
	AnimeYear  int    `json:"anime_year"`
}

type TrackFilters struct {
	TrackType string // "OP", "ED", ou "" pour tout
	MinYear   int    // 0 = pas de filtre
	MaxYear   int    // 0 = pas de filtre
}

type Score struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
	Score  int `json:"score"`
}

type GameResult struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Score    int       `json:"score"`
	XPGained int       `json:"xp_gained"`
	PlayedAt time.Time `json:"played_at"`
}

type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	Level      int    `json:"level"`
	XP         int    `json:"xp"`
	TotalGames int    `json:"total_games"`
	BestScore  int    `json:"best_score"`
}
