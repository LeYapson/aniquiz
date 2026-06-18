package handlers

import "github.com/LeYapson/aniquiz/internal/database"
import "github.com/LeYapson/aniquiz/internal/models"

// PgStore implements Store using the real PostgreSQL database package.
type PgStore struct{}

func (s *PgStore) GetRandomTrack() (*models.Track, error)     { return database.GetRandomTrack() }
func (s *PgStore) GetTrackByID(id int) (*models.Track, error) { return database.GetTrackByID(id) }
func (s *PgStore) GetAllTracks() ([]models.Track, error)      { return database.GetAllTracks() }
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
