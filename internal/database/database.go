package database

import (
	"context"
	"fmt"
	"os"

	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() (*pgxpool.Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter : %v", err)
	}

	Pool = pool
	return pool, nil
}

// CreateUser insère un nouvel utilisateur dans la base de données.
func CreateUser(username, email, passwordHash string) error {
	query := `
		INSERT INTO users (username, email, password_hash, xp, level, created_at)
		VALUES ($1, $2, $3, 0, 1, NOW())
	`
	_, err := Pool.Exec(context.Background(), query, username, email, passwordHash)
	return err
}

// GetUserByID récupère un utilisateur par son ID.
func GetUserByID(userID int) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, xp, level, anilist_username, anilist_user_id, anilist_token, created_at
		FROM users
		WHERE id = $1
	`
	err := Pool.QueryRow(context.Background(), query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Xp, &user.Level, &user.AnilistUsername, &user.AnilistUserID, &user.AnilistToken, &user.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserMAL enregistre le token et le profil MyAnimeList d'un utilisateur.
func UpdateUserMAL(userID, malUserID int, malUsername, token string) error {
	query := `
		UPDATE users
		SET mal_user_id = $1, mal_username = $2, mal_token = $3
		WHERE id = $4
	`
	_, err := Pool.Exec(context.Background(), query, malUserID, malUsername, token, userID)
	return err
}

// UpdateUserAnilist enregistre le token et le profil AniList d'un utilisateur.
func UpdateUserAnilist(userID, anilistUserID int, anilistUsername, token string) error {
	query := `
		UPDATE users
		SET anilist_user_id = $1, anilist_username = $2, anilist_token = $3
		WHERE id = $4
	`
	_, err := Pool.Exec(context.Background(), query, anilistUserID, anilistUsername, token, userID)
	return err
}

// GetUserByUsernameOrEmail récupère un utilisateur pour vérifier ses identifiants au login
func GetUserByUsernameOrEmail(identifier string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, xp, level, created_at
		FROM users
		WHERE username = $1 OR email = $1
	`

	err := Pool.QueryRow(context.Background(), query, identifier).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Xp, &user.Level, &user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
