package game

import(
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"github.com/LeYapson/aniquiz/internal/models"
)

// Client représente un joueur connecté via WebSocket
type Client struct {
	ID  string
	Username string
	Conn *websocket.Conn
	Room *Room
	Send chan []byte //canal pour envoyer les messages au joueur
	Score int
}

//ReadPump lit les messages envoyés par le joueur (ex: ses réponses)
func (c *Client) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_,message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// on decode le message JSON envoyé par le joueur
		var msg models.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		//si le joueur envoie une réponse
		if msg.Type == "ANSWER" {
			//on convertit le payload (qui est une interface) en string
			answerStr := fmt.Sprintf("%v", msg.Payload)

			//le salon vérifie la réponse
			c.Room.CheckAnswer(c, answerStr)
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
