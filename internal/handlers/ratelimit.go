package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
}

func newRateLimiterStore(r rate.Limit, burst int) *rateLimiterStore {
	s := &rateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		rate:     r,
		burst:    burst,
	}
	go s.cleanup()
	return s
}

func (s *rateLimiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.limiters[ip]; ok {
		entry.lastSeen = time.Now()
		return entry.limiter
	}
	l := rate.NewLimiter(s.rate, s.burst)
	s.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

// cleanup supprime les entrées inactives depuis plus de 5 minutes.
func (s *rateLimiterStore) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 5*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

// AuthRateLimiter limite à 10 requêtes par minute par IP sur les routes d'auth.
func AuthRateLimiter() gin.HandlerFunc {
	// 10 req/min = ~1 req toutes les 6s, burst de 10 pour absorber les pics légitimes
	store := newRateLimiterStore(rate.Every(6*time.Second), 10)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !store.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Trop de tentatives. Réessayez dans une minute.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
