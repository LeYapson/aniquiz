package game

import (
	"testing"
	"github.com/LeYapson/aniquiz/internal/models"
)

func TestVerifyAnswer(t *testing.T) {
	// Données de test (un faux morceau de musique)
	mockTrack := &models.Track{
		AnimeName: "Naruto Shippuden",
	}

	// Cas de test (Table-driven tests - la norme en Go)
	tests := []struct {
		name           string
		userAnswer     string
		expectedPoints int
		expectedCorrect bool
	}{
		{"Réponse exacte", "Naruto Shippuden", 10, true},
		{"Minuscules", "naruto shippuden", 10, true},
		{"Réponse partielle (proche)", "Naruto", 5, false}, // Si ta logique donne des points partiels
		{"Mauvaise réponse", "One Piece", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyAnswer(tt.userAnswer, mockTrack)

			if result.Points != tt.expectedPoints {
				t.Errorf("VerifyAnswer(%s) points = %d; attendu %d", tt.userAnswer, result.Points, tt.expectedPoints)
			}
			if result.IsCorrect != tt.expectedCorrect {
				t.Errorf("VerifyAnswer(%s) correct = %v; attendu %v", tt.userAnswer, result.IsCorrect, tt.expectedCorrect)
			}
		})
	}
}