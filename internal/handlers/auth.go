package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

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
				"id":         user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"xp":         user.Xp,
				"level":      user.Level,
				"created_at": user.CreatedAt,
			},
		})
	}
}
