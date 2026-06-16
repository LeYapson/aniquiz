package handlers

import (
	"net/http"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// importSemaphore limite le nombre d'imports Jikan simultanés sur tout le serveur.
var importSemaphore = make(chan struct{}, 2)

// importLimiter limite chaque utilisateur à 5 imports par minute.
var importLimiter = newRateLimiterStore(rate.Every(12*time.Second), 5)

type importResult struct {
	MalID   int    `json:"mal_id"`
	Title   string `json:"title,omitempty"`
	Error   string `json:"error,omitempty"`
	OPs     int    `json:"openings"`
	EDs     int    `json:"endings"`
	Skipped bool   `json:"skipped,omitempty"`
}

// BatchImportHandler importe des animes depuis leurs IDs MAL.
// POST /api/admin/import  {"ids": [20, 1, 21]}
func BatchImportHandler(c *gin.Context) {
	// Rate limit par utilisateur (JWT → IP comme clé)
	if !importLimiter.get(c.ClientIP()).Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Trop d'imports. Réessaie dans une minute."})
		return
	}

	var body struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids requis (tableau d'entiers MAL)"})
		return
	}
	if len(body.IDs) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 10 IDs par requête"})
		return
	}

	// Sémaphore global : max 2 imports simultanés sur tout le serveur
	importSemaphore <- struct{}{}
	defer func() { <-importSemaphore }()

	results := make([]importResult, len(body.IDs))

	for i, id := range body.IDs {
		if i > 0 {
			time.Sleep(1 * time.Second) // respecte le rate limit Jikan
		}

		// Skip si l'anime est déjà en base
		if already, err := database.IsAnimeImported(id); err == nil && already {
			results[i] = importResult{MalID: id, Skipped: true}
			continue
		}

		info, err := sourcing.ProcessAndSaveAnime(id)
		if err != nil {
			results[i] = importResult{MalID: id, Error: err.Error()}
			continue
		}
		results[i] = importResult{
			MalID: id,
			Title: info.Title,
			OPs:   len(info.Openings),
			EDs:   len(info.Endings),
		}
	}

	var imported, failed, skipped int
	for _, r := range results {
		switch {
		case r.Skipped:
			skipped++
		case r.Error != "":
			failed++
		default:
			imported++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": imported,
		"skipped":  skipped,
		"failed":   failed,
		"results":  results,
	})
}
