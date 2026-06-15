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
	MalUsername     string    `json:"mal_username" db:"mal_username"`
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
}

type Score struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
	Score  int `json:"score"`
}
