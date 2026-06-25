package handlers

import (
	"net/http"

	"github.com/LeYapson/aniquiz/internal/sourcing"
	"github.com/gin-gonic/gin"
)

// BackfillTitlesHandler — POST /api/admin/backfill-titles
// Recharge les titres alternatifs (anglais + synonymes) des animes déjà en base.
func BackfillTitlesHandler(c *gin.Context) {
	if !sourcing.StartTitleBackfill() {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "un backfill est déjà en cours",
			"progress": sourcing.TitleBackfillStatus(),
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Backfill des titres démarré en arrière-plan.",
		"progress": sourcing.TitleBackfillStatus(),
	})
}

// BackfillTitlesStatusHandler — GET /api/admin/backfill-titles/status
func BackfillTitlesStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, sourcing.TitleBackfillStatus())
}
