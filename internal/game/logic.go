package game

import (
	"strings"

	"github.com/LeYapson/aniquiz/internal/models"
)

type AnswerResult struct {
	IsCorrect bool   `json:"is_correct"`
	Points    int    `json:"points"`
	Message   string `json:"message"`
}

func VerifyAnswer(userInput string, track *models.Track) AnswerResult {
	user := strings.ToLower(strings.TrimSpace(userInput))
	target := strings.ToLower(track.AnimeName)

	if user == target {
		return AnswerResult{IsCorrect: true, Points: 10, Message: "Parfait !"}
	}

	userRunes := []rune(user)
	targetRunes := []rune(target)

	// Partial match: user input is a substring of the target or vice-versa.
	if len(userRunes) >= 3 && (strings.Contains(target, user) || strings.Contains(user, target)) {
		return AnswerResult{IsCorrect: false, Points: 5, Message: "C'est presque ça (nom de l'anime) !"}
	}

	// Fuzzy match: Levenshtein ≥80% similarity, operates on runes for correct Unicode handling.
	distance := levenshteinDistance(userRunes, targetRunes)
	longestLength := max(len(userRunes), len(targetRunes))
	if longestLength > 0 && float64(longestLength-distance)/float64(longestLength) >= 0.80 {
		return AnswerResult{IsCorrect: false, Points: 5, Message: "C'est presque ça (petite faute de frappe) !"}
	}

	return AnswerResult{IsCorrect: false, Points: 0, Message: ""}
}

// levenshteinDistance computes edit distance between two rune slices.
// Operating on runes (not bytes) ensures correct behaviour for multi-byte
// Unicode characters present in many anime titles.
func levenshteinDistance(s, t []rune) int {
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+cost))
		}
	}
	return d[len(s)][len(t)]
}
