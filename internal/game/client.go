package game

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

// Client représente un joueur connecté via WebSocket
type Client struct {
	ID       string
	UserID   int
	Username string
	Conn     *websocket.Conn
	Room     *Room
	Send     chan []byte //canal pour envoyer les messages au joueur
	Score    int
}

// ReadPump lit les messages envoyés par le joueur (ex: ses réponses)
func (c *Client) ReadPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// On décode le message JSON reçu du Front
		var msg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Erreur décodage JSON: %v", err)
			continue
		}

		// On réagit selon le type
		switch msg.Type {
		case "START_GAME":
			// On demande à la Room de démarrer
			c.Room.Start <- true
		case "SUBMIT_ANSWER":
			var answer string

			if err := json.Unmarshal(msg.Payload, &answer); err != nil {
				log.Printf("Erreur décodage payload string: %v", err)
				continue
			}

			fmt.Printf("Réponse reçue : %s\n", answer)
			c.Room.CheckAnswer(c, answer)
		case "UPDATE_SETTINGS":
			type SettingsPayload struct {
				MaxRounds     int    `json:"max_rounds"`
				RoundDuration int    `json:"round_duration"`
				IsPrivate     bool   `json:"is_private"`
				Password      string `json:"password"`
			}

			var settings SettingsPayload
			if err := json.Unmarshal(msg.Payload, &settings); err != nil {
				log.Printf("Erreur décodage settings: %v", err)
				continue
			}

			c.Room.Mu.Lock()
			// Sécurité : Seul le créateur peut modifier les paramètres
			if c.Room.CreatorID == c.ID && c.Room.State == StateLobby {
				if settings.MaxRounds > 0 {
					c.Room.MaxRounds = settings.MaxRounds
				}
				if settings.RoundDuration > 0 {
					c.Room.RoundDuration = settings.RoundDuration
				}
				c.Room.IsPrivate = settings.IsPrivate
				c.Room.Password = settings.Password

				// Notification générale aux clients du salon pour mettre à jour leur affichage
				c.Room.Mu.Unlock()
				//c.Room.broadcastSettingsUpdate()
			} else {
				c.Room.Mu.Unlock()
			}
		}
	}
}

// WritePump envoie les messages du serveur vers le client
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Le canal a été fermé par le salon
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// On envoie le message au format Texte
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}
