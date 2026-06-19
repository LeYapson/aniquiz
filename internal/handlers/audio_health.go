package handlers

import (
	"net/http"

	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
)

// AudioHealthcheckHandler démarre la vérification en masse des liens audio.
// Les liens morts (404/410) sont exclus du jeu ; le job tourne en arrière-plan.
//
// POST /api/admin/audio/healthcheck
func AudioHealthcheckHandler(c *gin.Context) {
	if !sourcing.StartAudioHealthcheck() {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "une vérification est déjà en cours",
			"progress": sourcing.AudioCheckStatus(),
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Vérification des liens audio démarrée en arrière-plan.",
		"progress": sourcing.AudioCheckStatus(),
	})
}

// AudioHealthStatusHandler retourne l'avancement de la vérification audio.
//
// GET /api/admin/audio/healthcheck/status
func AudioHealthStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, sourcing.AudioCheckStatus())
}
