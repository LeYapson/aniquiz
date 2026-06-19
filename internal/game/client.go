package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second // must be less than pongWait
)

// Client represents a player connected via WebSocket.
type Client struct {
	ID          string
	UserID      int
	Username    string
	Conn        *websocket.Conn
	Room        *Room
	Send        chan []byte
	Score       int
	IsSpectator bool
}

// ReadPump reads messages from the WebSocket connection and dispatches them to the room.
// A ping/pong deadline ensures dead connections are detected within pongWait.
func (c *Client) ReadPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Erreur décodage JSON: %v", err)
			continue
		}

		switch msg.Type {
		case "START_GAME":
			c.Room.Start <- true

		case "SUBMIT_ANSWER":
			if c.IsSpectator {
				continue
			}
			var answer string
			if err := json.Unmarshal(msg.Payload, &answer); err != nil {
				log.Printf("Erreur décodage payload string: %v", err)
				continue
			}
			fmt.Printf("Réponse reçue : %s\n", answer)
			c.Room.CheckAnswer(c, answer)

		case "CHAT":
			var text string
			if err := json.Unmarshal(msg.Payload, &text); err != nil || len([]rune(text)) == 0 {
				continue
			}
			if len([]rune(text)) > 200 {
				text = string([]rune(text)[:200])
			}
			chatMsg, _ := json.Marshal(map[string]interface{}{
				"type": "CHAT_MESSAGE",
				"payload": map[string]interface{}{
					"username": c.Username,
					"message":  text,
				},
			})
			c.Room.Broadcast <- chatMsg

		case "REACTION":
			var emoji string
			if err := json.Unmarshal(msg.Payload, &emoji); err != nil || len([]rune(emoji)) == 0 {
				continue
			}
			allowed := map[string]bool{"🔥": true, "🤔": true, "😱": true, "✅": true, "😭": true, "👏": true}
			if !allowed[emoji] {
				continue
			}
			reactionMsg, _ := json.Marshal(map[string]interface{}{
				"type": "REACTION_BROADCAST",
				"payload": map[string]interface{}{
					"username": c.Username,
					"emoji":    emoji,
				},
			})
			c.Room.Broadcast <- reactionMsg

		case "VOTE_SKIP":
			if c.IsSpectator {
				continue
			}
			c.Room.Mu.Lock()
			if !c.Room.IsPlaying {
				c.Room.Mu.Unlock()
				continue
			}
			c.Room.SkipVotes[c.ID] = true
			votes := len(c.Room.SkipVotes)
			activeCount := 0
			for cl := range c.Room.Clients {
				if !cl.IsSpectator {
					activeCount++
				}
			}
			needed := (activeCount + 1) / 2
			c.Room.Mu.Unlock()

			voteMsg, _ := json.Marshal(map[string]interface{}{
				"type": "SKIP_VOTE_UPDATE",
				"payload": map[string]interface{}{
					"votes":  votes,
					"needed": needed,
				},
			})
			c.Room.Broadcast <- voteMsg

			if votes >= needed {
				go c.Room.EndRound("Vote majoritaire !")
			}

		case "FORCE_SKIP":
			c.Room.Mu.Lock()
			isHost := c.Room.CreatorID == c.Username
			isPlaying := c.Room.IsPlaying
			c.Room.Mu.Unlock()
			if isHost && isPlaying {
				go c.Room.EndRound("Piste passée par l'hôte")
			}

		case "KICK_PLAYER":
			if c.Room.CreatorID != c.Username {
				continue
			}
			var targetUsername string
			if err := json.Unmarshal(msg.Payload, &targetUsername); err != nil || targetUsername == c.Username {
				continue
			}
			var target *Client
			c.Room.Mu.Lock()
			for cl := range c.Room.Clients {
				if cl.Username == targetUsername {
					target = cl
					break
				}
			}
			c.Room.Mu.Unlock()
			if target != nil {
				kickMsg, _ := json.Marshal(map[string]interface{}{
					"type":    "KICKED",
					"payload": "Vous avez été expulsé par l'hôte.",
				})
				select {
				case target.Send <- kickMsg:
				default:
				}
				go func(t *Client) {
					time.Sleep(300 * time.Millisecond)
					t.Conn.Close()
				}(target)
			}

		case "UPDATE_SETTINGS":
			type SettingsPayload struct {
				MaxRounds     int    `json:"max_rounds"`
				RoundDuration int    `json:"round_duration"`
				IsPrivate     bool   `json:"is_private"`
				Password      string `json:"password"`
				FilterType    string `json:"filter_type"`
				MinYear       int    `json:"min_year"`
				MaxYear       int    `json:"max_year"`
				FilterMalIDs  []int  `json:"filter_mal_ids"`
			}

			var settings SettingsPayload
			if err := json.Unmarshal(msg.Payload, &settings); err != nil {
				log.Printf("Erreur décodage settings: %v", err)
				continue
			}

			c.Room.Mu.Lock()
			if c.Room.CreatorID == c.Username && c.Room.State == StateLobby {
				if settings.MaxRounds > 0 {
					c.Room.MaxRounds = settings.MaxRounds
				}
				if settings.RoundDuration > 0 {
					c.Room.RoundDuration = settings.RoundDuration
				}
				c.Room.IsPrivate = settings.IsPrivate
				c.Room.Password = settings.Password
				c.Room.FilterType = settings.FilterType
				c.Room.MinYear = settings.MinYear
				c.Room.MaxYear = settings.MaxYear
				c.Room.FilterMalID = settings.FilterMalIDs
				c.Room.Mu.Unlock()

				settingsMsg, _ := json.Marshal(map[string]interface{}{
					"type": "SETTINGS_UPDATED",
					"payload": map[string]interface{}{
						"max_rounds":     c.Room.MaxRounds,
						"round_duration": c.Room.RoundDuration,
						"is_private":     c.Room.IsPrivate,
						"filter_type":    c.Room.FilterType,
						"min_year":       c.Room.MinYear,
						"max_year":       c.Room.MaxYear,
						"filter_mal_ids": c.Room.FilterMalID,
					},
				})
				c.Room.Broadcast <- settingsMsg
			} else {
				c.Room.Mu.Unlock()
			}
		}
	}
}

// WritePump delivers messages from the room to the WebSocket connection.
// A ticker sends WebSocket pings to detect dead connections.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
