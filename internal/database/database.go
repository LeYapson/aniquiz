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
		// Base partagée accessible via WireGuard (collaborateurs).
		// Override avec la variable d'env DATABASE_URL en local/prod si besoin.
		connStr = "postgres://postgres@192.168.27.74:5432/postgres?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter : %v", err)
	}

	Pool = pool
	return pool, nil
}

// Migrate applique les évolutions de schéma manquantes de façon idempotente.
// À appeler une fois au démarrage, après Connect().
func Migrate() error {
	migrations := []string{
		// Index pour les filtres de jeu fréquents
		`CREATE INDEX IF NOT EXISTS idx_tracks_mal_id    ON tracks(mal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_track_type ON tracks(track_type)`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_anime_year ON tracks(anime_year)`,
		// Table speed run
		`CREATE TABLE IF NOT EXISTS speedrun_results (
			id         SERIAL PRIMARY KEY,
			user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			score      INT NOT NULL DEFAULT 0,
			played_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_speedrun_user   ON speedrun_results(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_speedrun_score  ON speedrun_results(score DESC)`,
		// Système d'amis : une ligne par relation, statut 'pending' puis 'accepted'.
		`CREATE TABLE IF NOT EXISTS friendships (
			id           SERIAL PRIMARY KEY,
			requester_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			addressee_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status       TEXT NOT NULL DEFAULT 'pending',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT friendships_no_self CHECK (requester_id <> addressee_id),
			CONSTRAINT friendships_unique_pair UNIQUE (requester_id, addressee_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships(addressee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_friendships_requester ON friendships(requester_id)`,
		// Invitations à rejoindre un salon (éphémères, consommées à l'acceptation).
		`CREATE TABLE IF NOT EXISTS room_invites (
			id           SERIAL PRIMARY KEY,
			from_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			to_user_id   INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			room_id      TEXT NOT NULL,
			password     TEXT NOT NULL DEFAULT '',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT room_invites_unique UNIQUE (from_user_id, to_user_id, room_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_room_invites_to ON room_invites(to_user_id)`,
	}
	for _, q := range migrations {
		if _, err := Pool.Exec(context.Background(), q); err != nil {
			return fmt.Errorf("migration échouée (%q) : %w", q, err)
		}
	}
	return nil
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
//
// Les colonnes OAuth (anilist_*, mal_*) sont NULL tant que l'utilisateur n'a
// pas lié de compte : CreateUser ne les renseigne pas. Sans COALESCE, scanner
// un NULL dans un string/int Go échoue, ce qui faisait renvoyer 404 à
// /api/profile pour tout compte non lié (et masquait tout le profil, panneau
// d'amis compris).
func GetUserByID(userID int) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, xp, level,
		       COALESCE(anilist_username, ''), COALESCE(anilist_user_id, 0), COALESCE(anilist_token, ''),
		       COALESCE(mal_username, ''),     COALESCE(mal_user_id, 0),     COALESCE(mal_token, ''),
		       created_at
		FROM users
		WHERE id = $1
	`
	err := Pool.QueryRow(context.Background(), query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Xp, &user.Level,
		&user.AnilistUsername, &user.AnilistUserID, &user.AnilistToken,
		&user.MalUsername, &user.MalUserID, &user.MalToken, &user.CreatedAt,
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

// SaveGameResult enregistre le résultat d'un joueur pour une partie terminée.
func SaveGameResult(userID, score, xpGained int) error {
	query := `
		INSERT INTO game_results (user_id, score, xp_gained, played_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := Pool.Exec(context.Background(), query, userID, score, xpGained)
	return err
}

// GetUserHistory retourne les 20 dernières parties d'un utilisateur.
func GetUserHistory(userID int) ([]models.GameResult, error) {
	query := `
		SELECT id, user_id, score, xp_gained, played_at
		FROM game_results
		WHERE user_id = $1
		ORDER BY played_at DESC
		LIMIT 20
	`
	rows, err := Pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.GameResult
	for rows.Next() {
		var r models.GameResult
		if err := rows.Scan(&r.ID, &r.UserID, &r.Score, &r.XPGained, &r.PlayedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// AddUserXP ajoute de l'XP à un utilisateur et recalcule son niveau.
// Formule niveau : floor(sqrt(xp / 100)) + 1 (progression exponentielle).
func AddUserXP(userID, xpGained int) (newXP, newLevel int, err error) {
	query := `
		UPDATE users
		SET xp    = xp + $1,
		    level = FLOOR(SQRT((xp + $1)::float / 100))::int + 1
		WHERE id = $2
		RETURNING xp, level
	`
	err = Pool.QueryRow(context.Background(), query, xpGained, userID).Scan(&newXP, &newLevel)
	return
}

// GetLeaderboard retourne les N meilleurs joueurs triés par XP décroissant.
func GetLeaderboard(limit int) ([]models.LeaderboardEntry, error) {
	query := `
		SELECT
			ROW_NUMBER() OVER (ORDER BY u.xp DESC) AS rank,
			u.id, u.username, u.level, u.xp,
			COUNT(g.id)         AS total_games,
			COALESCE(MAX(g.score), 0) AS best_score
		FROM users u
		LEFT JOIN game_results g ON g.user_id = u.id
		GROUP BY u.id, u.username, u.level, u.xp
		ORDER BY u.xp DESC
		LIMIT $1
	`
	rows, err := Pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.Level, &e.XP, &e.TotalGames, &e.BestScore); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// SaveSpeedrunResult enregistre le score d'une partie speed run.
func SaveSpeedrunResult(userID, score int) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO speedrun_results (user_id, score, played_at) VALUES ($1, $2, NOW())`,
		userID, score,
	)
	return err
}

// GetSpeedrunLeaderboard retourne le meilleur score par joueur, trié par score décroissant.
func GetSpeedrunLeaderboard(limit int) ([]models.SpeedrunLeaderboardEntry, error) {
	query := `
		SELECT
			ROW_NUMBER() OVER (ORDER BY best.score DESC) AS rank,
			u.id, u.username, best.score, best.played_at
		FROM (
			SELECT DISTINCT ON (user_id)
				user_id, score, played_at
			FROM speedrun_results
			ORDER BY user_id, score DESC
		) best
		JOIN users u ON u.id = best.user_id
		ORDER BY best.score DESC
		LIMIT $1
	`
	rows, err := Pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.SpeedrunLeaderboardEntry
	for rows.Next() {
		var e models.SpeedrunLeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.BestScore, &e.PlayedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
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
