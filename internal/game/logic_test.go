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

func TestVerifyAnswerMode(t *testing.T) {
	track := &models.Track{
		AnimeName: "Cowboy Bebop",
		Title:     "Tank!",
		Artist:    "The Seatbelts",
	}

	tests := []struct {
		name       string
		input      string
		mode       string
		wantPoints int
	}{
		{"anime mode exact", "Cowboy Bebop", GuessModeAnime, 10},
		{"title mode exact", "Tank!", GuessModeTitle, 10},
		{"artist mode exact", "The Seatbelts", GuessModeArtist, 10},
		{"title mode but gave anime", "Cowboy Bebop", GuessModeTitle, 0},
		{"artist mode case insensitive", "the seatbelts", GuessModeArtist, 10},
		{"unknown mode falls back to anime", "Cowboy Bebop", "bogus", 10},
		{"empty mode falls back to anime", "Cowboy Bebop", "", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyAnswerMode(tt.input, track, tt.mode)
			if got.Points != tt.wantPoints {
				t.Errorf("VerifyAnswerMode(%q, %q) points = %d, want %d", tt.input, tt.mode, got.Points, tt.wantPoints)
			}
		})
	}
}

func TestVerifyAnswerAltTitles(t *testing.T) {
	track := &models.Track{
		AnimeName: "Isekai Maou to Shoukan Shoujo no Dorei Majutsu",
		AltTitles: []string{"How NOT to Summon a Demon King", "The Otherworldly Demon King"},
	}

	tests := []struct {
		name       string
		input      string
		wantPoints int
	}{
		{"titre japonais exact", "Isekai Maou to Shoukan Shoujo no Dorei Majutsu", 10},
		{"titre anglais exact", "How NOT to Summon a Demon King", 10},
		{"titre anglais insensible à la casse", "how not to summon a demon king", 10},
		{"synonyme exact", "The Otherworldly Demon King", 10},
		{"correspondance partielle sur l'anglais", "How NOT to Summon", 5},
		{"réponse fausse", "Naruto", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyAnswer(tt.input, track); got.Points != tt.wantPoints {
				t.Errorf("VerifyAnswer(%q) points = %d, want %d", tt.input, got.Points, tt.wantPoints)
			}
		})
	}
}

// Une cible vide (ex. artiste inconnu) ne doit jamais valider une réponse vide.
func TestMatchAnswerEmptyTarget(t *testing.T) {
	track := &models.Track{AnimeName: "X", Title: "Y", Artist: ""}
	if got := VerifyAnswerMode("", track, GuessModeArtist); got.Points != 0 || got.IsCorrect {
		t.Errorf("cible vide : got %+v, want 0 points / incorrect", got)
	}
}
