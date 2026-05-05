package handlers

import "github.com/LeYapson/aniquiz/internal/database"
import "github.com/LeYapson/aniquiz/internal/models"

// PgStore implements Store using the real PostgreSQL database package.
type PgStore struct{}

func (s *PgStore) GetRandomTrack() (*models.Track, error) { return database.GetRandomTrack() }
func (s *PgStore) GetTrackByID(id int) (*models.Track, error) { return database.GetTrackByID(id) }
func (s *PgStore) GetAllTracks() ([]models.Track, error)      { return database.GetAllTracks() }
