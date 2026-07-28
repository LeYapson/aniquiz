package handlers

import (
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
)

// PgStore implements Store using the real PostgreSQL database package.
type PgStore struct{}

func (s *PgStore) GetRandomTrack() (*models.Track, error)     { return database.GetRandomTrack() }
func (s *PgStore) GetTrackByID(id int) (*models.Track, error) { return database.GetTrackByID(id) }
func (s *PgStore) GetAllTracks() ([]models.Track, error)      { return database.GetAllTracks() }
func (s *PgStore) GetDistinctAnimeNames() ([]string, error)   { return database.GetDistinctAnimeNames() }
func (s *PgStore) CreateUser(username, email, passwordHash string) error {
	return database.CreateUser(username, email, passwordHash)
}

func (s *PgStore) GetUserByUsernameOrEmail(identifier string) (*models.User, error) {
	return database.GetUserByUsernameOrEmail(identifier)
}

func (s *PgStore) SaveSpeedrunResult(userID, score int) error {
	return database.SaveSpeedrunResult(userID, score)
}

func (s *PgStore) GetSpeedrunLeaderboard(limit int) ([]models.SpeedrunLeaderboardEntry, error) {
	return database.GetSpeedrunLeaderboard(limit)
}

func (s *PgStore) GetUserByUsernameAndEmail(username, email string) (*models.User, error) {
	return database.GetUserByUsernameAndEmail(username, email)
}

func (s *PgStore) CreatePasswordResetToken(userID int, token string, expiresAt time.Time) error {
	return database.CreatePasswordResetToken(userID, token, expiresAt)
}

func (s *PgStore) GetPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	return database.GetPasswordResetToken(token)
}

func (s *PgStore) DeletePasswordResetToken(token string) error {
	return database.DeletePasswordResetToken(token)
}

func (s *PgStore) UpdateUserPassword(userID int, passwordHash string) error {
	return database.UpdateUserPassword(userID, passwordHash)
}
