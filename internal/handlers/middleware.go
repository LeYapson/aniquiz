package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// IsAdmin indique si un pseudo figure dans l'allowlist ADMIN_USERNAMES
// (liste séparée par des virgules, insensible à la casse). Pas de rôle en base :
// les admins sont configurés via l'environnement de déploiement.
func IsAdmin(username string) bool {
	raw := os.Getenv("ADMIN_USERNAMES")
	if raw == "" {
		return false
	}
	for _, name := range strings.Split(raw, ",") {
		if strings.EqualFold(strings.TrimSpace(name), username) {
			return true
		}
	}
	return false
}

// AdminMiddleware restreint l'accès aux administrateurs. À chaîner APRÈS
// AuthMiddleware (qui renseigne "username" dans le contexte).
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := c.Get("username")
		if !ok || !IsAdmin(username.(string)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "accès réservé aux administrateurs"})
			return
		}
		c.Next()
	}
}

// AuthMiddleware valide le Bearer token JWT sur les routes protégées.
// Les claims extraites sont disponibles via c.Get("userID") et c.Get("username").
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token d'authentification requis"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
