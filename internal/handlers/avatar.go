package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/gin-gonic/gin"
)

// maxAvatarBytes borne la taille de la photo de profil. On la stocke en data URL
// base64 dans la base et on la renvoie dans /api/profile et au login : la garder
// petite évite d'alourdir ces réponses (et Nginx limite déjà le corps à 1 Mo).
const maxAvatarBytes = 512 * 1024 // 512 Ko

// allowedAvatarTypes : types MIME image acceptés (détectés par contenu, pas par
// extension, pour éviter qu'un fichier soit renommé en .png).
var allowedAvatarTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// UploadAvatarHandler — POST /api/me/avatar (multipart, champ "avatar").
// Valide type et taille, encode en data URL base64, et persiste en base.
func UploadAvatarHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fichier 'avatar' manquant"})
		return
	}
	if fileHeader.Size > maxAvatarBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image trop lourde (max 512 Ko)"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lecture du fichier impossible"})
		return
	}
	defer file.Close()

	// LimitReader : garde-fou même si Size mentait (Content-Length falsifié).
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lecture du fichier impossible"})
		return
	}
	if len(data) > maxAvatarBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image trop lourde (max 512 Ko)"})
		return
	}

	// Détection du type par le contenu (magic bytes), pas par l'extension.
	contentType := http.DetectContentType(data)
	if !allowedAvatarTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format non supporté (JPEG, PNG, GIF ou WebP)"})
		return
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data))

	if err := database.SetAvatarURL(userID.(int), dataURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_url": dataURL})
}

// DeleteAvatarHandler — DELETE /api/me/avatar. Retire la photo de profil.
func DeleteAvatarHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	if err := database.SetAvatarURL(userID.(int), ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_url": ""})
}