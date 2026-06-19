package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/LeYapson/aniquiz/internal/database"
	"github.com/LeYapson/aniquiz/internal/models"
)

var (
	ActiveRooms = make(map[string]*Room)
	RoomsMu     sync.Mutex
)

type RoomState string

const (
	StateLobby   RoomState = "LOBBY"
	StatePlaying RoomState = "PLAYING"
)

type RoundAnswer struct {
	Username string `json:"username"`
	TimeMs   int64  `json:"time_ms"`
	Bonus    int    `json:"bonus"`
}

type RoundSummaryItem struct {
	Round     int           `json:"round"`
	AnimeName string        `json:"anime_name"`
	Title     string        `json:"title"`
	Artist    string        `json:"artist"`
	TrackType string        `json:"track_type"`
	FoundBy   []RoundAnswer `json:"found_by"`
}

type Room struct {
	ID            string
	Clients       map[*Client]bool
	Broadcast     chan []byte
	Register      chan *Client
	Unregister    chan *Client
	State         RoomState
	CurrentTrack  *models.Track
	Start         chan bool
	IsPlaying     bool
	CurrentRound  int
	MaxRounds     int
	RoundDuration int
	IsPrivate     bool
	IsSolo        bool
	Password      string
	CreatorID     string
	HasAnswered   map[string]bool
	RoundAnswers  []RoundAnswer
	RoundStart    time.Time
	SkipVotes     map[string]bool
	RoundHistory  []RoundSummaryItem
	FilterType    string
	MinYear       int
	MaxYear       int
	FilterMalID   []int

	Mu   sync.Mutex
	done chan struct{} // closed when Run() exits; signals background goroutines to stop
}

type RoomSummary struct {
	ID           string    `json:"id"`
	State        RoomState `json:"state"`
	PlayersCount int       `json:"players_count"`
	IsPrivate    bool      `json:"is_private"`
	MaxRounds    int       `json:"max_rounds"`
}

func GetPublicRooms() []RoomSummary {
	RoomsMu.Lock()
	defer RoomsMu.Unlock()

	var list []RoomSummary
	for id, room := range ActiveRooms {
		room.Mu.Lock()
		if room.IsSolo {
			room.Mu.Unlock()
			continue
		}
		list = append(list, RoomSummary{
			ID:           id,
			State:        room.State,
			PlayersCount: len(room.Clients),
			IsPrivate:    room.IsPrivate,
			MaxRounds:    room.MaxRounds,
		})
		room.Mu.Unlock()
	}
	return list
}

func CreateRoom(id string, creatorID string, isSolo bool) *Room {
	return &Room{
		ID:            id,
		Clients:       make(map[*Client]bool),
		Broadcast:     make(chan []byte, 64),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Start:         make(chan bool),
		IsPlaying:     false,
		State:         StateLobby,
		CurrentRound:  0,
		MaxRounds:     5,
		RoundDuration: 20,
		Password:      "",
		IsPrivate:     false,
		IsSolo:        isSolo,
		CreatorID:     creatorID,
		HasAnswered:   make(map[string]bool),
		RoundAnswers:  []RoundAnswer{},
		SkipVotes:     make(map[string]bool),
		RoundHistory:  []RoundSummaryItem{},
		done:          make(chan struct{}),
	}
}

