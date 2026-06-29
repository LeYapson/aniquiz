package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// botTestRouter monte une route protégée par BotAuthMiddleware pour les tests.
func botTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/bot")
	g.Use(BotAuthMiddleware())
	g.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func TestBotAuthMiddleware(t *testing.T) {
	cases := []struct {
		name     string
		envKey   string
		header   string
		wantHTTP int
	}{
		{"clé absente de l'env", "", "whatever", http.StatusServiceUnavailable},
		{"bonne clé", "s3cret", "s3cret", http.StatusOK},
		{"mauvaise clé", "s3cret", "nope", http.StatusUnauthorized},
		{"header manquant", "s3cret", "", http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("BOT_API_KEY", c.envKey)
			req := httptest.NewRequest(http.MethodGet, "/api/bot/ping", nil)
			if c.header != "" {
				req.Header.Set("X-Bot-Key", c.header)
			}
			w := httptest.NewRecorder()
			botTestRouter().ServeHTTP(w, req)
			if w.Code != c.wantHTTP {
				t.Errorf("got %d, want %d", w.Code, c.wantHTTP)
			}
		})
	}
}
