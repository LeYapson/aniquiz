package game

import (
	"testing"

	"github.com/LeYapson/aniquiz/internal/models"
)

func TestXPToLevel(t *testing.T) {
	tests := []struct {
		xp        int
		wantLevel int
	}{
		{0, 1},
		{-10, 1},
		{99, 1},
		{100, 2},
		{399, 2},
		{400, 3},
		{900, 4},
		{1600, 5},
	}
	for _, tt := range tests {
		got := XPToLevel(tt.xp)
		if got != tt.wantLevel {
			t.Errorf("XPToLevel(%d) = %d, want %d", tt.xp, got, tt.wantLevel)
		}
	}
}

func TestXPForScore(t *testing.T) {
	tests := []struct {
		score  int
		wantXP int
	}{
		{0, 5}, // participation minimum
		{1, 10},
		{10, 100},
		{50, 500},
	}
	for _, tt := range tests {
		got := XPForScore(tt.score)
		if got != tt.wantXP {
			t.Errorf("XPForScore(%d) = %d, want %d", tt.score, got, tt.wantXP)
		}
	}
}

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