// Run is the room's event loop. All state mutations go through this goroutine,
// eliminating the need for locks on Clients map access.
// Closes r.done when it exits so background goroutines can detect shutdown.
func (r *Room) Run() {
	defer close(r.done)

	for {
		select {
		case client := <-r.Register:
			if r.State == StatePlaying {
				client.IsSpectator = true
			}
			r.Clients[client] = true

			statusMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "SPECTATOR_STATUS",
				"payload": client.IsSpectator,
			})
			r.safeSend(client, statusMsg)

			go r.BroadcastPlayerList()

			msgState, _ := json.Marshal(map[string]interface{}{
				"type":    "GAME_STATE",
				"payload": r.State,
			})
			r.safeSend(client, msgState)

			if r.State == StatePlaying && r.CurrentTrack != nil {
				msgTrack, _ := json.Marshal(map[string]interface{}{
					"type": "NewQuestion",
					"payload": map[string]interface{}{
						"audio_url": r.CurrentTrack.AudioURL,
						"room_id":   r.ID,
					},
				})
				r.safeSend(client, msgTrack)
			}

		case <-r.Start:
			if r.State != StatePlaying {
				r.State = StatePlaying
				r.broadcastGameState()
				go r.nextRound()
				log.Println("La partie commence dans le salon:", r.ID)
			}

		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)

				if len(r.Clients) == 0 {
					r.Mu.Lock()
					r.IsPlaying = false
					r.Mu.Unlock()
					RoomsMu.Lock()
					delete(ActiveRooms, r.ID)
					RoomsMu.Unlock()
					fmt.Printf("Salon %s vide, supprimé.\n", r.ID)
					return
				}

				r.Mu.Lock()
				wasHost := r.CreatorID == client.Username
				if wasHost {
					for c := range r.Clients {
						if !c.IsSpectator {
							r.CreatorID = c.Username
							break
						}
					}
				}
				newHost := r.CreatorID
				r.Mu.Unlock()

				if wasHost {
					hostMsg, _ := json.Marshal(map[string]interface{}{
						"type":    "HOST_CHANGED",
						"payload": newHost,
					})
					for c := range r.Clients {
						r.safeSend(c, hostMsg)
					}
				}

				go r.BroadcastPlayerList()
			}

		case message := <-r.Broadcast:
			for client := range r.Clients {
				// Non-blocking send: drop message to a client whose buffer is full
				// and disconnect them rather than stalling the entire room.
				select {
				case client.Send <- message:
				default:
					delete(r.Clients, client)
					close(client.Send)
				}
			}
		}
	}
}

// broadcastNotice sends an informational banner to every client in the room.
// Safe to call from background goroutines (e.g. nextRound): it never blocks and
// bails out if the room is being torn down.
func (r *Room) broadcastNotice(text string) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":    "NOTICE",
		"payload": text,
	})
	select {
	case r.Broadcast <- data:
	case <-r.done:
	}
}

// safeSend delivers a message to a client without blocking the caller.
func (r *Room) safeSend(c *Client, msg []byte) {
	select {
	case c.Send <- msg:
	default:
	}
}

func (r *Room) broadcastGameState() {
	msg := map[string]interface{}{
		"type":    "GAME_STATE",
		"payload": r.State,
	}
	data, _ := json.Marshal(msg)
	// broadcastGameState is called from Run() where r.Broadcast is consumed;
	// send directly to each client to avoid a deadlock on the Broadcast channel.
	for c := range r.Clients {
		r.safeSend(c, data)
	}
}

// BroadcastPlayerList sends an updated player list to all clients.
// Must be called in a goroutine when originating from Run() to avoid deadlock
// on the Broadcast channel.
func (r *Room) BroadcastPlayerList() {
	var players []models.PlayerInfo
	spectatorCount := 0
	for c := range r.Clients {
		if c.IsSpectator {
			spectatorCount++
			continue
		}
		players = append(players, models.PlayerInfo{
			ID:       c.ID,
			Username: c.Username,
			Score:    c.Score,
		})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type": "PLAYER_LIST",
		"payload": map[string]interface{}{
			"players":         players,
			"spectator_count": spectatorCount,
		},
	})
	select {
	case r.Broadcast <- data:
	case <-r.done:
	}
}

