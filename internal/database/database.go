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
	connStr := "postgres://postgres@localhost:5432/aniquiz_db"

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("Impossible de se connecter : %v", err)
	}

	return conn, nil
}

