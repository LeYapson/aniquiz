package game

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"fmt"
)

// Client représente un joueur connecté via WebSocket
type Client struct {
	ID       string
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