func (r *Room) EndRound(reason string) {
	r.Mu.Lock()
	if !r.IsPlaying {
		r.Mu.Unlock()
		return
	}

	track := r.CurrentTrack
	currentRound := r.CurrentRound
	answers := make([]RoundAnswer, len(r.RoundAnswers))
	copy(answers, r.RoundAnswers)
	r.IsPlaying = false
	r.RoundHistory = append(r.RoundHistory, RoundSummaryItem{
		Round:     currentRound,
		AnimeName: track.AnimeName,
		Title:     track.Title,
		Artist:    track.Artist,
		TrackType: track.TrackType,
		FoundBy:   answers,
	})
	r.Mu.Unlock()

	msg := map[string]interface{}{
		"type": "ROUND_ENDED",
		"payload": map[string]interface{}{
			"reason":     reason,
			"answer":     track.AnimeName,
			"title":      track.Title,
			"artist":     track.Artist,
			"track_type": track.TrackType,
			"difficulty": track.Difficulty,
			"video_url":  track.AudioURL,
			"found_by":   answers,
		},
	}

	data, _ := json.Marshal(msg)
	select {
	case r.Broadcast <- data:
	case <-r.done:
		return
	}

	// Wait before starting next round; abort if room is destroyed.
	select {
	case <-time.After(10 * time.Second):
	case <-r.done:
		return
	}

	go r.nextRound()
}

const bonusPremier = 10

func (r *Room) CheckAnswer(client *Client, answer string) {
	r.Mu.Lock()
	if !r.IsPlaying || r.CurrentTrack == nil || r.HasAnswered[client.ID] {
		r.Mu.Unlock()
		return
	}
	track := r.CurrentTrack
	elapsed := time.Since(r.RoundStart).Milliseconds()
	r.Mu.Unlock()

	result := VerifyAnswer(answer, track)
	if result.Points == 0 {
		return
	}

	// Re-acquire lock to atomically check + record the answer, preventing two
	// simultaneous correct submissions both seeing isFirst=true (TOCTOU).
	r.Mu.Lock()
	if r.HasAnswered[client.ID] {
		r.Mu.Unlock()
		return
	}
	isFirst := len(r.RoundAnswers) == 0
	bonus := 0
	if isFirst {
		bonus = bonusPremier
	}
	r.HasAnswered[client.ID] = true
	r.RoundAnswers = append(r.RoundAnswers, RoundAnswer{
		Username: client.Username,
		TimeMs:   elapsed,
		Bonus:    bonus,
	})
	r.Mu.Unlock()

	client.Score += result.Points + bonus

	msg := map[string]interface{}{
		"type": "PLAYER_GUESS",
		"payload": map[string]interface{}{
			"username": client.Username,
			"is_first": isFirst,
		},
	}
	data, _ := json.Marshal(msg)
	select {
	case r.Broadcast <- data:
	case <-r.done:
		return
	}

	go r.BroadcastPlayerList()
}

