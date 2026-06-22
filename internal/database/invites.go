package database

import (
	"context"

	"github.com/LeYapson/aniquiz/internal/models"
)

// AreFriends indique si deux utilisateurs ont une amitié confirmée.
func AreFriends(a, b int) (bool, error) {
	var n int
	err := Pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM friendships
		WHERE status = 'accepted'
		  AND ((requester_id = $1 AND addressee_id = $2)
		    OR (requester_id = $2 AND addressee_id = $1))`,
		a, b).Scan(&n)
	return n > 0, err
}

// CreateRoomInvite enregistre (ou rafraîchit) une invitation à rejoindre un
// salon. Une seule invitation vit par triplet (expéditeur, destinataire, salon).
func CreateRoomInvite(fromID, toID int, roomID, password string) error {
	_, err := Pool.Exec(context.Background(), `
		INSERT INTO room_invites (from_user_id, to_user_id, room_id, password)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (from_user_id, to_user_id, room_id)
		DO UPDATE SET password = EXCLUDED.password, created_at = NOW()`,
		fromID, toID, roomID, password)
	return err
}

// ListRoomInvites retourne les invitations récentes (< 30 min) reçues par un
// utilisateur, avec le pseudo de l'expéditeur. Purge au passage les plus vieilles.
func ListRoomInvites(userID int) ([]models.RoomInvite, error) {
	// Nettoyage opportuniste des invitations expirées.
	_, _ = Pool.Exec(context.Background(),
		`DELETE FROM room_invites WHERE created_at < NOW() - INTERVAL '30 minutes'`)

	rows, err := Pool.Query(context.Background(), `
		SELECT i.id, u.username, i.room_id, i.password, i.created_at
		FROM room_invites i
		JOIN users u ON u.id = i.from_user_id
		WHERE i.to_user_id = $1
		ORDER BY i.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []models.RoomInvite
	for rows.Next() {
		var inv models.RoomInvite
		if err := rows.Scan(&inv.ID, &inv.FromUsername, &inv.RoomID, &inv.Password, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, nil
}

// DeleteRoomInvite supprime une invitation reçue (le destinataire l'accepte ou la rejette).
func DeleteRoomInvite(userID, inviteID int) error {
	_, err := Pool.Exec(context.Background(),
		`DELETE FROM room_invites WHERE id = $1 AND to_user_id = $2`, inviteID, userID)
	return err
}
