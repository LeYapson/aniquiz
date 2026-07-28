package database

import (
	"context"
	"fmt"
	"os"
	"time"

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
		// Tables de base — créées en premier pour qu'une base vierge (nouveau
		// déploiement) se bootstrape seule. Idempotent : no-op si déjà présentes.
		`CREATE TABLE IF NOT EXISTS users (
			id               SERIAL PRIMARY KEY,
			username         TEXT UNIQUE NOT NULL,
			email            TEXT UNIQUE NOT NULL,
			password_hash    TEXT NOT NULL,
			xp               INT DEFAULT 0,
			level            INT DEFAULT 1,
			anilist_username TEXT DEFAULT '',
			anilist_user_id  INT DEFAULT 0,
			anilist_token    TEXT DEFAULT '',
			mal_username     TEXT DEFAULT '',
			mal_user_id      INT DEFAULT 0,
			mal_token        TEXT DEFAULT '',
			created_at       TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tracks (
			id          SERIAL PRIMARY KEY,
			title       TEXT NOT NULL,
			anime_name  TEXT NOT NULL,
			artist      TEXT DEFAULT '',
			audio_url   TEXT NOT NULL,
			difficulty  INT DEFAULT 1,
			mal_id      INT DEFAULT 0,
			track_type  TEXT DEFAULT '',
			anime_year  INT DEFAULT 0,
			CONSTRAINT tracks_unique_track UNIQUE (mal_id, title, track_type)
		)`,
		`CREATE TABLE IF NOT EXISTS game_results (
			id         SERIAL PRIMARY KEY,
			user_id    INT REFERENCES users(id),
			score      INT NOT NULL,
			xp_gained  INT NOT NULL,
			played_at  TIMESTAMPTZ DEFAULT NOW()
		)`,
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
		// Cosmétique : cadre d'avatar sélectionné (débloqué par niveau).
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_frame TEXT NOT NULL DEFAULT ''`,
		// Photo de profil : stockée en data URL base64 (pas de volume d'upload
		// à gérer, persiste avec la base). Vide = avatar généré à partir de l'initiale.
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT ''`,
		// Titres alternatifs d'un anime (anglais + synonymes) pour accepter
		// "How NOT to Summon a Demon King" en plus du titre japonais.
		`ALTER TABLE tracks ADD COLUMN IF NOT EXISTS anime_titles TEXT[] NOT NULL DEFAULT '{}'`,
		// Liaison du compte Discord (id + pseudo), pour le bot Discord.
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS discord_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS discord_username TEXT NOT NULL DEFAULT ''`,
		// Quiz du jour : résultats quotidiens (une tentative par joueur par jour).
		`CREATE TABLE IF NOT EXISTS daily_results (
			id         SERIAL PRIMARY KEY,
			user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			date       DATE NOT NULL,
			found      BOOLEAN NOT NULL DEFAULT false,
			time_ms    INT NOT NULL DEFAULT 0,
			played_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT daily_results_unique UNIQUE (user_id, date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_results_date ON daily_results(date)`,
		// Tokens de réinitialisation de mot de passe (un seul actif par utilisateur).
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id         SERIAL PRIMARY KEY,
			user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token      TEXT UNIQUE NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_prt_user_id ON password_reset_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_prt_token   ON password_reset_tokens(token)`,
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
		       COALESCE(avatar_frame, ''), COALESCE(avatar_url, ''),
		       COALESCE(discord_id, ''), COALESCE(discord_username, ''),
		       created_at
		FROM users
		WHERE id = $1
	`
	err := Pool.QueryRow(context.Background(), query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Xp, &user.Level,
		&user.AnilistUsername, &user.AnilistUserID, &user.AnilistToken,
		&user.MalUsername, &user.MalUserID, &user.MalToken,
		&user.AvatarFrame, &user.AvatarURL, &user.DiscordID, &user.DiscordUsername, &user.CreatedAt,
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

// GetUserByDiscordID résout un compte AniQuiz à partir de l'id Discord lié.
// Retourne (nil, nil) si aucun compte n'est lié à cet id.
func GetUserByDiscordID(discordID string) (*models.User, error) {
	var u models.User
	err := Pool.QueryRow(context.Background(), `
		SELECT id, username, COALESCE(email, ''), xp, level
		FROM users
		WHERE discord_id = $1`, discordID).Scan(&u.ID, &u.Username, &u.Email, &u.Xp, &u.Level)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUserDiscord enregistre (ou met à jour) le compte Discord lié.
func UpdateUserDiscord(userID int, discordID, discordUsername string) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE users SET discord_id = $1, discord_username = $2 WHERE id = $3`,
		discordID, discordUsername, userID)
	return err
}

// SetAvatarFrame met à jour le cadre d'avatar cosmétique sélectionné.
func SetAvatarFrame(userID int, frame string) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE users SET avatar_frame = $1 WHERE id = $2`, frame, userID)
	return err
}

// SetAvatarURL met à jour la photo de profil (data URL base64, ou "" pour retirer).
func SetAvatarURL(userID int, url string) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE users SET avatar_url = $1 WHERE id = $2`, url, userID)
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