func (r *Room) nextRound() {
	r.Mu.Lock()
	if r.CurrentRound >= r.MaxRounds {
		r.Mu.Unlock()
		r.finishGame()
		return
	}
	r.CurrentRound++
	r.HasAnswered = make(map[string]bool)
	r.RoundAnswers = []RoundAnswer{}
	r.SkipVotes = make(map[string]bool)
	r.RoundStart = time.Now()
	duration := r.RoundDuration
	filters := models.TrackFilters{
		TrackType: r.FilterType,
		MinYear:   r.MinYear,
		MaxYear:   r.MaxYear,
		MalIDs:    r.FilterMalID,
	}
	r.Mu.Unlock()

	track, err := database.GetRandomTrackFiltered(filters)

	// La librairie est petite : filtrer sur la liste perso de l'utilisateur peut
	// ne renvoyer aucune piste. Plutôt que de laisser la partie bloquée sans
	// musique, on retire le filtre « liste perso » et on prévient les joueurs.
	if errors.Is(err, database.ErrNoTrack) && len(filters.MalIDs) > 0 {
		r.broadcastNotice("Aucune musique de votre liste perso n'est disponible pour le moment — filtre ignoré pour cette partie.")
		filters.MalIDs = nil
		r.Mu.Lock()
		r.FilterMalID = nil
		r.Mu.Unlock()
		track, err = database.GetRandomTrackFiltered(filters)
	}

	if err != nil {
		// Même sans filtre liste perso, aucune piste : librairie vide ou filtres
		// type/année trop restrictifs. On termine proprement au lieu de figer.
		log.Printf("Erreur récup musique (salon %s): %v", r.ID, err)
		if errors.Is(err, database.ErrNoTrack) {
			r.broadcastNotice("Aucune musique ne correspond aux paramètres de la partie.")
		}
		r.finishGame()
		return
	}

	r.Mu.Lock()
	r.CurrentTrack = track
	r.IsPlaying = true
	r.Mu.Unlock()

	log.Printf("Nouveau round dans %s : %s", r.ID, track.AnimeName)

	msg := map[string]interface{}{
		"type": "NewQuestion",
		"payload": map[string]interface{}{
			"audio_url": track.AudioURL,
			"room_id":   r.ID,
			"duration":  duration,
		},
	}
	data, _ := json.Marshal(msg)

	select {
	case r.Broadcast <- data:
	case <-r.done:
		return
	}

	// Round timer: abort cleanly when the room is destroyed mid-round.
	go func() {
		select {
		case <-time.After(time.Duration(duration) * time.Second):
			r.EndRound("Temps écoulé !")
		case <-r.done:
		}
	}()
}

func (r *Room) finishGame() {
	r.Mu.Lock()
	r.State = StateLobby
	r.IsPlaying = false
	r.CurrentRound = 0
	r.CurrentTrack = nil
	history := make([]RoundSummaryItem, len(r.RoundHistory))
	copy(history, r.RoundHistory)
	r.RoundHistory = []RoundSummaryItem{}
	r.Mu.Unlock()

	r.grantXP()

	for c := range r.Clients {
		if c.IsSpectator {
			c.IsSpectator = false
			msg, _ := json.Marshal(map[string]interface{}{
				"type":    "SPECTATOR_STATUS",
				"payload": false,
			})
			r.safeSend(c, msg)
		}
	}

	msg := map[string]interface{}{
		"type": "GAME_OVER",
		"payload": map[string]interface{}{
			"message": "La partie est terminée !",
			"history": history,
		},
	}
	data, _ := json.Marshal(msg)
	select {
	case r.Broadcast <- data:
	case <-r.done:
		return
	}

	r.resetScores()
	r.broadcastGameState()
	go r.BroadcastPlayerList()
}

// XPToLevel converts a total XP amount to a level using floor(sqrt(xp/100)) + 1.
func XPToLevel(xp int) int {
	if xp <= 0 {
		return 1
	}
	level := int(math.Sqrt(float64(xp)/100)) + 1
	if level < 1 {
		return 1
	}
	return level
}

// XPForScore computes XP earned for a given game score (minimum 5 for participation).
func XPForScore(score int) int {
	xp := score * 10
	if xp < 5 {
		return 5
	}
	return xp
}

func (r *Room) grantXP() {
	for c := range r.Clients {
		if c.UserID == 0 || c.IsSpectator {
			continue
		}
		xpGained := XPForScore(c.Score)
		newXP, newLevel, err := database.AddUserXP(c.UserID, xpGained)
		if err != nil {
			log.Printf("Erreur AddUserXP pour %s: %v", c.Username, err)
			continue
		}
		if err := database.SaveGameResult(c.UserID, c.Score, xpGained); err != nil {
			log.Printf("Erreur SaveGameResult pour %s: %v", c.Username, err)
		}
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "XP_GAINED",
			"payload": map[string]interface{}{
				"xp_gained": xpGained,
				"new_xp":    newXP,
				"new_level": newLevel,
			},
		})
		r.safeSend(c, msg)
	}
}

func (r *Room) resetScores() {
	for c := range r.Clients {
		if !c.IsSpectator {
			c.Score = 0
		}
	}
}
