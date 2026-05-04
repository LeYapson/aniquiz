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

	// 2. Match Partiel (Tolérance)
	// Si la réponse utilisateur est contenue dans le titre (ex: "Naruto" dans "Naruto Shippuden")
	// On vérifie que la réponse fait au moins 3 caractères pour éviter les points gratuits sur "The" ou "a"
	if len(user) >= 3 && (strings.Contains(target, user) || strings.Contains(user, target)) {
		return AnswerResult{IsCorrect: false, Points: 5, Message: "C'est presque ça (nom de l'anime) !"}
	}

	return AnswerResult{IsCorrect: false, Points: 0, Message: ""}
}
