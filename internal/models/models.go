package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` //le "-" indique que ce champ ne sera pas inclus dans les réponses JSON
	AvatarUrl    string    `json:"avatar_url"`
	XP           int       `json:"xp"`
	CreatedAt    time.Time `json:"created_at"`
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
