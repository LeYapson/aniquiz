package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // Peut être le pseudo ou l'email
	Password   string `json:"password" binding:"required"`
}

// RegisterHandler gère la création de compte
func RegisterHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides ou mot de passe trop court (min 8 car.)"})
			return
		}

		// 1. Vérifier si l'utilisateur existe déjà
		existingUser, err := store.GetUserByUsernameOrEmail(req.Username)
		if err == nil && existingUser != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Ce pseudo est déjà utilisé"})
			return
		}
		existingEmail, err := store.GetUserByUsernameOrEmail(req.Email)
		if err == nil && existingEmail != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Cet email est déjà utilisé"})
			return
		}

		// 2. Hachage du mot de passe avec Bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du traitement du mot de passe"})
			return
		}

		// 3. Sauvegarde en Base de données
		err = store.CreateUser(req.Username, req.Email, string(hashedPassword))
		if err != nil {
			log.Printf("CreateUser error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer le compte"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Inscription réussie ! Vous pouvez maintenant vous connecter."})
	}
}

type ForgotPasswordRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ForgotPasswordHandler — POST /api/auth/forgot-password
// Vérifie que le pseudo ET l'email correspondent, puis envoie un lien de reset par mail.
// Répond toujours 200 pour éviter d'énumérer les comptes existants.
func ForgotPasswordHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pseudo et email valides requis"})
			return
		}

		successMsg := gin.H{"message": "Si un compte correspondant existe, un email de réinitialisation a été envoyé."}

		user, err := store.GetUserByUsernameAndEmail(req.Username, req.Email)
		if err != nil || user == nil {
			c.JSON(http.StatusOK, successMsg)
			return
		}

		token, err := generateResetToken()
		if err != nil {
			log.Printf("generateResetToken error: %v", err)
			c.JSON(http.StatusOK, successMsg)
			return
		}

		if err := store.CreatePasswordResetToken(user.ID, token, time.Now().Add(time.Hour)); err != nil {
			log.Printf("CreatePasswordResetToken error: %v", err)
			c.JSON(http.StatusOK, successMsg)
			return
		}

		appURL := os.Getenv("APP_URL")
		if appURL == "" {
			appURL = "https://aniquiz.fr"
		}
		resetURL := appURL + "/reset-password?token=" + token

		if err := SendPasswordResetEmail(user.Email, user.Username, resetURL); err != nil {
			log.Printf("SendPasswordResetEmail error: %v", err)
		}

		c.JSON(http.StatusOK, successMsg)
	}
}

// ResetPasswordHandler — POST /api/auth/reset-password
// Valide le token et met à jour le mot de passe.
func ResetPasswordHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token et nouveau mot de passe requis (min 8 caractères)"})
			return
		}

		resetToken, err := store.GetPasswordResetToken(req.Token)
		if err != nil || resetToken == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lien invalide ou expiré"})
			return
		}

		if time.Now().After(resetToken.ExpiresAt) {
			_ = store.DeletePasswordResetToken(req.Token)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lien expiré, veuillez faire une nouvelle demande"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du traitement du mot de passe"})
			return
		}

		if err := store.UpdateUserPassword(resetToken.UserID, string(hashedPassword)); err != nil {
			log.Printf("UpdateUserPassword error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre à jour le mot de passe"})
			return
		}

		_ = store.DeletePasswordResetToken(req.Token)

		c.JSON(http.StatusOK, gin.H{"message": "Mot de passe mis à jour avec succès !"})
	}
}

// LoginHandler gère la vérification des identifiants et la connexion
func LoginHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Identifiants requis"})
			return
		}

		// 1. Récupérer l'utilisateur en BDD via son pseudo ou son email
		user, err := store.GetUserByUsernameOrEmail(req.Identifier)
		if err != nil {
			// Par sécurité, on utilise un message générique pour ne pas indiquer si le pseudo existe ou non
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants incorrects"})
			return
		}

		// 2. Comparer le mot de passe reçu avec le hash stocké en BDD
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants incorrects"})
			return
		}

		// 3. Génération du token JWT
		token, err := GenerateToken(user.ID, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
			return
		}

		// 4. Connexion réussie : On renvoie les infos et le token au Front-end
		c.JSON(http.StatusOK, gin.H{
			"message": "Connexion réussie !",
			"token":   token,
			"user": gin.H{
				"id":               user.ID,
				"username":         user.Username,
				"email":            user.Email,
				"xp":               user.Xp,
				"level":            user.Level,
				"avatar_frame":     user.AvatarFrame,
				"avatar_url":       user.AvatarURL,
				"anilist_username": user.AnilistUsername,
				"mal_username":     user.MalUsername,
				"discord_username": user.DiscordUsername,
				"is_admin":         IsAdmin(user.Username),
				"created_at":       user.CreatedAt,
			},
		})
	}
}