// GetUserByUsernameAndEmail récupère un utilisateur uniquement si le pseudo ET l'email correspondent.
// Utilisé pour la récupération de mot de passe : les deux doivent correspondre.
func GetUserByUsernameAndEmail(username, email string) (*models.User, error) {
	var user models.User
	err := Pool.QueryRow(context.Background(), `
		SELECT id, username, email
		FROM users
		WHERE username = $1 AND LOWER(email) = LOWER($2)
	`, username, email).Scan(&user.ID, &user.Username, &user.Email)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreatePasswordResetToken supprime tout token existant pour cet utilisateur puis en crée un nouveau.
func CreatePasswordResetToken(userID int, token string, expiresAt time.Time) error {
	_, err := Pool.Exec(context.Background(), `
		DELETE FROM password_reset_tokens WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}
	_, err = Pool.Exec(context.Background(), `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, token, expiresAt)
	return err
}

// GetPasswordResetToken retourne le token s'il existe (expiré ou non — la vérification d'expiry est en Go).
func GetPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := Pool.QueryRow(context.Background(), `
		SELECT id, user_id, token, expires_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`, token).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DeletePasswordResetToken supprime un token (après usage ou expiry).
func DeletePasswordResetToken(token string) error {
	_, err := Pool.Exec(context.Background(), `
		DELETE FROM password_reset_tokens WHERE token = $1
	`, token)
	return err
}

// UpdateUserPassword met à jour le hash du mot de passe d'un utilisateur.
func UpdateUserPassword(userID int, passwordHash string) error {
	_, err := Pool.Exec(context.Background(), `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, passwordHash, userID)
	return err
}

// GetUserByUsernameOrEmail récupère un utilisateur pour vérifier ses identifiants au login
func GetUserByUsernameOrEmail(identifier string) (*models.User, error) {
	var user models.User
	// On charge aussi les comptes liés (AniList/MAL/Discord) : sinon le login
	// renvoie un user sans ces champs, et le client croit les connexions tierces
	// perdues à chaque reconnexion alors qu'elles sont bien en base.
	query := `
		SELECT id, username, email, password_hash, xp, level,
		       COALESCE(anilist_username, ''), COALESCE(anilist_user_id, 0), COALESCE(anilist_token, ''),
		       COALESCE(mal_username, ''),     COALESCE(mal_user_id, 0),     COALESCE(mal_token, ''),
		       COALESCE(avatar_frame, ''), COALESCE(avatar_url, ''),
		       COALESCE(discord_id, ''), COALESCE(discord_username, ''),
		       created_at
		FROM users
		WHERE username = $1 OR email = $1
	`

	err := Pool.QueryRow(context.Background(), query, identifier).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Xp, &user.Level,
		&user.AnilistUsername, &user.AnilistUserID, &user.AnilistToken,
		&user.MalUsername, &user.MalUserID, &user.MalToken,
		&user.AvatarFrame, &user.AvatarURL, &user.DiscordID, &user.DiscordUsername, &user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
