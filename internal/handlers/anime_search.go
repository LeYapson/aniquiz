package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type AnimeSearchResult struct {
	MalID int    `json:"mal_id"`
	Title string `json:"title"`
	Year  int    `json:"year"`
	Type  string `json:"type"`
	Image string `json:"image"`
}

// AnimeSearchHandler proxifie la recherche Jikan pour éviter les problèmes CORS et centraliser le rate limit.
// GET /api/anime/search?q=naruto
func AnimeSearchHandler(c *gin.Context) {
	q := c.Query("q")
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête trop courte (minimum 2 caractères)"})
		return
	}

	jikanURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s&limit=8&sfw=true", url.QueryEscape(q))
	resp, err := http.Get(jikanURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "impossible de contacter Jikan"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "API surchargée, réessaie dans quelques secondes"})
		return
	}

	var raw struct {
		Data []struct {
			MalID  int    `json:"mal_id"`
			Title  string `json:"title"`
			Year   int    `json:"year"`
			Type   string `json:"type"`
			Images struct {
				JPG struct {
					SmallImageURL string `json:"small_image_url"`
				} `json:"jpg"`
			} `json:"images"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "réponse Jikan invalide"})
		return
	}

	results := make([]AnimeSearchResult, 0, len(raw.Data))
	for _, d := range raw.Data {
		results = append(results, AnimeSearchResult{
			MalID: d.MalID,
			Title: d.Title,
			Year:  d.Year,
			Type:  d.Type,
			Image: d.Images.JPG.SmallImageURL,
		})
	}

	c.JSON(http.StatusOK, results)
}
