package database

import (
	"context"
	"errors"
	"strings"

	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/jackc/pgx/v5"
)

// Erreurs sentinelles du système d'amis, pour des réponses HTTP propres.
var (
	ErrFriendSelf         = errors.New("impossible de s'ajouter soi-même")
	ErrFriendUserNotFound = errors.New("utilisateur introuvable")
	ErrAlreadyFriends     = errors.New("vous êtes déjà amis")
	ErrRequestExists      = errors.New("demande déjà envoyée")
	ErrRequestNotFound    = errors.New("demande introuvable")
)

// SendFriendRequest crée une demande d'ami du requester vers l'utilisateur
// nommé. Si une demande inverse existe déjà (l'autre nous a demandé), elle est
// acceptée automatiquement (amitié mutuelle).
func SendFriendRequest(requesterID int, addresseeUsername string) error {
	ctx := context.Background()

	var addresseeID int
	err := Pool.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, addresseeUsername).Scan(&addresseeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrFriendUserNotFound
	}
	if err != nil {
		return err
	}
	if addresseeID == requesterID {
		return ErrFriendSelf
	}

	// Relation existante dans un sens ou l'autre ?
	var existingID, existingRequester int
	var status string
	err = Pool.QueryRow(ctx, `
		SELECT id, requester_id, status FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
		LIMIT 1`, requesterID, addresseeID).Scan(&existingID, &existingRequester, &status)
	if err == nil {
		if status == "accepted" {
			return ErrAlreadyFriends
		}
		if existingRequester == requesterID {
			return ErrRequestExists
		}
		// Demande inverse en attente → on l'accepte.
		_, err = Pool.Exec(ctx, `UPDATE friendships SET status = 'accepted' WHERE id = $1`, existingID)
		return err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = Pool.Exec(ctx,
		`INSERT INTO friendships (requester_id, addressee_id) VALUES ($1, $2)`,
		requesterID, addresseeID)
	return err
}

// RespondFriendRequest accepte ou refuse une demande reçue. Seul le destinataire
// (addressee) d'une demande encore 'pending' peut y répondre.
func RespondFriendRequest(userID, requestID int, accept bool) error {
	query := `DELETE FROM friendships
	          WHERE id = $1 AND addressee_id = $2 AND status = 'pending'`
	if accept {
		query = `UPDATE friendships SET status = 'accepted'
		         WHERE id = $1 AND addressee_id = $2 AND status = 'pending'`
	}

	tag, err := Pool.Exec(context.Background(), query, requestID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// RemoveFriend supprime une amitié confirmée entre deux utilisateurs (peu
// importe qui avait initié la demande).
func RemoveFriend(userID, friendID int) error {
	_, err := Pool.Exec(context.Background(), `
		DELETE FROM friendships
		WHERE status = 'accepted'
		  AND ((requester_id = $1 AND addressee_id = $2)
		    OR (requester_id = $2 AND addressee_id = $1))`,
		userID, friendID)
	return err
}

// ListFriends retourne les amis confirmés d'un utilisateur, triés par pseudo.
func ListFriends(userID int) ([]models.Friend, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT u.id, u.username, u.level
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE f.status = 'accepted'
		  AND (f.requester_id = $1 OR f.addressee_id = $1)
		ORDER BY u.username`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []models.Friend
	for rows.Next() {
		var f models.Friend
		if err := rows.Scan(&f.UserID, &f.Username, &f.Level); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}
	return friends, nil
}

// SearchUsers retourne jusqu'à `limit` utilisateurs dont le pseudo contient
// `query` (insensible à la casse), en excluant l'utilisateur courant. Sert à
// l'auto-complétion lors de l'ajout d'un ami.
func SearchUsers(query string, excludeUserID, limit int) ([]models.Friend, error) {
	// Échappe les jokers ILIKE pour qu'ils soient traités littéralement.
	esc := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
	pattern := "%" + esc + "%"

	rows, err := Pool.Query(context.Background(), `
		SELECT id, username, level
		FROM users
		WHERE id <> $1 AND username ILIKE $2 ESCAPE '\'
		ORDER BY username
		LIMIT $3`, excludeUserID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.Friend
	for rows.Next() {
		var u models.Friend
		if err := rows.Scan(&u.UserID, &u.Username, &u.Level); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ListPendingRequests retourne les demandes d'ami reçues et encore en attente.
func ListPendingRequests(userID int) ([]models.FriendRequest, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT f.id, u.id, u.username, u.level, f.created_at
		FROM friendships f
		JOIN users u ON u.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []models.FriendRequest
	for rows.Next() {
		var r models.FriendRequest
		if err := rows.Scan(&r.RequestID, &r.UserID, &r.Username, &r.Level, &r.CreatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, r)
	}
	return reqs, nil
}
