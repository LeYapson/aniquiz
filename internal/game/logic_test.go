package game

import (
	"testing"

	"github.com/LeYapson/aniquiz/internal/models"
)

func TestVerifyAnswer(t *testing.T) {
	track := &models.Track{AnimeName: "Naruto Shippuden"}

	tests := []struct {
		name        string
		input       string
		wantPoints  int
		wantCorrect bool
	}{
		{"exact match", "Naruto Shippuden", 10, true},
		{"case insensitive", "naruto shippuden", 10, true},
		{"leading/trailing spaces", "  Naruto Shippuden  ", 10, true},
		{"partial match", "Naruto", 5, false},
		{"target contained in input", "c'est Naruto Shippuden je crois", 5, false},
		{"wrong answer", "One Piece", 0, false},
		{"empty input", "", 0, false},
		{"too short input", "Na", 0, false},
		{"spaces only", "   ", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyAnswer(tt.input, track)
			if got.Points != tt.wantPoints {
				t.Errorf("points: got %d, want %d", got.Points, tt.wantPoints)
			}
			if got.IsCorrect != tt.wantCorrect {
				t.Errorf("correct: got %v, want %v", got.IsCorrect, tt.wantCorrect)
			}
		})
	}
}
