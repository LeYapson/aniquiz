package handlers

import (
	"net/http"
	"strconv"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
	"github.com/gin-gonic/gin"
)

// LibraryStatsHandler — GET /api/admin/stats
func LibraryStatsHandler(c *gin.Context) {
	stats, err := database.GetLibraryStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// BrowseTracksHandler — GET /api/admin/tracks?q=&limit=&offset=
// Navigateur de pistes en lecture seule (recherche par nom d'anime / titre).
func BrowseTracksHandler(c *gin.Context) {
	q := c.Query("q")

	limit := atoiDefault(c.Query("limit"), 25)
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	offset := atoiDefault(c.Query("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	rows, err := database.BrowseTracks(q, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if rows == nil {
		rows = []models.AdminTrackRow{}
	}
	c.JSON(http.StatusOK, rows)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
