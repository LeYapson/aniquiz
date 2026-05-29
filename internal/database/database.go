package database

import (
	"context"
	"fmt"

	//"os"

	"github.com/jackc/pgx/v5"
)

func Connect() (*pgx.Conn, error) {
	// Dans un premier temps, on peut mettre la chaine en dur pour tester
	// Format : postgres://utilisateur:motdepasse@localhost:5432/nom_bdd
	connStr := "postgres://postgres@localhost:5432/postgres?sslmode=disable"

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("Impossible de se connecter : %v", err)
	}

	return conn, nil
}

// CreateUser insère un nouvel utilisateur dans la base de données.
func CreateUser(username, email, passwordHash string) error {
	query := `
		INSERT INTO users (username, email, password_hash, xp, level, created_at)
		VALUES ($1, $2, $3, 0, 1, NOW())
	`
	_, err := conn.Exec(context.Background(), query, username, email, passwordHash)
	return err
}

// UserExists vérifie si un pseudo ou un email est déjà utilisé.
func UserExists(username, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)`
	err := conn.Slice(context.Background(), query, username, email).Scan(&exists)
	// Remplace par la méthode de scan standard de ton driver si tu n'utilises pas pgx/v5 directement :
	// err := conn.QueryRow(context.Background(), query, username, email).Scan(&exists)
	return exists, err
}