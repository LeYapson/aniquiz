package game

import (
	"fmt"
	"strings"

	"github.com/LeYapson/aniquiz/internal/models"
)

// Modes de réponse : ce que les joueurs doivent deviner.
const (
	GuessModeAnime  = "anime"  // nom de l'anime (défaut)
	GuessModeTitle  = "title"  // titre de la musique
	GuessModeArtist = "artist" // artiste / interprète
)

type AnswerResult struct {
	IsCorrect bool   `json:"is_correct"`
	Points    int    `json:"points"`
	Message   string `json:"message"`
}

// VerifyAnswer vérifie une réponse contre le nom de l'anime (mode par défaut).
// Conservé pour les appelants existants (speedrun, endpoint /quiz/answer).
func VerifyAnswer(userInput string, track *models.Track) AnswerResult {
	return VerifyAnswerMode(userInput, track, GuessModeAnime)
}

// VerifyAnswerMode vérifie une réponse selon le mode de jeu : nom de l'anime,
// titre de la musique ou artiste. Un mode inconnu/vide retombe sur l'anime.
func VerifyAnswerMode(userInput string, track *models.Track, mode string) AnswerResult {
	switch mode {
	case GuessModeTitle:
		return matchAnswer(userInput, track.Title, "titre")
	case GuessModeArtist:
		return matchAnswer(userInput, track.Artist, "artiste")
	default:
		// Anime : on accepte le titre principal OU un titre alternatif
		// (anglais, synonymes). On retient la meilleure correspondance.
		candidates := append([]string{track.AnimeName}, track.AltTitles...)
		best := AnswerResult{}
		for _, c := range candidates {
			r := matchAnswer(userInput, c, "nom de l'anime")
			if r.IsCorrect {
				return r
			}
			if r.Points > best.Points {
				best = r
			}
		}
		return best
	}
}

// matchAnswer applique la logique de correspondance (exacte / partielle / floue)
// d'une saisie contre une cible donnée.
func matchAnswer(userInput, targetRaw, label string) AnswerResult {
	user := strings.ToLower(strings.TrimSpace(userInput))
	target := strings.ToLower(strings.TrimSpace(targetRaw))

	if target == "" {
		return AnswerResult{IsCorrect: false, Points: 0, Message: ""}
	}

	if user == target {
		return AnswerResult{IsCorrect: true, Points: 10, Message: "Parfait !"}
	}

	userRunes := []rune(user)
	targetRunes := []rune(target)

	// Partial match: user input is a substring of the target or vice-versa.
	if len(userRunes) >= 3 && (strings.Contains(target, user) || strings.Contains(user, target)) {
		return AnswerResult{IsCorrect: false, Points: 5, Message: fmt.Sprintf("C'est presque ça (%s) !", label)}
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
