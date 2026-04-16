package game

import (
	"strings"

	"github.com/LeYapson/aniquiz/internal/models"
)

type CheckResult struct {
    IsCorrect bool   `json:"is_correct"`
    Points    int    `json:"points"`
    Message   string `json:"message"`
}

func VerifyAnswer(playerAnswer string, track *models.Track) CheckResult {
    player := strings.ToLower(strings.TrimSpace(playerAnswer))
    
    // 1. Check Anime Name (10 points)
    correctAnime := strings.ToLower(strings.TrimSpace(strings.Split(track.AnimeName, "(")[0]))
    if player == correctAnime {
        return CheckResult{IsCorrect: true, Points: 10, Message: "Bravo ! C'est le bon anime."}
    }

    // 2. Check Artist Name (5 points bonus)
    correctArtist := strings.ToLower(strings.TrimSpace(track.Artist))
    if player == correctArtist {
        return CheckResult{IsCorrect: false, Points: 5, Message: "C'est bien l'artiste ! Mais quel est l'anime ?"}
    }

    return CheckResult{IsCorrect: false, Points: 0, Message: "Dommage, ce n'est pas ça."}
}