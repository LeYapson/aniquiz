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

	// 1. Match Parfait
	if user == target {
		return AnswerResult{IsCorrect: true, Points: 10, Message: "Parfait !"}
	}

	// 2. Match Partiel (Tolérance de ton code d'origine)
	if len(user) >= 3 && (strings.Contains(target, user) || strings.Contains(user, target)) {
		return AnswerResult{IsCorrect: false, Points: 5, Message: "C'est presque ça (nom de l'anime) !"}
	}

	// 3. Match Flou (Recherche Levenshtein contre les fautes de frappe)
	distance := levenshteinDistance(user, target)
	longestLength := len(target)
	if len(user) > longestLength {
		longestLength = len(user)
	}

	if longestLength > 0 {
		similarity := float64(longestLength-distance) / float64(longestLength)
		// Si l'utilisateur est proche à 80% ou plus du titre exact
		if similarity >= 0.80 {
			return AnswerResult{IsCorrect: false, Points: 5, Message: "C'est presque ça (petite faute de frappe) !"}
		}
	}

	return AnswerResult{IsCorrect: false, Points: 0, Message: ""}
}

// Fonction interne pour calculer la distance de Levenshtein
func levenshteinDistance(s, t string) int {
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
			cost := 0
			if s[i-1] != t[j-1] {
				cost = 1
			}
			d[i][j] = min(
				d[i-1][j]+1,      // Suppression
				min(d[i][j-1]+1,  // Insertion
					d[i-1][j-1]+cost), // Substitution
			)
		}
	}
	return d[len(s)][len(t)]
}

// Fonction utilitaire pour trouver le minimum
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}