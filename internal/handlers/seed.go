package handlers

import (
	"net/http"

	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
)

// SeedHandler démarre l'import en masse des animes les plus populaires de MAL.
// Le job tourne en arrière-plan ; suivre l'avancement via /api/admin/seed/status.
//
// POST /api/admin/seed   {"pages": 4}   (défaut : 4 pages ≈ 100 animes)
func SeedHandler(c *gin.Context) {
	var body struct {
		Pages int `json:"pages"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Pages == 0 {
		body.Pages = 4
	}

	if !sourcing.StartSeed(body.Pages) {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "un import en masse est déjà en cours",
			"progress": sourcing.SeedStatus(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Import en masse démarré en arrière-plan.",
		"progress": sourcing.SeedStatus(),
	})
}

// SeedStatusHandler retourne l'avancement du job d'import en masse.
//
// GET /api/admin/seed/status
func SeedStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, sourcing.SeedStatus())
}
