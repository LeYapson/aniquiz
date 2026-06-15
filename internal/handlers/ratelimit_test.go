package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestAuthRateLimiter_AllowsInitialRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newRateLimiterStore(rate.Every(6*time.Second), 10)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if !store.get(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			c.Abort()
			return
		}
		c.Next()
	})
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Les 10 premières requêtes (burst) doivent passer
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("requête %d bloquée (code %d), attendu 200", i+1, w.Code)
		}
	}
}

func TestAuthRateLimiter_BlocksAfterBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Burst de 3 pour accélérer le test
	store := newRateLimiterStore(rate.Every(time.Hour), 3)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if !store.get(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			c.Abort()
			return
		}
		c.Next()
	})
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Épuiser le burst
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		router.ServeHTTP(w, req)
	}

	// La 4ème doit être bloquée
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("attendu 429, obtenu %d", w.Code)
	}
}

func TestAuthRateLimiter_DifferentIPsAreIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newRateLimiterStore(rate.Every(time.Hour), 1)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if !store.get(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			c.Abort()
			return
		}
		c.Next()
	})
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// IP A épuise son quota
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
	}

	// IP B doit encore pouvoir passer
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("IP B bloquée à tort (code %d)", w.Code)
	}
}
